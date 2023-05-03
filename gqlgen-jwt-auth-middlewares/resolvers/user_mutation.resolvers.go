package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"net/mail"
	"server/api/graphql/directives"
	"server/api/graphql/graph"
	"server/api/graphql/middleware"
	"server/internal/omnisend"
	"server/internal/subscription"
	"server/internal/user"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

func (r *mutationResolver) CreateUser(ctx context.Context, input graph.CreateUserInput) (*graph.User, error) {
	_, err := mail.ParseAddress(input.Email)
	if err != nil {
		return nil, errors.New("invalid email address")
	}

	// create user
	userDB, err := r.UserRepository.Create(input)
	if err != nil {
		return nil, err
	}

	err = r.SubscriptionRepository.Create(&subscription.Subscription{
		Email:              input.Email,
		SubscriptionTypeID: subscription.SubscriptionTypeIDNewsletter,
	})
	if err != nil {
		fmt.Println(err)
	}

	err = r.UserRepository.GenerateAndSendConfirmationKey(*userDB)
	if err != nil {
		return nil, err
	}

	if input.ApplyAsDesigner {
		httpAccess := middleware.GetHttpAccess(ctx)
		if httpAccess.IP == "" {
			_ = r.UserRepository.Delete(userDB.ID)
			return nil, errors.New("failed to apply as designer: failed to get IP from request for TOS acceptance")
		}

		_, err := r.UserRepository.SubmitDesignerApplication(userDB.ID, httpAccess.IP)
		if err != nil {
			return nil, err
		}
	}

	// subscribe to Omnisend
	omnisendContact, err := omnisend.GetContact(userDB.Email)
	if err != nil {
		fmt.Println("failed to upsert Omnisend contact for user with id " + strconv.Itoa(userDB.ID) + ": " + err.Error())
	}

	contactToUpsert := omnisend.Contact{
		CreatedAt: omnisend.JSONTime(time.Now()),
		FirstName: userDB.FirstName,
		LastName:  userDB.LastName,
		Tags: []omnisend.Tag{
			omnisend.TagCustomer,
		},
		Identifiers: []omnisend.Identifier{{
			Type: omnisend.IdentifierTypeEmail,
			Id:   userDB.Email,
			Channels: omnisend.Channel{Email: &omnisend.ChannelDetails{
				Status:     omnisend.ChannelStatusSubscribed,
				StatusDate: omnisend.JSONTime(time.Now()),
			}},
		}},
		CountryCode: userDB.Country,
	}

	if omnisendContact == nil {
		err = omnisend.CreateContact(contactToUpsert)
	} else {
		err = omnisend.PatchContact(contactToUpsert, omnisendContact.ContactID)
	}
	if err != nil {
		fmt.Println("failed to upsert Omnisend contact for user with id " + strconv.Itoa(userDB.ID) + ": " + err.Error())
	}

	return userDB.ToGraph(), nil
}

func (r *mutationResolver) ConfirmEmail(ctx context.Context, input graph.ConfirmEmailInput) (*string, error) {
	userID, err := r.UserRepository.UserIdFromConfirmationKey(input.Key)
	if err != nil {
		if errors.Is(err, user.ErrConfirmationKeyNoLongerValid) {
			usersDB, err := r.UserRepository.GetUsers(graph.UsersFilter{
				Ids: []int{*userID},
			})
			if err != nil {
				return nil, err
			}
			err = r.UserRepository.GenerateAndSendConfirmationKey(*usersDB[0])
			if err != nil {
				return nil, err
			}
			return nil, errors.Wrap(err, "please check your email for a new confirmation link")
		} else {
			return nil, err
		}
	}

	usersDB, err := r.UserRepository.GetUsers(graph.UsersFilter{
		Ids: []int{*userID},
	})
	if err != nil {
		return nil, err
	} else if len(usersDB) == 0 {
		return nil, errors.New("user doesn't exist")
	}
	userDB := usersDB[0]

	/*	customer, err := stripe.CreateCustomer(userDB.Email, userDB.FirstName+" "+userDB.LastName)
		if err != nil {
			return nil, err
		}

		_, err = r.UserRepository.CreateStripeID(&user.StripeID{
			ID:             customer.ID,
			UserID:         userDB.ID,
			StripeIDTypeID: user.TypeStripeIDCustomer,
		})
		if err != nil {
			return nil, err
		}*/

	// subscribe to Omnisend
	omnisendContact, err := omnisend.GetContact(userDB.Email)
	if err != nil {
		fmt.Println("failed to upsert Omnisend contact for user with id " + strconv.Itoa(userDB.ID) + ": " + err.Error())
	}

	contactToUpsert := omnisend.Contact{
		CreatedAt: omnisend.JSONTime(time.Now()),
		FirstName: userDB.FirstName,
		LastName:  userDB.LastName,
		Tags: []omnisend.Tag{
			omnisend.TagCustomer,
		},
		Identifiers: []omnisend.Identifier{{
			Type: omnisend.IdentifierTypeEmail,
			Id:   userDB.Email,
			Channels: omnisend.Channel{Email: &omnisend.ChannelDetails{
				Status:     omnisend.ChannelStatusSubscribed,
				StatusDate: omnisend.JSONTime(time.Now()),
			}},
		}},
		CountryCode: userDB.Country,
	}

	if omnisendContact == nil {
		err = omnisend.CreateContact(contactToUpsert)
	} else {
		err = omnisend.PatchContact(contactToUpsert, omnisendContact.ContactID)
	}
	if err != nil {
		fmt.Println("failed to upsert Omnisend contact for user with id " + strconv.Itoa(userDB.ID) + ": " + err.Error())
	}

	err = r.UserRepository.ConfirmUser(*userID)
	if err != nil {
		return nil, err
	}

	err = r.UserRepository.RevokeConfirmationKey(input.Key)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *mutationResolver) ResendConfirmationEmail(ctx context.Context, input *graph.ResendConfirmationEmailInput) (*string, error) {
	usersDB, err := r.UserRepository.GetUsers(graph.UsersFilter{
		Email: input.Email,
	})
	if err != nil {
		return nil, err
	} else if len(usersDB) == 0 {
		return nil, errors.New("email not registered")
	} else if usersDB[0].RegisterTime != nil {
		return nil, errors.New("email already confirmed")
	}

	err = r.UserRepository.GenerateAndSendConfirmationKey(*usersDB[0])
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *mutationResolver) SetRole(ctx context.Context, input *graph.SetRoleInput) (*string, error) {
	err := r.UserRepository.SetRole(input.UserID, input.RoleID, input.Activate)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *mutationResolver) UpdateProfile(ctx context.Context, input *graph.ProfileInput) (*string, error) {
	userDB := directives.GetLoggedUser(ctx)
	if input.Image != nil {
		err := user.UpdateProfileImage(userDB.ID, input.Image.Filename, input.Image.File)
		if err != nil {
			return nil, errors.Wrap(err, "failed to resolve UpdateProfile mutation")
		}
	}

	profile, err := r.UserRepository.UpsertProfile(userDB.ID, input)
	if err != nil {
		return nil, err
	}

	// check if profile is complete and update user accordingly
	hasCompleteProfile := true
	if input.Image == nil {
		if path, _ := user.ProfileImagePath(userDB.ID); path == nil {
			hasCompleteProfile = false
		}
	}

	if input.Description == nil {
		if profile == nil || profile.Description == "" {
			hasCompleteProfile = false
		}
	}

	if userDB.HasCompleteProfile != hasCompleteProfile {
		userDB.HasCompleteProfile = hasCompleteProfile
		_, err = r.UserRepository.Save(userDB)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (r *mutationResolver) AssignOffChainNfts(ctx context.Context, input graph.AssignOffChainNftsInput) (*string, error) {
	var err error
	userDB := directives.GetLoggedUser(ctx)

	alreadyAssigned := 0
	offChainNfts, err := r.UserRepository.OffchainNfts(user.OffchainNftsFilter{UserID: input.UserID, NftId: input.NftID})
	if err != nil {
		return nil, err
	} else if len(offChainNfts) == 1 {
		alreadyAssigned = offChainNfts[0].Amount
	} else if len(offChainNfts) > 1 {
		return nil, errors.New("more than one offChainNft record found")
	}

	if alreadyAssigned < input.Amount {
		err = r.CheckNftAvailability(input.NftID, input.Amount, userDB.ID)
		if err != nil {
			return nil, err
		}
	}

	userHasOffchainNfts := &user.UserHasOffchainNfts{
		UserID: input.UserID,
		NftID:  input.NftID,
		Amount: input.Amount,
	}
	err = r.UserRepository.UpsertOffchainNfts(userHasOffchainNfts)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

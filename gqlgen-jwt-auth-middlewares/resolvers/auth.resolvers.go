package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"server/api/graphql/directives"
	"server/api/graphql/graph"
	"server/api/graphql/middleware"
	"server/internal/auth"
	"server/pkg/blockchain"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
)

func (r *mutationResolver) Login(ctx context.Context, input *graph.LoginInput) (*graph.Authentication, error) {
	userDB, err := r.UserRepository.Login(input.Email, input.Password)
	if err != nil {
		return nil, err
	} else if userDB.RegisterTime == nil {
		return nil, errors.New("email not confirmed")
	}

	// create authorization token
	jwt, err := r.Auth.GenerateJWT(&auth.Claims{
		UserID:     userDB.ID,
		Expiration: r.Auth.CalculateExpiration(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve Login query: failed to generate jwt")
	}

	// set authorization cookie
	httpAccess := middleware.GetHttpAccess(ctx)
	httpAccess.SetAuthorizationCookie(jwt, r.Auth.TimeToLive)

	return &graph.Authentication{
		Jwt:  jwt,
		User: userDB.ToGraph(),
	}, nil
}

func (r *mutationResolver) LoginBlockchainInitialize(ctx context.Context) (*string, error) {
	key, err := r.UserRepository.GenerateLoginBlockchainMessage()
	if err != nil {
		err = errors.Wrap(err, "failed to resolve LoginBlockchainInitialize mutation")
		return nil, err
	}

	return key, nil
}

func (r *mutationResolver) LoginBlockchainEnd(ctx context.Context, input *graph.LoginBlockchainEndInput) (*graph.Authentication, error) {
	_, err := r.UserRepository.LoginBlockchainMessage(input.Message)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve LoginBlockchainEndInput mutation")
	}

	err = r.UserRepository.RevokeLoginBlockchainMessageKey(input.Message)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve LoginBlockchainEndInput mutation")
	}

	messageHex, err := hexutil.Decode(input.SignedMessage)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode message to hex")
	}
	address, err := blockchain.AddressFromSignature([]byte(input.Message), messageHex, true)
	if err != nil {
		return nil, err
	}

	userDB, err := r.UserRepository.UserFromAddress(address.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve LoginBlockchainEndInput mutation")
	} else if userDB == nil {
		return nil, errors.New("address is not associated to any account")
	}

	// create authorization token
	jwt, err := r.Auth.GenerateJWT(&auth.Claims{
		UserID:     userDB.ID,
		Expiration: r.Auth.CalculateExpiration(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve LoginBlockchainEndInput mutation: failed to generate jwt")
	}

	// set authorization cookie
	httpAccess := middleware.GetHttpAccess(ctx)
	httpAccess.SetAuthorizationCookie(jwt, r.Auth.TimeToLive)

	return &graph.Authentication{
		Jwt:  jwt,
		User: userDB.ToGraph(),
	}, nil
}

func (r *mutationResolver) Logout(ctx context.Context) (*string, error) {
	jwt, err := directives.GetJWT(ctx)
	if err != nil {
		m := "error"
		return &m, err
	}

	err = r.Auth.RevokeToken(*jwt)
	if err != nil {
		m := "error"
		return &m, err
	}

	return nil, nil
}

func (r *mutationResolver) ForgotPasswordInitialize(ctx context.Context, input *graph.ForgotPasswordInitialize) (*string, error) {
	usersDB, err := r.UserRepository.GetUsers(graph.UsersFilter{
		Email: &input.Email,
	})
	if err != nil {
		err = errors.Wrap(err, "failed to resolve ForgotPasswordInitialize mutation")
		return nil, err
	} else if len(usersDB) == 0 {
		return nil, errors.New("email not registered")
	}

	err = r.UserRepository.GenerateAndSendPassResetKey(*usersDB[0])
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *mutationResolver) ForgotPasswordEnd(ctx context.Context, input *graph.ForgotPasswordEnd) (*string, error) {
	userID, err := r.UserRepository.UserIDFromPassResetKey(input.Key)
	if err != nil {
		return nil, err
	}

	err = r.UserRepository.RevokePassResetKey(*userID, input.Key)
	if err != nil {
		return nil, err
	}

	err = r.UserRepository.ChangePassword(*userID, input.NewPassword)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve ForgotPasswordEnd mutation")
	}

	return nil, nil
}

func (r *mutationResolver) AssociateAddressInitialize(ctx context.Context, input *graph.AssociateAddressInitialize) (*string, error) {
	userDB, err := r.UserRepository.UserFromAddress(input.Address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve AssociateAddressInitialize mutation")
	} else if userDB != nil {
		return nil, errors.New("that address is associated to another account")
	}

	userDB = directives.GetLoggedUser(ctx)
	associateAddressMessage, err := r.UserRepository.GenerateAssociateAddressMessage(userDB.ID, input.Address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve AssociateAddressInitialize mutation")
	}

	return associateAddressMessage, nil
}

func (r *mutationResolver) AssociateAddressEnd(ctx context.Context, input *graph.AssociateAddressEnd) (*string, error) {
	userDB := directives.GetLoggedUser(ctx)

	associateAddress, err := r.UserRepository.AssociateAddressMessage(userDB.ID)
	if err != nil {
		return nil, err
	}

	err = r.UserRepository.RevokeAssociateMessageKey(userDB.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve AssociateAddressEnd mutation")
	}

	signedMessageHex, err := hexutil.Decode(input.SignedMessage)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode message to hex")
	}
	address, err := blockchain.AddressFromSignature([]byte(associateAddress.Message), signedMessageHex, true)
	if err != nil {
		return nil, err
	} else if associateAddress.Address != address.String() {
		return nil, errors.New("invalid signature")
	}

	err = r.UserRepository.AssociateAddress(userDB.ID, address.String())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

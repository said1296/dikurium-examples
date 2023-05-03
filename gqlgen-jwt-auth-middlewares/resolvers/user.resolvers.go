package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"server/api/graphql/graph"
	"server/internal/conversion"
	"server/internal/nft"
	"server/internal/thegraph"
	"server/internal/user"

	"github.com/pkg/errors"
)

func (r *userResolver) Addresses(ctx context.Context, obj *graph.User) ([]string, error) {
	userHasAddressesDB, err := r.UserRepository.Addresses(obj.ID)
	if err != nil {
		err = errors.Wrap(err, "failed to resolve Addresses field")
		return nil, err
	}

	addressesArray := make([]string, 0)
	for _, userHasAddressDB := range userHasAddressesDB {
		addressesArray = append(addressesArray, userHasAddressDB.AddressID)
	}

	return addressesArray, nil
}

func (r *userResolver) Owned(ctx context.Context, obj *graph.User) ([]*graph.UserHasNfts, error) {
	// get user addresses
	userHasAddressesDB, err := r.UserRepository.Addresses(obj.ID)
	if err != nil {
		err = errors.Wrap(err, "failed to resolve Owned field: failed to get user addresses")
		return nil, err
	}

	// get addresses nfts
	var addressesHaveNftsDB []*thegraph.AddressHasNfts
	for _, userHasAddressDB := range userHasAddressesDB {
		addressHasNftsDB, err := r.UserRepository.AddressNfts(user.AddressNftsFilter{Address: userHasAddressDB.AddressID})
		if err != nil {
			err = errors.Wrap(err, "failed to resolve Owned field: failed to get address Nfts")
			return nil, err
		}
		addressesHaveNftsDB = append(addressesHaveNftsDB, addressHasNftsDB...)
	}

	// convert to graph models
	userHasNftsGraph := make([]*graph.UserHasNfts, 0)
	for _, addressHasNftsDB := range addressesHaveNftsDB {
		nftID, err := conversion.StringToInt(addressHasNftsDB.Nft)
		if err != nil {
			err = errors.Wrap(err, "failed to resolve Owned field: failed to get Nft")
			return nil, err
		}

		nftDB, err := r.NftRepository.GetNfts(&graph.NftsFilter{
			Ids: []int{
				*nftID,
			},
		})
		if err != nil {
			err = errors.Wrap(err, "failed to resolve Owned field: failed to get Nft")
			return nil, err
		} else if len(nftDB) == 0 {
			continue
		}

		userHasNftGraph, err := addressHasNftsDB.ToGraph(*nftDB[0])
		if err != nil {
			err = errors.Wrap(err, "failed to resolve Owned field: failed to convert to graph model")
			return nil, err
		}
		userHasNftsGraph = append(userHasNftsGraph, userHasNftGraph)
	}

	// get offchain nfts
	offchainNfts, err := r.UserRepository.OffchainNfts(user.OffchainNftsFilter{
		UserID: obj.ID,
	})

	for _, offchainNft := range offchainNfts {
		seen := false
		for i, userHasNftGraph := range userHasNftsGraph {
			if userHasNftGraph.Nft.ID == offchainNft.NftID {
				seen = true
				userHasNftsGraph[i].OffChain += offchainNft.Amount
			}
		}

		if !seen {
			nftsDB, err := r.NftRepository.GetNfts(&graph.NftsFilter{
				Ids: []int{
					offchainNft.NftID,
				},
			})
			if err != nil {
				err = errors.Wrap(err, "failed to resolve Owned field: failed to get Nft")
				return nil, err
			} else if len(nftsDB) == 0 {
				continue
			}

			nftGraph, err := nftsDB[0].ToGraph()
			if err != nil {
				return nil, err
			}

			userHasNftsGraph = append(userHasNftsGraph, &graph.UserHasNfts{
				Address:  "",
				Nft:      nftGraph,
				OffChain: offchainNft.Amount,
			})
		}
	}

	return userHasNftsGraph, nil
}

func (r *userResolver) Designed(ctx context.Context, obj *graph.User, filter *graph.DesignedFieldFilter) ([]*graph.Nft, error) {
	nftsDB, err := r.UserRepository.UserCreatedNfts(obj.ID, nil, &user.NftsFilter{OnSale: filter.OnSale})
	if err != nil {
		err = errors.Wrap(err, "failed to resolve Designed field")
	}

	// convert to graph models
	nftsGraph, err := nft.NftsDBToGraph(nftsDB)
	if err != nil {
		err = errors.Wrap(err, "failed to resolve Designed field")
	}

	return nftsGraph, nil
}

func (r *userResolver) Roles(ctx context.Context, obj *graph.User) ([]*graph.Role, error) {
	rolesDB, err := r.UserRepository.GetRoles(graph.RolesFilter{
		User: &graph.RolesUserFilter{
			ID: &obj.ID,
		},
	})
	if err != nil {
		err = errors.Wrap(err, "failed to resolve Roles field")
		return nil, err
	}

	return user.RolesDBToGraph(rolesDB), nil
}

func (r *userResolver) Profile(ctx context.Context, obj *graph.User) (*graph.Profile, error) {
	profileDB, err := r.UserRepository.GetProfile(obj.ID)
	if err != nil {
		return nil, err
	}

	image, err := user.ProfileImageURL(obj.ID)
	if err != nil {
		return nil, err
	}

	var profileGraph graph.Profile
	if profileDB != nil {
		profileGraph = profileDB.ToGraph()
	}

	profileGraph.Image = image

	return &profileGraph, nil
}

func (r *userResolver) Address(ctx context.Context, obj *graph.User) (*graph.Address, error) {
	usersDB, err := r.UserRepository.GetUsers(graph.UsersFilter{
		Ids: []int{obj.ID},
	})

	if err != nil {
		return nil, err
	} else if len(usersDB) == 0 {
		return nil, errors.New("user doesn't exist")
	}

	userDB := usersDB[0]

	country, err := r.Gountries.FindCountryByAlpha(userDB.Country)
	if err != nil {
		return nil, err
	}

	return &graph.Address{
		CountryCode: userDB.Country,
		Country:     country.Name.Common,
	}, nil
}

// User returns graph.UserResolver implementation.
func (r *Resolver) User() graph.UserResolver { return &userResolver{r} }

type userResolver struct{ *Resolver }

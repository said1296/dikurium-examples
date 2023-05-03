package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"net/mail"
	"server/api/graphql/directives"
	"server/api/graphql/graph"
	"server/internal/user"

	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
)

func (r *queryResolver) Users(ctx context.Context, filter graph.UsersFilter) (*graph.UsersResult, error) {
	var err error

	usersDB, err := r.UserRepository.GetUsers(filter)
	if err != nil {
		err = errors.Wrap(err, "failed to resolve Users query")
		return nil, err
	}

	return &graph.UsersResult{
		Users: user.UsersDBToGraph(usersDB),
	}, nil
}

func (r *queryResolver) OffchainNfts(ctx context.Context) ([]*graph.OffchainNft, error) {
	userDB := directives.GetLoggedUser(ctx)
	offchainNfts, err := r.UserRepository.OffchainNfts(user.OffchainNftsFilter{
		UserID: userDB.ID,
	})
	if err != nil {
		return nil, err
	}

	offchainNftsGraph := make([]*graph.OffchainNft, len(offchainNfts))

	for i, offchainNft := range offchainNfts {
		offchainNftsGraph[i] = &graph.OffchainNft{
			ID:     offchainNft.NftID,
			Amount: offchainNft.Amount,
		}
	}

	return offchainNftsGraph, nil
}

func (r *queryResolver) ValidateEmail(ctx context.Context, email string) (*string, error) {
	_, err := mail.ParseAddress(email)
	if err != nil {
		m := "invalid email"
		return &m, nil
	}

	usersDB, err := r.UserRepository.GetUsers(graph.UsersFilter{
		Email: &email,
	})
	if err != nil {
		err = errors.Wrap(err, "failed to resolve ValidateEmail query")
		return nil, err
	} else if len(usersDB) > 0 {
		m := "email already registered"
		return &m, nil
	}

	return nil, nil
}

func (r *usersResultResolver) Count(ctx context.Context, obj *graph.UsersResult) (int, error) {
	var err error
	rctx := graphql.GetFieldContext(ctx)
	filter, ok := rctx.Parent.Args["filter"].(graph.UsersFilter)
	count := 0
	if ok {
		count, err = r.UserRepository.Count(filter)
		if err != nil {
			return 0, nil
		}
	} else {
		return 0, errors.New("Failed to resolve count field")
	}

	return count, nil
}

// UsersResult returns graph.UsersResultResolver implementation.
func (r *Resolver) UsersResult() graph.UsersResultResolver { return &usersResultResolver{r} }

type usersResultResolver struct{ *Resolver }

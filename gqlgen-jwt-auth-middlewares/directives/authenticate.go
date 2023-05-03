package directives

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"net/http"
	"server/api/graphql/graph"
	"server/api/graphql/grapherrors"
	"server/api/graphql/middleware"
	"server/internal/auth"
	"server/internal/user"
	"strings"
	"time"
)

const (
	authCookieName   = "Authorization"
	authCookiePrefix = "Bearer "
)

var (
	ctxKeyLoggedUser = &middleware.ContextKey{Name: "LoggedUser"}
	ctxKeyJWT        = &middleware.ContextKey{Name: "JWT"}
	ErrAuthorizationNotValid = grapherrors.NewError("AUTHORIZATION_INVALID")
	ErrAuthorizationFailed = grapherrors.NewError("AUTHORIZATION_FAILED")
	ErrAuthorizationNoPermission = grapherrors.NewError("AUTHORIZATION_NO_PERMISSION")
	ErrNoJWT = grapherrors.NewError("NO_JWT")
)



func Authenticate(ctx context.Context, obj interface{}, next graphql.Resolver, rules []graph.Rule, enforce *bool) (res interface{}, err error) {
	dependencies := middleware.GetDependencies(ctx)
	enforceAsserted := true
	if enforce != nil && *enforce == false {
		enforceAsserted = false
	}

	jwt, err := jwtFromCookie(ctx)
	if err != nil {
		if !errors.Is(err, ErrNoJWT) {
			return nil, ErrAuthorizationFailed.CompleteError(ctx, err)
		} else {
			if enforceAsserted {
				return nil, ErrNoJWT.CompleteError(ctx, errors.New("failed to get authorization cookie"))
			} else {
				return next(ctx)
			}
		}
	}

	claims, err := dependencies.Auth.ValidateAndGetClaims(*jwt)
	if err != nil {
		if !enforceAsserted {
			return next(ctx)
		}
		return nil, ErrAuthorizationNotValid.CompleteError(ctx, err)
	}

	userDB, err := retrieveUser(claims.UserID, dependencies)
	if err != nil {
		if !enforceAsserted {
			return next(ctx)
		}
		return nil, ErrAuthorizationFailed.CompleteError(ctx, err)
	}

	for _, rule := range rules {
		switch rule {
		case graph.RuleAdminRole:
			if !userDB.HasRole(user.RoleAdminID) {
				return nil, ErrAuthorizationNoPermission.CompleteError(ctx, errors.New("only admins have access"))
			}
		case graph.RuleDesignerOf:
			input, ok := graphql.GetFieldContext(ctx).Args["input"]
			if !ok {
				return nil, ErrAuthorizationFailed.CompleteError(ctx, errors.New("failed to parse input field from payload"))
			}
			var args map[string]interface{}

			err = mapstructure.Decode(input, &args)
			if err != nil {
				return nil, ErrAuthorizationFailed.CompleteError(ctx, errors.New("failed to convert input field to map"))
			}

			nftID, ok := args["NftID"].(int)
			if !ok {
				return nil, ErrAuthorizationFailed.CompleteError(ctx, errors.New("no NftID field to check permission"))
			}

			dependencies := middleware.GetDependencies(ctx)

			nftsDB, err := dependencies.UserRepository.UserCreatedNfts(userDB.ID, []*int{&nftID}, nil)
			if err != nil {
				return nil, err
			} else if len(nftsDB) == 0 {
				return nil, ErrAuthorizationNoPermission.CompleteError(ctx, errors.New("user is not designer of that nft"))
			}
		}
	}

	// check if jwt needs to be refreshed
	timeToLive := time.Unix(claims.Expiration, 0).Sub(time.Now())
	minimumTimeToLive := dependencies.Auth.TimeToLive.Seconds() * 0.2
	if timeToLive.Seconds() < minimumTimeToLive {
		jwt, err := dependencies.Auth.GenerateJWT(&auth.Claims{
			UserID:     userDB.ID,
			Expiration: dependencies.Auth.CalculateExpiration(),
		})
		if err != nil {
			return nil, ErrAuthorizationFailed.CompleteError(ctx, errors.Wrap(err, "failed to refresh jwt"))
		}

		httpAccess := middleware.GetHttpAccess(ctx)
		httpAccess.SetAuthorizationCookie(jwt, dependencies.Auth.TimeToLive)
	}

	// set authorizations in context
	ctx = context.WithValue(ctx, ctxKeyLoggedUser, userDB)
	ctx = context.WithValue(ctx, ctxKeyJWT, jwt)

	return next(ctx)
}

func jwtFromCookie(ctx context.Context) (*string, error) {
	httpAccess := middleware.GetHttpAccess(ctx)
	authCookie, err := httpAccess.Request.Cookie(authCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, ErrNoJWT
		}
		err = errors.Wrap(err, "failed to get authorization cookie")
		return nil, err
	} else if authCookie == nil {
		return nil, ErrNoJWT
	}

	jwt := strings.Replace(authCookie.Value, authCookiePrefix, "", 1)

	return &jwt, nil
}

func retrieveUser(userID int, dependencies *middleware.Dependencies) (*user.User, error) {
	// get the user from the database
	usersDB, err := dependencies.UserRepository.GetUsers(graph.UsersFilter{
		Ids: []int{
			userID,
		},
	})
	if len(usersDB) == 0 {
		return nil, errors.New("user doesn't exist")
	}

	// get user's roles
	usersDB[0].Roles, err = dependencies.UserRepository.GetRoles(graph.RolesFilter{
		User: &graph.RolesUserFilter{
			ID: &usersDB[0].ID,
		},
	})
	if err != nil {
		err = errors.Wrap(err, "failed to retrieve user's roles")
		return nil, err
	}

	return usersDB[0], nil
}

// GetLoggedUser finds the user from the context. REQUIRES AuthMiddleware to have run.
func GetLoggedUser(ctx context.Context) *user.User {
	userDB, _ := ctx.Value(ctxKeyLoggedUser).(*user.User)
	return userDB
}

func GetJWT(ctx context.Context) (*string, error) {
	jwt, ok := ctx.Value(ctxKeyJWT).(*string)
	if !ok {
		return nil, errors.New("failed to get jwt from context")
	}
	return jwt, nil
}
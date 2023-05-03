package middleware

import (
	"context"
	"net/http"
	"server/internal/auth"
	"server/internal/nft"
	"server/internal/user"
)

type Dependencies struct {
	NftRepository *nft.Repository
	UserRepository *user.Repository
	Auth *auth.Auth
}

var (
	ctxKeyRepositories = &ContextKey{"Cookies"}
)

// Inject repositories
func RepositoriesMiddleware(nftRepository *nft.Repository, userRepository *user.Repository, auth *auth.Auth, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dependencies := &Dependencies{
			NftRepository: nftRepository,
			UserRepository: userRepository,
			Auth: auth,
		}

		ctx := context.WithValue(r.Context(), ctxKeyRepositories, dependencies)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})

}

func GetDependencies(ctx context.Context) *Dependencies {
	repositories, _ := ctx.Value(ctxKeyRepositories).(*Dependencies)
	return repositories
}

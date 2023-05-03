package resolvers

import (
	"github.com/pariz/gountries"
	"server/internal/auth"
	"server/internal/authorization"
	"server/internal/blog"
	"server/internal/ipfs"
	"server/internal/nft"
	"server/internal/sales"
	"server/internal/subscription"
	"server/internal/user"
	"server/pkg/blockchain"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	NftRepository           *nft.Repository
	UserRepository          *user.Repository
	SubscriptionRepository  *subscription.Repository
	BlogRepository          *blog.Repository
	AuthorizationRepository *authorization.Repository
	SalesRepository         *sales.Repository
	IPFS                    *ipfs.IPFS
	Auth                    *auth.Auth
	Blockchain              *blockchain.Blockchain
	Gountries               *gountries.Query
}

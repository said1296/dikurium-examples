package resolvers

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"server/api/graphql/graph"
	"server/internal/authorization"
	"server/internal/constant"
	"server/internal/envs"
	"server/internal/stripe"
	"server/internal/user"
	"server/pkg/blockchain"
	"strconv"
	"time"
)

func (r *mutationResolver) CheckNftAvailability(nftID int, amountRequested int, userID int) error {
	// check if nft exists
	nftsDB, err := r.NftRepository.GetNfts(&graph.NftsFilter{
		Ids: []int{nftID},
	})
	if err != nil {
		return err
	} else if len(nftsDB) == 0 {
		return errors.New("requested NFT with id " + strconv.Itoa(nftID) + " doesn't exist")
	}

	// cancel old active payment intents
	cancelCreatedBefore := time.Now().Add(-constant.PaymentIntentExpiryTime)
	paymentIntentsToCancel, err := r.UserRepository.GetPaymentIntents(
		user.GetPaymentIntentsFilter{
			NftID:         nftID,
			StatusID:      user.StatusPaymentIntentActiveID,
			UserID:        userID,
			CreatedBefore: &cancelCreatedBefore,
		},
		user.GetPaymentIntentsFilter{
			OrCondition: true,
			StatusID:      user.StatusPaymentIntentActiveID,
			UserID:        userID,
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to get expired payment intents")
	}

	for _, paymentIntentToCancel := range paymentIntentsToCancel {
		if paymentIntentToCancel.PaymentIntentID == "" {
			paymentIntentToCancel.PaymentIntentStatusID = user.StatusPaymentIntentCancelledID
		} else {
			err = stripe.CancelPaymentIntent(paymentIntentToCancel.PaymentIntentID)
			if errors.Is(err, stripe.ErrCancelSucceededPaymentIntent) {
				paymentIntentToCancel.PaymentIntentStatusID = user.StatusPaymentIntentSuccessID
			} else if err != nil {
				return errors.Wrap(err, "failed to cancel expired payment intent in Stripe")
			} else {
				paymentIntentToCancel.PaymentIntentStatusID = user.StatusPaymentIntentCancelledID
			}
		}
		_, err = r.UserRepository.UpsertPaymentIntent(paymentIntentToCancel)
		if err != nil {
			return errors.Wrap(err, "failed to delete expired payment intent from DB")
		}
	}

	amountAvailable, err := r.GetAvailableNfts(nftID)
	if err != nil {
		return err
	} else if amountAvailable < amountRequested {
		return errors.New("there isn't enough nfts available")
	}

	return nil
}

func (r *Resolver) GetAvailableNfts(nftID int) (int, error) {
	// get nfts in process of onchain buying
	amountAuthorized, err := r.AuthorizationRepository.CountAuthorizedNft(authorization.GetAuthorizationFilter{
		NftID:               nftID,
		AuthorizationTypeID: authorization.TypeBuyID,
		Current:             true,
		Used:                &[]bool{false}[0],
	})
	if err != nil {
		return 0, err
	}

	// get nfts in process of offchain buying
	amountInPaymentIntents := 0
	paymentIntents, err := r.UserRepository.GetPaymentIntents(user.GetPaymentIntentsFilter{
		NftID:    nftID,
		StatusID: user.StatusPaymentIntentActiveID,
	})
	for i := range paymentIntents {
		for _, paymentIntentHasNft := range paymentIntents[i].HasNfts {
			amountInPaymentIntents += paymentIntentHasNft.Amount
		}
	}

	// get nfts owned offchain
	amountOwnedOffchain := 0
	offchainNfts, err := r.UserRepository.OffchainNfts(user.OffchainNftsFilter{
		NftId: nftID,
	})
	if err != nil {
		return 0, err
	}

	for _, offchainNft := range offchainNfts {
		amountOwnedOffchain += offchainNft.Amount
	}

	// get nfts owned by the sales contract
	addressHasNfts, err := r.UserRepository.AddressNfts(user.AddressNftsFilter{
		Address: envs.GetSalesAddress(),
		NftID:   nftID,
	})
	var amountOnSale int
	if len(addressHasNfts) > 0 {
		amountOnSale = addressHasNfts[0].Amount
	}

	// check availability
	amountUnavailable := amountAuthorized + amountInPaymentIntents + amountOwnedOffchain
	return amountOnSale - amountUnavailable, nil
}

func (r *mutationResolver) SignAuthorization(authorizationEncoded []byte) (blockchain.Signature, error) {
	signerPrivateKey, err := blockchain.PrivateKeyFromPem(constant.SignerKeyPemPath)
	if err != nil {
		return blockchain.Signature{}, errors.Wrap(err, "failed to get signer private key")
	}

	signature, err := blockchain.SignMessage(authorizationEncoded, signerPrivateKey)
	if err != nil {
		return signature, errors.Wrap(err, "failed to sign authorization")
	}

	return signature, nil
}

func GraphQLError(ctx context.Context, err error, code string) *gqlerror.Error {
	return &gqlerror.Error{
		Path:       graphql.GetPath(ctx),
		Message:    err.Error(),
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
}

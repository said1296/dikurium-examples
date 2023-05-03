package mappers

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"server/api/graphql/graph"
	"server/pkg/blockchain"
	"strconv"
)

func SignatureToGraph(signature blockchain.Signature) *graph.Signature {
	return &graph.Signature{
		R: hexutil.Encode(signature.R),
		S: hexutil.Encode(signature.S),
		V: strconv.Itoa(signature.V),
	}
}

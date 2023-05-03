package chainconfig

import (
	"bytes"
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	configAws = map[string]string{
		"AWS_1": "AWS_1",
		"AWS_2": "AWS_2",
		"AWS_3": "AWS_3",
	}
	configContracts = []ContractConfig{
		{
			Address:    "0x0",
			Type:       "ERC721",
			StartBlock: 0,
		},
		{
			Address:    "0x1",
			Type:       "ERC1155",
			StartBlock: 1,
		},
		{
			Address:    "0x2",
			Type:       "ERC721",
			StartBlock: 2,
		},
	}
)

type ChainConfigSuite struct {
	suite.Suite
	chainConfig ChainConfig
}

func (c *ChainConfigSuite) SetupTest() {
	c.chainConfig = ChainConfig{
		Id:                1,
		Chain:             "eth_main",
		Rpc:               "https://localhost:3001",
		Relay:             "https://localhost:3000",
		PollingSecs:       1,
		BlockBatchSize:    1,
		MaxRetries:        1,
		MaxRetryDelaySecs: 1,
		Aws:               configAws,
		Contracts:         configContracts,
	}
}

func (c *ChainConfigSuite) TestChainConfig_GenerateProviderClients() {
	ctx := context.Background()
	_, _, err := c.chainConfig.GenerateProviderClients(ctx)
	c.Assert().NoError(err)

	c.chainConfig.Rpc = "https://ujinaidnfjdsa.ethereum.managedblockchain.us-east-1.amazonaws.com"
	_, _, err = c.chainConfig.GenerateProviderClients(ctx)
	for envKey, envValue := range configAws {
		c.Assert().Equal(os.Getenv(strings.ToUpper(envKey)), envValue)
	}
	c.Assert().NoError(err)
}

func (c *ChainConfigSuite) TestChainConfig_GetAddressesSlice() {
	expected := make([]string, len(configContracts))
	for i, contract := range configContracts {
		expected[i] = contract.Address
	}

	actual := c.chainConfig.GetAddressesSlice()
	c.Assert().ElementsMatch(expected, actual)
}

func (c *ChainConfigSuite) TestChainConfig_GetCommonAddresses() {
	expected := make([]common.Address, len(configContracts))
	for i, contract := range configContracts {
		expected[i] = common.HexToAddress(contract.Address)
	}

	actual := c.chainConfig.GetCommonAddresses()
	c.Assert().ElementsMatch(expected, actual)
}

func (c *ChainConfigSuite) TestChainConfig_OldestCreationBlock() {
	expected := big.NewInt(0)
	actual, err := c.chainConfig.OldestCreationBlock()
	c.Assert().NoError(err)
	c.Assert().ElementsMatch(expected, actual)
}

func TestChainConfig(t *testing.T) {
	suite.Run(t, new(ChainConfigSuite))
}

func TestUnmarshallConfig(t *testing.T) {
	expectedChainConfigs := []ChainConfig{
		{
			Id:                0,
			Chain:             "eth_main",
			Rpc:               "https://ujinaidnfjdsa.ethereum.managedblockchain.us-east-1.amazonaws.com",
			Relay:             "http://localhost:3000",
			PollingSecs:       2 * time.Second,
			BlockBatchSize:    1000,
			MaxRetries:        3,
			MaxRetryDelaySecs: 5 * time.Second,
			Aws: map[string]string{
				"aws_region":            "us-east-1",
				"aws_access_key_id":     "aunsdifun43uin",
				"aws_secret_access_key": "fd+saubnfu342nn12321",
			},
			Contracts: []ContractConfig{
				{
					Address:    "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
					Type:       "ERC721",
					StartBlock: 14822196,
				},
				{
					Address:    "0x282bdd42f4eb70e7a9d9f40c8fea0825b7f68c5d",
					Type:       "ERC721",
					StartBlock: 14812191,
				},
			},
		},
		{
			Id:                1,
			Chain:             "eth_goerly",
			Rpc:               "https://eth-goerli.g.alchemy.com/v2/fdbsuiafbu1hbuhbfuhdsabu5h32",
			Relay:             "http://localhost:3000",
			PollingSecs:       4 * time.Second,
			BlockBatchSize:    2000,
			MaxRetries:        6,
			MaxRetryDelaySecs: 10 * time.Second,
			Aws:               nil,
			Contracts: []ContractConfig{
				{
					Address:    "0xb2a2c7fb3e326c5ef282cb78207fbd9dcba8e983",
					Type:       "ERC721",
					StartBlock: 13822102,
				},
				{
					Address:    "0x19b86299c21505cdf59ce63740b240a9c822b5e4",
					Type:       "ERC721",
					StartBlock: 14522219,
				},
				{
					Address:    "0xbce3781ae7ca1a5e050bd9c4c77369867ebc307e",
					Type:       "ERC721",
					StartBlock: 14822196,
				},
			},
		},
	}
	byteData := []byte(`
CHAINS:
  - CHAIN: eth_main
    RPC: https://ujinaidnfjdsa.ethereum.managedblockchain.us-east-1.amazonaws.com
    RELAY: http://localhost:3000
    POLLING_SECS: 2
    BLOCK_BATCH_SIZE: 1000
    MAX_RETRIES: 3
    MAX_RETRY_DELAY_SECS: 5
    AWS:
      AWS_REGION: us-east-1
      AWS_ACCESS_KEY_ID: aunsdifun43uin
      AWS_SECRET_ACCESS_KEY: fd+saubnfu342nn12321
    CONTRACTS:
      - ADDRESS: 0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d
        TYPE: ERC721
        START_BLOCK: 14822196
      - ADDRESS: 0x282bdd42f4eb70e7a9d9f40c8fea0825b7f68c5d
        TYPE: ERC721
        START_BLOCK: 14812191
  - CHAIN: eth_goerly
    RPC: https://eth-goerli.g.alchemy.com/v2/fdbsuiafbu1hbuhbfuhdsabu5h32
    RELAY: http://localhost:3000
    POLLING_SECS: 4
    BLOCK_BATCH_SIZE: 2000
    MAX_RETRIES: 6
    MAX_RETRY_DELAY_SECS: 10
    CONTRACTS:
      - ADDRESS: 0xb2a2c7fb3e326c5ef282cb78207fbd9dcba8e983
        TYPE: ERC721
        START_BLOCK: 13822102
      - ADDRESS: 0x19b86299c21505cdf59ce63740b240a9c822b5e4
        TYPE: ERC721
        START_BLOCK: 14522219
      - ADDRESS: 0xbce3781ae7ca1a5e050bd9c4c77369867ebc307e
        TYPE: ERC721
        START_BLOCK: 14822196
`)

	r := bytes.NewReader(byteData)

	viper.SetConfigType("yaml")
	err := viper.ReadConfig(r)
	assert.NoError(t, err)

	actualChainConfigs, err := unmarshallConfig()
	assert.NoError(t, err)

	for i, expectedChainConfig := range expectedChainConfigs {
		assert.Equal(t, expectedChainConfig, actualChainConfigs[i])
	}
}

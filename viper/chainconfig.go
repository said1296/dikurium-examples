package chainconfig

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/said1296/gethaws"
	"github.com/spf13/viper"
	"math"
	"math/big"
	"os"
	"strings"
	"time"
)

type ProviderType string

const (
	ProviderAws     ProviderType = "Aws"
	ProviderRegular ProviderType = "regular"
)

// configWrapper is the upmost level of the config structure
type configWrapper struct {
	Chains []ChainConfig
}

// ChainConfig holds the config for a single chain
type ChainConfig struct {
	Id                int
	Chain             string
	Rpc               string
	Relay             string            `mapstructure:"relay"`
	PollingSecs       time.Duration     `mapstructure:"polling_secs"`
	BlockBatchSize    uint64            `mapstructure:"block_batch_size"`
	MaxRetries        uint              `mapstructure:"max_retries"`
	MaxRetryDelaySecs time.Duration     `mapstructure:"max_retry_delay_secs"`
	Aws               map[string]string // map[envVariableName]envValue
	Contracts         []ContractConfig
}

// LoadChainConfigs loads the config from a config.yaml file into an array of ChainConfig
func LoadChainConfigs() ([]ChainConfig, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/oracle/")
	viper.AddConfigPath("$HOME/.oracle")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("[main][LoadChainConfigs] %w", err)
	}

	return unmarshallConfig()
}

func unmarshallConfig() ([]ChainConfig, error) {
	c := &configWrapper{}
	err := viper.Unmarshal(c)
	if err != nil {
		return nil, fmt.Errorf("[main][LoadChainConfigs] Failed to unmarshal config yaml: %w", err)
	}

	for i, _ := range c.Chains {
		c.Chains[i].PollingSecs = c.Chains[i].PollingSecs * time.Second
		c.Chains[i].MaxRetryDelaySecs = c.Chains[i].MaxRetryDelaySecs * time.Second
		c.Chains[i].Id = i
	}

	return c.Chains, nil
}

// GenerateProviderClients uses the provider url to generate geth's ethclient.Client and rpc.Client
func (c *ChainConfig) GenerateProviderClients(ctx context.Context) (client *ethclient.Client, rpcClient *rpc.Client, err error) {
	for envKey, envValue := range c.Aws {
		err := os.Setenv(strings.ToUpper(envKey), envValue)
		if err != nil {
			return nil, nil, err
		}
	}
	return gethaws.CreateClients(ctx, c.Rpc, nil)
}

// GetAddressesSlice returns the chain contracts' addresses as a string slice
func (c *ChainConfig) GetAddressesSlice() []string {
	var addresses []string
	for _, contract := range c.Contracts {
		addresses = append(addresses, strings.ToLower(contract.Address))
	}
	return addresses
}

// GetCommonAddresses returns the chain contracts' addresses as a geth's common.Address slice
func (c *ChainConfig) GetCommonAddresses() []common.Address {
	var addresses []common.Address
	for _, contract := range c.Contracts {
		addresses = append(addresses, common.HexToAddress(contract.Address))
	}
	return addresses
}

// OldestCreationBlock gets the oldest synced block from the passed contracts
func (c *ChainConfig) OldestCreationBlock() (*big.Int, error) {
	oldestBlock := uint64(math.MaxUint64)
	for _, contract := range c.Contracts {
		startBlock := uint64(contract.StartBlock)
		if startBlock < oldestBlock {
			oldestBlock = startBlock
		}
	}
	return big.NewInt(int64(oldestBlock)), nil
}

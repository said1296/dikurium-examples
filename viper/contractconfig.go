package chainconfig

// ContractConfig holds the config for a single contract
type ContractConfig struct {
	Address    string
	Type       string
	StartBlock int64 `mapstructure:"start_block"`
}

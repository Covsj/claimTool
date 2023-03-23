package config

import "github.com/BurntSushi/toml"

type ClaimArbConfig struct {
	PrivateKeys        []string
	ArbiturmRpc        string
	EthereumRpc        string
	DistributorAddress string
	TokenAddress       string
	CollectAddress     string
	StartBlockNumber   int
	GasLimit           int
	GasPrice           int
}

func (c *ClaimArbConfig) LoadFromFile(path string) error {
	_, err := toml.DecodeFile(path, c)
	return err
}

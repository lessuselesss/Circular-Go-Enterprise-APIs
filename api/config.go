package api

import (
	"encoding/json"
	"os"
)

// NetworkConfig holds the configuration for different networks
type NetworkConfig struct {
	MainAccount struct {
		PrivateKey string `json:"private_key"`
		SeedPhrase string `json:"seed_phrase"`
	} `json:"main_account"`
	SecondaryAccount struct {
		PrivateKey string `json:"private_key"`
		SeedPhrase string `json:"seed_phrase"`
	} `json:"secondary_account"`
	Network string            `json:"network"`
	NagURLs map[string]string `json:"nag_urls"`
}

// Config holds the complete configuration
type Config struct {
	Testnet NetworkConfig `json:"testnet"`
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// GetNAGURL returns the NAG URL for a given network
func (c *Config) GetNAGURL(network string) string {
	switch network {
	case "testnet":
		return c.Testnet.NagURLs["testnet"]
	default:
		return "https://nag-testnet.circular.io" // Default fallback
	}
}
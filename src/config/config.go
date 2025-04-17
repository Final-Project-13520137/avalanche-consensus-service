package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/Final-Project-13520137/avalanche-consensus-service/src/models/consensus"
)

// Config represents the application configuration
type Config struct {
	ServerPort     int                      `json:"server_port"`
	NodeID         string                   `json:"node_id"`
	PeerAddresses  []string                 `json:"peer_addresses"`
	ConsensusParams consensus.AvalancheParams `json:"consensus_params"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		ServerPort:     8080,
		NodeID:         "node-1",
		PeerAddresses:  []string{},
		ConsensusParams: consensus.DefaultParams(),
	}
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return config, nil
	}

	// Read file
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// SaveConfig saves configuration to a JSON file
func SaveConfig(config *Config, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
} 
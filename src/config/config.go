package config

import (
	"github.com/spf13/viper"
	"log"
	"os/user"
	"path/filepath"
)

// Global configuration
var config Config

// Config definition
type Config struct {
	Eth *EthConfig `json:"eth"`
}

type EthConfig struct {
	Network *Network `json:"network"`
	Wallet  *Wallet  `json:"wallet"`
}

// Network configuration
type Network struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

// Wallet configuration
type Wallet struct {
	Password string `json:"password"`
	Keystore string `json:"keystore"`
}

// Loads the configuration from config URL
func Load(path string, name string) {
	viper.SetConfigName(name)
	viper.AddConfigPath(path)

	// TODO: verify missing parameters maybe using default/required struct tags
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading configuration file, %s", err)
	}
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("unable to decode config, %v", err)
	}
}

// Get configuration
func Get() *Config {
	return &config
}

func GetEthConfig() *EthConfig {
	return config.Eth
}

func GetNetworkConfig() *Network {
	return GetEthConfig().Network
}

func GetWalletConfig() *Wallet {
	return GetEthConfig().Wallet
}

func DefaultKeyStore() string {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(currentUser.HomeDir, "/dcvp_keystore")
}

package config

import (
	"github.com/spf13/viper"
	"log"
)

// Global configuration
var config Config

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

// Config definition
type Config struct {
	Network *Network `json:"network"`
}

// Network configuration
type Network struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

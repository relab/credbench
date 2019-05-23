package main

import (
	"flag"
	"fmt"

	"github.com/r0qs/dvcp/src/accounts"
	"github.com/r0qs/dvcp/src/client"
	"github.com/r0qs/dvcp/src/config"
)

var (
	configFilePath string
	configFileName string
)

func init() {
	parseFlags()
	config.Load(configFilePath, configFileName)
}

func parseFlags() {
	flag.StringVar(&configFilePath, "config-path", ".", "Config file path")
	flag.StringVar(&configFileName, "config-name", "dev-config", "Config file name")
	flag.Parse()
}

func main() {
	chainConfig := config.GetNetworkConfig()
	backend := fmt.Sprintf("http://%s:%d", chainConfig.Address, chainConfig.Port)
	key, _ := accounts.OpenKeystore()
	client.Connect(backend, key)
}

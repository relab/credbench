package helm

import (
	"os"
	"path/filepath"
	"text/template"

	log "github.com/sirupsen/logrus"
)

var (
	chartURL          = "https://bbchain.no/charts"
	imageRepo         = "ethereum/client-go"
	imageTag          = "alltools"
	imageVersion      = "v1.10.2"
	chart             = "bbchain-charts/geth"
	chartName         = "geth"
	chartVersion      = "0.0.2"
	namespace         = "ethereum"
	verbosity         = 3
	timeout           = 120
	enablePersistence = false
	enableClique      = true
	cliquePeriod      = 15
	cliqueEpoch       = 30000
)

const helmTmpl = `namespaces:
  {{ .Namespace }}:

helmRepos:
  bbchain-charts: "{{ .ChartURL }}"

name: {{ .ChartName }}
namespace: {{ .Namespace }}
description: "Ethereum Go Client"
chart: {{ .Chart }}
version: {{ .ChartVersion }}
enabled: true
wait: true
timeout: {{ .Timeout }}

imagePullPolicy: Always

bootnode:
  enabled: true
  image:
    repository: {{ .ImageRepo }}
    tag: {{ .ImageTag }}-{{ .ImageVersion }}
  config:
    verbosity: {{ .Verbosity }}
  resources: {}
  nodeSelector: {}
  tolerations: []
  affinity: {}

geth:
  image:
    repository: {{ .ImageRepo }}
    tag: {{ .ImageVersion }}
  replicaCount: {{ .ReplicaCount }}
  service:
    type: ClusterIP
  config:
    verbosity: {{ .Verbosity }}
    http:
      enable: true
      port: 8545
      api: debug,web3,eth,txpool,personal,clique,miner,net
      corsdomains: "localhost"
      vhosts: "localhost"
    cache: "1024"
    ws:
      enable: true
      port: 8546
      api: eth,net,web3,txpool
      origins: "*"
    mine: true
    extraFlags: [--nousb --allow-insecure-unlock]
  genesis:
    difficulty: "1"
    gasLimit: {{ .GasLimit }}
    gasTarget: {{ .GasLimit }}
    gasPrice: {{ .GasPrice }}
    networkId: {{ .NetworkID }}
    clique:
      enable: {{ .EnableClique }}
      period: {{ .CliquePeriod }}
      epoch: {{ .CliqueEpoch }}
    # This is the initial validator without the 0x prefix
    # Other validators can be added using the clique api method clique_propose
    # https://geth.ethereum.org/docs/rpc/ns-clique
    validatorAddress: {{ .ValidatorAddress }}
    validatorKey: {{ .ValidatorKey }}
  accounts:
    {{- range $account := .Accounts }}
    - address: "{{ $account.Address }}"
      balance: "{{ $account.Balance }}"
    {{- end }}
  persistence:
    enabled: {{ .EnablePersistence }}
    # storageClass: "-"
    accessMode: ReadWriteOnce
    size: 20Gi
  resources: {}
  nodeSelector: {
    chain: geth
  }
  tolerations: []
  affinity: {}
`

type AccountData struct {
	Address string
	Balance string
}

type HelmChartData struct {
	ChartURL          string
	ImageRepo         string
	ImageTag          string
	ImageVersion      string
	Chart             string
	ChartName         string
	ChartVersion      string
	Namespace         string
	EnablePersistence bool
	Verbosity         int
	Timeout           int
	ReplicaCount      int
	NetworkID         int
	GasLimit          string
	GasPrice          string
	EnableClique      bool
	CliquePeriod      int
	CliqueEpoch       int
	ValidatorAddress  string
	ValidatorKey      string
	Accounts          []AccountData
}

func newHelmChartData(networkID int, gasLimit, gasPrice string, validatorAddress, validatorKey string, replicaCount int, accounts []AccountData) *HelmChartData {
	return &HelmChartData{
		ChartURL:          chartURL,
		ImageRepo:         imageRepo,
		ImageTag:          imageTag,
		ImageVersion:      imageVersion,
		Chart:             chart,
		ChartName:         chartName,
		ChartVersion:      chartVersion,
		Namespace:         namespace,
		EnablePersistence: enablePersistence,
		Verbosity:         verbosity,
		Timeout:           timeout,
		ReplicaCount:      replicaCount,
		NetworkID:         networkID,
		GasLimit:          gasLimit,
		GasPrice:          gasPrice,
		EnableClique:      enableClique,
		CliquePeriod:      cliquePeriod,
		CliqueEpoch:       cliqueEpoch,
		ValidatorAddress:  validatorAddress,
		ValidatorKey:      validatorKey,
		Accounts:          accounts,
	}
}

func createDeploymentFile(path string, data *HelmChartData) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	tmpl := template.Must(template.New("").Parse(helmTmpl))
	if err := tmpl.Execute(f, data); err != nil {
		return err
	}

	log.Infof("Helm file successfully exported to %s", path)
	return nil
}

// ExportHelmFile creates a helm file ready to be used in the deploy based
// on the generated genesis accounts.
func ExportHelmFile(exportPath string, networkID int, gasLimit, gasPrice, defaultBalance string, validatorAddress, validatorKey string, replicaCount int, accounts []string) error {
	accountsData := make([]AccountData, len(accounts))
	for i, addr := range accounts {
		accountsData[i] = AccountData{Address: addr, Balance: defaultBalance}
	}

	helmFile := filepath.Join(exportPath, "deploy-values.yaml")
	return createDeploymentFile(helmFile, newHelmChartData(networkID, gasLimit, gasPrice, validatorAddress, validatorKey, replicaCount, accountsData))
}

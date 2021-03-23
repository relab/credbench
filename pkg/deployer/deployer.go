package deployer

import (
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func DeployContract(auth *bind.TransactOpts, backend bind.ContractBackend, contractABI string, contractCode string, params ...interface{}) (common.Address, *types.Transaction, *bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, c, err := bind.DeployContract(auth, parsed, common.FromHex(contractCode), backend, params...)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, c, nil
}

func LinkContract(contractCode string, libs map[string]string) string {
	if len(libs) == 0 {
		return contractCode
	}
	// map of libname to deployed address
	code := []byte(contractCode)
	for name, address := range libs {
		re := regexp.MustCompile("_*" + name + "_*_")
		address = strings.ToLower(address)
		code = re.ReplaceAll(code, []byte(address[2:]))
	}
	return string(code)
}

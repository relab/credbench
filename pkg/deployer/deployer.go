package deployer

import (
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
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

// FindLinkReferences returns a map of the reference hash to the
// deployed address for easy linking.
func FindLinkReferences(deployedLibs map[string]string, linkReferences map[string]string) (map[string]string, error) {
	if len(deployedLibs) != len(linkReferences) {
		return nil, errors.New("deployed libraries do not match given references")
	}
	libs := make(map[string]string, len(deployedLibs))
	for libName, libAddress := range deployedLibs {
		ref, exists := linkReferences[libName]
		if !exists {
			return nil, errors.New("library reference required but not found")
		}
		libs[ref] = libAddress
	}
	return libs, nil
}

// LinkContract links the libraries in the contract bytecode
// by replacing the libraries placeholders by their deployed address
// The parameter libs contains a map of references to deployed address
func LinkContract(contractCode string, libs map[string]string) string {
	if len(libs) == 0 {
		return contractCode
	}
	code := []byte(contractCode)
	for ref, address := range libs {
		re := regexp.MustCompile("_*\\$" + ref + "\\$_*_")
		address = strings.ToLower(address)
		code = re.ReplaceAll(code, []byte(address[2:]))
	}
	return string(code)
}

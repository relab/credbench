# Decentralized Verifiable Credential Platform

## Dependencies

| Software                                         | Version  | Required for |
| ------------------------------------------------ | -------- | ------------ |
| [Solidity](https://github.com/ethereum/solidity) | >=0.8.2  | contracts    |
| [NPM](https://github.com/npm/cli)                | >=7.6.3  | contracts    |
| [Go](https://golang.org/doc/go1.16)              | >=1.16.2 | client       |
| [GNU Make](http://ftp.gnu.org/gnu/make/)         | >=4.3.0  | client       |
| [geth](https://github.com/ethereum/go-ethereum)  | >=1.9.24 | client       |

## Contracts

The contracts are available in the repository [bbchain-contracts](https://github.com/relab/bbchain-contracts) and its go-ethereum bindings are available in the [bbchain-bindings](https://github.com/relab/bbchain-bindings) repository.

The current bindings are available in a private repository, thus to be able to install the bindings using go, please do the procedure described below.

Add the following to your `~/.gitconfig`
```
[url "ssh://git@github.com/"]
	insteadOf = https://github.com/
```

Then run the command below to allow go to install from a private repository:
```
go env -w GOPRIVATE="github.com/relab/bbchain-bindings"
```

Then compile the code:
```
make all
```

* For more information please see: https://golang.org/ref/mod#private-modules

## Running ganache installed in the local machine
```
$ ganache-cli --deterministic --host 127.0.0.1 --port 8545 --networkId 5777 --gasLimit 6721975 --gasPrice 20000000000 --verbose
```

If you prefer, you can use geth instead of [ganache](https://truffleframework.com/ganache) using the command below.
```
$ ./genesis-gen.sh
$ geth --networkid=42 --nodiscover --rpc --rpcport=8545 --ws --wsport=8546 --rpccorsdomain="*" --datadir ./dev_datadir --dev.period 0 --miner.gasprice=20000000000 --miner.gastarget=6721975 --miner.gaslimit=6721975 --verbosity 5 --mine --miner.etherbase $(jq -r '.alloc | keys_unsorted[0]' dev_datadir/genesis.json) --miner.noverify --maxpeers 0 --password ./dev_datadir/password.txt --unlock $(jq -r '.alloc | keys_unsorted[0]' dev_datadir/genesis.json) --allow-insecure-unlock
```

## Command-line Client

See the available options by running the benchmark command-line tool:

```
./dist/ctbench
```

You can attach to the ganache-cli console using geth:
```
$ geth attach http://127.0.0.1:8545
```

### Testing

#### Issuing a Credential

1. Import a testing ganache account that already have funds
```
./dist/ctbench --config dev-config.json account import <hex_private_key>
```

2. Deploy libraries
```
./dist/ctbench --config dev-config.json deploy libs
```

1. Deploy a course
```
./dist/ctbench --config dev-config.json deploy course --owners=<teacher_address>,<another_teacher_address> --quorum=2
```

4. Adding a student

```
./dist/ctbench --config dev-config.json course addStudent <student_address> <course_address>
```

5. Issuing a credential
```
./dist/ctbench --config dev-config.json course issue <student_address> <course_address> credential.json
```

To see all available commands, please type:
```
./dist/ctbench help
```

### Running Test Scenarios

```
./dist/ctbench --config dev-config.json test generate
./dist/ctbench --config dev-config.json test run
```

### Inspecting the Test Database

Install boltbrowser
```
go get github.com/br0xen/boltbrowser
```

Run passing the db
```
boltbrowser dev_datadir/database/cteth.db
```
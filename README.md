# Decentralized Verifiable Credential Platform

## Dependencies

| Software                                         | Version             | Notes         |
| ------------------------------------------------ | ------------------- | ------------- |
| [Solidity](https://github.com/ethereum/solidity) | >=0.6.4             | contracts     |
| [NPM](https://github.com/npm/cli)                | >=6.14.4            | contracts     |
| [Go](https://golang.org/doc/go1.12)              | >=1.14.1            | client        |
| [GNU Make](http://ftp.gnu.org/gnu/make/)         | >=4.3.0             | client        |
| [geth](https://github.com/ethereum/go-ethereum)  | >=1.9.12 (optional) | tests on both |


## Contracts

Install and start a test chain.
```
cd src/ethereum
npm install
npm run compile
npm run generate
npm run ganache-cli
```

In other terminal run the command below to run the contract tests.
```
npm run test:ganache
```

If you prefer, you can use geth instead of [ganache](https://truffleframework.com/ganache) using the command below.
```
geth --networkid=42 --nodiscover --rpc --rpcport=8545 --ws --wsport=8546 --rpccorsdomain="*" --datadir ./dev_datadir --dev.period 0 --miner.gasprice=20000000000 --miner.gastarget=6721975 --miner.gaslimit=6721975 --verbosity 5 --mine --miner.etherbase $(jq -r '.alloc | keys_unsorted[0]' dev_datadir/genesis.json) --miner.noverify --maxpeers 0 --password ./dev_datadir/password.txt --unlock $(jq -r '.alloc | keys_unsorted[0]' dev_datadir/genesis.json) --allow-insecure-unlock
```


## Client

```
cd src
make all
```

Then run a test chain using the ganache-cli or geth as previously and execute from the source folder the client binary.

```
./bin/ctethapp
```

You can attach to the ganache-cli console using geth:
```
$ geth attach http://127.0.0.1:8545
```

### Testing

#### Issuing a credential

0. Run ganache ethereum client
```
$ ganache-cli --deterministic --host 127.0.0.1 --port 8545 --networkId 5777 --gasLimit 6721975 --gasPrice 20000000000 --verbose
```

1. Import a testing ganache account that already have funds
```
./bin/ctethapp --config dev-config.json account import 4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d
```

2. Deploy libraries
```
./bin/ctethapp --config dev-config.json deploy all
```

3. Deploy a course using the imported account
```
./bin/ctethapp --config dev-config.json course deploy --owners=0x90f8bf6a479f320ead074411a4b0e7944ea8c9c1 --quorum=1
```

4. Initializing issuer
```
./bin/ctethapp --config dev-config.json course initialize 0xCfEB869F69431e42cdB54A4F4f105C19C080A601
```

5. Adding a student
```
./bin/ctethapp --config dev-config.json course addStudent 0xCfEB869F69431e42cdB54A4F4f105C19C080A601 0xffcf8fdee72ac11b5c542428b35eef5769c409f0
```

6. Issuing a credential
```
./bin/ctethapp --config dev-config.json course issue 0xCfEB869F69431e42cdB54A4F4f105C19C080A601 0xFFcf8FDEE72ac11b5c542428B35EEF5769C409f0 exam_1_sample.json
```
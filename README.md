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
geth --networkid=42 --nodiscover \
     --rpc --rpcport=8545 --ws --wsport=8546 --rpccorsdomain="*" \
     --dev --dev.period 0 \
     --datadir=/<YOUR_PATH_TO>/devchain console 2>>dev.log
```

## Client

```
cd src
make all
```

Then run a test chain using the ganache-cli or geth as previously and execute from the source folder the client binary.

```
./bin/bbchain
```

You can attach to the ganache-cli console using geth:
```
$ geth attach http://127.0.0.1:8545
```

### Testing

#### Issuing a credential

0. Run ganache ethereum client
```
$ ganache-cli --deterministic --host 127.0.0.1 --port 8545
```

1. Import a testing ganache account that already have funds
```
./bin/bbchain --config dev-config.json account import 4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d
```

2. Deploy a course using the imported account
```
./bin/bbchain --config dev-config.json course deploy --owners=0x90f8bf6a479f320ead074411a4b0e7944ea8c9c1 --quorum=1
```

3. Adding a student
```
./bin/bbchain --config dev-config.json course addStudent 0x5b1869D9A4C187F2EAa108f3062412ecf0526b24 0xffcf8fdee72ac11b5c542428b35eef5769c409f0
```

4. Issuing a credential

```
./bin/bbchain --config dev-config.json course issue 0x5b1869D9A4C187F2EAa108f3062412ecf0526b24 0xFFcf8FDEE72ac11b5c542428B35EEF5769C409f0 exam_1_sample.json
```

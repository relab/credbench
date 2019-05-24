# Decentralized Verifiable Credential Platform

## Dependencies

| Software  | Version | Notes |
| --------- | ------- | ----- |
| [Solidity](https://github.com/ethereum/solidity) | >=0.5.7 | contracts |
| [NPM](https://github.com/npm/cli) | >=6.9.0 | contracts |
| [Go](https://golang.org/doc/go1.12) | >=1.12.5 | client |
| [GNU Make](http://ftp.gnu.org/gnu/make/) | >=4.2.1 | client |
| [geth](https://github.com/ethereum/go-ethereum) | >=1.8.27 (optional) | tests on both |


## Contracts

Install and start a test chain.
```
cd src/ethereum
npm install
npm run ganache-cli
```

In other terminal run the command below to run the contract tests.
```
npm test
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
make deps
make
```

Then run a test chain using the ganache-cli or geth as previously and execute from the source folder the client binary.

```
./bin/client
```
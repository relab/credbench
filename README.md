# Decentralized Verifiable Credential Platform

## Contracts
Install and start a test chain
```
cd src/ethereum
npm install
npm run ganache-cli
```

In other terminal run the command below to run the contract tests
```
npm test
```

If you prefer, you can use geth
```
geth --networkid=42 --nodiscover \
     --rpc --rpcport=8545 --ws --wsport=8546 --rpccorsdomain="*" \
     --dev --dev.period 0 \
     --datadir=/<YOUR_PATH_TO>/devchain console 2>>dev.log
```

## Client

```
cd src
make generate
make
./client-linux
```
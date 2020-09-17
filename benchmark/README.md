# Benchmark generator

## How to run

### Compiling protobuf files 
```console
$ protoc -I=./proto --go_out=paths=source_relative:./proto ./proto/*.proto
```

### Generate genesis

Initial setup to generate 100 accounts and give ether to them.
```console
$ ./benchmark --config dev-config.json genesis 20
$ geth --networkid=42 --nodiscover --datadir=./dev_datadir init dev_datadir/genesis.json
```

### Run an ethereum client
```console
$ geth --networkid=42 --nodiscover \
     --rpc --rpcport=8545 --ws --wsport=8546 --rpccorsdomain="*" \
     --dev.period 0 \
     --datadir=./dev_datadir --miner.gasprice=20000000000 --miner.gastarget=6721975 --miner.gaslimit=6721975 --verbosity 6
```

### Generate test cases

```console
$ ./benchmark --config dev-config.json generate
```

### Deploy a course

Deploy a course contract with 2 distinct evaluators and 20 distinct students

```console
$ ./benchmark --config dev-config.json course generate 2 20
```

#### TODO
- parse test file with configured distribution of accounts and contracts(random, predefined, etc)
- start/deploy eth clients from cmd line
- update accounts and faculty bucket with deployed course data
- deploy faculty + update accounts and faculty bucket with deployed data
- start certification process following configured distribution
- monitor/simulate costs (change gas according with mainnet) but using [Proof-of-Authority consensus](https://github.com/ethereum/EIPs/issues/225)
- verify credentials
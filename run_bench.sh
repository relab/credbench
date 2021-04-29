#!/bin/bash

# Project directory
BASE_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
BIN_DIR=${BASE_PATH}/dist
BENCH_BIN=${BIN_DIR}/ctbench

GAS_LIMIT=6721975
GAS_PRICE=20000000000

BENCH_CONFIG="${BENCH_CONFIG:-dev-config.json}"
echo "Using benchmark config: ${BENCH_CONFIG}"


function generate_genesis() {
  echo "Removing development datadir..."
  rm -rf dev_datadir

  TEST_CONFIG="${1:-test-config.json}"
  echo "Using test config: ${TEST_CONFIG}"

  ${BENCH_BIN} --config ${BENCH_CONFIG} genesis $(jq -r '.total_accounts' ${TEST_CONFIG})

  geth --networkid=42 --nodiscover --datadir=./dev_datadir init dev_datadir/genesis.json
}

# | Courses | Exams per Course | Students |
# |---------|------------------|----------|
# | 10      | 2, 5             | 200      |
# | 100     | 2, 5             | 200      |
# | 1000    | 2                | 100      |

function generate_test_config() {
  echo "Creatinf test cases directory..."
  mkdir -p ./testcases
  for c in 10 100 1000; do
    for x in 2 5; do
      for s in 100 200; do
        echo "Generating config for test case: $c-$x-$s"
        ${BENCH_BIN} --config ${BENCH_CONFIG} test generate case "./testcases/test_${c}_${x}_${s}.json" -t 1000 -c $c -x $x -s $s
      done
    done
  done
}

function run_geth() {
  geth --networkid=42 --nodiscover --rpc --rpcport=8545 --ws --wsport=8546 --rpccorsdomain="*" --datadir ./dev_datadir --dev.period 0 --miner.gasprice=${GAS_PRICE} --miner.gastarget=${GAS_LIMIT} --miner.gaslimit=${GAS_LIMIT} --verbosity 5 --mine --miner.etherbase $(jq -r '.alloc | keys_unsorted[0]' dev_datadir/genesis.json) --miner.noverify --maxpeers 0 --password ./dev_datadir/password.txt --unlock $(jq -r '.alloc | keys_unsorted[0]' dev_datadir/genesis.json) --allow-insecure-unlock
}

function run_ganache() {
  ganache-cli --deterministic --host 127.0.0.1 --port 8545 --networkId 5777 --gasLimit ${GAS_LIMIT} --gasPrice ${GAS_PRICE}
}

usage() {
    echo "usage: ${0} [option]"
    echo 'options:'
    echo '    -genesis Generate genesis file'
    echo '    -testcase Generate test case configuration'
    echo '    -ganache Runs pre-configure ganache'
    echo '    -geth Runs pre-configure geth'
    echo
}

option="${1}"
case ${option} in
    -genesis) generate_genesis "${@:2}";;
    -testcase) generate_test_config "${@:2}";;
    -ganache) run_ganache "${@:2}";;
    -geth) run_geth "${@:2}";;
    *) usage; exit 1;;
esac
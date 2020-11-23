rm -rf dev_datadir
./dist/ctbench --config dev-config.json genesis $(jq -r '.total_accounts' test-config.json)
geth --networkid=42 --nodiscover --datadir=./dev_datadir init dev_datadir/genesis.json
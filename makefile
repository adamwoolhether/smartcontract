# #######################################################################
# Install dependencies
# https://geth.ethereum.org/docs/install-and-build/installing-geth
# https://docs.soliditylang.org/en/v0.8.17/installing-solidity.html

dev.setup:
	brew update
	brew list ethereum || brew install ethereum
	brew list solidity || brew install solidity

dev.update:
	brew update
	brew upgrade ethereum
	brew upgrade solidity

# #######################################################################
# Go-Ethereum Commands

# Start in developer mode, open UNIX socket, http calls, and JSONRPC
# A Coinbase account is given to unlock in the dev env.
geth-up:
	geth --dev --ipcpath zarf/ethereum/geth.ipc \
	--http.corsdomain '*' --http --allow-insecure-unlock --rpc.allow-unprotected-txs \
	--mine --miner.threads 1 --verbosity 5 --datadir "zarf/ethereum/" \
	--unlock 0x6327A38415C53FFb36c11db55Ea74cc9cB4976Fd --password zarf/ethereum/password

geth-down:
	kill -INT $(shell ps -eo pid,comm | grep " geth" |awk '{print $$1}')1

geth-reset:
	rm -rf zarf/ethereum/geth/

# Open a JS console environment for geth API calls.
geth-attach:
	geth attach --datadir zarf/ethereum/

# Add a new account to the keystore with zero balance.
geth-new-account:
	get --datadir zarf/ethereum/ account new

geth-deposit:
	curl -H 'Content-Type: application/json' --data '{"jsonrpc":"2.0","method":"eth_sendTransaction", "params": [{"from":"0x6327A38415C53FFb36c11db55Ea74cc9cB4976Fd", "to":"0x8E113078ADF6888B7ba84967F299F29AeCe24c55", "value":"0x1000000000000000000"}], "id":1}' localhost:8545
	curl -H 'Content-Type: application/json' --data '{"jsonrpc":"2.0","method":"eth_sendTransaction", "params": [{"from":"0x6327A38415C53FFb36c11db55Ea74cc9cB4976Fd", "to":"0x0070742FF6003c3E809E78D524F0Fe5dcc5BA7F7", "value":"0x1000000000000000000"}], "id":1}' localhost:8545
	curl -H 'Content-Type: application/json' --data '{"jsonrpc":"2.0","method":"eth_sendTransaction", "params": [{"from":"0x6327A38415C53FFb36c11db55Ea74cc9cB4976Fd", "to":"0x7FDFc99999f1760e8dBd75a480B93c7B8386B79a", "value":"0x1000000000000000000"}], "id":1}' localhost:8545
	curl -H 'Content-Type: application/json' --data '{"jsonrpc":"2.0","method":"eth_sendTransaction", "params": [{"from":"0x6327A38415C53FFb36c11db55Ea74cc9cB4976Fd", "to":"0x000cF95cB5Eb168F57D0bEFcdf6A201e3E1acea9", "value":"0x1000000000000000000"}], "id":1}' localhost:8545


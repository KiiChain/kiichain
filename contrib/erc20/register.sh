# Apply the upgrade proposal
kiichaind tx gov submit-proposal contrib/erc20/MsgRegisterERC20.json --keyring-backend test --from mykey --fees 1000000000000000000akii -y
sleep 5

# Vote for the proposal
kiichaind tx gov vote 1 yes --keyring-backend test --from mykey --fees 1000000000000000000akii -y
sleep 5
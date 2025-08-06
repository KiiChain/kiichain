package main

import (
	"os"

	// Import the params to set the onchain config
	_ "github.com/kiichain/kiichain/v3/app/params"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	app "github.com/kiichain/kiichain/v3/app"
	"github.com/kiichain/kiichain/v3/cmd/kiichaind/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()

	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}

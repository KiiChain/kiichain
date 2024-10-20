package main

import (
	"os"

	"github.com/kiichain/kiichain3/app/params"
	"github.com/kiichain/kiichain3/cmd/kiichaind/cmd"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/kiichain/kiichain3/app"
)

func main() {
	params.SetAddressPrefixes()
	rootCmd, _ := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}

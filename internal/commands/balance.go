package commands

import (
	"github.com/spf13/cobra"
	"github.com/theblitlabs/parity-client/cmd/cli"
)

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Check token balances and stake status",
	Run: func(cmd *cobra.Command, args []string) {
		cli.RunBalance()
	},
}

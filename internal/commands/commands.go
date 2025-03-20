package commands

import "github.com/spf13/cobra"

func AddCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(stakeCmd)
	rootCmd.AddCommand(balanceCmd)
}

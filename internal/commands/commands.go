package commands

import "github.com/spf13/cobra"

// AddCommands adds all CLI commands to the root command
func AddCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(stakeCmd)
	rootCmd.AddCommand(balanceCmd)
}

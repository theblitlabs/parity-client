package commands

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/theblitlabs/parity-client/cmd/cli"
)

var stakeCmd = &cobra.Command{
	Use:   "stake",
	Short: "Stake tokens in the network",
	Run: func(cmd *cobra.Command, args []string) {
		cli.RunStake(cmd, args)
	},
}

func init() {
	stakeCmd.Flags().Float64P("amount", "a", 1.0, "Amount of tokens to stake")
	if err := stakeCmd.MarkFlagRequired("amount"); err != nil {
		log.Error().Err(err).Msg("Failed to mark amount flag as required")
	}
}

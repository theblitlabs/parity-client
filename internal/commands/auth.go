package commands

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/theblitlabs/parity-client/cmd/cli"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with the network",
	Run: func(cmd *cobra.Command, args []string) {
		cli.RunAuth()
	},
}

func init() {
	authCmd.Flags().String("private-key", "", "Private key in hex format")
	if err := authCmd.MarkFlagRequired("private-key"); err != nil {
		log.Error().Err(err).Msg("Failed to mark private-key flag as required")
	}
}

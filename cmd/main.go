package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/cmd/cli"
	"github.com/theblitlabs/parity-client/internal/commands"
	"github.com/theblitlabs/parity-client/internal/config"
	"github.com/theblitlabs/parity-client/internal/utils"
)

var (
	logMode    string
	configPath string
)

var rootCmd = &cobra.Command{
	Use:   "parity-client",
	Short: "Parity Protocol CLI",
	Long:  `A decentralized computing network powered by blockchain and secure enclaves`,
	Run: func(cmd *cobra.Command, args []string) {
		configManager := config.NewConfigManager(configPath)
		cfg, err := configManager.GetConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		cli.RunChain(cfg.Server.Port, cmd)
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		switch logMode {
		case "debug", "pretty", "info", "prod", "test":
			gologger.InitWithMode(gologger.LogMode(logMode))
		default:
			gologger.InitWithMode(gologger.LogModePretty)
		}
		if configPath == "" {
			configPath = utils.GetDefaultConfigPath()
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logMode, "log", "pretty", "Log mode: debug, pretty, info, prod, test")
	rootCmd.PersistentFlags().StringVar(&configPath, "config-path", "", "Path to config file (default: ~/.parity/.env or ./.env)")

	commands.AddCommands(rootCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

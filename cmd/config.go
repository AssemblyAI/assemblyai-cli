/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"fmt"

	U "github.com/AssemblyAI/assemblyai-cli/utils"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config [token]",
	Short: "Authenticate the CLI",
	Long:  `This command will validate your account and store your token safely, later to be used when transcribing files.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please provide a token. If you don't have one, create an account at https://app.assemblyai.com")
			return
		} else if len(args) > 1 {
			fmt.Println("Too many arguments. Please provide a single token.")
			return
		}
		U.Token = args[0]

		checkToken := U.CheckIfTokenValid()
		if !checkToken {
			fmt.Println("Your token appears to be invalid. Try again, and if the problem persists, contact support at support@assemblyai.com")
			return
		}

		if U.GetConfigFileValue("config.new") == "true" {
			U.SetUserAlias()
		}

		U.CreateConfigFile()
		U.SetConfigFileValue("features.telemetry", "true")
		U.SetConfigFileValue("config.token", U.Token)
		U.SetConfigFileValue("config.distinct_id", U.DistinctId)
		U.SetConfigFileValue("config.new", "false")

		U.TelemetryCaptureEvent("CLI configured", nil)

		fmt.Println("You're now authenticated.")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}

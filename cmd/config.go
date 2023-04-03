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

		argsArray := cmd.Flags().Args()

		if len(argsArray) == 0 {
			fmt.Println("Please provide a token. If you don't have one, create an account at https://app.assemblyai.com")
			return
		} else if len(argsArray) > 1 {
			fmt.Println("Too many arguments. Please provide a single token.")
			return
		}
		U.Token = argsArray[0]

		checkToken := U.CheckIfTokenValid()
		if !checkToken {
			fmt.Println(U.INVALID_TOKEN)
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
	configCmd.Flags().Bool("test", false, "Flag for test executing purpose")
	configCmd.Flags().MarkHidden("test")

	rootCmd.AddCommand(configCmd)
}

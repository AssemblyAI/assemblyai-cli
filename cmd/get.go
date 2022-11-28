/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"errors"

	S "github.com/AssemblyAI/assemblyai-cli/schemas"
	U "github.com/AssemblyAI/assemblyai-cli/utils"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [transcription_id]",
	Short: "Get a transcription",
	Long:  `After submitting a file for transcription, you can fetch it by passing its ID.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("No transcription ID provided."),
				Message: "You must provide a transcription ID.",
			}
			U.PrintError(printErrorProps)
			return
		}
		id := args[0]

		U.Token = U.GetStoredToken()
		if U.Token == "" {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("No token found."),
				Message: "Please start by running \033[1m\033[34massemblyai config [token]\033[0m",
			}
			U.PrintError(printErrorProps)
			return
		}

		U.PollTranscription(id, flags)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.PersistentFlags().BoolVarP(&flags.Poll, "poll", "p", true, "The CLI will poll the transcription until it's complete.")
	getCmd.PersistentFlags().BoolVarP(&flags.Json, "json", "j", false, "If true, the CLI will output the JSON.")
	getCmd.PersistentFlags().StringVar(&flags.Csv, "csv", "", "Specify the filename to save the transcript result onto a .CSV file extension")
}

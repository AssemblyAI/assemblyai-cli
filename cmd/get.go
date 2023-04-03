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

// get represents the getTranscription command
var getCmd = &cobra.Command{
	Use:   "get [transcription_id]",
	Short: "Get a transcription",
	Long:  `After submitting a file for transcription, you can fetch it by passing its ID.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var flags S.TranscribeFlags
		args = cmd.Flags().Args()
		if len(args) == 0 {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("No transcription ID provided."),
				Message: "You must provide a transcription ID.",
			}
			U.PrintError(printErrorProps)
			return
		}
		id := args[0]
		flags.Poll, _ = cmd.Flags().GetBool("poll")
		flags.Json, _ = cmd.Flags().GetBool("json")
		flags.Srt, _ = cmd.Flags().GetBool("srt")

		U.Token = U.GetStoredToken()
		if U.Token == "" {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("No token found."),
				Message: "Please start by running \033[1m\033[34massemblyai config [token]\033[0m",
			}
			U.PrintError(printErrorProps)
			return
		}

		checkToken := U.CheckIfTokenValid()
		if !checkToken {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("Invalid token"),
				Message: U.INVALID_TOKEN,
			}
			U.PrintError(printErrorProps)
			return
		}

		U.PollTranscription(id, flags)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().BoolP("json", "j", false, "If true, the CLI will output the JSON.")
	getCmd.Flags().BoolP("poll", "p", true, "The CLI will poll the transcription until it's complete.")
	getCmd.Flags().Bool("test", false, "Flag for test executing purpose")
	getCmd.PersistentFlags().BoolP("srt", "", false, "Generate an SRT file for the audio file transcribed.")
	getCmd.Flags().MarkHidden("test")
}

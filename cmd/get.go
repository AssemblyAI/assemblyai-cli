/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// get represents the getTranscription command
var getCmd = &cobra.Command{
	Use:   "get [transcription_id]",
	Short: "Get a transcription",
	Long:  `After submitting a file for transcription, you can fetch it by passing its ID.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var flags TranscribeFlags
		args = cmd.Flags().Args()
		if len(args) == 0 {
			fmt.Println("You must provide a transcription ID.")
			return
		}
		id := args[0]
		flags.Poll, _ = cmd.Flags().GetBool("poll")
		flags.Json, _ = cmd.Flags().GetBool("json")

		Token = GetStoredToken()
		if Token == "" {
			fmt.Println("Please start by running \033[1m\033[34massemblyai config [token]\033[0m")
			return
		}

		PollTranscription(id, flags)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().BoolP("json", "j", false, "If true, the CLI will output the JSON.")
	getCmd.Flags().BoolP("poll", "p", true, "The CLI will poll the transcription until it's complete.")
}

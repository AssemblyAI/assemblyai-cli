/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// get represents the getTranscription command
var get = &cobra.Command{
	Use:   "get [transcription ID]",
	Short: "Getting the Transcription Result",
	Long:  `After you submit an audio file for processing, get transcription response including the results of any Audio Intelligence features you enabled, using your unique transcription ID as a parameter.`,
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
			fmt.Println("You must login first. Run `assemblyai config <token>`")
			return
		}

		PollTranscription(id, flags)
	},
}

func init() {
	rootCmd.AddCommand(get)
	get.Flags().BoolP("json", "j", false, "If true, the CLI will output the JSON.")
	get.Flags().BoolP("poll", "p", true, "The CLI will poll the transcription until it's complete.")
}

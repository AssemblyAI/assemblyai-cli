/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// getTranscriptionCmd represents the getTranscription command
var getTranscriptionCmd = &cobra.Command{
	Use:   "getTranscription [transcription ID]",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var flags TranscribeFlags
		args = cmd.Flags().Args()
		if len(args) == 0 {
			fmt.Println("You must provide an audio URL")
			return
		}
		id := args[0]
		flags.Poll, _ = cmd.Flags().GetBool("poll")
		flags.Json, _ = cmd.Flags().GetBool("json")

		db := GetOpenDatabase()
		token := GetStoredToken(db)
		defer db.Close()

		if token == "" {
			fmt.Println("You must login first")
		}
		PollTranscription(token, id, flags)
	},
}

func init() {
	rootCmd.AddCommand(getTranscriptionCmd)
	getTranscriptionCmd.Flags().BoolP("json", "j", false, "If true, the CLI will output the JSON.")
	getTranscriptionCmd.Flags().BoolP("poll", "p", true, "The CLI will poll the transcription until it's complete.")
}

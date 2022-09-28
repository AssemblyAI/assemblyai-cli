/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// transcribeCmd represents the transcribe command
var transcribeCmd = &cobra.Command{
	Use:   "transcribe",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
		and usage of using your command. For example:

		Cobra is a CLI library for Go that empowers applications.
		This application is a tool to generate the needed files
		to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var flags TranscribeFlags

		flags.Poll, _ = cmd.Flags().GetBool("poll")
		flags.Punctuate, _ = cmd.Flags().GetBool("punctuate")
		flags.FormatText, _ = cmd.Flags().GetBool("format_text")
		flags.DualChannel, _ = cmd.Flags().GetBool("dual_channel")
		flags.Json, _ = cmd.Flags().GetBool("json")
		flags.RedactPii, _ = cmd.Flags().GetBool("redact_pii")
		flags.PiiPolicies, _ = cmd.Flags().GetString("pii_policies")
		flags.AutoHighlights, _ = cmd.Flags().GetBool("auto_highlights")
		flags.ContentModeration, _ = cmd.Flags().GetBool("content_moderation")
		flags.TopicDetection, _ = cmd.Flags().GetBool("topic_detection")
		flags.SentimentAnalysis, _ = cmd.Flags().GetBool("sentiment_analysis")
		flags.AutoChapters, _ = cmd.Flags().GetBool("auto_chapters")
		flags.EntityDetection, _ = cmd.Flags().GetBool("entity_detection")

		Transcribe(flags)
	},
}

func init() {
	rootCmd.AddCommand(transcribeCmd)
	transcribeCmd.PersistentFlags().BoolP("poll", "p", true, "The CLI will poll the transcription until it's complete.")
	transcribeCmd.PersistentFlags().BoolP("punctuate", "u", true, "Enable automatic punctuation.")
	transcribeCmd.PersistentFlags().BoolP("format_text", "f", true, "Enable text formatting")
	transcribeCmd.PersistentFlags().BoolP("dual_channel", "d", false, "Enable dual channel")
	transcribeCmd.PersistentFlags().BoolP("json", "j", false, "If true, the CLI will output the JSON.")
	transcribeCmd.PersistentFlags().BoolP("redact_pii", "r", false, "Remove personally identifiable information from the transcription.")
	transcribeCmd.PersistentFlags().StringP("pii_policies", "i", "drug,number_sequence,person_name", "The list of PII policies to redact (source), comma-separated. Required if the redact_pii flag is true, with the default value including drugs, number sequences, and person names.")
	transcribeCmd.PersistentFlags().BoolP("auto_highlights", "a", false, "Automatically detect important phrases and words in the text.")
	transcribeCmd.PersistentFlags().BoolP("content_moderation", "c", false, "Detect if sensitive content is spoken in the file.")
	transcribeCmd.PersistentFlags().BoolP("topic_detection", "t", false, "Label the topics that are spoken in the file.")
	transcribeCmd.PersistentFlags().BoolP("sentiment_analysis", "x", false, "Detect the sentiment of each sentence of speech spoken in the file.")
	transcribeCmd.PersistentFlags().BoolP("auto_chapters", "s", false, "A \"summary over time\" for the audio file transcribed.")
	transcribeCmd.PersistentFlags().BoolP("entity_detection", "e", false, "Identify a wide range of entities that are spoken in the audio file.")
}

func Transcribe(flags TranscribeFlags) {
	db := GetOpenDatabase()
	token := GetStoredToken(db)
	if token != "" {
		fmt.Printf("Your Token is %s\n", token)
		defer db.Close()

		fmt.Println(flags.Poll)
		// var transcribeResponse TranscriptResponse
		// response := QueryApi(token, "/account", "GET", nil)

		return
	}
}

type TranscribeFlags struct {
	Poll bool
	Punctuate bool
	FormatText bool
	DualChannel bool
	Json bool
	RedactPii bool
	PiiPolicies string
	AutoHighlights bool
	ContentModeration bool
	TopicDetection bool
	SentimentAnalysis bool
	AutoChapters bool
	EntityDetection bool
}
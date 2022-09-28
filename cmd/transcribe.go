/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// transcribeCmd represents the transcribe command
var transcribeCmd = &cobra.Command{
	Use:   "transcribe [file path or URL]",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
		and usage of using your command. For example:

		Cobra is a CLI library for Go that empowers applications.
		This application is a tool to generate the needed files
		to quickly create a Cobra application.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var params TranscribeParams
		var flags TranscribeFlags

		args = cmd.Flags().Args()
		if len(args) == 0 {
			fmt.Println("You must provide an audio URL")
			return
		}
		params.AudioURL = args[0]

		flags.Poll, _ = cmd.Flags().GetBool("poll")
		flags.Json, _ = cmd.Flags().GetBool("json")
		// params.PiiPolicies, _ = cmd.Flags().GetString("pii_policies")
		params.AutoChapters, _ = cmd.Flags().GetBool("auto_chapters")
		params.AutoHighlights, _ = cmd.Flags().GetBool("auto_highlights")
		params.ContentModeration, _ = cmd.Flags().GetBool("content_moderation")
		params.DualChannel, _ = cmd.Flags().GetBool("dual_channel")
		params.EntityDetection, _ = cmd.Flags().GetBool("entity_detection")
		params.FormatText, _ = cmd.Flags().GetBool("format_text")
		params.Punctuate, _ = cmd.Flags().GetBool("punctuate")
		params.RedactPii, _ = cmd.Flags().GetBool("redact_pii")
		params.SentimentAnalysis, _ = cmd.Flags().GetBool("sentiment_analysis")
		params.SpeakerLabels, _ = cmd.Flags().GetBool("speaker_labels")
		params.TopicDetection, _ = cmd.Flags().GetBool("topic_detection")

		Transcribe(params, flags)
	},
}

func init() {
	// transcribeCmd.PersistentFlags().StringP("pii_policies", "i", "drug,number_sequence,person_name", "The list of PII policies to redact (source), comma-separated. Required if the redact_pii flag is true, with the default value including drugs, number sequences, and person names.")
	transcribeCmd.Flags().BoolP("auto_chapters", "s", false, "A \"summary over time\" for the audio file transcribed.")
	transcribeCmd.Flags().BoolP("auto_highlights", "a", false, "Automatically detect important phrases and words in the text.")
	transcribeCmd.Flags().BoolP("content_moderation", "c", false, "Detect if sensitive content is spoken in the file.")
	transcribeCmd.Flags().BoolP("dual_channel", "d", false, "Enable dual channel")
	transcribeCmd.Flags().BoolP("entity_detection", "e", false, "Identify a wide range of entities that are spoken in the audio file.")
	transcribeCmd.Flags().BoolP("format_text", "f", true, "Enable text formatting")
	transcribeCmd.Flags().BoolP("json", "j", false, "If true, the CLI will output the JSON.")
	transcribeCmd.Flags().BoolP("poll", "p", true, "The CLI will poll the transcription until it's complete.")
	transcribeCmd.Flags().BoolP("punctuate", "u", true, "Enable automatic punctuation.")
	transcribeCmd.Flags().BoolP("redact_pii", "r", false, "Remove personally identifiable information from the transcription.")
	transcribeCmd.Flags().BoolP("sentiment_analysis", "x", false, "Detect the sentiment of each sentence of speech spoken in the file.")
	transcribeCmd.Flags().BoolP("speaker_labels", "l", true, "Automatically detect the number of speakers in your audio file, and each word in the transcription text can be associated with its speaker.")
	transcribeCmd.Flags().BoolP("topic_detection", "t", false, "Label the topics that are spoken in the file.")
	rootCmd.AddCommand(transcribeCmd)
}

func Transcribe(params TranscribeParams, flags TranscribeFlags) {
	db := GetOpenDatabase()
	token := GetStoredToken(db)
	defer db.Close()

	if token == "" {
		fmt.Println("You must login first")
	}

	paramsJSON, err := json.Marshal(params)
	PrintError(err)
	body := bytes.NewReader(paramsJSON)

	response := QueryApi(token, "/transcript", "POST", body)
	var transcriptResponse TranscriptResponse
	if err := json.Unmarshal(response, &transcriptResponse); err != nil {
		fmt.Println("Can not unmarshal JSON")
		return
	}

	id := transcriptResponse.ID
	if !flags.Poll {
		if flags.Json {
			fmt.Println(string(response))
		}

		fmt.Printf("Your transcription was created (id %s)\n", id)
		return
	}

	PollTranscription(token, id, flags)
}

type TranscribeFlags struct {
	Poll bool `json:"poll"`
	Json bool `json:"json"`
}

type TranscribeParams struct {
	// PiiPolicies       string `json:"pii_policies"`
	AudioURL          string `json:"audio_url"`
	AutoChapters      bool   `json:"auto_chapters"`
	AutoHighlights    bool   `json:"auto_highlights"`
	ContentModeration bool   `json:"content_safety"`
	DualChannel       bool   `json:"dual_channel"`
	EntityDetection   bool   `json:"entity_detection"`
	FormatText        bool   `json:"format_text"`
	Punctuate         bool   `json:"punctuate"`
	RedactPii         bool   `json:"redact_pii"`
	SentimentAnalysis bool   `json:"sentiment_analysis"`
	SpeakerLabels     bool   `json:"speaker_labels"`
	TopicDetection    bool   `json:"iab_categories"`
}

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

var flags S.TranscribeFlags
var params S.TranscribeParams

var transcribeCmd = &cobra.Command{
	Use:   "transcribe <url | path | youtube URL>",
	Short: "Transcribe and understand audio with a single AI-powered API",
	Long: `Automatically convert audio and video files and live audio streams to text with AssemblyAI's Speech-to-Text APIs. 
	Do more with Audio Intelligence - summarization, content moderation, topic detection, and more. 
	Powered by cutting-edge AI models.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("Please provide a URL, path, or YouTube URL"),
				Message: "Please provide a local file, a file URL or a YouTube URL to be transcribed.",
			}
			U.PrintError(printErrorProps)
			return
		}
		params.AudioURL = args[0]

		U.ValidateParams(params, cmd.Flags())
		U.ValidateFlags(flags)

		U.Transcribe(params, flags)
	},
}

func init() {
	transcribeCmd.PersistentFlags().BoolVarP(&flags.Poll, "poll", "p", true, "The CLI will poll the transcription until it's complete.")
	transcribeCmd.PersistentFlags().BoolVarP(&flags.Json, "json", "j", false, "If true, the CLI will output the JSON.")
	transcribeCmd.PersistentFlags().StringVar(&flags.Csv, "csv", "", "Specify the filename to save the transcript result onto a .CSV file extension")
	transcribeCmd.PersistentFlags().BoolVarP(&params.AutoChapters, "auto_chapters", "s", false, "A \"summary over time\" for the audio file transcribed.")
	transcribeCmd.PersistentFlags().BoolVarP(&params.AutoHighlights, "auto_highlights", "a", false, "Automatically detect important phrases and words in the text.")
	transcribeCmd.PersistentFlags().BoolVarP(&params.ContentModeration, "content_moderation", "c", false, "Detect if sensitive content is spoken in the file.")
	transcribeCmd.PersistentFlags().BoolVarP(&params.DualChannel, "dual_channel", "d", false, "Enable dual channel")
	transcribeCmd.PersistentFlags().BoolVarP(&params.EntityDetection, "entity_detection", "e", false, "Identify a wide range of entities that are spoken in the audio file.")
	transcribeCmd.PersistentFlags().BoolVarP(&params.FormatText, "format_text", "f", true, "Enable text formatting")
	transcribeCmd.PersistentFlags().BoolVarP(&params.LanguageDetection, "language_detection", "n", false, "Identify the dominant language that’s spoken in an audio file.")
	transcribeCmd.PersistentFlags().BoolVarP(&params.Punctuate, "punctuate", "u", true, "Enable automatic punctuation.")
	transcribeCmd.PersistentFlags().BoolVarP(&params.RedactPii, "redact_pii", "r", false, "Remove personally identifiable information from the transcription.")
	transcribeCmd.PersistentFlags().BoolVarP(&params.SentimentAnalysis, "sentiment_analysis", "x", false, "Detect the sentiment of each sentence of speech spoken in the file.")
	transcribeCmd.PersistentFlags().BoolVarP(&params.SpeakerLabels, "speaker_labels", "l", true, "Automatically detect the number of speakers in your audio file, and each word in the transcription text can be associated with its speaker.")
	transcribeCmd.PersistentFlags().BoolVarP(&params.Summarization, "summarization", "m", false, "Generate a single abstractive summary of the entire audio.")
	transcribeCmd.PersistentFlags().BoolVarP(&params.TopicDetection, "topic_detection", "t", false, "Label the topics that are spoken in the file.")
	transcribeCmd.PersistentFlags().StringSliceVarP(&params.RedactPiiPolicies, "redact_pii_policies", "i", []string{"drug", "number_sequence", "person_name"}, "The list of PII policies to redact, comma-separated without space in-between. Required if the redact_pii flag is true.")
	transcribeCmd.PersistentFlags().StringSliceVarP(&params.WordBoost, "word_boost", "k", nil, "The value of this flag MUST be used surrounded by quotes. Any term included will have its likelihood of being transcribed boosted.")
	transcribeCmd.PersistentFlags().StringVarP(&params.BoostParam, "boost_param", "z", "", "Control how much weight should be applied to your boosted keywords/phrases. This value can be either low, default, or high.")
	transcribeCmd.PersistentFlags().StringVarP(&params.LanguageCode, "language_code", "g", "", "Specify the language of the speech in your audio file.")
	transcribeCmd.PersistentFlags().StringVarP(&params.SummaryModel, "summary_model", "q", "informative", "The model used to generate the summary.")
	transcribeCmd.PersistentFlags().StringVarP(&params.SummaryType, "summary_type", "y", "bullets", "Type of summary generated.")
	transcribeCmd.PersistentFlags().StringVarP(&params.WebhookAuthHeaderName, "webhook_auth_header_name", "b", "", "Containing the header's name which will be inserted into the webhook request")
	transcribeCmd.PersistentFlags().StringVarP(&params.WebhookAuthHeaderValue, "webhook_auth_header_value", "o", "", "The value of the header that will be inserted into the webhook request.")
	transcribeCmd.PersistentFlags().StringVarP(&params.WebhookURL, "webhook_url", "w", "", "Receive a webhook once your transcript is complete.")

	transcribeCmd.PersistentFlags().String("custom_spelling", "", "Specify how words are spelled or formatted in the transcript text.")

	rootCmd.AddCommand(transcribeCmd)
}

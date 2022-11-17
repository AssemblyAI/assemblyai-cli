/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"errors"
	"fmt"
	"strings"

	S "github.com/AssemblyAI/assemblyai-cli/schemas"
	U "github.com/AssemblyAI/assemblyai-cli/utils"
	"github.com/spf13/cobra"
)

var transcribeCmd = &cobra.Command{
	Use:   "transcribe <url | path | youtube URL>",
	Short: "Transcribe and understand audio with a single AI-powered API",
	Long: `Automatically convert audio and video files and live audio streams to text with AssemblyAI's Speech-to-Text APIs. 
	Do more with Audio Intelligence - summarization, content moderation, topic detection, and more. 
	Powered by cutting-edge AI models.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var params S.TranscribeParams
		var flags S.TranscribeFlags

		args = cmd.Flags().Args()
		if len(args) == 0 {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("Please provide a URL, path, or YouTube URL"),
				Message: "Please provide a local file, a file URL or a YouTube URL to be transcribed.",
			}
			U.PrintError(printErrorProps)
			return
		}
		params.AudioURL = args[0]

		flags.Json, _ = cmd.Flags().GetBool("json")
		flags.Poll, _ = cmd.Flags().GetBool("poll")
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
		params.Summarization, _ = cmd.Flags().GetBool("summarization")
		if params.Summarization {
			params.SummaryType, _ = cmd.Flags().GetString("summary_type")
			if _, ok := S.SummarizationTypeMapReverse[params.SummaryType]; !ok {
				printErrorProps := S.PrintErrorProps{
					Error:   errors.New("Invalid summary type"),
					Message: "Invalid summary type. To know more about Summarization, head over to https://assemblyai.com/docs/audio-intelligence#summarization",
				}
				U.PrintError(printErrorProps)
				return
			}

		}

		if params.RedactPii {
			policies, _ := cmd.Flags().GetString("redact_pii_policies")
			policiesArray := strings.Split(policies, ",")

			for _, policy := range policiesArray {
				if _, ok := S.PIIRedactionPolicyMap[policy]; !ok {
					printErrorProps := S.PrintErrorProps{
						Error:   errors.New("Invalid redaction policy"),
						Message: fmt.Sprintf("%s is not a valid policy. See https://www.assemblyai.com/docs/audio-intelligence#pii-redaction for the complete list of supported policies.", policy),
					}
					U.PrintError(printErrorProps)
					return
				}
			}

			params.RedactPiiPolicies = policiesArray
		}
		webhook := cmd.Flags().Lookup("webhook_url").Value.String()
		if webhook != "" {
			params.WebhookURL = webhook
			webhookHeaderName := cmd.Flags().Lookup("webhook_auth_header_name").Value.String()
			webhookHeaderValue := cmd.Flags().Lookup("webhook_auth_header_value").Value.String()
			if webhookHeaderName != "" {
				params.WebhookAuthHeaderName = webhookHeaderName
			}
			if webhookHeaderValue != "" {
				params.WebhookAuthHeaderValue = webhookHeaderValue
			}
		}
		languageDetection, _ := cmd.Flags().GetBool("language_detection")
		languageCode, _ := cmd.Flags().GetString("language_code")
		if (languageCode != "" || languageDetection) && params.SpeakerLabels {
			if cmd.Flags().Lookup("speaker_labels").Changed {
				printErrorProps := S.PrintErrorProps{
					Error:   errors.New("Speaker labels are not supported for languages other than English"),
					Message: "Speaker labels are not supported for languages other than English.",
				}
				U.PrintError(printErrorProps)
				return
			} else {
				params.SpeakerLabels = false
			}
		}
		if languageDetection && languageCode == "" {
			params.LanguageDetection = true
		}
		if languageCode != "" {
			if _, ok := S.LanguageMap[languageCode]; !ok {
				printErrorProps := S.PrintErrorProps{
					Error:   errors.New("Invalid language code"),
					Message: "Invalid language code. See https://www.assemblyai.com/docs#supported-languages for supported languages.",
				}
				U.PrintError(printErrorProps)
				return
			}
			params.LanguageCode = &languageCode
			params.LanguageDetection = false
		}

		U.Transcribe(params, flags)
	},
}

func init() {
	transcribeCmd.PersistentFlags().BoolP("auto_chapters", "s", false, "A \"summary over time\" for the audio file transcribed.")
	transcribeCmd.PersistentFlags().BoolP("auto_highlights", "a", false, "Automatically detect important phrases and words in the text.")
	transcribeCmd.PersistentFlags().BoolP("content_moderation", "c", false, "Detect if sensitive content is spoken in the file.")
	transcribeCmd.PersistentFlags().BoolP("dual_channel", "d", false, "Enable dual channel")
	transcribeCmd.PersistentFlags().BoolP("entity_detection", "e", false, "Identify a wide range of entities that are spoken in the audio file.")
	transcribeCmd.PersistentFlags().BoolP("format_text", "f", true, "Enable text formatting")
	transcribeCmd.PersistentFlags().BoolP("json", "j", false, "If true, the CLI will output the JSON.")
	transcribeCmd.PersistentFlags().BoolP("language_detection", "n", false, "Identify the dominant language that’s spoken in an audio file.")
	transcribeCmd.PersistentFlags().BoolP("poll", "p", true, "The CLI will poll the transcription until it's complete.")
	transcribeCmd.PersistentFlags().BoolP("punctuate", "u", true, "Enable automatic punctuation.")
	transcribeCmd.PersistentFlags().BoolP("redact_pii", "r", false, "Remove personally identifiable information from the transcription.")
	transcribeCmd.PersistentFlags().BoolP("sentiment_analysis", "x", false, "Detect the sentiment of each sentence of speech spoken in the file.")
	transcribeCmd.PersistentFlags().BoolP("speaker_labels", "l", true, "Automatically detect the number of speakers in your audio file, and each word in the transcription text can be associated with its speaker.")
	transcribeCmd.PersistentFlags().BoolP("summarization", "m", false, "Generate a single abstractive summary of the entire audio.")
	transcribeCmd.PersistentFlags().BoolP("topic_detection", "t", false, "Label the topics that are spoken in the file.")
	transcribeCmd.PersistentFlags().StringP("language_code", "g", "", "Specify the language of the speech in your audio file.")
	transcribeCmd.PersistentFlags().StringP("redact_pii_policies", "i", "drug,number_sequence,person_name", "The list of PII policies to redact, comma-separated without space in-between. Required if the redact_pii flag is true.")
	transcribeCmd.PersistentFlags().StringP("summary_type", "y", "bullets", "Type of summary generated.")
	transcribeCmd.PersistentFlags().StringP("webhook_auth_header_name", "b", "", "Containing the header's name which will be inserted into the webhook request")
	transcribeCmd.PersistentFlags().StringP("webhook_auth_header_value", "o", "", "The value of the header that will be inserted into the webhook request.")
	transcribeCmd.PersistentFlags().StringP("webhook_url", "w", "", "Receive a webhook once your transcript is complete.")
	rootCmd.AddCommand(transcribeCmd)
}

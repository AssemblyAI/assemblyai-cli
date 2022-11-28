/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

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

		if params.WordBoost == nil && params.BoostParam != "" {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("Please provide a valid word boost"),
				Message: "To boost a word, please provide a valid list of words to boost. For example: --word_boost \"word1,word2,word3\"  --boost_param high",
			}
			U.PrintError(printErrorProps)
			return
		} else if params.BoostParam != "" && params.BoostParam != "low" && params.BoostParam != "default" && params.BoostParam != "high" {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("Invalid boost_param"),
				Message: "Please provide a valid boost_param. Valid values are low, default, or high.",
			}
			U.PrintError(printErrorProps)
			return
		}

		if !params.Summarization {
			params.SummaryType = ""
			params.SummaryModel = ""
		} else {
			params.Punctuate = true
			params.FormatText = true
			if _, ok := S.SummarizationTypeMapReverse[params.SummaryType]; !ok {
				printErrorProps := S.PrintErrorProps{
					Error:   errors.New("Invalid summary type"),
					Message: "Invalid summary type. To know more about Summarization, head over to https://assemblyai.com/docs/audio-intelligence#summarization",
				}
				U.PrintError(printErrorProps)
				return
			}
			if _, ok := S.SummarizationModelMap[params.SummaryModel]; !ok {
				printErrorProps := S.PrintErrorProps{
					Error:   errors.New("Invalid summary model"),
					Message: "Invalid summary model. To know more about Summarization, head over to https://assemblyai.com/docs/audio-intelligence#summarization",
				}
				U.PrintError(printErrorProps)
				return
			}
			if !U.Contains(S.SummarizationModelMap[params.SummaryModel], params.SummaryType) {
				printErrorProps := S.PrintErrorProps{
					Error:   errors.New("Invalid summary model"),
					Message: "Cant use summary model " + params.SummaryModel + " with summary type " + params.SummaryType + ". To know more about Summarization, head over to https://assemblyai.com/docs/audio-intelligence#summarization",
				}
				U.PrintError(printErrorProps)
				return
			}
			if params.SummaryModel == "conversational" && !params.SpeakerLabels {
				printErrorProps := S.PrintErrorProps{
					Error:   errors.New("Speaker labels required for conversational summary model"),
					Message: "Speaker labels are required for conversational summarization. To know more about Summarization, head over to https://assemblyai.com/docs/audio-intelligence#summarization",
				}
				U.PrintError(printErrorProps)
				return
			}
		}
		if !params.RedactPii {
			params.RedactPiiPolicies = nil
		} else {
			for _, policy := range params.RedactPiiPolicies {
				if _, ok := S.PIIRedactionPolicyMap[policy]; !ok {
					printErrorProps := S.PrintErrorProps{
						Error:   errors.New("Invalid redaction policy"),
						Message: fmt.Sprintf("%s is not a valid policy. See https://www.assemblyai.com/docs/audio-intelligence#pii-redaction for the complete list of supported policies.", policy),
					}
					U.PrintError(printErrorProps)
					return
				}
			}
		}

		if params.LanguageDetection && params.LanguageCode != "" {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("Language detection and language code cannot be used together"),
				Message: "Language detection and language code cannot be used together.",
			}
			U.PrintError(printErrorProps)
			return
		}
		if (params.LanguageCode != "" || params.LanguageDetection) && params.SpeakerLabels {
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
		if params.LanguageCode != "" {
			if _, ok := S.LanguageMap[params.LanguageCode]; !ok {
				printErrorProps := S.PrintErrorProps{
					Error:   errors.New("Invalid language code"),
					Message: "Invalid language code. See https://www.assemblyai.com/docs#supported-languages for supported languages.",
				}
				U.PrintError(printErrorProps)
				return
			}
		}

		customSpelling, _ := cmd.Flags().GetString("custom_spelling")
		if customSpelling != "" {
			parsedCustomSpelling := []S.CustomSpelling{}

			_, err := os.Stat(customSpelling)

			if !os.IsNotExist(err) {
				file, err := os.Open(customSpelling)
				if err != nil {
					printErrorProps := S.PrintErrorProps{
						Error:   err,
						Message: "Error opening custom spelling file",
					}
					U.PrintError(printErrorProps)
					return
				}
				defer file.Close()
				byteCustomSpelling, err := ioutil.ReadAll(file)
				if err != nil {
					printErrorProps := S.PrintErrorProps{
						Error:   err,
						Message: "Error reading custom spelling file",
					}
					U.PrintError(printErrorProps)
					return
				}

				err = json.Unmarshal(byteCustomSpelling, &parsedCustomSpelling)
				if err != nil {
					printErrorProps := S.PrintErrorProps{
						Error:   err,
						Message: "Error parsing custom spelling file",
					}
					U.PrintError(printErrorProps)
					return
				}
			} else {
				err = json.Unmarshal([]byte(customSpelling), &parsedCustomSpelling)
				if err != nil {
					printErrorProps := S.PrintErrorProps{
						Error:   err,
						Message: "Invalid custom spelling. Please provide a valid custom spelling JSON.",
					}
					U.PrintError(printErrorProps)
					return
				}
			}

			err = U.ValidateCustomSpelling(parsedCustomSpelling)
			if err != nil {
				printErrorProps := S.PrintErrorProps{
					Error:   err,
					Message: "Invalid custom spelling. Please provide a valid custom spelling JSON.",
				}
				U.PrintError(printErrorProps)
				return
			}
			params.CustomSpelling = parsedCustomSpelling
		}

		if flags.Csv != "" && !flags.Poll {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("CSV output is only supported with polling"),
				Message: "CSV output is only supported with polling.",
			}
			U.PrintError(printErrorProps)
			return
		}

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

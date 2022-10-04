/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/posthog/posthog-go"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// transcribeCmd represents the transcribe command
var transcribeCmd = &cobra.Command{
	Use:   "transcribe <url | path>",
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
			fmt.Println("Please provide a local file or a URL to be transcribed.")
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
	token := GetStoredToken()

	if token == "" {
		fmt.Println("You must login first. Run `assemblyai config <token>`")
		return
	}

	isYoutubeLink := isYoutubeLink(params.AudioURL)

	if isYoutubeLink {
		fmt.Println("Youtube link is not yet supported, please provide a file Url or path")
		return
	}

	_, err := url.ParseRequestURI(params.AudioURL)
	if err != nil {
		uploadedURL := UploadFile(token, params.AudioURL)
		if uploadedURL == "" {
			fmt.Println("The file doesn't exist. Please try again with a different one.")
			return
		}
		params.AudioURL = uploadedURL
	}

	isAAICDN := checkAAICDN(params.AudioURL)

	if !isAAICDN {
		resp, err := http.Get(params.AudioURL)
		if err != nil || resp.StatusCode != 200 {
			fmt.Println("We couldn't transcribe the file in the URL. Please try again with a different one.")
			return
		}
	}

	paramsJSON, err := json.Marshal(params)
	PrintError(err)

	TelemetryCaptureEvent("CLI transcription created", map[string]interface{}{})
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
			print := BeutifyJSON(response)
			fmt.Println(string(print))
			return
		}
		fmt.Printf("Your transcription was created (id %s)\n", *id)
		return
	}

	PollTranscription(token, *id, flags)
}

func isYoutubeLink(url string) bool {
	return strings.HasPrefix(url, "https://www.youtube.com/watch?v=")
}

func checkAAICDN(url string) bool {
	return strings.HasPrefix(url, "https://cdn.assemblyai.com/")
}

func PollTranscription(token string, id string, flags TranscribeFlags) {
	s := CallSpinner(" Your file is being transcribed (id " + id + ")... Processing time is usually 20% of the file's duration.")
	for {
		response := QueryApi(token, "/transcript/"+id, "GET", nil)

		if response == nil {
			s.Stop()
			fmt.Println("Something went wrong. Please try again later.")
			return
		}

		var transcript TranscriptResponse
		if err := json.Unmarshal(response, &transcript); err != nil {
			fmt.Println(err)
			s.Stop()
			return
		}

		if transcript.Error != nil {
			s.Stop()
			fmt.Println(*transcript.Error)
			return
		}
		if *transcript.Status == "completed" {
			s.Stop()

			properties := posthog.NewProperties().
				Set("poll", flags.Poll).
				Set("json", flags.Json).
				Set("speaker_labels", *transcript.SpeakerLabels).
				Set("punctuate", *transcript.Punctuate).
				Set("format_text", *transcript.FormatText).
				Set("dual_channel", *transcript.DualChannel).
				Set("redact_pii", *transcript.RedactPii).
				Set("auto_highlights", *transcript.AutoHighlights).
				Set("content_moderation", *transcript.ContentSafety).
				Set("topic_detection", *transcript.IabCategories).
				Set("sentiment_analysis", *transcript.SentimentAnalysis).
				Set("auto_chapters", *transcript.AutoChapters).
				Set("entity_detection", *transcript.EntityDetection)

			TelemetryCaptureEvent("CLI transcription finished", properties)

			if flags.Json {
				print := BeutifyJSON(response)
				fmt.Println(string(print))
				return
			}
			GetFormattedOutput(transcript, flags)
			return
		}
		time.Sleep(5 * time.Second)
	}
}

func GetFormattedOutput(transcript TranscriptResponse, flags TranscribeFlags) {
	width, _, err := term.GetSize(0)
	if err != nil {
		width = 512
	}

	fmt.Println("Transcript")
	if *transcript.SpeakerLabels == true {
		GetFormattedUtterances(*transcript.Utterances, width)
	} else {
		fmt.Println(*transcript.Text)
	}
	if *transcript.DualChannel == true {
		fmt.Println("\nDual Channel")
		GetFormattedDualChannel(*transcript.Utterances, width)
	}
	if *transcript.AutoHighlights == true {
		fmt.Println("Highlights")
		GetFormattedHighlights(*transcript.AutoHighlightsResult)
	}
	if *transcript.ContentSafety == true {
		fmt.Println("Content Moderation")
		GetFormattedContentSafety(*transcript.ContentSafetyLabels, width)
	}
	if *transcript.IabCategories == true {
		fmt.Println("Topic Detection")
		GetFormattedTopicDetection(*transcript.IabCategoriesResult, width)
	}
	if *transcript.SentimentAnalysis == true {
		fmt.Println("Sentiment Analysis")
		GetFormattedSentimentAnalysis(*transcript.SentimentAnalysisResults, width)
	}
	if *transcript.AutoChapters == true {
		fmt.Println("Chapters")
		GetFormattedChapters(*transcript.Chapters, width)
	}
	if *transcript.EntityDetection == true {
		fmt.Println("Entity Detection")
		GetFormattedEntityDetection(*transcript.Entities, width)
	}
}

func GetFormattedDualChannel(utterances []SentimentAnalysisResult, width int) {
	textWidth := width - 21
	w := tabwriter.NewWriter(os.Stdout, 10, 1, 1, ' ', 0)
	for _, utterance := range utterances {
		duration := time.Duration(utterance.Start) * time.Millisecond
		start := fmt.Sprintf("%02d:%02d", int(duration.Minutes()), int(duration.Seconds())%60)
		speaker := fmt.Sprintf("(Channel %s)", *utterance.Channel)

		if len(utterance.Text) > textWidth {
			for i := 0; i < len(utterance.Text); i += textWidth {
				end := i + textWidth
				if end > len(utterance.Text) {
					end = len(utterance.Text)
				}
				fmt.Fprintf(w, "%s  %s  %s\n", start, speaker, utterance.Text[i:end])
				start = "        "
				speaker = "        "
			}
		} else {
			fmt.Fprintf(w, "%s  %s  %s\n", start, speaker, utterance.Text)
		}
	}
	fmt.Fprintln(w)
	w.Flush()
}

func GetFormattedUtterances(utterances []SentimentAnalysisResult, width int) {
	textWidth := width - 21
	w := tabwriter.NewWriter(os.Stdout, 10, 1, 1, ' ', 0)
	for _, utterance := range utterances {
		duration := time.Duration(utterance.Start) * time.Millisecond
		start := fmt.Sprintf("%02d:%02d", int(duration.Minutes()), int(duration.Seconds())%60)
		speaker := fmt.Sprintf("(Speaker %s)", utterance.Speaker)

		if len(utterance.Text) > textWidth {
			for i := 0; i < len(utterance.Text); i += textWidth {
				end := i + textWidth
				if end > len(utterance.Text) {
					end = len(utterance.Text)
				}
				fmt.Fprintf(w, "%s  %s  %s\n", start, speaker, utterance.Text[i:end])
				start = "        "
				speaker = "        "
			}
		} else {
			fmt.Fprintf(w, "%s  %s  %s\n", start, speaker, utterance.Text)
		}
	}
	fmt.Fprintln(w)
	w.Flush()
}

func GetFormattedHighlights(highlights AutoHighlightsResult) {
	if highlights.Status != "success" {
		fmt.Println("Could not retrieve highlights")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 10, 10, 1, '\t', 0)
	fmt.Fprintf(w, "| COUNT\t | TEXT\t\n")
	for _, highlight := range highlights.Results {
		fmt.Fprintf(w, "| %s\t | %s\t\n", strconv.FormatInt(highlight.Count, 10), highlight.Text)
	}
	fmt.Fprintln(w)
	w.Flush()
}

func GetFormattedContentSafety(labels ContentSafetyLabels, width int) {
	if labels.Status != "success" {
		fmt.Println("Could not retrieve content safety labels")
		return
	}
	textWidth := width - 20
	labelWidth := 13

	w := tabwriter.NewWriter(os.Stdout, 1, 10, 1, '\t', 0)
	fmt.Fprintf(w, "| LABEL\t | TEXT\t\n")
	for _, label := range labels.Results {
		var labelString string
		for _, innerLabel := range label.Labels {
			labelString = innerLabel.Label + " " + labelString
		}

		if len(label.Text) > textWidth || len(labelString) > 30 {
			maxLength := int(math.Max(float64(len(label.Text)), float64(len(labelString))))

			x := 0
			for i := 0; i < maxLength; i += textWidth {
				labelStart := x
				labelEnd := x + labelWidth
				if labelEnd > len(labelString) {
					if x > len(labelString) {
						labelStart = len(labelString)
					}
					labelEnd = len(labelString)
				}
				textStart := i
				textEnd := i + textWidth
				if textEnd > len(label.Text) {
					if i > len(label.Text) {
						textStart = len(label.Text)
					}
					textEnd = len(label.Text)
				}
				fmt.Fprintf(w, "| %s\t | %s\t\n", labelString[labelStart:labelEnd], label.Text[textStart:textEnd])

				x += labelWidth
			}

		} else {
			fmt.Fprintf(w, "| %s\t | %s\t\n", labelString, label.Text)
		}

	}
	fmt.Fprintln(w)
	w.Flush()
}

func GetFormattedTopicDetection(categories IabCategoriesResult, width int) {
	if categories.Status != "success" {
		fmt.Println("Could not retrieve topic detection")
		return
	}
	textWidth := int(math.Abs(float64(width - 80)))

	w := tabwriter.NewWriter(os.Stdout, 40, 8, 1, '\t', 0)
	fmt.Fprintf(w, "| TOPIC \t| TEXT\n")
	for _, category := range categories.Results {
		if textWidth < 20 {
			fmt.Fprintf(w, "| %s\n", category.Labels[0].Label)
			fmt.Fprintf(w, "| %s\n", category.Text)
		} else if len(category.Text) > textWidth || len(category.Labels) > 1 {
			labelWidth := 0
			for i, innerLabel := range category.Labels {
				if i < 3 {
					labelWidth = int(math.Max(float64(len(innerLabel.Label)), float64(labelWidth)))
				}
			}
			maxLength := int(math.Max(float64(len(category.Text)), float64(labelWidth)))
			x := 0
			for i := 0; i < maxLength; i += textWidth {
				label := ""
				if x < 3 && x < len(category.Labels) {
					label = category.Labels[x].Label
				}
				textStart := i
				textEnd := i + textWidth
				if textEnd > len(category.Text) {
					if i > len(category.Text) {
						textStart = len(category.Text)
					}
					textEnd = len(category.Text)
				}

				fmt.Fprintf(w, "| %s\t| %s\n", label, category.Text[textStart:textEnd])
				x += 1
			}
		} else {
			fmt.Fprintf(w, "| %s\t| %s\n", category.Labels[0].Label, category.Text)
		}

		fmt.Fprintf(w, "| \t| \n")
	}
	fmt.Fprintln(w)
	w.Flush()
}

func GetFormattedSentimentAnalysis(sentiments []SentimentAnalysisResult, width int) {
	if len(sentiments) == 0 {
		fmt.Println("Could not retrieve sentiment analysis")
		return
	}
	textWidth := width - 20

	w := tabwriter.NewWriter(os.Stdout, 10, 8, 1, '\t', 0)
	fmt.Fprintf(w, "| SENTIMENT\t | TEXT\t\n")
	for _, sentiment := range sentiments {
		sentimentStatus := *sentiment.Sentiment
		if len(sentiment.Text) > textWidth {
			maxLength := len(sentiment.Text)

			for i := 0; i < maxLength; i += textWidth {
				textStart := i
				textEnd := i + textWidth
				if textEnd > len(sentiment.Text) {
					if i > len(sentiment.Text) {
						textStart = len(sentiment.Text)
					}
					textEnd = len(sentiment.Text)
				}
				fmt.Fprintf(w, "| %s\t | %s\t\n", sentimentStatus, sentiment.Text[textStart:textEnd])
				sentimentStatus = ""
			}

		} else {
			fmt.Fprintf(w, "| %s\t | %s\t\n", sentimentStatus, sentiment.Text)
		}

	}
	fmt.Fprintln(w)
	w.Flush()
}

func GetFormattedChapters(chapters []Chapter, width int) {
	if len(chapters) == 0 {
		fmt.Println("Could not retrieve chapters")
		return
	}
	textWidth := width - 21

	w := tabwriter.NewWriter(os.Stdout, 10, 8, 1, '\t', 0)
	for _, chapter := range chapters {
		// Gist
		fmt.Fprintf(w, "| Gist\t | %s\n", chapter.Gist)
		fmt.Fprintf(w, "| \t | \n")

		// Headline
		headline := "Headline"
		if len(chapter.Headline) > textWidth {
			maxLength := len(chapter.Headline)

			for i := 0; i < maxLength; i += textWidth {
				textStart := i
				textEnd := i + textWidth
				if textEnd > len(chapter.Headline) {
					if i > len(chapter.Headline) {
						textStart = len(chapter.Headline)
					}
					textEnd = len(chapter.Headline)
				}
				fmt.Fprintf(w, "| %s\t | %s\n", headline, chapter.Headline[textStart:textEnd])
				headline = ""
			}

		} else {
			fmt.Fprintf(w, "| %s\t | %s\n", headline, chapter.Headline)
		}

		fmt.Fprintf(w, "| \t | \n")
		// Summary
		summary := "Summary"
		if len(chapter.Summary) > textWidth {
			maxLength := len(chapter.Summary)

			for i := 0; i < maxLength; i += textWidth {
				textStart := i
				textEnd := i + textWidth
				if textEnd > len(chapter.Summary) {
					if i > len(chapter.Summary) {
						textStart = len(chapter.Summary)
					}
					textEnd = len(chapter.Summary)
				}
				fmt.Fprintf(w, "| %s\t | %s\n", summary, chapter.Summary[textStart:textEnd])
				summary = ""
			}

		} else {
			fmt.Fprintf(w, "| %s\t | %s\n", summary, chapter.Summary)
		}
		fmt.Fprintf(w, "| \t | \n")
	}
	fmt.Fprintln(w)
	w.Flush()
}

func GetFormattedEntityDetection(entities []Entity, width int) {
	if len(entities) == 0 {
		fmt.Println("Could not retrieve entity detection")
		return
	}
	textWidth := width - 20

	w := tabwriter.NewWriter(os.Stdout, 10, 8, 1, '\t', 0)
	fmt.Fprintf(w, "| TYPE\t | TEXT\t\n")
	for _, entity := range entities {
		if len(entity.Text) > textWidth {
			maxLength := len(entity.Text)

			for i := 0; i < maxLength; i += textWidth {
				textStart := i
				textEnd := i + textWidth
				if textEnd > len(entity.Text) {
					if i > len(entity.Text) {
						textStart = len(entity.Text)
					}
					textEnd = len(entity.Text)
				}
				fmt.Fprintf(w, "| %s\t | %s\t\n", entity.EntityType, entity.Text[textStart:textEnd])
			}

		} else {
			fmt.Fprintf(w, "| %s\t | %s\t\n", entity.EntityType, entity.Text)
		}

	}
	fmt.Fprintln(w)
	w.Flush()
}

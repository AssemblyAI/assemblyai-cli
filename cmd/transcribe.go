/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	uitable "github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	pb "gopkg.in/cheggaaa/pb.v1"
)

var transcribeCmd = &cobra.Command{
	Use:   "transcribe <url | path | youtube URL>",
	Short: "Transcribe and understand audio with a single AI-powered API",
	Long: `Automatically convert audio and video files and live audio streams to text with AssemblyAI's Speech-to-Text APIs. 
	Do more with Audio Intelligence - summarization, content moderation, topic detection, and more. 
	Powered by cutting-edge AI models.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var params TranscribeParams
		var flags TranscribeFlags

		args = cmd.Flags().Args()
		if len(args) == 0 {
			fmt.Println("Please provide a local file, a file URL or a YouTube URL to be transcribed.")
			return
		}
		params.AudioURL = args[0]

		flags.Poll, _ = cmd.Flags().GetBool("poll")
		flags.Json, _ = cmd.Flags().GetBool("json")
		params.AutoChapters, _ = cmd.Flags().GetBool("auto_chapters")
		params.AutoHighlights, _ = cmd.Flags().GetBool("auto_highlights")
		params.ContentModeration, _ = cmd.Flags().GetBool("content_moderation")
		params.EntityDetection, _ = cmd.Flags().GetBool("entity_detection")
		params.FormatText, _ = cmd.Flags().GetBool("format_text")
		params.Punctuate, _ = cmd.Flags().GetBool("punctuate")
		params.SentimentAnalysis, _ = cmd.Flags().GetBool("sentiment_analysis")
		params.TopicDetection, _ = cmd.Flags().GetBool("topic_detection")
		params.RedactPii, _ = cmd.Flags().GetBool("redact_pii")
		if params.RedactPii {
			policies, _ := cmd.Flags().GetString("redact_pii_policies")
			policiesArray := strings.Split(policies, ",")

			for _, policy := range policiesArray {
				if _, ok := PIIRedactionPolicyMap[policy]; !ok {
					fmt.Printf("%s is not a valid policy. See https://www.assemblyai.com/docs/audio-intelligence#pii-redaction for the complete list of supported policies.", policy)
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
		speakerLabels, _ := cmd.Flags().GetBool("speaker_labels")
		params.DualChannel, _ = cmd.Flags().GetBool("dual_channel")

		languageCode, _ := cmd.Flags().GetString("language_code")
		if languageCode == "" {
			languageDetection, _ := cmd.Flags().GetBool("language_detection")
			if languageDetection && speakerLabels {
				fmt.Println("Speaker Labels is currently supported in English. Please disable language_detection if you’d like to use it.")
				return
			} else if languageDetection {
				params.LanguageDetection = true
			} else {
				if cmd.Flags().Lookup("speaker_labels").Changed {
					params.SpeakerLabels = speakerLabels
				} else {
					params.SpeakerLabels = true
				}
			}
		} else {
			if speakerLabels {
				fmt.Println("Speaker Labels is currently supported in English. Please disable language_code if you’d like to use it.")
				return
			}
			if _, ok := LanguageMap[languageCode]; !ok {
				fmt.Println("Invalid language code. See https://www.assemblyai.com/docs#supported-languages for supported languages.")
				return
			}
			params.LanguageDetection = false
			params.LanguageCode = &languageCode
		}

		transcribe(params, flags)
	},
}

func init() {
	transcribeCmd.PersistentFlags().StringP("redact_pii_policies", "i", "drug,number_sequence,person_name", "The list of PII policies to redact, comma-separated without space in-between. Required if the redact_pii flag is true.")
	transcribeCmd.PersistentFlags().BoolP("auto_chapters", "s", false, "A \"summary over time\" for the audio file transcribed.")
	transcribeCmd.PersistentFlags().BoolP("auto_highlights", "a", false, "Automatically detect important phrases and words in the text.")
	transcribeCmd.PersistentFlags().BoolP("content_moderation", "c", false, "Detect if sensitive content is spoken in the file.")
	transcribeCmd.PersistentFlags().BoolP("dual_channel", "d", false, "Enable dual channel")
	transcribeCmd.PersistentFlags().BoolP("entity_detection", "e", false, "Identify a wide range of entities that are spoken in the audio file.")
	transcribeCmd.PersistentFlags().BoolP("format_text", "f", true, "Enable text formatting")
	transcribeCmd.PersistentFlags().BoolP("language_detection", "n", false, "Identify the dominant language that’s spoken in an audio file.")
	transcribeCmd.PersistentFlags().StringP("language_code", "g", "", "Specify the language of the speech in your audio file.")
	transcribeCmd.PersistentFlags().BoolP("json", "j", false, "If true, the CLI will output the JSON.")
	transcribeCmd.PersistentFlags().BoolP("poll", "p", true, "The CLI will poll the transcription until it's complete.")
	transcribeCmd.PersistentFlags().BoolP("punctuate", "u", true, "Enable automatic punctuation.")
	transcribeCmd.PersistentFlags().BoolP("redact_pii", "r", false, "Remove personally identifiable information from the transcription.")
	transcribeCmd.PersistentFlags().BoolP("sentiment_analysis", "x", false, "Detect the sentiment of each sentence of speech spoken in the file.")
	transcribeCmd.PersistentFlags().BoolP("speaker_labels", "l", false, "Automatically detect the number of speakers in your audio file, and each word in the transcription text can be associated with its speaker.")
	transcribeCmd.PersistentFlags().BoolP("topic_detection", "t", false, "Label the topics that are spoken in the file.")
	transcribeCmd.PersistentFlags().StringP("webhook_url", "w", "", "Receive a webhook once your transcript is complete.")
	transcribeCmd.PersistentFlags().StringP("webhook_auth_header_name", "b", "", "Containing the header's name which will be inserted into the webhook request")
	transcribeCmd.PersistentFlags().StringP("webhook_auth_header_value", "o", "", "The value of the header that will be inserted into the webhook request.")
	rootCmd.AddCommand(transcribeCmd)
}

func transcribe(params TranscribeParams, flags TranscribeFlags) {
	Token = GetStoredToken()
	if Token == "" {
		fmt.Println("You must login first. Run `assemblyai config <token>`")
		return
	}

	if isUrl(params.AudioURL) {
		if isYoutubeLink(params.AudioURL) {
			if isShortenedYoutubeLink(params.AudioURL) {
				params.AudioURL = strings.Replace(params.AudioURL, "youtu.be/", "www.youtube.com/watch?v=", 1)
			}
			u, err := url.Parse(params.AudioURL)
			if err != nil {
				fmt.Println("Error parsing URL")
				return
			}
			youtubeId := u.Query().Get("v")
			if youtubeId == "" {
				fmt.Println("Could not find YouTube ID in URL")
				return
			}
			youtubeVideoURL := YoutubeDownload(youtubeId)
			if youtubeVideoURL == "" {
				fmt.Println(" Please try again with a different one.")
				return
			}
			params.AudioURL = youtubeVideoURL
		}
		if !checkAAICDN(params.AudioURL) {
			resp, err := http.Get(params.AudioURL)
			if err != nil || resp.StatusCode != 200 {
				fmt.Println("We couldn't transcribe the file in the URL. Please try again with a different one.")
				return
			}
		}
	} else {
		uploadedURL := UploadFile(params.AudioURL)
		if uploadedURL == "" {
			fmt.Println("The file doesn't exist. Please try again with a different one.")
			return
		}
		params.AudioURL = uploadedURL
	}

	paramsJSON, err := json.Marshal(params)
	PrintError(err)

	TelemetryCaptureEvent("CLI transcription created", nil)
	body := bytes.NewReader(paramsJSON)

	response := QueryApi("/transcript", "POST", body)
	var transcriptResponse TranscriptResponse
	if err := json.Unmarshal(response, &transcriptResponse); err != nil {
		PrintError(err)
		return
	}
	if transcriptResponse.Error != nil {
		fmt.Println(*transcriptResponse.Error)
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

	PollTranscription(*id, flags)
}

func isUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func isShortenedYoutubeLink(url string) bool {
	regex := regexp.MustCompile(`^(https?\:\/\/)?(youtu\.?be)\/.+$`)
	return regex.MatchString(url)
}

func isFullLengthYoutubeLink(url string) bool {
	regex := regexp.MustCompile(`^(https?\:\/\/)?(www\.youtube\.com)\/.+$`)
	return regex.MatchString(url)
}

func isYoutubeLink(url string) bool {
	return isFullLengthYoutubeLink(url) || isShortenedYoutubeLink(url)
}

func checkAAICDN(url string) bool {
	return strings.HasPrefix(url, "https://cdn.assemblyai.com/")
}

func UploadFile(path string) string {
	isAbs := filepath.IsAbs(path)
	if !isAbs {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Println("Error getting current directory")
			return ""
		}
		path = filepath.Join(wd, path)
	}

	file, err := os.Open(path)
	if err != nil {
		return ""
	}

	TelemetryCaptureEvent("CLI upload started", nil)

	fileInfo, _ := file.Stat()
	bar := pb.New(int(fileInfo.Size()))
	bar.SetUnits(pb.U_BYTES_DEC)
	bar.Prefix("  Uploading file to our servers: ")
	bar.ShowBar = false
	bar.ShowTimeLeft = false
	bar.Start()

	response := QueryApi("/upload", "POST", bar.NewProxyReader(file))

	bar.Finish()
	var uploadResponse UploadResponse
	if err := json.Unmarshal(response, &uploadResponse); err != nil {
		return ""
	}
	TelemetryCaptureEvent("CLI upload ended", nil)

	return uploadResponse.UploadURL
}

func PollTranscription(id string, flags TranscribeFlags) {
	fmt.Println("  Transcribing file with id " + id)

	s := CallSpinner(" Processing time is usually 20% of the file's duration.")

	for {
		response := QueryApi("/transcript/"+id, "GET", nil)
		if response == nil {
			s.Stop()
			fmt.Println("Something went wrong. Please try again later.")
			return
		}
		var transcript TranscriptResponse
		if err := json.Unmarshal(response, &transcript); err != nil {
			PrintError(err)
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
			var properties *PostHogProperties = new(PostHogProperties)
			properties.Poll = flags.Poll
			properties.Json = flags.Json
			properties.AutoChapters = *transcript.AutoChapters
			properties.AutoHighlights = *transcript.AutoHighlights
			properties.ContentModeration = *transcript.ContentSafety
			properties.DualChannel = transcript.DualChannel
			properties.EntityDetection = *transcript.EntityDetection
			properties.FormatText = *transcript.FormatText
			properties.Punctuate = *transcript.Punctuate
			properties.RedactPii = *transcript.RedactPii
			properties.SentimentAnalysis = *transcript.SentimentAnalysis
			properties.SpeakerLabels = transcript.SpeakerLabels
			properties.TopicDetection = *transcript.IabCategories

			TelemetryCaptureEvent("CLI transcription finished", properties)

			if flags.Json {
				print := BeutifyJSON(response)
				fmt.Println(string(print))
				return
			}
			getFormattedOutput(transcript, flags)
			return
		}
		time.Sleep(3 * time.Second)
	}
}

func getFormattedOutput(transcript TranscriptResponse, flags TranscribeFlags) {
	width, _, err := term.GetSize(0)
	if err != nil {
		width = 512
	}
	fmt.Printf("\033[1m%s\033[0m\n", "Transcript")
	if transcript.SpeakerLabels == true {
		speakerLabelsPrintFormatted(transcript.Utterances, width)
	} else {
		textPrintFormatted(*transcript.Text, width)
	}
	if transcript.DualChannel != nil && *transcript.DualChannel == true {
		fmt.Printf("\033[1m%s\033[0m\n", "\nDual Channel")
		dualChannelPrintFormatted(transcript.Utterances, width)
	}
	if *transcript.AutoHighlights == true {
		fmt.Printf("\033[1m%s\033[0m\n", "Highlights")
		highlightsPrintFormatted(*transcript.AutoHighlightsResult)
	}
	if *transcript.ContentSafety == true {
		fmt.Printf("\033[1m%s\033[0m\n", "Content Moderation")
		contentSafetyPrintFormatted(*transcript.ContentSafetyLabels, width)
	}
	if *transcript.IabCategories == true {
		fmt.Printf("\033[1m%s\033[0m\n", "Topic Detection")
		topicDetectionPrintFormatted(*transcript.IabCategoriesResult, width)
	}
	if *transcript.SentimentAnalysis == true {
		fmt.Printf("\033[1m%s\033[0m\n", "Sentiment Analysis")
		sentimentAnalysisPrintFormatted(transcript.SentimentAnalysisResults, width)
	}
	if *transcript.AutoChapters == true {
		fmt.Printf("\033[1m%s\033[0m\n", "Chapters")
		chaptersPrintFormatted(transcript.Chapters, width)
	}
	if *transcript.EntityDetection == true {
		fmt.Printf("\033[1m%s\033[0m\n", "Entity Detection")
		entityDetectionPrintFormatted(transcript.Entities, width)
	}
}

func textPrintFormatted(text string, width int) {
	words := strings.Split(text, " ")
	var line string
	for _, word := range words {
		if len(line)+len(word) > width-1 {
			fmt.Println(line)
			line = ""
		}
		line += word + " "
	}
	fmt.Println(line)
	fmt.Println()
}

func dualChannelPrintFormatted(utterances []SentimentAnalysisResult, width int) {
	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 21)
	for _, utterance := range utterances {
		duration := time.Duration(*utterance.Start) * time.Millisecond
		start := fmt.Sprintf("%02d:%02d", int(duration.Minutes()), int(duration.Seconds())%60)
		speaker := fmt.Sprintf("(Channel %s)", utterance.Channel)

		table.AddRow(start, speaker, utterance.Text)
	}
	fmt.Println(table)
	fmt.Println()
}

func speakerLabelsPrintFormatted(utterances []SentimentAnalysisResult, width int) {
	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 25)

	//  get utterances length
	singleSpeaker := len(utterances) == 1

	for _, utterance := range utterances {
		duration := time.Duration(*utterance.Start) * time.Millisecond
		start := fmt.Sprintf("%02d:%02d", int(duration.Minutes()), int(duration.Seconds())%60)
		speaker := fmt.Sprintf("(Speaker %s)", utterance.Speaker)
		if singleSpeaker {
			words := strings.Split(utterance.Text, ".")
			text := ""
			for i, word := range words {
				if i%3 == 0 {
					table.AddRow(start, speaker, text)
					start = ""
					speaker = ""
					text = ""
				} else {
					if strings.HasPrefix(word, " ") && len(text) == 0 {
						word = word[1:]
					}
					text = text + word + "."
				}
			}
		} else {
			table.AddRow(start, speaker, utterance.Text)
		}
	}
	fmt.Println(table)
	fmt.Println()
}

func highlightsPrintFormatted(highlights AutoHighlightsResult) {
	if *highlights.Status != "success" {
		fmt.Println("Could not retrieve highlights")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.Separator = " |\t"
	table.AddRow("| count", "text")
	sort.SliceStable(highlights.Results, func(i, j int) bool {
		return int(*highlights.Results[i].Count) > int(*highlights.Results[j].Count)
	})
	for _, highlight := range highlights.Results {
		table.AddRow("| "+strconv.FormatInt(*highlight.Count, 10), highlight.Text)
	}
	fmt.Println(table)
	fmt.Println()
}

func contentSafetyPrintFormatted(labels ContentSafetyLabels, width int) {
	if *labels.Status != "success" {
		fmt.Println("Could not retrieve content safety labels")
		return
	}
	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 24)
	table.Separator = " |\t"
	table.AddRow("| label", "text")
	for _, label := range labels.Results {
		var labelString string
		for _, innerLabel := range label.Labels {
			labelString = innerLabel.Label + " " + labelString
		}
		table.AddRow("| "+labelString, label.Text)
	}
	fmt.Println(table)
	fmt.Println()
}

func topicDetectionPrintFormatted(categories IabCategoriesResult, width int) {
	if *categories.Status != "success" {
		fmt.Println("Could not retrieve topic detection")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint((width / 2) - 5)
	table.Separator = " |\t"
	table.AddRow("| rank", "topic")
	var ArrayCategoriesSorted []ArrayCategories
	for category, i := range categories.Summary {
		add := ArrayCategories{
			Category: category,
			Score:    i,
		}
		ArrayCategoriesSorted = append(ArrayCategoriesSorted, add)
	}
	sort.SliceStable(ArrayCategoriesSorted, func(i, j int) bool {
		return ArrayCategoriesSorted[i].Score > ArrayCategoriesSorted[j].Score
	})

	for i, category := range ArrayCategoriesSorted {
		table.AddRow(fmt.Sprintf("| %o", i+1), category.Category)
	}
	fmt.Println(table)
	fmt.Println()
}

func sentimentAnalysisPrintFormatted(sentiments []SentimentAnalysisResult, width int) {
	if len(sentiments) == 0 {
		fmt.Println("Could not retrieve sentiment analysis")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 20)
	table.Separator = " |\t"
	table.AddRow("| sentiment", "text")
	for _, sentiment := range sentiments {
		sentimentStatus := sentiment.Sentiment
		table.AddRow("| "+sentimentStatus, sentiment.Text)
	}
	fmt.Println(table)
	fmt.Println()
}

func chaptersPrintFormatted(chapters []Chapter, width int) {
	if len(chapters) == 0 {
		fmt.Println("Could not retrieve chapters")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 17)
	table.Separator = " |\t"
	for _, chapter := range chapters {
		start := time.Duration(*chapter.Start) * time.Millisecond
		end := time.Duration(*chapter.End) * time.Millisecond
		table.AddRow("| timestamp", fmt.Sprintf("%02d:%02d-%02d:%02d", int(start.Minutes()), int(start.Seconds())%60, int(end.Minutes()), int(end.Seconds())%60))
		table.AddRow("| Gist", chapter.Gist)
		table.AddRow("| Headline", chapter.Headline)
		table.AddRow("| Summary", chapter.Summary)
		table.AddRow("", "")
	}
	fmt.Println(table)
	fmt.Println()
}

func entityDetectionPrintFormatted(entities []Entity, width int) {
	if len(entities) == 0 {
		fmt.Println("Could not retrieve entity detection")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 20)
	table.Separator = " |\t"
	table.AddRow("| type", "text")
	entityMap := make(map[string][]string)
	for _, entity := range entities {
		isAlreadyInMap := false
		for _, text := range entityMap[entity.EntityType] {
			if text == entity.Text {
				isAlreadyInMap = true
				break
			}
		}
		if !isAlreadyInMap {
			entityMap[entity.EntityType] = append(entityMap[entity.EntityType], entity.Text)
		}
	}
	for entityType, entityTexts := range entityMap {
		table.AddRow("| "+entityType, strings.Join(entityTexts, ", "))
	}
	fmt.Println(table)
	fmt.Println()
}

type ArrayCategories struct {
	Score    float64 `json:"score"`
	Category string  `json:"category"`
}

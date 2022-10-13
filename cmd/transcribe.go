/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
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
		params.DualChannel, _ = cmd.Flags().GetBool("dual_channel")
		params.EntityDetection, _ = cmd.Flags().GetBool("entity_detection")
		params.FormatText, _ = cmd.Flags().GetBool("format_text")
		params.Punctuate, _ = cmd.Flags().GetBool("punctuate")
		params.SentimentAnalysis, _ = cmd.Flags().GetBool("sentiment_analysis")
		params.SpeakerLabels, _ = cmd.Flags().GetBool("speaker_labels")
		params.TopicDetection, _ = cmd.Flags().GetBool("topic_detection")
		params.RedactPii, _ = cmd.Flags().GetBool("redact_pii")
		params.Summarization, _ = cmd.Flags().GetBool("summarization")
		if params.RedactPii {
			policies, _ := cmd.Flags().GetString("redact_pii_policies")
			params.RedactPiiPolicies = strings.Split(policies, ",")
		}
		if params.Summarization {
			params.SummaryType, _ = cmd.Flags().GetString("summary_type")
		}

		transcribe(params, flags)
	},
}

func init() {
	transcribeCmd.PersistentFlags().StringP("redact_pii_policies", "i", "drug,number_sequence,person_name", "The list of PII policies to redact, comma-separated without space in-between.")
	transcribeCmd.PersistentFlags().BoolP("auto_chapters", "s", false, "A \"summary over time\" for the audio file transcribed.")
	transcribeCmd.PersistentFlags().BoolP("auto_highlights", "a", false, "Automatically detect important phrases and words in the text.")
	transcribeCmd.PersistentFlags().BoolP("content_moderation", "c", false, "Detect if sensitive content is spoken in the file.")
	transcribeCmd.PersistentFlags().BoolP("dual_channel", "d", false, "Enable dual channel")
	transcribeCmd.PersistentFlags().BoolP("entity_detection", "e", false, "Identify a wide range of entities that are spoken in the audio file.")
	transcribeCmd.PersistentFlags().BoolP("format_text", "f", true, "Enable text formatting")
	transcribeCmd.PersistentFlags().BoolP("json", "j", false, "If true, the CLI will output the JSON.")
	transcribeCmd.PersistentFlags().BoolP("poll", "p", true, "The CLI will poll the transcription until it's complete.")
	transcribeCmd.PersistentFlags().BoolP("punctuate", "u", true, "Enable automatic punctuation.")
	transcribeCmd.PersistentFlags().BoolP("redact_pii", "r", false, "Remove personally identifiable information from the transcription.")
	transcribeCmd.PersistentFlags().BoolP("sentiment_analysis", "x", false, "Detect the sentiment of each sentence of speech spoken in the file.")
	transcribeCmd.PersistentFlags().BoolP("speaker_labels", "l", true, "Automatically detect the number of speakers in your audio file, and each word in the transcription text can be associated with its speaker.")
	transcribeCmd.PersistentFlags().BoolP("topic_detection", "t", false, "Label the topics that are spoken in the file.")
	transcribeCmd.PersistentFlags().BoolP("summarization", "z", false, "Automatically summarize audio and video files at scale.")
	transcribeCmd.PersistentFlags().StringP("summary_type", "y", "bullets", "Presentation way of summarization. Available: bullets, paragraph, headline or gist")
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
		fmt.Println("Can not unmarshal JSON")
		return
	}
	if transcriptResponse.Error != nil || *transcriptResponse.Error != "" {
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
	bar.Prefix(" Uploading file to our servers: ")
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
	fmt.Println(" Transcribing file with id " + id)
	showProgressBar := TranscriptionLength != 0
	timePercentage := (TranscriptionLength * 30) / 100

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	var s *spinner.Spinner
	var bar *pb.ProgressBar

	if showProgressBar {
		fmt.Println(" Processing time is usually 20% of the file's duration.")
		bar = pb.StartNew(timePercentage)
		go showProgress(timePercentage, ctx, bar)
	} else {
		s = CallSpinner(" Processing time is usually 20% of the file's duration.")
	}

	for {
		response := QueryApi("/transcript/"+id, "GET", nil)
		if response == nil {
			if showProgressBar {
				cancelCtx()
				bar.Set(timePercentage)
				bar.Finish()
			} else {
				s.Stop()
			}
			fmt.Println("Something went wrong. Please try again later.")
			return
		}
		var transcript TranscriptResponse
		if err := json.Unmarshal(response, &transcript); err != nil {
			fmt.Println(err)
			if showProgressBar {
				cancelCtx()
				bar.Set(timePercentage)
				bar.Finish()
			} else {
				s.Stop()
			}
			return
		}
		if transcript.Error != nil {
			if showProgressBar {
				cancelCtx()
				bar.Set(timePercentage)
				bar.Finish()
			} else {
				s.Stop()
			}
			fmt.Println(*transcript.Error)
			return
		}
		if *transcript.Status == "completed" {
			if showProgressBar {
				cancelCtx()
				bar.Set(timePercentage)
				bar.Finish()
			} else {
				s.Stop()
			}
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
	fmt.Print("\033[H\033[2J")
	fmt.Println("Transcript")
	if transcript.SpeakerLabels == true {
		speakerLabelsPrintFormatted(transcript.Utterances, width)
	} else {
		textPrintFormatted(*transcript.Text, width)
	}
	if transcript.DualChannel != nil && *transcript.DualChannel == true {
		fmt.Println("\nDual Channel")
		dualChannelPrintFormatted(transcript.Utterances, width)
	}
	if *transcript.AutoHighlights == true {
		fmt.Println("Highlights")
		highlightsPrintFormatted(*transcript.AutoHighlightsResult)
	}
	if *transcript.ContentSafety == true {
		fmt.Println("Content Moderation")
		contentSafetyPrintFormatted(*transcript.ContentSafetyLabels, width)
	}
	if *transcript.IabCategories == true {
		fmt.Println("Topic Detection")
		topicDetectionPrintFormatted(*transcript.IabCategoriesResult, width)
	}
	if *transcript.SentimentAnalysis == true {
		fmt.Println("Sentiment Analysis")
		sentimentAnalysisPrintFormatted(transcript.SentimentAnalysisResults, width)
	}
	if *transcript.AutoChapters == true {
		fmt.Println("Chapters")
		chaptersPrintFormatted(transcript.Chapters, width)
	}
	if *transcript.EntityDetection == true {
		fmt.Println("Entity Detection")
		entityDetectionPrintFormatted(transcript.Entities, width)
	}
	if *transcript.Summary != "" {
		fmt.Println("Entity Detection")
		summaryPrintFormatted(*transcript.Summary, width)
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
	table.AddRow("", "", "")
	fmt.Println(table)
}

func speakerLabelsPrintFormatted(utterances []SentimentAnalysisResult, width int) {
	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 25)
	for _, utterance := range utterances {
		duration := time.Duration(*utterance.Start) * time.Millisecond
		start := fmt.Sprintf("%02d:%02d", int(duration.Minutes()), int(duration.Seconds())%60)
		speaker := fmt.Sprintf("(Speaker %s)", utterance.Speaker)
		table.AddRow(start, speaker, utterance.Text)
	}
	table.AddRow("", "", "")
	fmt.Println(table)
}

func highlightsPrintFormatted(highlights AutoHighlightsResult) {
	if *highlights.Status != "success" {
		fmt.Println("Could not retrieve highlights")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.Separator = "|"
	table.AddRow("COUNT", "TEXT")
	for _, highlight := range highlights.Results {
		table.AddRow(strconv.FormatInt(*highlight.Count, 10), highlight.Text)
	}
	table.AddRow("", "")
	fmt.Println(table)
}

func contentSafetyPrintFormatted(labels ContentSafetyLabels, width int) {
	if *labels.Status != "success" {
		fmt.Println("Could not retrieve content safety labels")
		return
	}
	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 20)
	table.Separator = "|"
	table.AddRow("LABEL", "TEXT")
	for _, label := range labels.Results {
		var labelString string
		for _, innerLabel := range label.Labels {
			labelString = innerLabel.Label + " " + labelString
		}
		table.AddRow(labelString, label.Text)
	}
	table.AddRow("", "")
	fmt.Println(table)
}

func topicDetectionPrintFormatted(categories IabCategoriesResult, width int) {
	if *categories.Status != "success" {
		fmt.Println("Could not retrieve topic detection")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint((width / 2) - 5)
	table.Separator = "|"
	table.AddRow("TOPIC", "TEXT")
	for _, category := range categories.Results {
		categories := ""
		for i, innerCategory := range category.Labels {
			categories += innerCategory.Label + " "
			if i == 2 {
				break
			}
		}
		table.AddRow(categories, category.Text)
		table.AddRow("", "")
	}
	fmt.Println(table)
}

func sentimentAnalysisPrintFormatted(sentiments []SentimentAnalysisResult, width int) {
	if len(sentiments) == 0 {
		fmt.Println("Could not retrieve sentiment analysis")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 20)
	table.Separator = "|"
	table.AddRow("SENTIMENT", "TEXT")
	for _, sentiment := range sentiments {
		sentimentStatus := sentiment.Sentiment
		table.AddRow(sentimentStatus, sentiment.Text)
	}
	table.AddRow("", "")
	fmt.Println(table)
}

func chaptersPrintFormatted(chapters []Chapter, width int) {
	if len(chapters) == 0 {
		fmt.Println("Could not retrieve chapters")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 15)
	table.Separator = "|"
	for _, chapter := range chapters {
		// Gist
		table.AddRow("Gist", chapter.Gist)
		table.AddRow("", "")

		// Headline
		table.AddRow("Headline", chapter.Headline)

		table.AddRow("", "")
		table.AddRow("Summary", chapter.Summary)
	}
	table.AddRow("", "")
	fmt.Println(table)
}

func entityDetectionPrintFormatted(entities []Entity, width int) {
	if len(entities) == 0 {
		fmt.Println("Could not retrieve entity detection")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 20)
	table.Separator = "|"
	table.AddRow("TYPE", "TEXT")
	for _, entity := range entities {
		table.AddRow(entity.EntityType, entity.Text)
	}
	table.AddRow("", "")
	fmt.Println(table)
}

func summaryPrintFormatted(summary string, width int) {
	if summary == "" {
		fmt.Println("Could not retrieve summary")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 5)
	table.Separator = "|"
	table.AddRow("Summary")
	table.AddRow(summary)
	table.AddRow("")
	fmt.Println(table)
}

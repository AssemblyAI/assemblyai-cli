package utils

import (
	"bytes"
	"encoding/json"
	"errors"
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

	S "github.com/AssemblyAI/assemblyai-cli/schemas"
	"github.com/gosuri/uitable"
	"golang.org/x/term"
	"gopkg.in/cheggaaa/pb.v1"
)

var width int

func Transcribe(params S.TranscribeParams, flags S.TranscribeFlags) {
	Token = GetStoredToken()
	if Token == "" {
		printErrorProps := S.PrintErrorProps{
			Error:   errors.New("No token found"),
			Message: "Please start by running \033[1m\033[34massemblyai config [token]\033[0m",
		}
		PrintError(printErrorProps)
		return
	}

	if isUrl(params.AudioURL) {
		if isYoutubeLink(params.AudioURL) {
			if isYoutubeShortLink(params.AudioURL) {
				printErrorProps := S.PrintErrorProps{
					Error:   errors.New("short links are not supported"),
					Message: "The AssemblyAI CLI doesn’t support YouTube Shorts yet.",
				}
				PrintError(printErrorProps)
				return
			}
			if isShortenedYoutubeLink(params.AudioURL) {
				params.AudioURL = strings.Replace(params.AudioURL, "youtu.be/", "www.youtube.com/watch?v=", 1)
			}
			u, err := url.Parse(params.AudioURL)
			if err != nil {
				printErrorProps := S.PrintErrorProps{
					Error:   errors.New("invalid url"),
					Message: "Error parsing URL",
				}
				PrintError(printErrorProps)
				return
			}
			youtubeId := u.Query().Get("v")
			if youtubeId == "" {
				printErrorProps := S.PrintErrorProps{
					Error:   errors.New("invalid youtube url"),
					Message: "Could not find YouTube ID in URL",
				}
				PrintError(printErrorProps)
				return
			}
			youtubeVideoURL := YoutubeDownload(youtubeId)
			if youtubeVideoURL == "" {
				printErrorProps := S.PrintErrorProps{
					Error:   errors.New("invalid youtube url"),
					Message: "Please try again with a different one.",
				}
				PrintError(printErrorProps)
				return
			}
			params.AudioURL = youtubeVideoURL
		}
		if !checkAAICDN(params.AudioURL) {
			resp, err := http.Get(params.AudioURL)
			if err != nil || resp.StatusCode != 200 {
				printErrorProps := S.PrintErrorProps{
					Error:   errors.New("unreachable url"),
					Message: "We couldn't transcribe the file in the URL. Please try again with a different one.",
				}
				PrintError(printErrorProps)
				return
			}
		}
	} else {
		uploadedURL := UploadFile(params.AudioURL)
		if uploadedURL == "" {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("invalid file"),
				Message: "The file doesn't exist. Please try again with a different one.",
			}
			PrintError(printErrorProps)
			return
		}
		params.AudioURL = uploadedURL
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
	}

	TelemetryCaptureEvent("CLI transcription created", nil)
	body := bytes.NewReader(paramsJSON)

	response := QueryApi("/transcript", "POST", body, nil)
	var transcriptResponse S.TranscriptResponse
	if err := json.Unmarshal(response, &transcriptResponse); err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
		return
	}
	if transcriptResponse.Error != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   errors.New(*transcriptResponse.Error),
			Message: *transcriptResponse.Error,
		}
		PrintError(printErrorProps)
		return
	}
	id := transcriptResponse.ID
	if !flags.Poll {
		if flags.Json {
			print := BeutifyJSON(response)
			fmt.Println(string(print))
			return
		}
		fmt.Fprintf(os.Stdin, "Your transcription was created (id %s)\n", *id)
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

func isYoutubeShortLink(url string) bool {
	regex := regexp.MustCompile(`^(https?\:\/\/)?(www\.youtube\.com)\/shorts\/.+$`)
	regexShare := regexp.MustCompile(`^(https?\:\/\/)?(youtube\.com)\/shorts\/.+$`)

	return regex.MatchString(url) || regexShare.MatchString(url)
}

func isYoutubeLink(url string) bool {
	return isFullLengthYoutubeLink(url) || isShortenedYoutubeLink(url) || isYoutubeShortLink(url)
}

func checkAAICDN(url string) bool {
	return strings.HasPrefix(url, "https://cdn.assemblyai.com/")
}

func UploadFile(path string) string {
	isAbs := filepath.IsAbs(path)
	if !isAbs {
		wd, err := os.Getwd()
		if err != nil {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("Error getting current directory"),
				Message: "Error getting current directory",
			}
			PrintError(printErrorProps)
			return ""
		}
		path = filepath.Join(wd, path)
	}

	file, err := os.Open(path)
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   errors.New("Error opening file"),
			Message: "Error opening file",
		}
		PrintError(printErrorProps)
		return ""
	}

	TelemetryCaptureEvent("CLI upload started", nil)

	fileInfo, _ := file.Stat()
	bar := pb.New(int(fileInfo.Size()))
	bar.Output = os.Stdin
	bar.SetUnits(pb.U_BYTES_DEC)
	bar.Prefix("Uploading file to our servers: ")
	bar.ShowBar = false
	bar.ShowTimeLeft = false
	bar.Start()

	response := QueryApi("/upload", "POST", bar.NewProxyReader(file), nil)

	bar.Finish()
	var uploadResponse S.UploadResponse
	if err := json.Unmarshal(response, &uploadResponse); err != nil {
		return ""
	}
	TelemetryCaptureEvent("CLI upload ended", nil)

	return uploadResponse.UploadURL
}

func PollTranscription(id string, flags S.TranscribeFlags) {
	fmt.Fprintln(os.Stdin, "Transcribing file with id "+id)

	s := CallSpinner(" Processing time is usually 20% of the file's duration.")

	for {
		response := QueryApi("/transcript/"+id, "GET", nil, s)
		if response == nil {
			s.Stop()
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("Error getting transcription"),
				Message: "Something went wrong. Please try again later.",
			}
			PrintError(printErrorProps)
			return
		}
		var transcript S.TranscriptResponse
		if err := json.Unmarshal(response, &transcript); err != nil {
			printErrorProps := S.PrintErrorProps{
				Error:   err,
				Message: "Something went wrong. Please try again.",
			}
			PrintError(printErrorProps)
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
			var properties *S.PostHogProperties = new(S.PostHogProperties)
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

func getFormattedOutput(transcript S.TranscriptResponse, flags S.TranscribeFlags) {
	getWidth, _, err := term.GetSize(0)
	if err != nil {
		width = 512
	} else {
		width = getWidth
	}
	fmt.Fprintf(os.Stdin, "\033[1m%s\033[0m\n", "Transcript")
	if transcript.SpeakerLabels == true {
		speakerLabelsPrintFormatted(transcript.Utterances)
	} else {
		textPrintFormatted(*transcript.Text, transcript.Words)
	}
	if transcript.DualChannel != nil && *transcript.DualChannel == true {
		fmt.Fprintf(os.Stdin, "\033[1m%s\033[0m\n", "\nDual Channel")
		dualChannelPrintFormatted(transcript.Utterances)
	}
	if *transcript.AutoHighlights == true {
		fmt.Fprintf(os.Stdin, "\033[1m%s\033[0m\n", "Highlights")
		highlightsPrintFormatted(*transcript.AutoHighlightsResult)
	}
	if *transcript.ContentSafety == true {
		fmt.Fprintf(os.Stdin, "\033[1m%s\033[0m\n", "Content Moderation")
		contentSafetyPrintFormatted(*transcript.ContentSafetyLabels)
	}
	if *transcript.IabCategories == true {
		fmt.Fprintf(os.Stdin, "\033[1m%s\033[0m\n", "Topic Detection")
		topicDetectionPrintFormatted(*transcript.IabCategoriesResult)
	}
	if *transcript.SentimentAnalysis == true {
		fmt.Fprintf(os.Stdin, "\033[1m%s\033[0m\n", "Sentiment Analysis")
		sentimentAnalysisPrintFormatted(transcript.SentimentAnalysisResults)
	}
	if *transcript.AutoChapters == true {
		fmt.Fprintf(os.Stdin, "\033[1m%s\033[0m\n", "Chapters")
		chaptersPrintFormatted(transcript.Chapters)
	}
	if *transcript.EntityDetection == true {
		fmt.Fprintf(os.Stdin, "\033[1m%s\033[0m\n", "Entity Detection")
		entityDetectionPrintFormatted(transcript.Entities)
	}
	if transcript.Summarization != nil && *transcript.Summarization == true {
		fmt.Fprintf(os.Stdin, "\033[1m%s\033[0m\n", "Summary")
		summaryPrintFormatted(transcript.Summary)
	}
}

func textPrintFormatted(text string, words []S.SentimentAnalysisResult) {
	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 10)
	sentences := SplitSentences(text, true)
	timestamps := GetSentenceTimestamps(sentences, words)
	for index, sentence := range sentences {
		if sentence != "" {
			stamp := ""
			if len(timestamps) > index {
				stamp = timestamps[index]
			}
			table.AddRow(stamp, sentence)
		}
	}
	fmt.Println(table)
	fmt.Println()
}

func dualChannelPrintFormatted(utterances []S.SentimentAnalysisResult) {
	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 21)
	for _, utterance := range utterances {
		start := TransformMsToTimestamp(*utterance.Start)
		speaker := fmt.Sprintf("(Channel %s)", utterance.Channel)

		sentences := SplitSentences(utterance.Text, false)
		for _, sentence := range sentences {
			table.AddRow(start, speaker, sentence)
			start = ""
			speaker = ""
		}
	}
	fmt.Println(table)
	fmt.Println()
}

func speakerLabelsPrintFormatted(utterances []S.SentimentAnalysisResult) {
	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 27)

	for _, utterance := range utterances {
		sentences := SplitSentences(utterance.Text, false)
		timestamps := GetSentenceTimestampsAndSpeaker(sentences, utterance.Words)
		for index, sentence := range sentences {
			if sentence != "" {
				info := []string{"", ""}
				if len(timestamps) > index {
					info = timestamps[index]
				}
				table.AddRow(info[0], info[1], sentence)
			}
		}
	}
	fmt.Println(table)
	fmt.Println()
}

func highlightsPrintFormatted(highlights S.AutoHighlightsResult) {
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

func contentSafetyPrintFormatted(labels S.ContentSafetyLabels) {
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

func topicDetectionPrintFormatted(categories S.IabCategoriesResult) {
	if *categories.Status != "success" {
		fmt.Println("Could not retrieve topic detection")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 20)
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

func sentimentAnalysisPrintFormatted(sentiments []S.SentimentAnalysisResult) {
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

func chaptersPrintFormatted(chapters []S.Chapter) {
	if len(chapters) == 0 {
		fmt.Println("Could not retrieve chapters")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 19)
	table.Separator = " |\t"
	for _, chapter := range chapters {
		start := TransformMsToTimestamp(*chapter.Start)
		end := TransformMsToTimestamp(*chapter.End)
		table.AddRow("| timestamp", fmt.Sprintf("%s-%s", start, end))
		table.AddRow("| Gist", chapter.Gist)
		table.AddRow("| Headline", chapter.Headline)
		table.AddRow("| Summary", chapter.Summary)
		table.AddRow("", "")
	}
	fmt.Println(table)
	fmt.Println()
}

func entityDetectionPrintFormatted(entities []S.Entity) {
	if len(entities) == 0 {
		fmt.Println("Could not retrieve entity detection")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 25)
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

func summaryPrintFormatted(summary *string) {
	if summary == nil {
		fmt.Println("Could not retrieve summary")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 20)
	table.Separator = " |\t"

	table.AddRow(*summary)

	fmt.Println(table)
	fmt.Println()
}

type ArrayCategories struct {
	Score    float64 `json:"score"`
	Category string  `json:"category"`
}

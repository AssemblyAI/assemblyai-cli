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
	"strings"
	"time"

	S "github.com/AssemblyAI/assemblyai-cli/schemas"
	"golang.org/x/term"
	"gopkg.in/cheggaaa/pb.v1"
)

var width int
var Flags S.TranscribeFlags

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
	Flags = flags
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
			if flags.Csv != "" {
				if filepath.Ext(flags.Csv) == "" {
					flags.Csv = flags.Csv + ".csv"
				}

				row := [][]string{}
				row = append(row, []string{"\"" + *transcript.Text + "\""})
				GenerateCsv(flags.Csv, []string{"text"}, row)
			}

			getFormattedOutput(transcript)
			return
		}
		time.Sleep(3 * time.Second)
	}
}

func getFormattedOutput(transcript S.TranscriptResponse) {
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

type ArrayCategories struct {
	Score    float64 `json:"score"`
	Category string  `json:"category"`
}

func ValidateCustomSpelling(customSpelling []S.CustomSpelling) error {
	for _, spelling := range customSpelling {
		if len(spelling.From) == 0 {
			return fmt.Errorf("from cannot be empty")
		}
		if spelling.To == "" {
			return fmt.Errorf("to cannot be empty")
		}
	}
	return nil
}

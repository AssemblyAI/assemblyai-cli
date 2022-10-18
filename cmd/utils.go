/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/posthog/posthog-go"
	"golang.org/x/term"
	pb "gopkg.in/cheggaaa/pb.v1"
)

var AAIURL = "https://api.assemblyai.com/v2"
var PH_TOKEN string

func TelemetryCaptureEvent(event string, properties *PostHogProperties) {
	isTelemetryEnabled := getConfigFileValue("features.telemetry")
	if isTelemetryEnabled == "true" {

		if PH_TOKEN == "" {
			godotenv.Load()
			PH_TOKEN = os.Getenv("POSTHOG_API_TOKEN")
		}

		client := posthog.New(PH_TOKEN)
		defer client.Close()

		distinctId := getConfigFileValue("config.distinct_id")

		if distinctId == "" {
			distinctId = uuid.New().String()
			setConfigFileValue("config.distinct_id", distinctId)
		}
		if properties != nil {
			var PhProperties posthog.Properties
			if properties.I == true {
				PhProperties = posthog.NewProperties().
					Set("OS", properties.OS).
					Set("Arch", properties.Arch).
					Set("Version", properties.Version).
					Set("Method", properties.Method)
			} else {
				PhProperties = posthog.NewProperties().
					Set("poll", properties.Poll).
					Set("json", properties.Json).
					Set("speaker_labels", properties.SpeakerLabels).
					Set("punctuate", properties.Punctuate).
					Set("format_text", properties.FormatText).
					Set("dual_channel", properties.DualChannel).
					Set("redact_pii", properties.RedactPii).
					Set("auto_highlights", properties.AutoHighlights).
					Set("content_moderation", properties.ContentModeration).
					Set("topic_detection", properties.TopicDetection).
					Set("sentiment_analysis", properties.SentimentAnalysis).
					Set("auto_chapters", properties.AutoChapters).
					Set("entity_detection", properties.EntityDetection)
			}
			client.Enqueue(posthog.Capture{
				DistinctId: distinctId,
				Event:      event,
				Properties: PhProperties,
			})
			return
		}

		client.Enqueue(posthog.Capture{
			DistinctId: distinctId,
			Event:      event,
		})
	}
}

func spinnerMessage(message string) string {
	width, _, err := term.GetSize(0)
	if err != nil {
		width = 512
	}
	words := strings.Split(message, "")
	if len(words) > 0 {
		message = ""
		for _, word := range words {
			if len(message)+len(word) > width-4 {
				message += "..."
				return message
			}
			message += word + ""
		}
	}
	return message
}

func CallSpinner(message string) *spinner.Spinner {
	newMessage := spinnerMessage(message)
	s := spinner.New(spinner.CharSets[7], 100*time.Millisecond, spinner.WithSuffix(newMessage))
	s.Start()
	return s
}

func PrintError(err error) {
	if err != nil {
		// fmt.Println(err)
		return
	}
}

func QueryApi(path string, method string, body io.Reader) []byte {
	resp, err := http.NewRequest(method, AAIURL+path, body)
	PrintError(err)

	resp.Header.Add("Accept", "application/json")
	resp.Header.Add("Authorization", Token)
	resp.Header.Add("Transfer-Encoding", "chunked")

	response, err := http.DefaultClient.Do(resp)
	PrintError(err)
	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	PrintError(err)
	return responseData
}

func BeutifyJSON(data []byte) []byte {
	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, data, "", "\t")
	if error != nil {
		return data
	}
	return prettyJSON.Bytes()
}

func showProgress(total int, ctx context.Context, bar *pb.ProgressBar) {
	for {
		bar.Prefix(" Transcribing file: ")
		bar.ShowBar = false
		bar.ShowTimeLeft = false
		bar.ShowCounters = false
		bar.ShowFinalTime = true
		for i := 0; i < total-1; i++ {
			if i*total/300 < total {
				bar.Set(i * total / 300)
			}
			time.Sleep(100 * time.Millisecond)
		}
		bar.Finish()
	}
}

func SplitSentences(wholeText string) []string {
	words := strings.Split(wholeText, ".")
	sentences := []string{}
	text := ""
	extra := "."
	if len(words) > 9 {
		extra = ".\n"
	}
	for i, word := range words {
		if i == len(words)-1 {
			text += word
			sentences = append(sentences, text)
			text = ""
		} else {
			if word != "" {
				if i%3 == 0 && i != 0 || i == len(words)-1 {
					text += word + extra
					sentences = append(sentences, text)
					text = ""
				} else {
					if strings.HasPrefix(word, " ") && len(text) == 0 {
						word = word[1:]
					}
					text += word + "."
				}
			}
		}
	}
	return sentences
}

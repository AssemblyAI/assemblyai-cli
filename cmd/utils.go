/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
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
		fmt.Println("Something just went wrong. Please try again.")
		os.Exit(1)
	}
}

func QueryApi(path string, method string, body io.Reader, s *spinner.Spinner) []byte {
	resp, err := http.NewRequest(method, AAIURL+path, body)
	PrintError(err)

	resp.Header.Add("Accept", "application/json")
	resp.Header.Add("Authorization", Token)
	resp.Header.Add("Transfer-Encoding", "chunked")

	response, err := http.DefaultClient.Do(resp)
	PrintError(err)
	defer response.Body.Close()
	if response.StatusCode != 200 {
		if s != nil {
			s.Stop()
		}
		fmt.Println("Something just went wrong. Please try again.")
		os.Exit(1)
	}

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

func SplitSentences(wholeText string, isLineBreeakEnabled bool) []string {
	for _, splitException := range splitExceptions {
		wholeText = strings.ReplaceAll(wholeText, splitException[0], splitException[1])
	}
	// reg := regexp.MustCompile(`\d\.\d`)

	words := strings.Split(wholeText, ".")
	sentences := []string{}
	text := ""
	extra := "."
	if isLineBreeakEnabled {
		extra = ".\n"
	}
	for i, word := range words {
		if i == len(words)-1 {
			text += word
			text = strings.ReplaceAll(text, "{{}}", ".")
			sentences = append(sentences, text)
			text = ""
		} else {
			if word != "" {
				if i%3 == 0 && i != 0 || i == len(words)-1 {
					text += word + extra
					text = strings.ReplaceAll(text, "{{}}", ".")
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
	for i, sentence := range sentences {
		if sentence == "" {
			sentences = append(sentences[:i], sentences[i+1:]...)
		}
	}
	return sentences
}

var splitExceptions = [][]string{
	{"Mr.", "Mr{{}}"},
	{"Mrs.", "Mrs{{}}"},
	{"Ms.", "Ms{{}}"},
	{"Dr.", "Dr{{}}"},
	{"Prof.", "Prof{{}}"},
	{"St.", "St{{}}"},
	{"Mt.", "Mt{{}}"},
	{"Gen.", "Gen{{}}"},
	{"Sen.", "Sen{{}}"},
	{"Rep.", "Rep{{}}"},
	{"Gov.", "Gov{{}}"},
	{"Pres.", "Pres{{}}"},
	{"Rev.", "Rev{{}}"},
	{"Fr.", "Fr{{}}"},
	{"Br.", "Br{{}}"},
	{"Jr.", "Jr{{}}"},
	{"Sr.", "Sr{{}}"},
	{"Ud.", "Ud{{}}"},
	{"Uds.", "Uds{{}}"},
	{" A.", " A{{}}"},
	{" I.", " I{{}}"},
	{" C.", " C{{}}"},
	{" R.", " R{{}}"},
	{" P.", " P{{}}"},
	{".com", "{{}}com"},
}

func TransformMsToTimestamp(ms int64) string {
	duration := time.Duration(ms) * time.Millisecond
	return fmt.Sprintf("%02d:%02d", int(duration.Minutes()), int(duration.Seconds())%60)
}

func GetSentenceTimestamps(sentences []string, words []SentimentAnalysisResult) []string {
	var lastIndex int
	timestamps := []string{}
	for index, sentence := range sentences {
		if index == 0 {
			timestamps = append(timestamps, TransformMsToTimestamp(*words[0].Start))
			lastIndex = 0
		} else {
			sentenceWords := strings.Split(sentence, " ")
			for i := lastIndex; i < len(words); i++ {
				if strings.Contains(sentence, words[i].Text) {
					if len(words) >= i+2 {
						if words[i].Text == sentenceWords[0] && words[i+1].Text == sentenceWords[1] && words[i+2].Text == sentenceWords[2] {
							timestamps = append(timestamps, TransformMsToTimestamp(*words[i].Start))
							lastIndex = i
							break
						}
					} else if len(words) >= i+1 {
						if words[i].Text == sentenceWords[0] && words[i+1].Text == sentenceWords[1] {
							timestamps = append(timestamps, TransformMsToTimestamp(*words[i].Start))
							lastIndex = i
							break
						}
					} else {
						if words[i].Text == sentenceWords[0] {
							timestamps = append(timestamps, TransformMsToTimestamp(*words[i].Start))
							lastIndex = i
							break
						}
					}
				}
			}

		}
	}
	return timestamps
}

func GetSentenceTimestampsAndSpeaker(sentences []string, words []SentimentAnalysisResult) [][]string {
	var lastIndex int
	timestamps := [][]string{}
	for index, sentence := range sentences {
		if sentence[0] == ' ' {
			sentence = sentence[1:]
		}
		if sentence != "" {
			if index == 0 {
				timestamps = append(timestamps, []string{TransformMsToTimestamp(*words[0].Start), fmt.Sprintf("(Speaker %s)", words[0].Speaker)})
				lastIndex = 0
			} else {
				sentenceWords := strings.Split(sentence, " ")
				for i := lastIndex; i < len(words); i++ {
					if strings.Contains(sentence, words[i].Text) {
						if len(words) >= i+1 {
							if words[i].Text == sentenceWords[0] {
								timestamps = append(timestamps, []string{TransformMsToTimestamp(*words[i].Start), fmt.Sprintf("(Speaker %s)", words[i].Speaker)})
								lastIndex = i
								break
							}
						} else {
							if words[i].Text == sentenceWords[0] && words[i+1].Text == sentenceWords[1] {
								timestamps = append(timestamps, []string{TransformMsToTimestamp(*words[i].Start), fmt.Sprintf("(Speaker %s)", words[i].Speaker)})
								lastIndex = i
								break
							}
						}
					}
				}

			}
		}
	}
	return timestamps
}

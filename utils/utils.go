/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	S "github.com/AssemblyAI/assemblyai-cli/schemas"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/posthog/posthog-go"
	"golang.org/x/term"
)

var AAIURL = "https://api.assemblyai.com/v2"
var PH_TOKEN string
var SENTRY_DNS string

func TelemetryCaptureEvent(event string, properties *S.PostHogProperties) {
	isTelemetryEnabled := GetConfigFileValue("features.telemetry")
	if isTelemetryEnabled == "true" {

		if PH_TOKEN == "" {
			godotenv.Load()
			PH_TOKEN = os.Getenv("POSTHOG_API_TOKEN")
		}

		client := posthog.New(PH_TOKEN)
		defer client.Close()

		distinctId := GetConfigFileValue("config.distinct_id")

		if distinctId == "" {
			distinctId = uuid.New().String()
			SetConfigFileValue("config.distinct_id", distinctId)
		}
		if properties != nil {
			var PhProperties posthog.Properties
			if properties.I == true {
				PhProperties = posthog.NewProperties().
					Set("OS", properties.OS).
					Set("Arch", properties.Arch).
					Set("Version", properties.Version).
					Set("Method", properties.Method)
			} else if properties.LatestVersion != "" {
				PhProperties = posthog.NewProperties().
					Set("latest_version", properties.LatestVersion).
					Set("current_version", properties.Version)
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

func PrintError(props S.PrintErrorProps) {
	err := props.Error
	message := props.Message
	if err != nil {
		isTelemetryEnabled := GetConfigFileValue("features.telemetry")
		if isTelemetryEnabled == "true" {
			InitSentry()
			sentry.CaptureException(err)
		}
		fmt.Println(message)
		os.Exit(1)
	}
}

func QueryApi(path string, method string, body io.Reader, s *spinner.Spinner) []byte {
	resp, err := http.NewRequest(method, AAIURL+path, body)
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
	}

	resp.Header.Add("Accept", "application/json")
	resp.Header.Add("Authorization", Token)
	resp.Header.Add("Transfer-Encoding", "chunked")

	response, err := http.DefaultClient.Do(resp)
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		if s != nil {
			s.Stop()
		}
		fmt.Println("Something just went wrong. Please try again.")
		os.Exit(1)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
	}
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

func GetSentenceTimestamps(sentences []string, words []S.SentimentAnalysisResult) []string {
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

func GetSentenceTimestampsAndSpeaker(sentences []string, words []S.SentimentAnalysisResult) [][]string {
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

func InitSentry() {
	isTelemetryEnabled := GetConfigFileValue("features.telemetry")
	if isTelemetryEnabled == "true" {
		if SENTRY_DNS == "" {
			godotenv.Load()
			SENTRY_DNS = os.Getenv("SENTRY_DNS")
		}
		sentrySyncTransport := sentry.NewHTTPSyncTransport()
		sentrySyncTransport.Timeout = time.Second * 3

		err := sentry.Init(sentry.ClientOptions{
			Dsn:              SENTRY_DNS,
			TracesSampleRate: 1.0,
			Transport:        sentrySyncTransport,
		})
		if err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}
		defer sentry.Flush(5 * time.Second)
	}
}

func CheckForUpdates(currentVersion string) {
	getWidth, _, err := term.GetSize(0)
	if err != nil {
		getWidth = 512
	}
	resp, err := http.Get("https://api.github.com/repos/assemblyai/assemblyai-cli/releases/latest")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var release S.Release
	err = json.Unmarshal(body, &release)
	if err != nil {
		return
	}
	if *release.TagName != currentVersion {
		underlinedBlue := color.New(color.FgHiBlue).Add(color.Underline).SprintFunc()
		blue := color.New(color.FgHiBlue).SprintFunc()
		black := color.New(color.FgHiBlack).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

		message := fmt.Sprintf("A new version of the CLI (%s) is available!", *release.TagName)
		instructions := "Run the installation scripts to upgrade it: https://github.com/AssemblyAI/assemblyai-cli#installation"

		messagePadding := (getWidth - len(message) - 2) / 2
		instructionsPadding := (getWidth - len(instructions) - 2) / 2
		if messagePadding < 0 {
			messagePadding = 0
		}
		if instructionsPadding < 0 {
			instructionsPadding = 0
		}
		fmt.Println(strings.Repeat(underlinedBlue(" "), getWidth))
		fmt.Printf("%s%s%s", blue("|"), strings.Repeat(" ", getWidth-2), blue("|"))
		fmt.Println()
		fmt.Printf("%s%s%s%s%s", blue("|"), strings.Repeat(" ", messagePadding), yellow(message), strings.Repeat(" ", messagePadding), blue("|"))
		fmt.Println()
		if len(instructions) < getWidth-2 {
			fmt.Printf("%s%s%s%s%s", blue("|"), strings.Repeat(" ", instructionsPadding), black(instructions), strings.Repeat(" ", instructionsPadding), blue("|"))
			fmt.Println()
		} else {
			instructions1 := "Run the installation scripts to upgrade it:"
			instructions2 := "https://github.com/AssemblyAI/assemblyai-cli#installation"
			instructions1Padding := (getWidth - len(instructions1) - 2) / 2
			instructions2Padding := (getWidth - len(instructions2) - 2) / 2
			fmt.Printf("%s%s%s%s%s", blue("|"), strings.Repeat(" ", instructions1Padding), black(instructions1), strings.Repeat(" ", instructions1Padding), blue("|"))
			fmt.Println()
			fmt.Printf("%s%s%s%s%s", blue("|"), strings.Repeat(" ", instructions2Padding), black(instructions2), strings.Repeat(" ", instructions2Padding), blue("|"))
			fmt.Println()
		}
		fmt.Printf("%s%s%s", blue("|"), strings.Repeat(underlinedBlue(" "), getWidth-2), blue("|"))
		fmt.Println()

		var properties *S.PostHogProperties = &S.PostHogProperties{
			Version:       currentVersion,
			LatestVersion: *release.TagName,
		}
		TelemetryCaptureEvent("CLI update available", properties)
	}
}

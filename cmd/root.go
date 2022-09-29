/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var rootCmd = &cobra.Command{
	Use:   "cli",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
		examples and usage of using your application. For example:

		Cobra is a CLI library for Go that empowers applications.
		This application is a tool to generate the needed files
		to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("help section")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

var AAITokenEnvName = "ASSMEBLYAI_TOKEN"
var AAIURL = "https://api.assemblyai.com/v2"

func GetDatabaseConfig() badger.Options {
	badgerCfg := badger.DefaultOptions("/tmp/badger")
	badgerCfg.Logger = nil
	return badgerCfg
}

func PrintError(err error) {
	if err != nil {
		// fmt.Println(err)
	}
}

func GetOpenDatabase() *badger.DB {
	badgerOptions := GetDatabaseConfig()
	db, err := badger.Open(badgerOptions)
	PrintError(err)
	return db
}

func GetStoredToken(db *badger.DB) string {
	var valCopy []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(AAITokenEnvName))
		PrintError(err)

		if err != nil {
			fmt.Println("You need to run `assemblyai config` first")
			return nil
		}

		err = item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)
			return nil
		})
		return nil
	})
	PrintError(err)

	return string(valCopy)
}

func QueryApi(token string, path string, method string, body io.Reader) []byte {
	resp, err := http.NewRequest(method, AAIURL+path, body)
	PrintError(err)

	resp.Header.Add("Accept", "application/json")
	resp.Header.Add("Authorization", token)

	response, err := http.DefaultClient.Do(resp)
	PrintError(err)
	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	PrintError(err)

	return responseData
}

func checkIfTokenValid(token string) CheckIfTokenValidResponse {
	var funcResponse CheckIfTokenValidResponse

	response := QueryApi(token, "/account", "GET", nil)

	if response == nil {
		return funcResponse
	}

	var result Account
	if err := json.Unmarshal(response, &result); err != nil {
		fmt.Println("Can not unmarshal JSON")
	}
	funcResponse.IsVerified = result.IsVerified
	funcResponse.CurrentBalance = fmt.Sprintf("%f", result.CurrentBalance.Amount)

	return funcResponse
}

func DeleteToken(db *badger.DB) {
	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(AAITokenEnvName))
		return err
	})
	PrintError(err)
}

func BeutifyJSON(data []byte) []byte {
	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, data, "", "\t")
	if error != nil {
		return data
	}
	return prettyJSON.Bytes()
}

func PollTranscription(token string, id string, flags TranscribeFlags) {
	fmt.Println("◑ We're processing your transcription...")

	for {
		response := QueryApi(token, "/transcript/"+id, "GET", nil)
		var transcript TranscriptResponse
		if err := json.Unmarshal(response, &transcript); err != nil {
			fmt.Println(err)
			return
		}
		if transcript.Status == "completed" {
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
	fmt.Print("\033[1A\033[K")

	width, _, err := term.GetSize(0)
	if err != nil {
		width = 512
	}

	fmt.Println("Transcript")
	if !transcript.SpeakerLabels && !*transcript.DualChannel {
		fmt.Println(transcript.Text)
	} else {
		GetFormattedUtterances(*transcript.Utterances, width)
	}
	if transcript.AutoHighlights {
		fmt.Println("Highlights")
		GetFormattedHighlights(*transcript.AutoHighlightsResult)
	}
	if transcript.ContentSafety {
		fmt.Println("Content Moderation")
		GetFormattedContentSafety(transcript.ContentSafetyLabels, width)
	}
	if transcript.IabCategories {
		fmt.Println("Topic Detection")
		GetFormattedTopicDetection(transcript.IabCategoriesResult, width)
	}
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
	textWidth := int(math.Abs(float64(width - 60)))

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

type CheckIfTokenValidResponse struct {
	IsVerified     bool   `json:"is_verified"`
	CurrentBalance string `json:"current_balance"`
}

type Account struct {
	IsVerified     bool           `json:"is_verified"`
	CurrentBalance CurrentBalance `json:"current_balance"`
}

type CurrentBalance struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type TranscriptResponse struct {
	AcousticModel            string                     `json:"acoustic_model"`
	AudioDuration            int64                      `json:"audio_duration"`
	AudioEndAt               interface{}                `json:"audio_end_at"`
	AudioStartFrom           interface{}                `json:"audio_start_from"`
	AudioURL                 string                     `json:"audio_url"`
	AutoChapters             bool                       `json:"auto_chapters"`
	AutoHighlights           bool                       `json:"auto_highlights"`
	AutoHighlightsResult     *AutoHighlightsResult      `json:"auto_highlights_result"`
	BoostParam               interface{}                `json:"boost_param"`
	Chapters                 *[]Chapter                 `json:"chapters"`
	ClusterID                interface{}                `json:"cluster_id"`
	Confidence               float64                    `json:"confidence"`
	ContentSafety            bool                       `json:"content_safety"`
	ContentSafetyLabels      ContentSafetyLabels        `json:"content_safety_labels"`
	CustomSpelling           interface{}                `json:"custom_spelling"`
	Disfluencies             bool                       `json:"disfluencies"`
	DualChannel              *bool                      `json:"dual_channel"`
	Entities                 *[]Entity                  `json:"entities"`
	EntityDetection          bool                       `json:"entity_detection"`
	FilterProfanity          bool                       `json:"filter_profanity"`
	FormatText               bool                       `json:"format_text"`
	IabCategories            bool                       `json:"iab_categories"`
	IabCategoriesResult      IabCategoriesResult        `json:"iab_categories_result"`
	ID                       string                     `json:"id"`
	LanguageCode             string                     `json:"language_code"`
	LanguageDetection        bool                       `json:"language_detection"`
	LanguageModel            string                     `json:"language_model"`
	Punctuate                bool                       `json:"punctuate"`
	RedactPii                bool                       `json:"redact_pii"`
	RedactPiiAudio           bool                       `json:"redact_pii_audio"`
	RedactPiiAudioQuality    interface{}                `json:"redact_pii_audio_quality"`
	RedactPiiPolicies        interface{}                `json:"redact_pii_policies"`
	RedactPiiSub             interface{}                `json:"redact_pii_sub"`
	SentimentAnalysis        bool                       `json:"sentiment_analysis"`
	SentimentAnalysisResults *[]SentimentAnalysisResult `json:"sentiment_analysis_results"`
	SpeakerLabels            bool                       `json:"speaker_labels"`
	SpeedBoost               bool                       `json:"speed_boost"`
	Status                   string                     `json:"status"`
	Text                     string                     `json:"text"`
	Throttled                interface{}                `json:"throttled"`
	Utterances               *[]SentimentAnalysisResult `json:"utterances"`
	WebhookAuth              bool                       `json:"webhook_auth"`
	WebhookAuthHeaderName    interface{}                `json:"webhook_auth_header_name"`
	WebhookStatusCode        interface{}                `json:"webhook_status_code"`
	WebhookURL               interface{}                `json:"webhook_url"`
	WordBoost                []interface{}              `json:"word_boost"`
	Words                    []SentimentAnalysisResult  `json:"words"`
}

type AutoHighlightsResult struct {
	Results []AutoHighlightsResultResult `json:"results"`
	Status  string                       `json:"status"`
}

type AutoHighlightsResultResult struct {
	Count      int64       `json:"count"`
	Rank       float64     `json:"rank"`
	Text       string      `json:"text"`
	Timestamps []Timestamp `json:"timestamps"`
}

type Timestamp struct {
	End   int64 `json:"end"`
	Start int64 `json:"start"`
}

type Chapter struct {
	End      int64  `json:"end"`
	Gist     string `json:"gist"`
	Headline string `json:"headline"`
	Start    int64  `json:"start"`
	Summary  string `json:"summary"`
}

type ContentSafetyLabels struct {
	Results              []ContentSafetyLabelsResult `json:"results"`
	SeverityScoreSummary SeverityScoreSummary        `json:"severity_score_summary"`
	Status               string                      `json:"status"`
	Summary              Summary                     `json:"summary"`
}

type ContentSafetyLabelsResult struct {
	Labels    []Label   `json:"labels"`
	Text      string    `json:"text"`
	Timestamp Timestamp `json:"timestamp"`
}

type Label struct {
	Confidence float64  `json:"confidence"`
	Label      string   `json:"label"`
	Severity   *float64 `json:"severity"`
}

type SeverityScoreSummary struct {
	Profanity Profanity `json:"profanity"`
}

type Profanity struct {
	Low    json.Number `json:"low"`
	Medium json.Number `json:"medium"`
	High   json.Number `json:"high"`
}

type Summary struct {
	Profanity float64 `json:"profanity"`
	Nsfw      float64 `json:"nsfw"`
}

type Entity struct {
	End        int64      `json:"end"`
	EntityType EntityType `json:"entity_type"`
	Start      int64      `json:"start"`
	Text       string     `json:"text"`
}

type IabCategoriesResult struct {
	Results []IabCategoriesResultResult `json:"results"`
	Status  string                      `json:"status"`
	Summary map[string]float64          `json:"summary"`
}

type IabCategoriesResultResult struct {
	Labels    []FluffyLabel `json:"labels"`
	Text      string        `json:"text"`
	Timestamp Timestamp     `json:"timestamp"`
}

type FluffyLabel struct {
	Label     string  `json:"label"`
	Relevance float64 `json:"relevance"`
}

type SentimentAnalysisResult struct {
	Confidence float64                   `json:"confidence"`
	End        int64                     `json:"end"`
	Sentiment  *Sentiment                `json:"sentiment,omitempty"`
	Speaker    Speaker                   `json:"speaker"`
	Start      int64                     `json:"start"`
	Text       string                    `json:"text"`
	Words      []SentimentAnalysisResult `json:"words,omitempty"`
}

type EntityType string

const (
	Location     EntityType = "location"
	Occupation   EntityType = "occupation"
	Organization EntityType = "organization"
	PersonName   EntityType = "person_name"
)

type Sentiment string

const (
	Negative Sentiment = "NEGATIVE"
	Neutral  Sentiment = "NEUTRAL"
	Positive Sentiment = "POSITIVE"
)

type Speaker string

const (
	A Speaker = "A"
	B Speaker = "B"
	C Speaker = "C"
	D Speaker = "D"
	E Speaker = "E"
	F Speaker = "F"
	G Speaker = "G"
	H Speaker = "H"
)

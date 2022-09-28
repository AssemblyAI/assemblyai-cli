/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/spf13/cobra"
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

// delete token
func DeleteToken(db *badger.DB) {
	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(AAITokenEnvName))
		return err
	})
	PrintError(err)
}

func PollTranscription(token string, id string, flags TranscribeFlags) {
	fmt.Println("◑ We're processing your transcription...")

	for {
		response := QueryApi(token, "/transcript/"+id, "GET", nil)
		var transcript TranscriptResponse
		if err := json.Unmarshal(response, &transcript); err != nil {
			fmt.Println("Can not unmarshal JSON")
			return
		}
		if transcript.Status == "completed" {
			if flags.Json {
				fmt.Println(string(response))
			}
			GetFormattedOutput(transcript, flags)
			return
		}
		time.Sleep(5 * time.Second)
	}
}

func GetFormattedOutput(transcript TranscriptResponse, flags TranscribeFlags) {
	if !transcript.SpeakerLabels {
		fmt.Println(transcript.Text)
		return
	}
	w := new(tabwriter.Writer)

	w.Init(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)

	for _, utterance := range transcript.Utterances {

		duration := time.Duration(utterance.Start) * time.Millisecond
		start := fmt.Sprintf("%02d:%02d", int(duration.Minutes()), int(duration.Seconds())%60)
		speaker := fmt.Sprintf("(Speaker %s)", utterance.Speaker)

		fmt.Fprintf(w, "%s\t%s\t%s\t\n", start, speaker, utterance.Text)
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
	ID                       string                    `json:"id"`
	LanguageModel            string                    `json:"language_model"`
	AcousticModel            string                    `json:"acoustic_model"`
	LanguageCode             string                    `json:"language_code"`
	Status                   string                    `json:"status"`
	AudioURL                 string                    `json:"audio_url"`
	Text                     string                    `json:"text"`
	Words                    []SentimentAnalysisResult `json:"words"`
	Utterances               []SentimentAnalysisResult `json:"utterances"`
	Confidence               float64                   `json:"confidence"`
	AudioDuration            int64                     `json:"audio_duration"`
	Punctuate                bool                      `json:"punctuate"`
	FormatText               bool                      `json:"format_text"`
	DualChannel              bool                      `json:"dual_channel"`
	WebhookURL               interface{}               `json:"webhook_url"`
	WebhookStatusCode        interface{}               `json:"webhook_status_code"`
	WebhookAuth              bool                      `json:"webhook_auth"`
	WebhookAuthHeaderName    interface{}               `json:"webhook_auth_header_name"`
	SpeedBoost               bool                      `json:"speed_boost"`
	AutoHighlightsResult     AutoHighlightsResult      `json:"auto_highlights_result"`
	AutoHighlights           bool                      `json:"auto_highlights"`
	AudioStartFrom           interface{}               `json:"audio_start_from"`
	AudioEndAt               interface{}               `json:"audio_end_at"`
	WordBoost                []interface{}             `json:"word_boost"`
	BoostParam               interface{}               `json:"boost_param"`
	FilterProfanity          bool                      `json:"filter_profanity"`
	RedactPii                bool                      `json:"redact_pii"`
	RedactPiiAudio           bool                      `json:"redact_pii_audio"`
	RedactPiiAudioQuality    string                    `json:"redact_pii_audio_quality"`
	RedactPiiPolicies        []string                  `json:"redact_pii_policies"`
	RedactPiiSub             string                    `json:"redact_pii_sub"`
	SpeakerLabels            bool                      `json:"speaker_labels"`
	ContentSafety            bool                      `json:"content_safety"`
	IabCategories            bool                      `json:"iab_categories"`
	ContentSafetyLabels      ContentSafetyLabels       `json:"content_safety_labels"`
	IabCategoriesResult      IabCategoriesResult       `json:"iab_categories_result"`
	LanguageDetection        bool                      `json:"language_detection"`
	CustomSpelling           interface{}               `json:"custom_spelling"`
	ClusterID                interface{}               `json:"cluster_id"`
	Throttled                interface{}               `json:"throttled"`
	Disfluencies             bool                      `json:"disfluencies"`
	SentimentAnalysis        bool                      `json:"sentiment_analysis"`
	AutoChapters             bool                      `json:"auto_chapters"`
	Chapters                 []Chapter                 `json:"chapters"`
	SentimentAnalysisResults []SentimentAnalysisResult `json:"sentiment_analysis_results"`
	EntityDetection          bool                      `json:"entity_detection"`
	Entities                 []Entity                  `json:"entities"`
}

type AutoHighlightsResult struct {
	Status  string                       `json:"status"`
	Results []AutoHighlightsResultResult `json:"results"`
}

type AutoHighlightsResultResult struct {
	Count      int64       `json:"count"`
	Rank       float64     `json:"rank"`
	Text       string      `json:"text"`
	Timestamps []Timestamp `json:"timestamps"`
}

type Timestamp struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

type Chapter struct {
	Summary  string `json:"summary"`
	Headline string `json:"headline"`
	Gist     string `json:"gist"`
	Start    int64  `json:"start"`
	End      int64  `json:"end"`
}

type ContentSafetyLabels struct {
	Status               string                      `json:"status"`
	Results              []ContentSafetyLabelsResult `json:"results"`
	Summary              Summary                     `json:"summary"`
	SeverityScoreSummary SeverityScoreSummary        `json:"severity_score_summary"`
}

type ContentSafetyLabelsResult struct {
	Text      string        `json:"text"`
	Labels    []PurpleLabel `json:"labels"`
	Timestamp Timestamp     `json:"timestamp"`
}

type PurpleLabel struct {
	Label      string   `json:"label"`
	Confidence float64  `json:"confidence"`
	Severity   *float64 `json:"severity"`
}

type SeverityScoreSummary struct {
	Profanity Profanity `json:"profanity"`
}

type Profanity struct {
	Low    int64 `json:"low"`
	Medium int64 `json:"medium"`
	High   int64 `json:"high"`
}

type Summary struct {
	Profanity float64 `json:"profanity"`
	Nsfw      float64 `json:"nsfw"`
}

type Entity struct {
	EntityType EntityType `json:"entity_type"`
	Text       string     `json:"text"`
	Start      int64      `json:"start"`
	End        int64      `json:"end"`
}

type IabCategoriesResult struct {
	Status  string                      `json:"status"`
	Results []IabCategoriesResultResult `json:"results"`
	Summary map[string]float64          `json:"summary"`
}

type IabCategoriesResultResult struct {
	Text      string        `json:"text"`
	Labels    []FluffyLabel `json:"labels"`
	Timestamp Timestamp     `json:"timestamp"`
}

type FluffyLabel struct {
	Relevance float64 `json:"relevance"`
	Label     string  `json:"label"`
}

type SentimentAnalysisResult struct {
	Text       string                    `json:"text"`
	Start      int64                     `json:"start"`
	End        int64                     `json:"end"`
	Sentiment  *Sentiment                `json:"sentiment,omitempty"`
	Confidence float64                   `json:"confidence"`
	Speaker    Speaker                   `json:"speaker"`
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

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
	"math"
	"net/http"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/briandowns/spinner"
	badger "github.com/dgraph-io/badger/v3"
	"golang.org/x/term"
)

var AAITokenEnvName = "ASSMEBLYAI_TOKEN"
var AAIURL = "https://api.assemblyai.com/v2"

func GetDatabaseConfig() badger.Options {
	badgerCfg := badger.DefaultOptions("/tmp/badger")
	badgerCfg.Logger = nil
	return badgerCfg
}

func callSpinner(message string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[7], 100*time.Millisecond, spinner.WithSuffix(message))
	s.Start()
	return s
}

func PrintError(err error) {
	if err != nil {
		fmt.Println(err)
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
	resp.Header.Add("Transfer-Encoding", "chunked")

	response, err := http.DefaultClient.Do(resp)
	PrintError(err)

	responseData, err := ioutil.ReadAll(response.Body)
	PrintError(err)
	defer response.Body.Close()

	return responseData
}

func CheckIfTokenValid(token string) CheckIfTokenValidResponse {
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

func UploadFile(token string, path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	s := callSpinner(" Your file is being uploaded...")
	response := QueryApi(token, "/upload", "POST", file)
	var uploadResponse UploadResponse
	if err := json.Unmarshal(response, &uploadResponse); err != nil {
		return ""
	}
	s.Stop()
	return uploadResponse.UploadURL
}

func PollTranscription(token string, id string, flags TranscribeFlags) {
	s := callSpinner(" Your file is being transcribed...")
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

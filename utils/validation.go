package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"

	S "github.com/AssemblyAI/assemblyai-cli/schemas"
	"github.com/spf13/pflag"
)

func ValidateParams(params S.TranscribeParams, flagSet *pflag.FlagSet) {
	if params.WordBoost == nil && params.BoostParam != "" {
		printErrorProps := S.PrintErrorProps{
			Error:   errors.New("Please provide a valid word boost"),
			Message: "To boost a word, please provide a valid list of words to boost. For example: --word_boost \"word1,word2,word3\"  --boost_param high",
		}
		PrintError(printErrorProps)
		return
	} else if params.BoostParam != "" && params.BoostParam != "low" && params.BoostParam != "default" && params.BoostParam != "high" {
		printErrorProps := S.PrintErrorProps{
			Error:   errors.New("Invalid boost_param"),
			Message: "Please provide a valid boost_param. Valid values are low, default, or high.",
		}
		PrintError(printErrorProps)
		return
	}

	if !params.Summarization {
		params.SummaryType = ""
		params.SummaryModel = ""
	} else {
		params.Punctuate = true
		params.FormatText = true
		if _, ok := S.SummarizationTypeMapReverse[params.SummaryType]; !ok {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("Invalid summary type"),
				Message: "Invalid summary type. To know more about Summarization, head over to https://assemblyai.com/docs/audio-intelligence#summarization",
			}
			PrintError(printErrorProps)
			return
		}
		if _, ok := S.SummarizationModelMap[params.SummaryModel]; !ok {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("Invalid summary model"),
				Message: "Invalid summary model. To know more about Summarization, head over to https://assemblyai.com/docs/audio-intelligence#summarization",
			}
			PrintError(printErrorProps)
			return
		}
		if !Contains(S.SummarizationModelMap[params.SummaryModel], params.SummaryType) {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("Invalid summary model"),
				Message: "Cant use summary model " + params.SummaryModel + " with summary type " + params.SummaryType + ". To know more about Summarization, head over to https://assemblyai.com/docs/audio-intelligence#summarization",
			}
			PrintError(printErrorProps)
			return
		}
		if params.SummaryModel == "conversational" && !params.SpeakerLabels {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("Speaker labels required for conversational summary model"),
				Message: "Speaker labels are required for conversational summarization. To know more about Summarization, head over to https://assemblyai.com/docs/audio-intelligence#summarization",
			}
			PrintError(printErrorProps)
			return
		}
	}

	if !params.RedactPii {
		params.RedactPiiPolicies = nil
	} else {
		for _, policy := range params.RedactPiiPolicies {
			if _, ok := S.PIIRedactionPolicyMap[policy]; !ok {
				printErrorProps := S.PrintErrorProps{
					Error:   errors.New("Invalid redaction policy"),
					Message: fmt.Sprintf("%s is not a valid policy. See https://www.assemblyai.com/docs/audio-intelligence#pii-redaction for the complete list of supported policies.", policy),
				}
				PrintError(printErrorProps)
				return
			}
		}
	}

	if params.LanguageCode != "" {
		if params.LanguageDetection {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("Language detection and language code cannot be used together"),
				Message: "Language detection and language code cannot be used together.",
			}
			PrintError(printErrorProps)
			return
		}
		if params.SpeakerLabels {
			if flagSet.Lookup("speaker_labels").Changed {
				printErrorProps := S.PrintErrorProps{
					Error:   errors.New("Speaker labels are not supported for languages other than English"),
					Message: "Speaker labels are not supported for languages other than English.",
				}
				PrintError(printErrorProps)
				return
			} else {
				params.SpeakerLabels = false
			}
		}
		if _, ok := S.LanguageMap[params.LanguageCode]; !ok {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("Invalid language code"),
				Message: "Invalid language code. See https://www.assemblyai.com/docs#supported-languages for supported languages.",
			}
			PrintError(printErrorProps)
			return
		}
	}
	if params.LanguageDetection && params.SpeakerLabels {
		if flagSet.Lookup("speaker_labels").Changed {
			printErrorProps := S.PrintErrorProps{
				Error:   errors.New("Speaker labels are not supported for languages other than English"),
				Message: "Speaker labels are not supported for languages other than English.",
			}
			PrintError(printErrorProps)
			return
		} else {
			params.SpeakerLabels = false
		}
	}

	customSpelling, _ := flagSet.GetString("custom_spelling")
	if customSpelling != "" {
		parsedCustomSpelling := []S.CustomSpelling{}

		_, err := os.Stat(customSpelling)

		if !os.IsNotExist(err) {
			file, err := os.Open(customSpelling)
			if err != nil {
				printErrorProps := S.PrintErrorProps{
					Error:   err,
					Message: "Error opening custom spelling file",
				}
				PrintError(printErrorProps)
				return
			}
			defer file.Close()
			byteCustomSpelling, err := ioutil.ReadAll(file)
			if err != nil {
				printErrorProps := S.PrintErrorProps{
					Error:   err,
					Message: "Error reading custom spelling file",
				}
				PrintError(printErrorProps)
				return
			}

			err = json.Unmarshal(byteCustomSpelling, &parsedCustomSpelling)
			if err != nil {
				printErrorProps := S.PrintErrorProps{
					Error:   err,
					Message: "Error parsing custom spelling file",
				}
				PrintError(printErrorProps)
				return
			}
		} else {
			err = json.Unmarshal([]byte(customSpelling), &parsedCustomSpelling)
			if err != nil {
				printErrorProps := S.PrintErrorProps{
					Error:   err,
					Message: "Invalid custom spelling. Please provide a valid custom spelling JSON.",
				}
				PrintError(printErrorProps)
				return
			}
		}

		err = validateCustomSpelling(parsedCustomSpelling)
		if err != nil {
			printErrorProps := S.PrintErrorProps{
				Error:   err,
				Message: "Invalid custom spelling. Please provide a valid custom spelling JSON.",
			}
			PrintError(printErrorProps)
			return
		}
		params.CustomSpelling = parsedCustomSpelling
	}
}

func ValidateFlags(flags S.TranscribeFlags) {
	if flags.Csv != "" && !flags.Poll {
		printErrorProps := S.PrintErrorProps{
			Error:   errors.New("CSV output is only supported with polling"),
			Message: "CSV output is only supported with polling.",
		}
		PrintError(printErrorProps)
		return
	}
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

func validateCustomSpelling(customSpelling []S.CustomSpelling) error {
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

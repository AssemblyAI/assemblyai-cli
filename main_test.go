package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"testing"

	S "github.com/AssemblyAI/assemblyai-cli/schemas"
	"github.com/AssemblyAI/assemblyai-cli/utils"
	U "github.com/AssemblyAI/assemblyai-cli/utils"
)

func TestVersion(t *testing.T) {
	out, err := exec.Command("go", "run", "main.go", "-v", "--test").Output()
	if err != nil {
		fmt.Println(err)
	}
	version := utils.GetEnvWithKey("VERSION")
	if version == nil {
		t.Error("VERSION not set")
	}
	if string(out) != "AssemblyAI CLI "+*version+"\n" {
		t.Errorf("Expected AssemblyAI CLI v1.13, got %s.", string(out))
	}
}

func TestValidate(t *testing.T) {
	out, err := exec.Command("go", "run", "main.go", "validate", "--test").Output()
	if err != nil {
		fmt.Println(err)
	}
	if string(out) != "Please start by running \033[1m\033[34massemblyai config [token]\033[0m\n" {
		t.Errorf("Expected Please start by running \033[1m\033[34massemblyai config [token]\033[0m, got %s.", string(out))
	}
}

func TestAuthBad(t *testing.T) {
	out, err := exec.Command("go", "run", "main.go", "config", "invalid", "--test").Output()
	if err != nil {
		fmt.Println(err)
	}
	if string(out) != U.INVALID_TOKEN+"\n" {
		t.Errorf("Expected Something just went wrong. Please try again., got %s.", string(out))
	}
}

func TestAuthCorrect(t *testing.T) {
	token := utils.GetEnvWithKey("TOKEN")
	out, err := exec.Command("go", "run", "main.go", "config", *token, "--test").Output()
	if err != nil {
		fmt.Println(err)
	}
	if string(out) != "You're now authenticated.\n" {
		t.Errorf("Expected You're now authenticated., got %s.", string(out))
	}
}

func TestTranscribeInvalidFlags(t *testing.T) {
	out, err := exec.Command("go", "run", "main.go", "transcribe", "-i", "invalid", "-o", "invalid", "--test").Output()
	if err != nil {
		fmt.Println(err)
	}
	if string(out) != "\nrequires at least 1 arg(s), only received 0\n" {
		t.Errorf("Expected requires at least 1 arg(s), only received 0, got %s.", string(out))
	}
}

func TestTranscribeBadYoutube(t *testing.T) {
	out, err := exec.Command("go", "run", "main.go", "transcribe", "https://www.youtube.com/watch?vs=m3cSH7jK3UU", "--test").Output()
	if err != nil {
		fmt.Println(err)
	}
	if string(out) != "\nCould not find YouTube ID in URL\n" {
		t.Errorf("Expected Could not find YouTube ID in URL, got %s.", string(out))
	}
}

func TestTranscribeBadFile(t *testing.T) {
	out, err := exec.Command("go", "run", "main.go", "transcribe", "invalid", "--test").Output()
	if err != nil {
		fmt.Println(err)
	}
	if string(out) != "\nError opening file\n" {
		t.Errorf("Expected Error opening file, got %s.", string(out))
	}
}

func TestTranscribeWithFlags(t *testing.T) {
	out, err := exec.Command(
		"go",
		"run",
		"main.go",
		"transcribe",
		"https://storage.googleapis.com/aai-web-samples/2%20min.ogg",
		"--auto_highlights",
		"--content_moderation",
		"--entity_detection",
		"--format_text",
		"--punctuate",
		"--redact_pii",
		"--sentiment_analysis",
		"--speaker_labels",
		"--summarization",
		"--topic_detection",
		"-p=false",
		"-j",
		"--test",
	).Output()
	if err != nil {
		fmt.Println(err)
	}

	var result S.TranscriptResponse
	json.Unmarshal(out, &result)

	if *result.Status != "queued" {
		t.Errorf("Expected queued, got %s.", *result.Status)
	}
	if *result.AutoHighlights != true {
		t.Errorf("Expected Auto Highlights true, got false.")
	}
	if *result.ContentSafety != true {
		t.Errorf("Expected Content Safety true, got false.")
	}
	if *result.EntityDetection != true {
		t.Errorf("Expected Entity Detection true, got false.")
	}
	if *result.FormatText != true {
		t.Errorf("Expected Format Text true, got false.")
	}
	if *result.Punctuate != true {
		t.Errorf("Expected Punctuate true, got false.")
	}
	if *result.RedactPii != true {
		t.Errorf("Expected RedactPII true, got false.")
	}
	if *result.SentimentAnalysis != true {
		t.Errorf("Expected Sentiment Analysis true, got false.")
	}
	if result.SpeakerLabels != true {
		t.Errorf("Expected Speaker Labels true, got false.")
	}
	if *result.Summarization != true {
		t.Errorf("Expected Summarization true, got false.")
	}
	if *result.IabCategories != true {
		t.Errorf("Expected IAB Categories(Topic detection) true, got false.")
	}
}

func TestTranscribeRestrictions(t *testing.T) {
	// Speaker Labels && Dual Channel
	out, err := exec.Command(
		"go",
		"run",
		"main.go",
		"transcribe",
		"https://storage.googleapis.com/aai-web-samples/2%20min.ogg",
		"--speaker_labels",
		"--dual_channel",
		"-p=false",
		"-j",
		"--test",
	).Output()
	if err != nil {
		fmt.Println(err)
	}
	if string(out) != "\nSpeaker labels are not supported for dual channel audio\n" {
		t.Errorf("Expected Speaker labels are not supported for dual channel audio, got %s.", string(out))
	}

	// Auto Chapters && Summarization
	out, err = exec.Command(
		"go",
		"run",
		"main.go",
		"transcribe",
		"https://storage.googleapis.com/aai-web-samples/2%20min.ogg",
		"--auto_chapters",
		"--summarization",
		"-p=false",
		"-j",
		"--test",
	).Output()
	if err != nil {
		fmt.Println(err)
	}
	if string(out) != "\nAuto chapters are not supported for summarization\n" {
		t.Errorf("Expected Auto chapters are not supported for summarization, got %s.", string(out))
	}

	// Language Detection && Language Code
	out, err = exec.Command(
		"go",
		"run",
		"main.go",
		"transcribe",
		"https://storage.googleapis.com/aai-web-samples/2%20min.ogg",
		"--language_detection",
		"--language_code=en-US",
		"-p=false",
		"-j",
		"--test",
	).Output()
	if err != nil {
		fmt.Println(err)
	}
	if string(out) != "\nPlease provide either language detection or language code, not both.\n" {
		t.Errorf("Expected Please provide either language detection or language code, not both., got %s.", string(out))
	}

	// Language Detection && Speaker labels
	out, err = exec.Command(
		"go",
		"run",
		"main.go",
		"transcribe",
		"https://storage.googleapis.com/aai-web-samples/2%20min.ogg",
		"--language_detection",
		"--speaker_labels",
		"-p=false",
		"-j",
		"--test",
	).Output()
	if err != nil {
		fmt.Println(err)
	}
	if string(out) != "\nSpeaker labels are not supported for languages other than English.\n" {
		t.Errorf("Expected Speaker labels are not supported for languages other than English., got %s.", string(out))
	}
}

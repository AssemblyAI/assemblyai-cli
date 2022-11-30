package main

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/AssemblyAI/assemblyai-cli/utils"
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
	token := utils.GetEnvWithKey("TOKEN")
	if token != nil {
		if string(out) != "Your Token is "+*token+"\n" {
			t.Errorf("Expected Your Token is , got %s.", string(out))
		}
	} else {
		if string(out) != "Please start by running \033[1m\033[34massemblyai config [token]\033[0m\n" {
			t.Errorf("Expected Please start by running \033[1m\033[34massemblyai config [token]\033[0m, got %s.", string(out))
		}
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

func TestAuthBad(t *testing.T) {
	out, err := exec.Command("go", "run", "main.go", "config", "invalid", "--test").Output()
	if err != nil {
		fmt.Println(err)
	}
	if string(out) != "Something just went wrong. Please try again.\n" {
		t.Errorf("Expected Something just went wrong. Please try again., got %s.", string(out))
	}
}

func TestTranscribeInvalidFlags(t *testing.T) {
	TestAuthCorrect(t)
	out, err := exec.Command("go", "run", "main.go", "transcribe", "-i", "invalid", "-o", "invalid", "--test").Output()
	if err != nil {
		fmt.Println(err)
	}
	if string(out) != "requires at least 1 arg(s), only received 0\n" {
		t.Errorf("Expected requires at least 1 arg(s), only received 0, got %s.", string(out))
	}
}

func TestTranscribeBadYoutube(t *testing.T) {
	TestAuthCorrect(t)
	out, err := exec.Command("go", "run", "main.go", "transcribe", "https://www.youtube.com/watch?vs=m3cSH7jK3UU", "--test").Output()
	if err != nil {
		fmt.Println(err)
	}
	if string(out) != "Could not find YouTube ID in URL\n" {
		t.Errorf("Expected Could not find YouTube ID in URL, got %s.", string(out))
	}
}

func TestTranscribeBadFile(t *testing.T) {
	out, err := exec.Command("go", "run", "main.go", "transcribe", "invalid", "--test").Output()
	if err != nil {
		fmt.Println(err)
	}
	if string(out) != "Error opening file\n" {
		t.Errorf("Expected Error opening file, got %s.", string(out))
	}
}

func TestTranscribeYoutube(t *testing.T) {

}

package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	S "github.com/AssemblyAI/assemblyai-cli/schemas"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/posthog/posthog-go"
	"github.com/spf13/viper"
)

var ConfigFolderPath = ".config/assemblyai"
var ConfigFileName = "config.toml"
var Token string
var DistinctId string

func CheckIfTokenValid() bool {
	response := QueryApi("/account", "GET", nil, nil)
	if response == nil {
		return false
	}
	var result S.Account
	if err := json.Unmarshal(response, &result); err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
	}
	if result.Error != nil {
		return false
	}
	if result.Id != nil {
		DistinctId = strconv.Itoa(*result.Id)
	} else {
		DistinctId = uuid.New().String()
	}

	return true
}

func ConfigFolderExist() bool {
	home, err := os.UserHomeDir()
	configFile := filepath.Join(home, ConfigFolderPath, ConfigFileName)
	_, err = os.Stat(configFile)
	return errors.Is(err, os.ErrNotExist) == false
}

func SetConfigFileValue(key string, value string) {
	home, err := os.UserHomeDir()
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
		return
	}
	configFolder := filepath.Join(home, ConfigFolderPath)
	configFile := filepath.Join(configFolder, ConfigFileName)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("Please start by running \033[1m\033[34massemblyai config [token]\033[0m")
		return
	}
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("toml")
	viper.AddConfigPath(configFolder)
	viper.Set(key, value)
	viper.WriteConfig()
}

func GetConfigFileValue(key string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	configFolder := filepath.Join(home, ConfigFolderPath)
	configFile := filepath.Join(configFolder, ConfigFileName)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return ""
	}
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("toml")
	viper.AddConfigPath(configFolder)
	viper.ReadInConfig()
	return viper.GetString(key)
}

func GetStoredToken() string {
	return GetConfigFileValue("config.token")
}

func SetUserAlias() {
	if GetConfigFileValue("features.telemetry") == "true" {
		tempID := GetConfigFileValue("config.distinct_id")
		if tempID != DistinctId {
			if PH_TOKEN == "" {
				godotenv.Load()
				PH_TOKEN = os.Getenv("POSTHOG_API_TOKEN")
			}

			client := posthog.New(PH_TOKEN)
			defer client.Close()

			client.Enqueue(posthog.Alias{
				DistinctId: DistinctId,
				Alias:      tempID,
			})
		}
	}
}

func CreateConfigFile() {
	home, err := os.UserHomeDir()
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
		return
	}
	configFolder := filepath.Join(home, ConfigFolderPath)
	err = os.MkdirAll(configFolder, 0755)
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
		return
	}

	configFile := filepath.Join(configFolder, ConfigFileName)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		_, err := os.Create(configFile)
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
		return
	}
}

func GetEnvWithKey(key string) *string {
	godotenv.Load()
	value := os.Getenv(key)
	return &value
}

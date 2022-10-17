/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/posthog/posthog-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFolderPath = ".config/assemblyai"
var configFileName = "config.toml"
var Token string
var distinctId string

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config [token]",
	Short: "Authenticate the CLI",
	Long:  `This command will validate your account and store your token safely, later to be used when transcribing files.`,
	Run: func(cmd *cobra.Command, args []string) {

		argsArray := cmd.Flags().Args()

		if len(argsArray) == 0 {
			fmt.Println("Please provide a token. If you don't have one, create an account at https://app.assemblyai.com")
			return
		} else if len(argsArray) > 1 {
			fmt.Println("Too many arguments. Please provide a single token.")
			return
		}
		Token = argsArray[0]

		checkToken := CheckIfTokenValid()
		if !checkToken {
			fmt.Println("Your token appears to be invalid. Try again, and if the problem persists, contact support at support@assemblyai.com")
			return
		}

		if getConfigFileValue("config.new") == "true" {
			SetUserAlias()
		}

		createConfigFile()
		setConfigFileValue("features.telemetry", "true")
		setConfigFileValue("config.token", Token)
		setConfigFileValue("config.distinct_id", distinctId)
		setConfigFileValue("config.new", "false")

		TelemetryCaptureEvent("CLI configured", nil)

		fmt.Println("You're now authenticated.")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func CheckIfTokenValid() bool {
	response := QueryApi("/account", "GET", nil)
	if response == nil {
		return false
	}
	var result Account
	if err := json.Unmarshal(response, &result); err != nil {
		PrintError(err)
	}
	if result.Error != nil {
		return false
	}
	if result.Id != nil {
		distinctId = strconv.Itoa(*result.Id)
	} else {
		distinctId = uuid.New().String()
	}

	return true
}

func createConfigFile() {
	home, err := os.UserHomeDir()
	if err != nil {
		PrintError(err)
		return
	}
	configFolder := filepath.Join(home, configFolderPath)
	err = os.MkdirAll(configFolder, 0755)
	if err != nil {
		PrintError(err)
		return
	}

	configFile := filepath.Join(configFolder, configFileName)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		_, err := os.Create(configFile)
		PrintError(err)
		return
	}
}

func setConfigFileValue(key string, value string) {
	home, err := os.UserHomeDir()
	if err != nil {
		PrintError(err)
		return
	}
	configFolder := filepath.Join(home, configFolderPath)
	configFile := filepath.Join(configFolder, configFileName)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("Config file does not exist. Please run `assemblyai config` first.")
		return
	}
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("toml")
	viper.AddConfigPath(configFolder)
	viper.Set(key, value)
	viper.WriteConfig()
}

func getConfigFileValue(key string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	configFolder := filepath.Join(home, configFolderPath)
	configFile := filepath.Join(configFolder, configFileName)
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
	return getConfigFileValue("config.token")
}

func SetUserAlias() {
	if getConfigFileValue("features.telemetry") == "true" {
		tempID := getConfigFileValue("config.distinct_id")
		if tempID != distinctId {
			if PH_TOKEN == "" {
				godotenv.Load()
				PH_TOKEN = os.Getenv("POSTHOG_API_TOKEN")
			}

			client := posthog.New(PH_TOKEN)
			defer client.Close()

			client.Enqueue(posthog.Alias{
				DistinctId: distinctId,
				Alias:      tempID,
			})
		}
	}
}

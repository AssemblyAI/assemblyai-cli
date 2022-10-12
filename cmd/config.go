/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFolderPath = ".config/assemblyai"
var configFileName = "config.toml"
var Token string

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config <token>",
	Short: "Authenticate with AssemblyAI",
	Long: `Authenticate with AssemblyAI. 
	You can get your token from https://app.assemblyai.com/account.`,
	Run: func(cmd *cobra.Command, args []string) {

		argsArray := cmd.Flags().Args()

		if len(argsArray) == 0 {
			fmt.Println("Please provide a token. If you don't have one, you can get it at https://app.assemblyai.com")
			return
		} else if len(argsArray) > 1 {
			fmt.Println("Too many arguments. Please provide a single token.")
			return
		}
		token := argsArray[0]

		checkToken := CheckIfTokenValid(token)

		if !checkToken {
			fmt.Println("Invalid token. Try again, and if the problem persists, contact support at support@assemblyai.com")
			return
		}
		distinctId := uuid.New().String()

		createConfigFile()
		setConfigFileValue("features.telemetry", "true")
		setConfigFileValue("config.token", token)
		setConfigFileValue("config.distinct_id", distinctId)

		TelemetryCaptureEvent("CLI configured", nil)

		fmt.Println("You're now authenticated.")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func CheckIfTokenValid(token string) bool {

	response := QueryApi("/account", "GET", nil)

	if response == nil {
		return false
	}

	var result Account
	if err := json.Unmarshal(response, &result); err != nil {
		fmt.Println("Can not unmarshal JSON")
	}
	if result.Error != nil {
		return false
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

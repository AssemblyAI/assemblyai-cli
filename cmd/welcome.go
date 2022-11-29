package cmd

import (
	"fmt"
	"runtime"

	S "github.com/AssemblyAI/assemblyai-cli/schemas"
	U "github.com/AssemblyAI/assemblyai-cli/utils"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// welcomeCmd represents the welcome command
var welcomeCmd = &cobra.Command{
	Hidden: true,
	Use:    "welcome",
	Short:  "Welcome to AssemblyAI CLI!",
	Long:   "We are excited to announce the AssemblyAI CLI, a quick way to test our latest models right from your terminal, with minimal installation required.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to the AssemblyAI CLI!")

		i, _ := cmd.Flags().GetBool("i")
		if i {
			isUpgrading := U.ConfigFolderExist()
			U.CreateConfigFile()

			if !isUpgrading {
				U.SetConfigFileValue("features.telemetry", "true")
				fmt.Println("Please start by running \033[1m\033[34massemblyai config [token]\033[0m")
			}

			distinctId := U.GetConfigFileValue("config.distinct_id")
			if isUpgrading == false && distinctId == "" {
				U.SetConfigFileValue("config.distinct_id", uuid.New().String())
				U.SetConfigFileValue("config.new", "true")
			}

			var properties *S.PostHogProperties = new(S.PostHogProperties)
			properties.I = i
			properties.OS, _ = cmd.Flags().GetString("os")
			properties.Arch, _ = cmd.Flags().GetString("arch")
			properties.Method, _ = cmd.Flags().GetString("method")
			properties.Version, _ = cmd.Flags().GetString("version")
			if properties.OS == "" {
				// get system os
				properties.OS = runtime.GOOS
			}
			if properties.Arch == "" {
				// get system arch
				properties.Arch = runtime.GOARCH
			}

			U.TelemetryCaptureEvent("CLI installed", properties)
		}

	},
}

func init() {
	rootCmd.AddCommand(welcomeCmd)
	welcomeCmd.PersistentFlags().BoolP("i", "i", false, "")
	welcomeCmd.PersistentFlags().StringP("os", "o", "", "")
	welcomeCmd.PersistentFlags().StringP("arch", "a", "", "")
	welcomeCmd.PersistentFlags().StringP("method", "m", "", "")
	welcomeCmd.PersistentFlags().StringP("version", "v", "", "")
	welcomeCmd.Flags().Bool("test", false, "Flag for test executing purpose")
	welcomeCmd.Flags().MarkHidden("test")
}

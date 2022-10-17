/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"runtime"

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
		fmt.Println("Please start by running `assemblyai config [token]`")

		i, _ := cmd.Flags().GetBool("i")
		if i {
			createConfigFile()
			setConfigFileValue("features.telemetry", "true")

			distinctId := getConfigFileValue("config.distinct_id")
			if distinctId == "" {
				setConfigFileValue("config.distinct_id", uuid.New().String())
				setConfigFileValue("config.new", "true")
			}
			var properties *PostHogProperties = new(PostHogProperties)
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

			TelemetryCaptureEvent("CLI installed", properties)
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
}

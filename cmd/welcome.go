/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// welcomeCmd represents the welcome command
var welcomeCmd = &cobra.Command{
	Hidden: true,
	Use:    "welcome",
	Short:  "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// welcome to cli
		fmt.Println("Welcome to the AssemblyAI CLI!")
		fmt.Println("Please start by running `assemblyai config <token>`")

		i, _ := cmd.Flags().GetBool("i")
		if i {
			createConfigFile()
			setConfigFileValue("features.telemetry", "true")
			setConfigFileValue("config.distinct_id", uuid.New().String())
			var properties *PostHogProperties = new(PostHogProperties)
			properties.I, _ = cmd.Flags().GetBool("i")
			properties.OS, _ = cmd.Flags().GetString("os")
			properties.Arch, _ = cmd.Flags().GetString("arch")
			properties.Method, _ = cmd.Flags().GetString("method")
			properties.Version, _ = cmd.Flags().GetString("version")

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

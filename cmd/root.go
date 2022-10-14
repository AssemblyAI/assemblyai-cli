/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var VERSION string

var rootCmd = &cobra.Command{
	Use:   "assemblyai",
	Short: "AssemblyAI CLI",
	Long: `Please authenticate with AssemblyAI to use this CLI.
assemblyai config {YOUR TOKEN}`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		versionFlag, _ := cmd.Flags().GetBool("version")
		if versionFlag {
			if VERSION == "" {
				godotenv.Load()
				VERSION = os.Getenv("VERSION")
			}
			fmt.Printf("AssemblyAI CLI %s\n", VERSION)
		} else {
			cmd.Help()
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Check current installed version.")
}

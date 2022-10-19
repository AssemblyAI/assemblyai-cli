/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Hidden: true,
	Use:    "validate",
	Short:  "Validate your token",
	Long: `Seamlessly validate your AssemblyAI token.`,
	Run: func(cmd *cobra.Command, args []string) {
		Token = GetStoredToken()
		if Token != "" {
			fmt.Printf("Your Token is %s\n", Token)
			return
		} else {
			fmt.Println("Please start by running \033[1m\033[34massemblyai config [token]\033[0m")
		}
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

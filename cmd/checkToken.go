/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// checkTokenCmd represents the checkToken command
var checkTokenCmd = &cobra.Command{
	Hidden: true,
	Use:    "checkToken",
	Short:  "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		token := GetStoredToken()
		if token != "" {
			fmt.Printf("Your Token is %s\n", token)
			return
		} else {
			fmt.Println("You must login first. Run `assemblyai config <token>`")
		}
	},
}

func init() {
	rootCmd.AddCommand(checkTokenCmd)
}

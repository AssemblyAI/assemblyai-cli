/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deleteTokenCmd represents the deleteToken command
var deleteTokenCmd = &cobra.Command{
	Use:   "deleteToken",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		db := GetOpenDatabase()
		DeleteToken(db)
		fmt.Println("Token deleted")
	},
}

func init() {
	rootCmd.AddCommand(deleteTokenCmd)
}

/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
		and usage of using your command. For example:

		Cobra is a CLI library for Go that empowers applications.
		This application is a tool to generate the needed files
		to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		token := os.Args[len(os.Args) - 1]

		checkToken := checkIfTokenValid(token)

		if checkToken.IsVerified {
			db :=  GetOpenDatabase()
			txn := db.NewTransaction(true)
			err := txn.Set([]byte(AAITokenEnvName), []byte(token))
			PrintError(err)
			
			defer db.Close()
			defer txn.Discard()

			if err := txn.Commit(); err != nil {
				fmt.Println(err)
			}
			fmt.Printf("You're now authenticated. Your current balance is $%s\n", checkToken.CurrentBalance)

		} else {
			fmt.Println("Invalid token. Try again, and if the problem persists, contact support at support@assemblyai.com")
		}

	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
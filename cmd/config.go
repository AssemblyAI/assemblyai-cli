/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"
	"os"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/spf13/cobra"
)

var AAITokenEnvName = "ASSMEBLYAI_TOKEN"
var AAIURL = "https://api.assemblyai.com/v2"

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
		if checkToken(token) {
			db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
			if err != nil {
				fmt.Println(err)
			}
			defer db.Close()
			txn := db.NewTransaction(true)
			defer txn.Discard()

			err = txn.Set([]byte(AAITokenEnvName), []byte(token))
			if err != nil {
				fmt.Println(err)
			}

			if err := txn.Commit(); err != nil {
				fmt.Println(err)
			}
			fmt.Println("You're now authenticated.")
		} else {
			fmt.Println("Something went wrong. Please try again.")
		}

	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func checkToken(token string) bool {
	resp, err := http.NewRequest("GET", AAIURL + "/account", nil)
	if err != nil {
		fmt.Println(err)
	}
	resp.Header.Add("Accept", "application/json")
	resp.Header.Add("Authorization", token)
	
	response, err := http.DefaultClient.Do(resp)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
    return false
	}
	return true
}

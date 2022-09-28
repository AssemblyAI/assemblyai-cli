/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cli",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
		examples and usage of using your application. For example:

		Cobra is a CLI library for Go that empowers applications.
		This application is a tool to generate the needed files
		to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("help section")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

var AAITokenEnvName = "ASSMEBLYAI_TOKEN"
var AAIURL = "https://api.assemblyai.com/v2"

func GetDatabaseConfig() badger.Options {
	badgerCfg := badger.DefaultOptions("/tmp/badger")
	badgerCfg.Logger = nil
	return badgerCfg
}

func PrintError(err error) {
	if err != nil {
		// fmt.Println(err)
	}
}

func GetOpenDatabase() *badger.DB {
	badgerOptions := GetDatabaseConfig()
	db, err := badger.Open(badgerOptions)
	PrintError(err)
	return db
}

func GetStoredToken(db *badger.DB) string {
	var valCopy []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(AAITokenEnvName))
		PrintError(err)

		if err != nil {
			fmt.Println("You need to run `assemblyai config` first")
			return nil
		}

		err = item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)
			return nil
		})
		return nil
	})
	PrintError(err)

	return string(valCopy)
}

func QueryApi(token string, path string, method string, body io.Reader) []byte {
	resp, err := http.NewRequest(method, AAIURL + path, body)
	PrintError(err)

	resp.Header.Add("Accept", "application/json")
	resp.Header.Add("Authorization", token)
	
	response, err := http.DefaultClient.Do(resp)
	PrintError(err)
	defer response.Body.Close()
	
	responseData, err := ioutil.ReadAll(response.Body)
	PrintError(err)

	return responseData
}

func checkIfTokenValid(token string) CheckIfTokenValidResponse {
	var funcResponse CheckIfTokenValidResponse

	response := QueryApi(token, "/account", "GET", nil)

	if response == nil {
		return funcResponse
	}

	var result Account
	if err := json.Unmarshal(response, &result); err != nil {
			fmt.Println("Can not unmarshal JSON")
	}
	funcResponse.IsVerified = result.IsVerified
	funcResponse.CurrentBalance = fmt.Sprintf("%f", result.CurrentBalance.Amount)
	
	return funcResponse
}

// delete token
func DeleteToken(db *badger.DB) {
	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(AAITokenEnvName))
		return err
	})
	PrintError(err)
}

type CheckIfTokenValidResponse struct {
	IsVerified     bool   `json:"is_verified"`
	CurrentBalance string `json:"current_balance"`
}

type Account struct {
	IsVerified     bool           `json:"is_verified"`    
	CurrentBalance CurrentBalance `json:"current_balance"`
}

type CurrentBalance struct {
	Amount   float64 `json:"amount"`  
	Currency string  `json:"currency"`
}

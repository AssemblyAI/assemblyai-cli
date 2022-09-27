/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"
)

// randomCmd represents the random command
var randomCmd = &cobra.Command{
	Use:   "random",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		getRandomJoke()
	},
}

func init() {
	rootCmd.AddCommand(randomCmd)
}

type Joke struct {
	ID     int    `json:"id"`
	Joke  string `json:"joke"`
	Status int    `json:"status"`
}

func getRandomJoke() {
	baseAPI := "https://icanhazdadjoke.com/"
	responseBytes := getJokeData(baseAPI)
	var joke Joke
	json.Unmarshal(responseBytes, &joke)
	fmt.Println(joke.Joke)
}

func getJokeData(baseAPI string) []byte {
	resp, err := http.NewRequest("GET", baseAPI, nil)
	if err != nil {
		fmt.Println(err)
	}
	resp.Header.Add("Accept", "application/json")
	resp.Header.Add("User-Agent", "curl/7.64.1")
	
	response, err := http.DefaultClient.Do(resp)
	if err != nil {
		fmt.Println(err)
	}

	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}
	return responseBytes

}
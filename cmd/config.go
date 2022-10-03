/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config <token>",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
		and usage of using your command. For example:

		Cobra is a CLI library for Go that empowers applications.
		This application is a tool to generate the needed files
		to quickly create a Cobra application.`,
	Example:       "assemblyai config <token>",
	SilenceUsage:  true,
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		argsArray := cmd.Flags().Args()

		if len(argsArray) == 0 {
			fmt.Println("Please provide a token. You can get one at https://app.assemblyai.com")
			return
		} else if len(argsArray) > 1 {
			fmt.Println("Too many arguments. Please provide a single token.")
			return
		}
		token := argsArray[0]

		checkToken := CheckIfTokenValid(token)

		if !checkToken.IsVerified {
			fmt.Println("Invalid token. Try again, and if the problem persists, contact support at support@assemblyai.com")
			return
		}
		Config(token)

		fmt.Println("You're now authenticated.")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func Config(token string) {
	db := GetOpenDatabase()
	txn := db.NewTransaction(true)
	err := txn.Set([]byte(AAITokenEnvName), []byte(token))
	PrintError(err)

	defer db.Close()
	defer txn.Discard()

	if err := txn.Commit(); err != nil {
		fmt.Println(err)
	}
}

/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/spf13/cobra"
)

// transcribeCmd represents the transcribe command
var transcribeCmd = &cobra.Command{
	Use:   "transcribe",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
		if err != nil {
			fmt.Println(err)
		}
		defer db.Close()

		err = db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(AAITokenEnvName))
			if err != nil {
				fmt.Println(err)
			}
			var valCopy []byte
			err = item.Value(func(val []byte) error {
				valCopy = append([]byte{}, val...)
				return nil
			})
			valCopy, err = item.ValueCopy(nil)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("The answer is: %s\n", valCopy)
		
			return nil
		})

    fmt.Printf("Database Password: ")
		
	},
}

func init() {
	rootCmd.AddCommand(transcribeCmd)
}

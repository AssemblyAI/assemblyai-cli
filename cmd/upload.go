/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:    "upload",
	Hidden: true,
	Short:  "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		db := GetOpenDatabase()
		token := GetStoredToken(db)
		defer db.Close()

		args = cmd.Flags().Args()
		if len(args) == 0 {
			fmt.Println("You must provide an audio URL")
			return
		}
		path := args[0]

		toPrint := UploadFile(token, path)
		fmt.Println(toPrint)
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
}

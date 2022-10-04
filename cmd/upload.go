/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

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
		token := GetStoredToken()
		if token == "" {
			fmt.Println("You must login first. Run `assemblyai config <token>`")
			return
		}

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

func UploadFile(token string, path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}

	TelemetryCaptureEvent("CLI upload started", map[string]interface{}{})
	s := CallSpinner(" Your file is being uploaded...")
	response := QueryApi(token, "/upload", "POST", file)
	var uploadResponse UploadResponse
	if err := json.Unmarshal(response, &uploadResponse); err != nil {
		return ""
	}
	s.Stop()

	TelemetryCaptureEvent("CLI upload ended", map[string]interface{}{})
	return uploadResponse.UploadURL
}

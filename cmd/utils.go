/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/posthog/posthog-go"
)

var AAITokenEnvName = "ASSMEBLYAI_TOKEN"
var AAIURL = "https://api.assemblyai.com/v2"

func TelemetryCaptureEvent(event string, properties map[string]interface{}) {
	isTelemetryEnabled := getConfigFileValue("features.telemetry")
	if isTelemetryEnabled == "true" {
		godotenv.Load()
		PH_TOKEN := os.Getenv("POSTHOG_API_TOKEN")

		client := posthog.New(PH_TOKEN)
		defer client.Close()

		distinctId := getConfigFileValue("config.distinct_id")

		if distinctId == "" {
			distinctId = uuid.New().String()
			setConfigFileValue("config.distinct_id", distinctId)
		}

		client.Enqueue(posthog.Capture{
			DistinctId: distinctId,
			Event:      event,
			Properties: properties,
		})
	}
}

func CallSpinner(message string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[7], 100*time.Millisecond, spinner.WithSuffix(message))
	s.Start()
	return s
}

func PrintError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func QueryApi(token string, path string, method string, body io.Reader) []byte {
	resp, err := http.NewRequest(method, AAIURL+path, body)
	PrintError(err)

	resp.Header.Add("Accept", "application/json")
	resp.Header.Add("Authorization", token)
	resp.Header.Add("Transfer-Encoding", "chunked")

	response, err := http.DefaultClient.Do(resp)
	PrintError(err)

	responseData, err := ioutil.ReadAll(response.Body)
	PrintError(err)
	defer response.Body.Close()

	return responseData
}

func BeutifyJSON(data []byte) []byte {
	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, data, "", "\t")
	if error != nil {
		return data
	}
	return prettyJSON.Bytes()
}

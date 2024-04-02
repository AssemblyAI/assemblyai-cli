/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package utils

import (
	"errors"
	"io"
	"os"

	S "github.com/AssemblyAI/assemblyai-cli/schemas"
	"github.com/kkdai/youtube/v2"
)

var Filename = os.TempDir() + "tmp-video."

func YoutubeDownload(id string) string {
	client := youtube.Client{}

	video, err := client.GetVideo(id)
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Failed to fetch YouTube video metadata.",
		}
		PrintError(printErrorProps)
	}

	formats := video.Formats.WithAudioChannels() // only get videos with audio
	stream, _, err := client.GetStream(video, &formats[0])
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Failed to get YouTube video audio channels.",
		}
		PrintError(printErrorProps)
	}
	defer stream.Close()

	file, err := os.Create(Filename)
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
	}
	defer file.Close()

	_, err = io.Copy(file, stream)
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
	}
	uploadedURL := UploadFile(Filename)
	if uploadedURL == "" {
		printErrorProps := S.PrintErrorProps{
			Error:   errors.New("The file does not exist. Please try again with a different one."),
			Message: "The file does not exist. Please try again with a different one.",
		}
		PrintError(printErrorProps)
	}
	err = os.Remove(Filename)
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
	}
	return uploadedURL
}

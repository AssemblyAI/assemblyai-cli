/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	S "github.com/AssemblyAI/assemblyai-cli/schemas"
	pb "gopkg.in/cheggaaa/pb.v1"
)

var key = "AIzaSyA8eiZmM1FaDVjRy-df2KTyQ_vz_yYM39w"
var Filename = os.TempDir() + "tmp-video.mp4"
var fileLength = 0
var percent = 0
var chunkSize = 8000000

func YoutubeDownload(id string) string {
	var body S.YoutubeBodyMetaInfo
	body.Context.Client.Hl = "en"
	body.Context.Client.ClientName = "ANDROID"
	body.Context.Client.ClientVersion = "17.31.35"
	body.Context.Client.AndroidSDKVersion = 30
	body.Context.Client.UserAgent = "com.google.android.youtube/17.31.35 (Linux; U; Android 11) gzip"
	body.Context.Client.TimeZone = "UTC"
	body.Context.Client.UtcOffsetMinutes = 0
	body.VideoID = id
	body.Params = "8AEB"
	body.PlaybackContext.ContentPlaybackContext.Html5Preference = "HTML5_PREF_WANTS"
	body.RacyCheckOk = true
	body.ContentCheckOk = true

	paramsJSON, err := json.Marshal(body)
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
		return ""
	}

	requestBody := bytes.NewReader(paramsJSON)
	fmt.Println("Transcribing Youtube video...")
	video := QueryYoutube(requestBody)
	if *video.PlayabilityStatus.Status != "OK" || video.StreamingData.Formats == nil {
		printErrorProps := S.PrintErrorProps{
			Error:   errors.New("Video is not available"),
			Message: "The video is not available for download.",
		}
		PrintError(printErrorProps)
		return ""
	}
	var idx int
	var lowestContentLength int
	for i, format := range video.StreamingData.Formats {
		if format.ContentLength != nil {
			length, _ := strconv.Atoi(*format.ContentLength)
			if i == 0 {
				fileLength = length
				lowestContentLength = length
				idx = i
			} else if length < lowestContentLength {
				lowestContentLength = length
				fileLength = length
				idx = i
			}
		}
	}
	if fileLength == 0 {
		for i, format := range video.StreamingData.Formats {
			length := int(*format.Bitrate)
			if i == 0 {
				lowestContentLength = length
				idx = i
			} else if length < lowestContentLength {
				lowestContentLength = length
				idx = i
			}
		}
	}
	videoUrl := ""
	if video.StreamingData.Formats[idx].URL != nil {
		videoUrl = *video.StreamingData.Formats[idx].URL
	} else {
		split := strings.Split(*video.StreamingData.Formats[idx].SignatureCipher, "&")
		youtubeUrl := ""
		for _, value := range split {
			if strings.HasPrefix(value, "url=") {
				youtubeUrl = strings.TrimPrefix(value, "url=")
				videoUrl, err = url.QueryUnescape(youtubeUrl)
				break
			}
		}
	}

	info, err := os.Stat(os.TempDir())
	if err != nil || !info.IsDir() || info.Mode().Perm()&(1<<uint(7)) == 0 {
		Filename = "./tmp-video.mp4"
		local, err := os.Stat("./")
		if err != nil || !local.IsDir() || local.Mode().Perm()&(1<<uint(7)) == 0 {
			err = os.Chmod("./", 0700)
			if err != nil {
				fmt.Println("Unable to create temporary file")
				return ""
			}
		}
	}

	DownloadVideo(videoUrl)
	uploadedURL := UploadFile(Filename)
	if uploadedURL == "" {
		printErrorProps := S.PrintErrorProps{
			Error:   errors.New("The file does not exist. Please try again with a different one."),
			Message: "The file does not exist. Please try again with a different one.",
		}
		PrintError(printErrorProps)
	}
	err = os.Remove(Filename)
	return uploadedURL
}

func QueryYoutube(body io.Reader) S.YoutubeMetaInfo {
	resp, err := http.NewRequest("POST", "https://www.youtube.com/youtubei/v1/player?key="+key, body)
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
	}

	resp.Header.Add("Accept", "application/json")

	response, err := http.DefaultClient.Do(resp)
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
	}
	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
		fmt.Println("Our Youtube transcribe service is currently unavailable. Please try again later.")
	}

	var videoResponse S.YoutubeMetaInfo
	if err := json.Unmarshal(responseData, &videoResponse); err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
	}

	return videoResponse
}

func DownloadVideo(url string) {
	resp, err := http.Head(url)
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
	}
	fileLength, err = strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
	}

	file, err := os.Create(Filename)
	if err != nil {
		printErrorProps := S.PrintErrorProps{
			Error:   err,
			Message: "Something went wrong. Please try again.",
		}
		PrintError(printErrorProps)
	}
	defer file.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	if fileLength > chunkSize {
		bar := pb.New(fileLength)
		bar.Prefix("Downloading video: ")
		bar.SetUnits(pb.U_BYTES_DEC)
		bar.ShowBar = false
		bar.ShowTimeLeft = false
		bar.Start()
		chunks := int(math.Ceil(float64(fileLength) / float64(chunkSize)))
		for i := 0; i < chunks; i++ {
			bar.Set(i * chunkSize)
			start := i * chunkSize
			end := start + chunkSize - 1
			if end > fileLength {
				end = fileLength
			}
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
			resp, err := client.Do(req)
			if err != nil {
				printErrorProps := S.PrintErrorProps{
					Error:   err,
					Message: "Something went wrong. Please try again.",
				}
				PrintError(printErrorProps)
			}
			defer resp.Body.Close()
			_, err = io.Copy(file, resp.Body)
			if err != nil {
				printErrorProps := S.PrintErrorProps{
					Error:   err,
					Message: "Something went wrong. Please try again.",
				}
				PrintError(printErrorProps)
			}
		}
		bar.Set(fileLength)
		bar.Finish()
	} else {
		req.Header.Set("Range", fmt.Sprintf("Bytes=0-%d", fileLength))
		resp, err = client.Do(req)
		if err != nil {
			printErrorProps := S.PrintErrorProps{
				Error:   err,
				Message: "Something went wrong. Please try again.",
			}
			PrintError(printErrorProps)
		}
		defer resp.Body.Close()
		go displayDownloadProgress()
		body := io.TeeReader(resp.Body, &writeCounter{0, int64(fileLength)})
		_, err = io.Copy(file, body)
		if err != nil {
			printErrorProps := S.PrintErrorProps{
				Error:   err,
				Message: "Something went wrong. Please try again.",
			}
			PrintError(printErrorProps)
		}
	}

}

func (pWc *writeCounter) Write(b []byte) (n int, err error) {
	n = len(b)
	pWc.BytesDownloaded += int64(n)
	percent = int(math.Round(float64(pWc.BytesDownloaded) * 100.0 / float64(pWc.TotalBytes)))
	return
}

func displayDownloadProgress() {
	bar := pb.New(fileLength)
	bar.Prefix("Downloading video: ")
	bar.SetUnits(pb.U_BYTES_DEC)
	bar.ShowBar = false
	bar.ShowTimeLeft = false
	bar.Start()
	for percent < 100 {
		bar.Set(percent * fileLength / 100)
	}
	bar.Set(fileLength)
	bar.Finish()
}

type writeCounter struct {
	BytesDownloaded int64
	TotalBytes      int64
}

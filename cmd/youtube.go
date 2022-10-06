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
	"strconv"
	"time"
)

var key = "AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"
var Filename = "tmp-video.mp4"

func YoutubeDownload(id string) bool {
	var body YoutubeBodyMetaInfo
	body.Context.Client.Hl = "en"
	body.Context.Client.ClientName = "WEB"
	body.Context.Client.ClientVersion = "2.20210721.08.00"
	body.Context.Client.ClientFormFactor = "UNKNOWN_FORM_FACTOR"
	body.Context.Client.ClientScreen = "WATCH"
	body.Context.Client.MainAppWebInfo.GraftURL = "/watch?v=" + id
	body.Context.User.LockedSafetyMode = false
	body.Context.Request.UseSSL = true
	body.Context.Request.InternalExperimentFlags = nil
	body.Context.Request.ConsistencyTokenJars = nil
	body.VideoID = id
	body.PlaybackContext.ContentPlaybackContext.Vis = 0
	body.PlaybackContext.ContentPlaybackContext.Splay = false
	body.PlaybackContext.ContentPlaybackContext.AutoCaptionsDefaultOn = false
	body.PlaybackContext.ContentPlaybackContext.AutonavState = "STATE_NONE"
	body.PlaybackContext.ContentPlaybackContext.Html5Preference = "HTML5_PREF_WANTS"
	body.PlaybackContext.ContentPlaybackContext.LactMilliseconds = "-1"
	body.RacyCheckOk = false
	body.ContentCheckOk = false

	paramsJSON, err := json.Marshal(body)
	if err != nil {
		return false
	}

	requestBody := bytes.NewReader(paramsJSON)

	s := CallSpinner(" The video is being downloaded...")
	video := QueryYoutube(requestBody)
	if *video.PlayabilityStatus.Status != "OK" {
		s.Stop()
		fmt.Println("The video is not available for download")
		return false
	}

	var idx int
	var itag int64
	for index, format := range video.StreamingData.Formats {
		if *format.MIMEType == "video/mp4; codecs=\"avc1.42001E, mp4a.40.2\"" {
			idx = index
			itag = *format.Itag
			break
		}
	}
	status := download(idx, itag, video)
	s.Stop()
	return status
}

func QueryYoutube(body io.Reader) YoutubeMetaInfo {
	resp, err := http.NewRequest("POST", "https://www.youtube.com/youtubei/v1/player?key="+key, body)
	PrintError(err)

	resp.Header.Add("Accept", "application/json")

	response, err := http.DefaultClient.Do(resp)
	PrintError(err)
	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	PrintError(err)

	var videoResponse YoutubeMetaInfo
	if err := json.Unmarshal(responseData, &videoResponse); err != nil {
		PrintError(err)
	}

	return videoResponse
}

func download(index int, itag int64, video YoutubeMetaInfo) bool {
	var (
		out    *os.File
		err    error
		offset int64
		length int64
	)

	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	out, err = os.OpenFile(Filename, flags, 0644)
	if err != nil {
		return false
	}

	defer out.Close()

	url := video.StreamingData.Formats[index].URL

	if resp, err := http.Head(*url); err != nil {
		return false
	} else {
		if resp.StatusCode == 403 {
			return false
		}

		if size := resp.Header.Get("Content-Length"); len(size) == 0 {
			return false
		} else if length, err = strconv.ParseInt(size, 10, 64); err != nil {
			PrintError(err)
			return false
		}

		if length <= offset {
			return false
		}
	}

	start := time.Now()
	resp, err := http.Get(*url)
	if err != nil {
		PrintError(err)
		return false
	}
	defer resp.Body.Close()

	if length, err = io.Copy(out, resp.Body); err != nil {
		PrintError(err)
		return false
	}

	duration := time.Now().Sub(start)
	if duration > time.Second {
		duration -= duration % time.Second
	}

	if err := out.Close(); err != nil {
		PrintError(err)
		return false
	}
	// TODO : transcode to mp3
	// ffmpeg, err := exec.LookPath("ffmpeg")
	// if err != nil {
	// 	fmt.Println("ffmpeg not found")
	// } else {
	// 	fmt.Println("Extracting audio ..")
	// 	mp3 := strings.TrimRight(filename, filepath.Ext(filename)) + ".mp3"
	// 	cmd := exec.Command(ffmpeg, "-y", "-loglevel", "quiet", "-i", filename, "-vn", mp3)
	// 	cmd.Stdin = os.Stdin
	// 	cmd.Stdout = os.Stdout
	// 	cmd.Stderr = os.Stderr
	// 	if err := cmd.Run(); err != nil {
	// 		fmt.Println("Failed to extract audio:", err)
	// 	} else {
	// 		fmt.Println()
	// 		fmt.Println("Extracted audio:", mp3)
	// 	}
	// }

	return true
}

const (
	KB float64 = 1 << (10 * (iota + 1))
	MB
	GB
)

type YoutubeBodyMetaInfo struct {
	Context         Context         `json:"context"`
	VideoID         string          `json:"videoId"`
	PlaybackContext PlaybackContext `json:"playbackContext"`
	RacyCheckOk     bool            `json:"racyCheckOk"`
	ContentCheckOk  bool            `json:"contentCheckOk"`
}

type Context struct {
	Client  Client  `json:"client"`
	User    User    `json:"user"`
	Request Request `json:"request"`
}

type Client struct {
	Hl               string         `json:"hl"`
	ClientName       string         `json:"clientName"`
	ClientVersion    string         `json:"clientVersion"`
	ClientFormFactor string         `json:"clientFormFactor"`
	ClientScreen     string         `json:"clientScreen"`
	MainAppWebInfo   MainAppWebInfo `json:"mainAppWebInfo"`
}

type MainAppWebInfo struct {
	GraftURL string `json:"graftUrl"`
}

type Request struct {
	UseSSL                  bool          `json:"useSsl"`
	InternalExperimentFlags []interface{} `json:"internalExperimentFlags"`
	ConsistencyTokenJars    []interface{} `json:"consistencyTokenJars"`
}

type User struct {
	LockedSafetyMode bool `json:"lockedSafetyMode"`
}

type PlaybackContext struct {
	ContentPlaybackContext ContentPlaybackContext `json:"contentPlaybackContext"`
}

type ContentPlaybackContext struct {
	Vis                   int64  `json:"vis"`
	Splay                 bool   `json:"splay"`
	AutoCaptionsDefaultOn bool   `json:"autoCaptionsDefaultOn"`
	AutonavState          string `json:"autonavState"`
	Html5Preference       string `json:"html5Preference"`
	LactMilliseconds      string `json:"lactMilliseconds"`
}

type YoutubeMetaInfo struct {
	ResponseContext   *ResponseContext   `json:"responseContext,omitempty"`
	PlayabilityStatus *PlayabilityStatus `json:"playabilityStatus,omitempty"`
	StreamingData     *StreamingData     `json:"streamingData,omitempty"`
	PlaybackTracking  *PlaybackTracking  `json:"playbackTracking,omitempty"`
	Captions          *Captions          `json:"captions,omitempty"`
	VideoDetails      *VideoDetails      `json:"videoDetails,omitempty"`
	PlayerConfig      *PlayerConfig      `json:"playerConfig,omitempty"`
	Storyboards       *Storyboards       `json:"storyboards,omitempty"`
	Microformat       *Microformat       `json:"microformat,omitempty"`
	TrackingParams    *string            `json:"trackingParams,omitempty"`
	Attestation       *Attestation       `json:"attestation,omitempty"`
	Endscreen         *Endscreen         `json:"endscreen,omitempty"`
	FrameworkUpdates  *FrameworkUpdates  `json:"frameworkUpdates,omitempty"`
}

type Attestation struct {
	PlayerAttestationRenderer *PlayerAttestationRenderer `json:"playerAttestationRenderer,omitempty"`
}

type PlayerAttestationRenderer struct {
	Challenge    *string       `json:"challenge,omitempty"`
	BotguardData *BotguardData `json:"botguardData,omitempty"`
}

type BotguardData struct {
	Program            *string             `json:"program,omitempty"`
	InterpreterSafeURL *InterpreterSafeURL `json:"interpreterSafeUrl,omitempty"`
	ServerEnvironment  *int64              `json:"serverEnvironment,omitempty"`
}

type InterpreterSafeURL struct {
	PrivateDoNotAccessOrElseTrustedResourceURLWrappedValue *string `json:"privateDoNotAccessOrElseTrustedResourceUrlWrappedValue,omitempty"`
}

type Captions struct {
	PlayerCaptionsTracklistRenderer *PlayerCaptionsTracklistRenderer `json:"playerCaptionsTracklistRenderer,omitempty"`
}

type PlayerCaptionsTracklistRenderer struct {
	CaptionTracks          []CaptionTrack        `json:"captionTracks,omitempty"`
	AudioTracks            []AudioTrack          `json:"audioTracks,omitempty"`
	TranslationLanguages   []TranslationLanguage `json:"translationLanguages,omitempty"`
	DefaultAudioTrackIndex *int64                `json:"defaultAudioTrackIndex,omitempty"`
}

type AudioTrack struct {
	CaptionTrackIndices []int64 `json:"captionTrackIndices,omitempty"`
}

type CaptionTrack struct {
	BaseURL        *string      `json:"baseUrl,omitempty"`
	Name           *Description `json:"name,omitempty"`
	VssID          *string      `json:"vssId,omitempty"`
	LanguageCode   *string      `json:"languageCode,omitempty"`
	Kind           *string      `json:"kind,omitempty"`
	IsTranslatable *bool        `json:"isTranslatable,omitempty"`
}

type Description struct {
	SimpleText *string `json:"simpleText,omitempty"`
}

type TranslationLanguage struct {
	LanguageCode *string      `json:"languageCode,omitempty"`
	LanguageName *Description `json:"languageName,omitempty"`
}

type Endscreen struct {
	EndscreenRenderer *EndscreenRenderer `json:"endscreenRenderer,omitempty"`
}

type EndscreenRenderer struct {
	Elements       []Element `json:"elements,omitempty"`
	StartMS        *string   `json:"startMs,omitempty"`
	TrackingParams *string   `json:"trackingParams,omitempty"`
}

type Element struct {
	EndscreenElementRenderer *EndscreenElementRenderer `json:"endscreenElementRenderer,omitempty"`
}

type EndscreenElementRenderer struct {
	Style             *string            `json:"style,omitempty"`
	Image             *ImageClass        `json:"image,omitempty"`
	Icon              *Icon              `json:"icon,omitempty"`
	Left              *float64           `json:"left,omitempty"`
	Width             *float64           `json:"width,omitempty"`
	Top               *float64           `json:"top,omitempty"`
	AspectRatio       *float64           `json:"aspectRatio,omitempty"`
	StartMS           *string            `json:"startMs,omitempty"`
	EndMS             *string            `json:"endMs,omitempty"`
	Title             *Metadata          `json:"title,omitempty"`
	Metadata          *Metadata          `json:"metadata,omitempty"`
	CallToAction      *Description       `json:"callToAction,omitempty"`
	Dismiss           *Description       `json:"dismiss,omitempty"`
	Endpoint          *Endpoint          `json:"endpoint,omitempty"`
	HovercardButton   *HovercardButton   `json:"hovercardButton,omitempty"`
	TrackingParams    *string            `json:"trackingParams,omitempty"`
	IsSubscribe       *bool              `json:"isSubscribe,omitempty"`
	ID                *string            `json:"id,omitempty"`
	ThumbnailOverlays []ThumbnailOverlay `json:"thumbnailOverlays,omitempty"`
}

type Endpoint struct {
	ClickTrackingParams *string                  `json:"clickTrackingParams,omitempty"`
	CommandMetadata     *EndpointCommandMetadata `json:"commandMetadata,omitempty"`
	BrowseEndpoint      *BrowseEndpoint          `json:"browseEndpoint,omitempty"`
	WatchEndpoint       *WatchEndpoint           `json:"watchEndpoint,omitempty"`
	URLEndpoint         *URLEndpoint             `json:"urlEndpoint,omitempty"`
}

type BrowseEndpoint struct {
	BrowseID *string `json:"browseId,omitempty"`
}

type EndpointCommandMetadata struct {
	WebCommandMetadata *PurpleWebCommandMetadata `json:"webCommandMetadata,omitempty"`
}

type PurpleWebCommandMetadata struct {
	URL         *string `json:"url,omitempty"`
	WebPageType *string `json:"webPageType,omitempty"`
	RootVe      *int64  `json:"rootVe,omitempty"`
	APIURL      *string `json:"apiUrl,omitempty"`
}

type URLEndpoint struct {
	URL    *string `json:"url,omitempty"`
	Target *string `json:"target,omitempty"`
}

type WatchEndpoint struct {
	VideoID                            *string                             `json:"videoId,omitempty"`
	WatchEndpointSupportedOnesieConfig *WatchEndpointSupportedOnesieConfig `json:"watchEndpointSupportedOnesieConfig,omitempty"`
}

type WatchEndpointSupportedOnesieConfig struct {
	Html5PlaybackOnesieConfig *Html5PlaybackOnesieConfig `json:"html5PlaybackOnesieConfig,omitempty"`
}

type Html5PlaybackOnesieConfig struct {
	CommonConfig *CommonConfigElement `json:"commonConfig,omitempty"`
}

type CommonConfigElement struct {
	URL *string `json:"url,omitempty"`
}

type HovercardButton struct {
	SubscribeButtonRenderer *SubscribeButtonRenderer `json:"subscribeButtonRenderer,omitempty"`
}

type SubscribeButtonRenderer struct {
	ButtonText               *ButtonText                  `json:"buttonText,omitempty"`
	Subscribed               *bool                        `json:"subscribed,omitempty"`
	Enabled                  *bool                        `json:"enabled,omitempty"`
	Type                     *string                      `json:"type,omitempty"`
	ChannelID                *string                      `json:"channelId,omitempty"`
	ShowPreferences          *bool                        `json:"showPreferences,omitempty"`
	SubscribedButtonText     *ButtonText                  `json:"subscribedButtonText,omitempty"`
	UnsubscribedButtonText   *ButtonText                  `json:"unsubscribedButtonText,omitempty"`
	TrackingParams           *string                      `json:"trackingParams,omitempty"`
	UnsubscribeButtonText    *ButtonText                  `json:"unsubscribeButtonText,omitempty"`
	ServiceEndpoints         []SubscribeCommand           `json:"serviceEndpoints,omitempty"`
	SubscribeAccessibility   *SubscribeAccessibilityClass `json:"subscribeAccessibility,omitempty"`
	UnsubscribeAccessibility *SubscribeAccessibilityClass `json:"unsubscribeAccessibility,omitempty"`
	SignInEndpoint           *SignInEndpoint              `json:"signInEndpoint,omitempty"`
}

type ButtonText struct {
	Runs []Run `json:"runs,omitempty"`
}

type Run struct {
	Text *string `json:"text,omitempty"`
}

type SubscribeCommand struct {
	ClickTrackingParams   *string                          `json:"clickTrackingParams,omitempty"`
	CommandMetadata       *SubscribeCommandCommandMetadata `json:"commandMetadata,omitempty"`
	SubscribeEndpoint     *SubscribeEndpoint               `json:"subscribeEndpoint,omitempty"`
	SignalServiceEndpoint *SignalServiceEndpoint           `json:"signalServiceEndpoint,omitempty"`
}

type SubscribeCommandCommandMetadata struct {
	WebCommandMetadata *FluffyWebCommandMetadata `json:"webCommandMetadata,omitempty"`
}

type FluffyWebCommandMetadata struct {
	SendPost *bool   `json:"sendPost,omitempty"`
	APIURL   *string `json:"apiUrl,omitempty"`
}

type SignalServiceEndpoint struct {
	Signal  *string                       `json:"signal,omitempty"`
	Actions []SignalServiceEndpointAction `json:"actions,omitempty"`
}

type SignalServiceEndpointAction struct {
	ClickTrackingParams *string          `json:"clickTrackingParams,omitempty"`
	OpenPopupAction     *OpenPopupAction `json:"openPopupAction,omitempty"`
}

type OpenPopupAction struct {
	Popup     *Popup  `json:"popup,omitempty"`
	PopupType *string `json:"popupType,omitempty"`
}

type Popup struct {
	ConfirmDialogRenderer *ConfirmDialogRenderer `json:"confirmDialogRenderer,omitempty"`
}

type ConfirmDialogRenderer struct {
	TrackingParams  *string      `json:"trackingParams,omitempty"`
	DialogMessages  []ButtonText `json:"dialogMessages,omitempty"`
	ConfirmButton   *Button      `json:"confirmButton,omitempty"`
	CancelButton    *Button      `json:"cancelButton,omitempty"`
	PrimaryIsCancel *bool        `json:"primaryIsCancel,omitempty"`
}

type Button struct {
	ButtonRenderer *ButtonRenderer `json:"buttonRenderer,omitempty"`
}

type ButtonRenderer struct {
	Style           *string                 `json:"style,omitempty"`
	Size            *string                 `json:"size,omitempty"`
	IsDisabled      *bool                   `json:"isDisabled,omitempty"`
	Text            *ButtonText             `json:"text,omitempty"`
	Accessibility   *AccessibilityDataClass `json:"accessibility,omitempty"`
	TrackingParams  *string                 `json:"trackingParams,omitempty"`
	ServiceEndpoint *UnsubscribeCommand     `json:"serviceEndpoint,omitempty"`
}

type AccessibilityDataClass struct {
	Label *string `json:"label,omitempty"`
}

type UnsubscribeCommand struct {
	ClickTrackingParams *string                          `json:"clickTrackingParams,omitempty"`
	CommandMetadata     *SubscribeCommandCommandMetadata `json:"commandMetadata,omitempty"`
	UnsubscribeEndpoint *SubscribeEndpoint               `json:"unsubscribeEndpoint,omitempty"`
}

type SubscribeEndpoint struct {
	ChannelIDS []string `json:"channelIds,omitempty"`
	Params     *string  `json:"params,omitempty"`
}

type SignInEndpoint struct {
	ClickTrackingParams *string                        `json:"clickTrackingParams,omitempty"`
	CommandMetadata     *SignInEndpointCommandMetadata `json:"commandMetadata,omitempty"`
}

type SignInEndpointCommandMetadata struct {
	WebCommandMetadata *CommonConfigElement `json:"webCommandMetadata,omitempty"`
}

type SubscribeAccessibilityClass struct {
	AccessibilityData *AccessibilityDataClass `json:"accessibilityData,omitempty"`
}

type Icon struct {
	Thumbnails []CommonConfigElement `json:"thumbnails,omitempty"`
}

type ImageClass struct {
	Thumbnails []ThumbnailThumbnail `json:"thumbnails,omitempty"`
}

type ThumbnailThumbnail struct {
	URL    *string `json:"url,omitempty"`
	Width  *int64  `json:"width,omitempty"`
	Height *int64  `json:"height,omitempty"`
}

type Metadata struct {
	SimpleText    *string                      `json:"simpleText,omitempty"`
	Accessibility *SubscribeAccessibilityClass `json:"accessibility,omitempty"`
}

type ThumbnailOverlay struct {
	ThumbnailOverlayTimeStatusRenderer *ThumbnailOverlayTimeStatusRenderer `json:"thumbnailOverlayTimeStatusRenderer,omitempty"`
}

type ThumbnailOverlayTimeStatusRenderer struct {
	Text  *Metadata `json:"text,omitempty"`
	Style *string   `json:"style,omitempty"`
}

type FrameworkUpdates struct {
	EntityBatchUpdate *EntityBatchUpdate `json:"entityBatchUpdate,omitempty"`
}

type EntityBatchUpdate struct {
	Mutations []Mutation  `json:"mutations,omitempty"`
	Timestamp *TimestampR `json:"timestamp,omitempty"`
}

type Mutation struct {
	EntityKey *string  `json:"entityKey,omitempty"`
	Type      *string  `json:"type,omitempty"`
	Payload   *Payload `json:"payload,omitempty"`
}

type Payload struct {
	OfflineabilityEntity *OfflineabilityEntity `json:"offlineabilityEntity,omitempty"`
}

type OfflineabilityEntity struct {
	Key                     *string `json:"key,omitempty"`
	AddToOfflineButtonState *string `json:"addToOfflineButtonState,omitempty"`
}

type TimestampR struct {
	Seconds *string `json:"seconds,omitempty"`
	Nanos   *int64  `json:"nanos,omitempty"`
}

type Microformat struct {
	PlayerMicroformatRenderer *PlayerMicroformatRenderer `json:"playerMicroformatRenderer,omitempty"`
}

type PlayerMicroformatRenderer struct {
	Thumbnail          *ImageClass  `json:"thumbnail,omitempty"`
	Embed              *Embed       `json:"embed,omitempty"`
	Title              *Description `json:"title,omitempty"`
	Description        *Description `json:"description,omitempty"`
	LengthSeconds      *string      `json:"lengthSeconds,omitempty"`
	OwnerProfileURL    *string      `json:"ownerProfileUrl,omitempty"`
	ExternalChannelID  *string      `json:"externalChannelId,omitempty"`
	IsFamilySafe       *bool        `json:"isFamilySafe,omitempty"`
	AvailableCountries []string     `json:"availableCountries,omitempty"`
	IsUnlisted         *bool        `json:"isUnlisted,omitempty"`
	HasYpcMetadata     *bool        `json:"hasYpcMetadata,omitempty"`
	ViewCount          *string      `json:"viewCount,omitempty"`
	Category           *string      `json:"category,omitempty"`
	PublishDate        *string      `json:"publishDate,omitempty"`
	OwnerChannelName   *string      `json:"ownerChannelName,omitempty"`
	UploadDate         *string      `json:"uploadDate,omitempty"`
}

type Embed struct {
	IframeURL      *string `json:"iframeUrl,omitempty"`
	FlashURL       *string `json:"flashUrl,omitempty"`
	Width          *int64  `json:"width,omitempty"`
	Height         *int64  `json:"height,omitempty"`
	FlashSecureURL *string `json:"flashSecureUrl,omitempty"`
}

type PlayabilityStatus struct {
	Status          *string     `json:"status,omitempty"`
	PlayableInEmbed *bool       `json:"playableInEmbed,omitempty"`
	Miniplayer      *Miniplayer `json:"miniplayer,omitempty"`
	ContextParams   *string     `json:"contextParams,omitempty"`
}

type Miniplayer struct {
	MiniplayerRenderer *MiniplayerRenderer `json:"miniplayerRenderer,omitempty"`
}

type MiniplayerRenderer struct {
	PlaybackMode *string `json:"playbackMode,omitempty"`
}

type PlaybackTracking struct {
	VideostatsPlaybackURL                   *URL    `json:"videostatsPlaybackUrl,omitempty"`
	VideostatsDelayplayURL                  *URL    `json:"videostatsDelayplayUrl,omitempty"`
	VideostatsWatchtimeURL                  *URL    `json:"videostatsWatchtimeUrl,omitempty"`
	PtrackingURL                            *URL    `json:"ptrackingUrl,omitempty"`
	QoeURL                                  *URL    `json:"qoeUrl,omitempty"`
	AtrURL                                  *AtrURL `json:"atrUrl,omitempty"`
	VideostatsScheduledFlushWalltimeSeconds []int64 `json:"videostatsScheduledFlushWalltimeSeconds,omitempty"`
	VideostatsDefaultFlushIntervalSeconds   *int64  `json:"videostatsDefaultFlushIntervalSeconds,omitempty"`
}

type AtrURL struct {
	BaseURL                 *string `json:"baseUrl,omitempty"`
	ElapsedMediaTimeSeconds *int64  `json:"elapsedMediaTimeSeconds,omitempty"`
}

type URL struct {
	BaseURL *string `json:"baseUrl,omitempty"`
}

type PlayerConfig struct {
	AudioConfig           *AudioConfig           `json:"audioConfig,omitempty"`
	StreamSelectionConfig *StreamSelectionConfig `json:"streamSelectionConfig,omitempty"`
	MediaCommonConfig     *MediaCommonConfig     `json:"mediaCommonConfig,omitempty"`
	WebPlayerConfig       *WebPlayerConfig       `json:"webPlayerConfig,omitempty"`
}

type AudioConfig struct {
	LoudnessDB              *float64 `json:"loudnessDb,omitempty"`
	PerceptualLoudnessDB    *float64 `json:"perceptualLoudnessDb,omitempty"`
	EnablePerFormatLoudness *bool    `json:"enablePerFormatLoudness,omitempty"`
}

type MediaCommonConfig struct {
	DynamicReadaheadConfig *DynamicReadaheadConfig `json:"dynamicReadaheadConfig,omitempty"`
}

type DynamicReadaheadConfig struct {
	MaxReadAheadMediaTimeMS *int64 `json:"maxReadAheadMediaTimeMs,omitempty"`
	MinReadAheadMediaTimeMS *int64 `json:"minReadAheadMediaTimeMs,omitempty"`
	ReadAheadGrowthRateMS   *int64 `json:"readAheadGrowthRateMs,omitempty"`
}

type StreamSelectionConfig struct {
	MaxBitrate *string `json:"maxBitrate,omitempty"`
}

type WebPlayerConfig struct {
	WebPlayerActionsPorting *WebPlayerActionsPorting `json:"webPlayerActionsPorting,omitempty"`
}

type WebPlayerActionsPorting struct {
	GetSharePanelCommand        *GetSharePanelCommand        `json:"getSharePanelCommand,omitempty"`
	SubscribeCommand            *SubscribeCommand            `json:"subscribeCommand,omitempty"`
	UnsubscribeCommand          *UnsubscribeCommand          `json:"unsubscribeCommand,omitempty"`
	AddToWatchLaterCommand      *AddToWatchLaterCommand      `json:"addToWatchLaterCommand,omitempty"`
	RemoveFromWatchLaterCommand *RemoveFromWatchLaterCommand `json:"removeFromWatchLaterCommand,omitempty"`
}

type AddToWatchLaterCommand struct {
	ClickTrackingParams  *string                                     `json:"clickTrackingParams,omitempty"`
	CommandMetadata      *SubscribeCommandCommandMetadata            `json:"commandMetadata,omitempty"`
	PlaylistEditEndpoint *AddToWatchLaterCommandPlaylistEditEndpoint `json:"playlistEditEndpoint,omitempty"`
}

type AddToWatchLaterCommandPlaylistEditEndpoint struct {
	PlaylistID *string        `json:"playlistId,omitempty"`
	Actions    []PurpleAction `json:"actions,omitempty"`
}

type PurpleAction struct {
	AddedVideoID *string `json:"addedVideoId,omitempty"`
	Action       *string `json:"action,omitempty"`
}

type GetSharePanelCommand struct {
	ClickTrackingParams                 *string                              `json:"clickTrackingParams,omitempty"`
	CommandMetadata                     *SubscribeCommandCommandMetadata     `json:"commandMetadata,omitempty"`
	WebPlayerShareEntityServiceEndpoint *WebPlayerShareEntityServiceEndpoint `json:"webPlayerShareEntityServiceEndpoint,omitempty"`
}

type WebPlayerShareEntityServiceEndpoint struct {
	SerializedShareEntity *string `json:"serializedShareEntity,omitempty"`
}

type RemoveFromWatchLaterCommand struct {
	ClickTrackingParams  *string                                          `json:"clickTrackingParams,omitempty"`
	CommandMetadata      *SubscribeCommandCommandMetadata                 `json:"commandMetadata,omitempty"`
	PlaylistEditEndpoint *RemoveFromWatchLaterCommandPlaylistEditEndpoint `json:"playlistEditEndpoint,omitempty"`
}

type RemoveFromWatchLaterCommandPlaylistEditEndpoint struct {
	PlaylistID *string        `json:"playlistId,omitempty"`
	Actions    []FluffyAction `json:"actions,omitempty"`
}

type FluffyAction struct {
	Action         *string `json:"action,omitempty"`
	RemovedVideoID *string `json:"removedVideoId,omitempty"`
}

type ResponseContext struct {
	VisitorData                     *string                          `json:"visitorData,omitempty"`
	ServiceTrackingParams           []ServiceTrackingParam           `json:"serviceTrackingParams,omitempty"`
	MainAppWebResponseContext       *MainAppWebResponseContext       `json:"mainAppWebResponseContext,omitempty"`
	WebResponseContextExtensionData *WebResponseContextExtensionData `json:"webResponseContextExtensionData,omitempty"`
}

type MainAppWebResponseContext struct {
	LoggedOut *bool `json:"loggedOut,omitempty"`
}

type ServiceTrackingParam struct {
	Service *string `json:"service,omitempty"`
	Params  []Param `json:"params,omitempty"`
}

type Param struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`
}

type WebResponseContextExtensionData struct {
	HasDecorated *bool `json:"hasDecorated,omitempty"`
}

type Storyboards struct {
	PlayerStoryboardSpecRenderer *PlayerStoryboardSpecRenderer `json:"playerStoryboardSpecRenderer,omitempty"`
}

type PlayerStoryboardSpecRenderer struct {
	Spec *string `json:"spec,omitempty"`
}

type StreamingData struct {
	ExpiresInSeconds *string  `json:"expiresInSeconds,omitempty"`
	Formats          []Format `json:"formats,omitempty"`
	AdaptiveFormats  []Format `json:"adaptiveFormats,omitempty"`
}

type Format struct {
	Itag             *int64          `json:"itag,omitempty"`
	URL              *string         `json:"url,omitempty"`
	MIMEType         *string         `json:"mimeType,omitempty"`
	Bitrate          *int64          `json:"bitrate,omitempty"`
	Width            *int64          `json:"width,omitempty"`
	Height           *int64          `json:"height,omitempty"`
	InitRange        *Range          `json:"initRange,omitempty"`
	IndexRange       *Range          `json:"indexRange,omitempty"`
	LastModified     *string         `json:"lastModified,omitempty"`
	ContentLength    *string         `json:"contentLength,omitempty"`
	Quality          *string         `json:"quality,omitempty"`
	FPS              *int64          `json:"fps,omitempty"`
	QualityLabel     *string         `json:"qualityLabel,omitempty"`
	ProjectionType   *ProjectionType `json:"projectionType,omitempty"`
	AverageBitrate   *int64          `json:"averageBitrate,omitempty"`
	ApproxDurationMS *string         `json:"approxDurationMs,omitempty"`
	ColorInfo        *ColorInfo      `json:"colorInfo,omitempty"`
	HighReplication  *bool           `json:"highReplication,omitempty"`
	AudioQuality     *string         `json:"audioQuality,omitempty"`
	AudioSampleRate  *string         `json:"audioSampleRate,omitempty"`
	AudioChannels    *int64          `json:"audioChannels,omitempty"`
	LoudnessDB       *float64        `json:"loudnessDb,omitempty"`
}

type ColorInfo struct {
	Primaries               *string `json:"primaries,omitempty"`
	TransferCharacteristics *string `json:"transferCharacteristics,omitempty"`
	MatrixCoefficients      *string `json:"matrixCoefficients,omitempty"`
}

type Range struct {
	Start *string `json:"start,omitempty"`
	End   *string `json:"end,omitempty"`
}

type VideoDetails struct {
	VideoID           *string     `json:"videoId,omitempty"`
	Title             *string     `json:"title,omitempty"`
	LengthSeconds     *string     `json:"lengthSeconds,omitempty"`
	ChannelID         *string     `json:"channelId,omitempty"`
	IsOwnerViewing    *bool       `json:"isOwnerViewing,omitempty"`
	ShortDescription  *string     `json:"shortDescription,omitempty"`
	IsCrawlable       *bool       `json:"isCrawlable,omitempty"`
	Thumbnail         *ImageClass `json:"thumbnail,omitempty"`
	AllowRatings      *bool       `json:"allowRatings,omitempty"`
	ViewCount         *string     `json:"viewCount,omitempty"`
	Author            *string     `json:"author,omitempty"`
	IsPrivate         *bool       `json:"isPrivate,omitempty"`
	IsUnpluggedCorpus *bool       `json:"isUnpluggedCorpus,omitempty"`
	IsLiveContent     *bool       `json:"isLiveContent,omitempty"`
}

type ProjectionType string

const (
	Rectangular ProjectionType = "RECTANGULAR"
)

package schemas

type YoutubeBodyMetaInfo struct {
	Context         Context         `json:"context"`
	VideoID         string          `json:"videoId"`
	Params          string          `json:"params"`
	PlaybackContext PlaybackContext `json:"playbackContext"`
	RacyCheckOk     bool            `json:"racyCheckOk"`
	ContentCheckOk  bool            `json:"contentCheckOk"`
}

type Context struct {
	Client Client `json:"client"`
}

type Client struct {
	Hl                string `json:"hl"`
	ClientName        string `json:"clientName"`
	ClientVersion     string `json:"clientVersion"`
	AndroidSDKVersion int64  `json:"androidSdkVersion"`
	UserAgent         string `json:"userAgent"`
	TimeZone          string `json:"timeZone"`
	UtcOffsetMinutes  int64  `json:"utcOffsetMinutes"`
}

type PlaybackContext struct {
	ContentPlaybackContext ContentPlaybackContext `json:"contentPlaybackContext"`
}

type ContentPlaybackContext struct {
	Html5Preference string `json:"html5Preference"`
}

type YoutubeMetaInfo struct {
	AdPlacements      []interface{}      `json:"adPlacements,omitempty"`
	Annotations       []interface{}      `json:"annotations,omitempty"`
	Attestation       *interface{}       `json:"attestation,omitempty"`
	Captions          *interface{}       `json:"captions,omitempty"`
	Endscreen         *interface{}       `json:"endscreen,omitempty"`
	FrameworkUpdates  *interface{}       `json:"frameworkUpdates,omitempty"`
	Microformat       *interface{}       `json:"microformat,omitempty"`
	PlayabilityStatus *PlayabilityStatus `json:"playabilityStatus,omitempty"`
	PlaybackTracking  *interface{}       `json:"playbackTracking,omitempty"`
	PlayerAds         []interface{}      `json:"playerAds,omitempty"`
	PlayerConfig      *interface{}       `json:"playerConfig,omitempty"`
	ResponseContext   *interface{}       `json:"responseContext,omitempty"`
	Storyboards       *interface{}       `json:"storyboards,omitempty"`
	StreamingData     *StreamingData     `json:"streamingData,omitempty"`
	TrackingParams    *string            `json:"trackingParams,omitempty"`
	VideoDetails      *interface{}       `json:"videoDetails,omitempty"`
}

type PlayabilityStatus struct {
	Status          *string      `json:"status,omitempty"`
	Reason          *string      `json:"reason,omitempty"`
	PlayableInEmbed *bool        `json:"playableInEmbed,omitempty"`
	Miniplayer      *interface{} `json:"miniplayer,omitempty"`
	ContextParams   *string      `json:"contextParams,omitempty"`
}

type StreamingData struct {
	ExpiresInSeconds *string  `json:"expiresInSeconds,omitempty"`
	Formats          []Format `json:"formats,omitempty"`
	AdaptiveFormats  []Format `json:"adaptiveFormats,omitempty"`
}

type Format struct {
	Itag             *int64       `json:"itag,omitempty"`
	URL              *string      `json:"url,omitempty"`
	MIMEType         *string      `json:"mimeType,omitempty"`
	Bitrate          *int64       `json:"bitrate,omitempty"`
	Width            *int64       `json:"width,omitempty"`
	Height           *int64       `json:"height,omitempty"`
	InitRange        *interface{} `json:"initRange,omitempty"`
	IndexRange       *interface{} `json:"indexRange,omitempty"`
	LastModified     *string      `json:"lastModified,omitempty"`
	ContentLength    *string      `json:"contentLength,omitempty"`
	Quality          *string      `json:"quality,omitempty"`
	FPS              *int64       `json:"fps,omitempty"`
	QualityLabel     *string      `json:"qualityLabel,omitempty"`
	ProjectionType   *interface{} `json:"projectionType,omitempty"`
	AverageBitrate   *int64       `json:"averageBitrate,omitempty"`
	ApproxDurationMS *string      `json:"approxDurationMs,omitempty"`
	ColorInfo        *interface{} `json:"colorInfo,omitempty"`
	HighReplication  *bool        `json:"highReplication,omitempty"`
	AudioQuality     *string      `json:"audioQuality,omitempty"`
	AudioSampleRate  *string      `json:"audioSampleRate,omitempty"`
	AudioChannels    *int64       `json:"audioChannels,omitempty"`
	LoudnessDB       *float64     `json:"loudnessDb,omitempty"`
	SignatureCipher  *string      `json:"signatureCipher,omitempty"`
}

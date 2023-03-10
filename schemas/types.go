/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package schemas

import (
	"encoding/json"
)

type CheckIfTokenValidResponse struct {
	IsVerified     bool   `json:"is_verified"`
	CurrentBalance string `json:"current_balance"`
}

type Account struct {
	Error          *string        `json:"error,omitempty"`
	IsVerified     bool           `json:"is_verified"`
	CurrentBalance CurrentBalance `json:"current_balance"`
	Id             *int           `json:"id,omitempty"`
}

type CurrentBalance struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type TranscriptResponse struct {
	AcousticModel            *string                    `json:"acoustic_model,omitempty"`
	AudioDuration            *int64                     `json:"audio_duration,omitempty"`
	AudioEndAt               *interface{}               `json:"audio_end_at,omitempty"`
	AudioStartFrom           *interface{}               `json:"audio_start_from,omitempty"`
	AudioURL                 *string                    `json:"audio_url,omitempty"`
	AutoChapters             *bool                      `json:"auto_chapters,omitempty"`
	AutoHighlights           *bool                      `json:"auto_highlights,omitempty"`
	AutoHighlightsResult     *AutoHighlightsResult      `json:"auto_highlights_result,omitempty"`
	BoostParam               interface{}                `json:"boost_param"`
	Chapters                 *[]Chapter                 `json:"chapters,omitempty"`
	ClusterID                interface{}                `json:"cluster_id"`
	Confidence               *float64                   `json:"confidence,omitempty"`
	ContentSafety            *bool                      `json:"content_safety,omitempty"`
	ContentSafetyLabels      *ContentSafetyLabels       `json:"content_safety_labels,omitempty"`
	CustomSpelling           interface{}                `json:"custom_spelling"`
	Disfluencies             *bool                      `json:"disfluencies,omitempty"`
	DualChannel              *bool                      `json:"dual_channel,omitempty"`
	Entities                 *[]Entity                  `json:"entities,omitempty"`
	EntityDetection          *bool                      `json:"entity_detection,omitempty"`
	Error                    *string                    `json:"error,omitempty"`
	FilterProfanity          *bool                      `json:"filter_profanity,omitempty"`
	FormatText               *bool                      `json:"format_text,omitempty"`
	IabCategories            *bool                      `json:"iab_categories,omitempty"`
	IabCategoriesResult      *IabCategoriesResult       `json:"iab_categories_result,omitempty"`
	ID                       *string                    `json:"id,omitempty"`
	LanguageCode             *string                    `json:"language_code,omitempty"`
	LanguageDetection        *bool                      `json:"language_detection,omitempty"`
	LanguageModel            *string                    `json:"language_model,omitempty"`
	Punctuate                *bool                      `json:"punctuate,omitempty"`
	RedactPii                *bool                      `json:"redact_pii,omitempty"`
	RedactPiiAudio           *bool                      `json:"redact_pii_audio,omitempty"`
	RedactPiiAudioQuality    interface{}                `json:"redact_pii_audio_quality"`
	RedactPiiPolicies        interface{}                `json:"redact_pii_policies"`
	RedactPiiSub             interface{}                `json:"redact_pii_sub"`
	SentimentAnalysis        *bool                      `json:"sentiment_analysis,omitempty"`
	SentimentAnalysisResults *[]SentimentAnalysisResult `json:"sentiment_analysis_results,omitempty"`
	SpeakerLabels            bool                       `json:"speaker_labels,omitempty"`
	SpeedBoost               *bool                      `json:"speed_boost,omitempty"`
	Status                   *string                    `json:"status,omitempty"`
	Summarization            *bool                      `json:"summarization,omitempty"`
	Summary                  *string                    `json:"summary,omitempty"`
	SummaryType              *string                    `json:"summary_type,omitempty"`
	Text                     *string                    `json:"text,omitempty"`
	Throttled                interface{}                `json:"throttled"`
	Utterances               *[]SentimentAnalysisResult `json:"utterances,omitempty"`
	WebhookAuth              *bool                      `json:"webhook_auth,omitempty"`
	WebhookAuthHeaderName    interface{}                `json:"webhook_auth_header_name"`
	WebhookStatusCode        interface{}                `json:"webhook_status_code"`
	WebhookURL               interface{}                `json:"webhook_url"`
	WordBoost                []interface{}              `json:"word_boost,omitempty"`
	Words                    []SentimentAnalysisResult  `json:"words,omitempty"`
}

type AutoHighlightsResult struct {
	Results []AutoHighlightsResultResult `json:"results,omitempty"`
	Status  *string                      `json:"status,omitempty"`
}

type AutoHighlightsResultResult struct {
	Count      *int64      `json:"count,omitempty"`
	Rank       *float64    `json:"rank,omitempty"`
	Text       string      `json:"text,omitempty"`
	Timestamps []Timestamp `json:"timestamps,omitempty"`
}

type Timestamp struct {
	Start *int64 `json:"start,omitempty"`
	End   *int64 `json:"end,omitempty"`
}

type Chapter struct {
	Summary  string `json:"summary,omitempty"`
	Headline string `json:"headline,omitempty"`
	Gist     string `json:"gist,omitempty"`
	Start    *int64 `json:"start,omitempty"`
	End      *int64 `json:"end,omitempty"`
}

type ContentSafetyLabels struct {
	Status               *string                     `json:"status,omitempty"`
	Results              []ContentSafetyLabelsResult `json:"results,omitempty"`
	Summary              *Summary                    `json:"summary,omitempty"`
	SeverityScoreSummary *SeverityScoreSummary       `json:"severity_score_summary,omitempty"`
}

type ContentSafetyLabelsResult struct {
	Text      string     `json:"text,omitempty"`
	Labels    []Label    `json:"labels,omitempty"`
	Timestamp *Timestamp `json:"timestamp,omitempty"`
}

type Label struct {
	Label      string   `json:"label,omitempty"`
	Confidence *float64 `json:"confidence,omitempty"`
	Severity   *float64 `json:"severity"`
}

type SeverityScoreSummary struct {
	Profanity *Profanity `json:"profanity,omitempty"`
}

type Profanity struct {
	Low    json.Number `json:"low,omitempty"`
	Medium json.Number `json:"medium,omitempty"`
	High   json.Number `json:"high,omitempty"`
}

type Summary struct {
	Profanity *float64 `json:"profanity,omitempty"`
	Nsfw      *float64 `json:"nsfw,omitempty"`
}

type Entity struct {
	EntityType string `json:"entity_type,omitempty"`
	Text       string `json:"text,omitempty"`
	Start      *int64 `json:"start,omitempty"`
	End        *int64 `json:"end,omitempty"`
}

type IabCategoriesResult struct {
	Status  *string                     `json:"status,omitempty"`
	Results []IabCategoriesResultResult `json:"results,omitempty"`
	Summary map[string]float64          `json:"summary,omitempty"`
}

type IabCategoriesResultResult struct {
	Text      string        `json:"text,omitempty"`
	Labels    []FluffyLabel `json:"labels,omitempty"`
	Timestamp *Timestamp    `json:"timestamp,omitempty"`
}

type FluffyLabel struct {
	Relevance *float64 `json:"relevance,omitempty"`
	Label     string   `json:"label,omitempty"`
}

type SentimentAnalysisResult struct {
	Channel    string                    `json:"channel,omitempty"`
	Text       string                    `json:"text,omitempty"`
	Start      *int64                    `json:"start,omitempty"`
	End        *int64                    `json:"end,omitempty"`
	Sentiment  string                    `json:"sentiment,omitempty"`
	Confidence *float64                  `json:"confidence,omitempty"`
	Speaker    string                    `json:"speaker,omitempty"`
	Words      []SentimentAnalysisResult `json:"words,omitempty"`
}

type UploadResponse struct {
	UploadURL string `json:"upload_url"`
}

type TranscribeFlags struct {
	Poll bool `json:"poll"`
	Json bool `json:"json"`
}

type TranscribeParams struct {
	AudioURL               string           `json:"audio_url"`
	AutoChapters           bool             `json:"auto_chapters"`
	AutoHighlights         bool             `json:"auto_highlights"`
	BoostParam             *string          `json:"boost_param,omitempty"`
	ContentModeration      bool             `json:"content_safety"`
	CustomSpelling         []CustomSpelling `json:"custom_spelling,omitempty"`
	DualChannel            bool             `json:"dual_channel"`
	EntityDetection        bool             `json:"entity_detection"`
	FormatText             bool             `json:"format_text"`
	LanguageCode           *string          `json:"language_code,omitempty"`
	LanguageDetection      bool             `json:"language_detection"`
	Punctuate              bool             `json:"punctuate"`
	RedactPii              bool             `json:"redact_pii"`
	RedactPiiPolicies      []string         `json:"redact_pii_policies"`
	SentimentAnalysis      bool             `json:"sentiment_analysis"`
	SpeakerLabels          bool             `json:"speaker_labels"`
	Summarization          bool             `json:"summarization,omitempty"`
	SummaryModel           string           `json:"summary_model,omitempty"`
	SummaryType            string           `json:"summary_type,omitempty"`
	TopicDetection         bool             `json:"iab_categories"`
	WebhookAuthHeaderName  string           `json:"webhook_auth_header_name,omitempty"`
	WebhookAuthHeaderValue string           `json:"webhook_auth_header_value,omitempty"`
	WebhookURL             string           `json:"webhook_url,omitempty"`
	WordBoost              []string         `json:"word_boost,omitempty"`
}

type CustomSpelling struct {
	From []string `json:"from"`
	To   string   `json:"to"`
}

type RedactPiiPolicy string

const (
	MedicalProcess         RedactPiiPolicy = "medical_process"
	MedicalCondition       RedactPiiPolicy = "medical_condition"
	BloodType              RedactPiiPolicy = "blood_type"
	Drug                   RedactPiiPolicy = "drug"
	Injury                 RedactPiiPolicy = "injury"
	NumberSequence         RedactPiiPolicy = "number_sequence"
	EmailAddress           RedactPiiPolicy = "email_address"
	DateOfBirth            RedactPiiPolicy = "date_of_birth"
	PhoneNumber            RedactPiiPolicy = "phone_number"
	USSocialSecurityNumber RedactPiiPolicy = "us_social_security_number"
	CreditCardNumber       RedactPiiPolicy = "credit_card_number"
	CreditCardExpiration   RedactPiiPolicy = "credit_card_expiration"
	Date                   RedactPiiPolicy = "date"
	Nationality            RedactPiiPolicy = "nationality"
	Event                  RedactPiiPolicy = "event"
	Language               RedactPiiPolicy = "language"
	Location               RedactPiiPolicy = "location"
	MoneyAmount            RedactPiiPolicy = "money_amount"
	PersonName             RedactPiiPolicy = "person_name"
	PersonAge              RedactPiiPolicy = "person_age"
	Organization           RedactPiiPolicy = "organization"
	PoliticalAffiliation   RedactPiiPolicy = "political_affiliation"
	Occupation             RedactPiiPolicy = "occupation"
	Religion               RedactPiiPolicy = "religion"
	DriversLicense         RedactPiiPolicy = "drivers_license"
	BankingInformation     RedactPiiPolicy = "banking_information"
)

type PostHogProperties struct {
	Poll              bool   `json:"poll,omitempty"`
	Json              bool   `json:"json,omitempty"`
	SpeakerLabels     bool   `json:"speaker_labels,omitempty"`
	Punctuate         bool   `json:"punctuate,omitempty"`
	FormatText        bool   `json:"format_text,omitempty"`
	DualChannel       *bool  `json:"dual_channel,omitempty"`
	RedactPii         bool   `json:"redact_pii,omitempty"`
	AutoHighlights    bool   `json:"auto_highlights,omitempty"`
	ContentModeration bool   `json:"content_safety,omitempty"`
	TopicDetection    bool   `json:"iab_categories,omitempty"`
	SentimentAnalysis bool   `json:"sentiment_analysis,omitempty"`
	AutoChapters      bool   `json:"auto_chapters,omitempty"`
	EntityDetection   bool   `json:"entity_detection,omitempty"`
	Version           string `json:"version,omitempty"`
	OS                string `json:"os,omitempty"`
	Arch              string `json:"arch,omitempty"`
	Method            string `json:"method,omitempty"`
	I                 bool   `json:"i,omitempty"`
	LatestVersion     string `json:"latest_version,omitempty"`
}

var LanguageMap = map[string]string{
	"en":    "Global English",
	"en-au": "Australian English",
	"en_uk": "British English",
	"en-US": "US English",
	"es":    "Spanish",
	"fr":    "French",
	"de":    "German",
	"it":    "Italian",
	"pt":    "Portuguese",
	"nl":    "Dutch",
	"hi":    "Hindi",
	"ja":    "Japanese",
}

var SummarizationTypeMap = map[string]string{
	"paragraph":       "Paragraph",
	"headline":        "Headline",
	"gist":            "Gist",
	"bullets":         "Bullets",
	"bullets_verbose": "Bullets Verbose",
}

var PIIRedactionPolicyMap = map[string]string{
	"banking_information":       "Banking Information",
	"blood_type":                "Blood Type",
	"credit_card_cvv":           "Credit Card CVV",
	"credit_card_expiration":    "Credit Card Expiration",
	"credit_card_number":        "Credit Card Number",
	"date":                      "Date",
	"drivers_license":           "Drivers License",
	"drug":                      "Drug",
	"email_address":             "Email Address",
	"event":                     "Event",
	"injury":                    "Injury",
	"language":                  "Language",
	"location":                  "Location",
	"medical_condition":         "Medical Condition",
	"medical_process":           "Medical Process",
	"money_amount":              "Money Amount",
	"nationality":               "Nationality",
	"number_sequence":           "Number Sequence",
	"occupation":                "Occupation",
	"organization":              "Organization",
	"person_age":                "Person Age",
	"person_name":               "Person Name",
	"phone_number":              "Phone Number",
	"political_affiliation":     "Political Affiliation",
	"religion":                  "Religion",
	"us_social_security_number": "US Social Security Number",
}

var SummarizationTypeMapReverse = map[string]string{
	"paragraph":       "Paragraph",
	"headline":        "Headline",
	"gist":            "Gist",
	"bullets":         "Bullets",
	"bullets_verbose": "Bullets Verbose",
}

var SummarizationModelMap = map[string][]string{
	"conversational": {"headline", "paragraph", "bullets", "bullets_verbose"},
	"catchy":         {"gist", "headline"},
	"informative":    {"headline", "paragraph", "bullets", "bullets_verbose"},
}

type PrintErrorProps struct {
	Error   error
	Message string
}

type Release struct {
	URL              *string      `json:"url,omitempty"`
	AssetsURL        *string      `json:"assets_url,omitempty"`
	UploadURL        *string      `json:"upload_url,omitempty"`
	HTMLURL          *string      `json:"html_url,omitempty"`
	ID               *int64       `json:"id,omitempty"`
	Author           *interface{} `json:"author,omitempty"`
	NodeID           *string      `json:"node_id,omitempty"`
	TagName          *string      `json:"tag_name,omitempty"`
	TargetCommitish  *string      `json:"target_commitish,omitempty"`
	Name             *string      `json:"name,omitempty"`
	Draft            *bool        `json:"draft,omitempty"`
	Prerelease       *bool        `json:"prerelease,omitempty"`
	CreatedAt        *string      `json:"created_at,omitempty"`
	PublishedAt      *string      `json:"published_at,omitempty"`
	Assets           *interface{} `json:"assets,omitempty"`
	TarballURL       *string      `json:"tarball_url,omitempty"`
	ZipballURL       *string      `json:"zipball_url,omitempty"`
	Body             *string      `json:"body,omitempty"`
	Message          *string      `json:"message,omitempty"`
	DocumentationUrl *string      `json:"documentation_url,omitempty"`
}

var ValidFileTypes = []string{
	"3ga",
	"8svx",
	"aac",
	"ac3",
	"aif",
	"aiff",
	"alac",
	"amr",
	"ape",
	"au",
	"dss",
	"flac",
	"flv",
	"m4a",
	"m4b",
	"m4p",
	"m4r",
	"mp3",
	"mpga",
	"ogg",
	"oga",
	"mogg",
	"opus",
	"qcp",
	"tta",
	"voc",
	"wav",
	"wma",
	"wv",
	"webm",
	"MTS",
	"M2TS",
	"TS",
	"mov",
	"mp2",
	"mp4",
	"m4p",
	"m4v",
	"mxf",
}

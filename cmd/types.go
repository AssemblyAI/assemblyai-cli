/*
Copyright © 2022 AssemblyAI support@assemblyai.com
*/
package cmd

import "encoding/json"

type CheckIfTokenValidResponse struct {
	IsVerified     bool   `json:"is_verified"`
	CurrentBalance string `json:"current_balance"`
}

type Account struct {
	IsVerified     bool           `json:"is_verified"`
	CurrentBalance CurrentBalance `json:"current_balance"`
}

type CurrentBalance struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type TranscriptResponse struct {
	Error                    *string                    `json:"error,omitempty"`
	AcousticModel            *string                    `json:"acoustic_model,omitempty"`
	AudioDuration            *int64                     `json:"audio_duration,omitempty"`
	AudioEndAt               *interface{}               `json:"audio_end_at,omitempty"`
	AudioStartFrom           *interface{}               `json:"audio_start_from,omitempty"`
	AudioURL                 *string                    `json:"audio_url,omitempty"`
	AutoChapters             *bool                      `json:"auto_chapters,omitempty"`
	AutoHighlights           *bool                      `json:"auto_highlights,omitempty"`
	AutoHighlightsResult     *AutoHighlightsResult      `json:"auto_highlights_result,omitempty"`
	BoostParam               *interface{}               `json:"boost_param,omitempty"`
	Chapters                 *[]Chapter                 `json:"chapters,omitempty"`
	ClusterID                *interface{}               `json:"cluster_id,omitempty"`
	Confidence               *float64                   `json:"confidence,omitempty"`
	ContentSafety            *bool                      `json:"content_safety,omitempty"`
	ContentSafetyLabels      *ContentSafetyLabels       `json:"content_safety_labels,omitempty"`
	CustomSpelling           *interface{}               `json:"custom_spelling,omitempty"`
	Disfluencies             *bool                      `json:"disfluencies,omitempty"`
	DualChannel              *bool                      `json:"dual_channel,omitempty"`
	Entities                 *[]Entity                  `json:"entities,omitempty"`
	EntityDetection          *bool                      `json:"entity_detection,omitempty"`
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
	RedactPiiAudioQuality    *interface{}               `json:"redact_pii_audio_quality,omitempty"`
	RedactPiiPolicies        *interface{}               `json:"redact_pii_policies,omitempty"`
	RedactPiiSub             *interface{}               `json:"redact_pii_sub,omitempty"`
	SentimentAnalysis        *bool                      `json:"sentiment_analysis,omitempty"`
	SentimentAnalysisResults *[]SentimentAnalysisResult `json:"sentiment_analysis_results,omitempty"`
	SpeakerLabels            *bool                      `json:"speaker_labels,omitempty"`
	SpeedBoost               *bool                      `json:"speed_boost,omitempty"`
	Status                   *string                    `json:"status,omitempty"`
	Text                     *string                    `json:"text,omitempty"`
	Throttled                *interface{}               `json:"throttled,omitempty"`
	Utterances               *[]SentimentAnalysisResult `json:"utterances,omitempty"`
	WebhookAuth              *bool                      `json:"webhook_auth,omitempty"`
	WebhookAuthHeaderName    *interface{}               `json:"webhook_auth_header_name,omitempty"`
	WebhookStatusCode        *interface{}               `json:"webhook_status_code,omitempty"`
	WebhookURL               *interface{}               `json:"webhook_url,omitempty"`
	WordBoost                *[]interface{}             `json:"word_boost,omitempty"`
	Words                    *[]SentimentAnalysisResult `json:"words,omitempty"`
}

type AutoHighlightsResult struct {
	Results []AutoHighlightsResultResult `json:"results"`
	Status  string                       `json:"status"`
}

type AutoHighlightsResultResult struct {
	Count      int64       `json:"count"`
	Rank       float64     `json:"rank"`
	Text       string      `json:"text"`
	Timestamps []Timestamp `json:"timestamps"`
}

type Timestamp struct {
	End   int64 `json:"end"`
	Start int64 `json:"start"`
}

type Chapter struct {
	End      int64  `json:"end"`
	Gist     string `json:"gist"`
	Headline string `json:"headline"`
	Start    int64  `json:"start"`
	Summary  string `json:"summary"`
}

type ContentSafetyLabels struct {
	Results              []ContentSafetyLabelsResult `json:"results"`
	SeverityScoreSummary SeverityScoreSummary        `json:"severity_score_summary"`
	Status               string                      `json:"status"`
	Summary              Summary                     `json:"summary"`
}

type ContentSafetyLabelsResult struct {
	Labels    []Label   `json:"labels"`
	Text      string    `json:"text"`
	Timestamp Timestamp `json:"timestamp"`
}

type Label struct {
	Confidence float64  `json:"confidence"`
	Label      string   `json:"label"`
	Severity   *float64 `json:"severity"`
}

type SeverityScoreSummary struct {
	Profanity Profanity `json:"profanity"`
}

type Profanity struct {
	Low    json.Number `json:"low"`
	Medium json.Number `json:"medium"`
	High   json.Number `json:"high"`
}

type Summary struct {
	Profanity float64 `json:"profanity"`
	Nsfw      float64 `json:"nsfw"`
}

type Entity struct {
	End        int64      `json:"end"`
	EntityType EntityType `json:"entity_type"`
	Start      int64      `json:"start"`
	Text       string     `json:"text"`
}

type IabCategoriesResult struct {
	Results []IabCategoriesResultResult `json:"results"`
	Status  string                      `json:"status"`
	Summary map[string]float64          `json:"summary"`
}

type IabCategoriesResultResult struct {
	Labels    []FluffyLabel `json:"labels"`
	Text      string        `json:"text"`
	Timestamp Timestamp     `json:"timestamp"`
}

type FluffyLabel struct {
	Label     string  `json:"label"`
	Relevance float64 `json:"relevance"`
}

type SentimentAnalysisResult struct {
	Channel    *string                   `json:"channel,omitempty"`
	Confidence float64                   `json:"confidence"`
	End        int64                     `json:"end"`
	Sentiment  *Sentiment                `json:"sentiment,omitempty"`
	Speaker    Speaker                   `json:"speaker"`
	Start      int64                     `json:"start"`
	Text       string                    `json:"text"`
	Words      []SentimentAnalysisResult `json:"words,omitempty"`
}

type EntityType string

const (
	Location     EntityType = "location"
	Occupation   EntityType = "occupation"
	Organization EntityType = "organization"
	PersonName   EntityType = "person_name"
)

type Sentiment string

const (
	Negative Sentiment = "NEGATIVE"
	Neutral  Sentiment = "NEUTRAL"
	Positive Sentiment = "POSITIVE"
)

type Speaker string

const (
	A Speaker = "A"
	B Speaker = "B"
	C Speaker = "C"
	D Speaker = "D"
	E Speaker = "E"
	F Speaker = "F"
	G Speaker = "G"
	H Speaker = "H"
)

type UploadResponse struct {
	UploadURL string `json:"upload_url"`
}

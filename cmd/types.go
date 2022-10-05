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
	AcousticModel            *string                   `json:"acoustic_model,omitempty"`
	AudioDuration            *int64                    `json:"audio_duration,omitempty"`
	AudioEndAt               *interface{}              `json:"audio_end_at,omitempty"`
	AudioStartFrom           *interface{}              `json:"audio_start_from,omitempty"`
	AudioURL                 *string                   `json:"audio_url,omitempty"`
	AutoChapters             *bool                     `json:"auto_chapters,omitempty"`
	AutoHighlights           *bool                     `json:"auto_highlights,omitempty"`
	AutoHighlightsResult     *AutoHighlightsResult     `json:"auto_highlights_result,omitempty"`
	BoostParam               interface{}               `json:"boost_param"`
	Chapters                 []Chapter                 `json:"chapters,omitempty"`
	ClusterID                interface{}               `json:"cluster_id"`
	Confidence               *float64                  `json:"confidence,omitempty"`
	ContentSafety            *bool                     `json:"content_safety,omitempty"`
	ContentSafetyLabels      *ContentSafetyLabels      `json:"content_safety_labels,omitempty"`
	CustomSpelling           interface{}               `json:"custom_spelling"`
	Disfluencies             *bool                     `json:"disfluencies,omitempty"`
	DualChannel              *bool                     `json:"dual_channel,omitempty"`
	Entities                 []Entity                  `json:"entities,omitempty"`
	EntityDetection          *bool                     `json:"entity_detection,omitempty"`
	Error                    *string                   `json:"error,omitempty"`
	FilterProfanity          *bool                     `json:"filter_profanity,omitempty"`
	FormatText               *bool                     `json:"format_text,omitempty"`
	IabCategories            *bool                     `json:"iab_categories,omitempty"`
	IabCategoriesResult      *IabCategoriesResult      `json:"iab_categories_result,omitempty"`
	ID                       *string                   `json:"id,omitempty"`
	LanguageCode             *string                   `json:"language_code,omitempty"`
	LanguageDetection        *bool                     `json:"language_detection,omitempty"`
	LanguageModel            *string                   `json:"language_model,omitempty"`
	Punctuate                *bool                     `json:"punctuate,omitempty"`
	RedactPii                *bool                     `json:"redact_pii,omitempty"`
	RedactPiiAudio           *bool                     `json:"redact_pii_audio,omitempty"`
	RedactPiiAudioQuality    interface{}               `json:"redact_pii_audio_quality"`
	RedactPiiPolicies        interface{}               `json:"redact_pii_policies"`
	RedactPiiSub             interface{}               `json:"redact_pii_sub"`
	SentimentAnalysis        *bool                     `json:"sentiment_analysis,omitempty"`
	SentimentAnalysisResults []SentimentAnalysisResult `json:"sentiment_analysis_results,omitempty"`
	SpeakerLabels            bool                      `json:"speaker_labels,omitempty"`
	SpeedBoost               *bool                     `json:"speed_boost,omitempty"`
	Status                   *string                   `json:"status,omitempty"`
	Text                     *string                   `json:"text,omitempty"`
	Throttled                interface{}               `json:"throttled"`
	Utterances               []SentimentAnalysisResult `json:"utterances,omitempty"`
	WebhookAuth              *bool                     `json:"webhook_auth,omitempty"`
	WebhookAuthHeaderName    interface{}               `json:"webhook_auth_header_name"`
	WebhookStatusCode        interface{}               `json:"webhook_status_code"`
	WebhookURL               interface{}               `json:"webhook_url"`
	WordBoost                []interface{}             `json:"word_boost,omitempty"`
	Words                    []SentimentAnalysisResult `json:"words,omitempty"`
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
	// PiiPolicies       string `json:"pii_policies"`
	AudioURL          string `json:"audio_url"`
	AutoChapters      bool   `json:"auto_chapters"`
	AutoHighlights    bool   `json:"auto_highlights"`
	ContentModeration bool   `json:"content_safety"`
	DualChannel       bool   `json:"dual_channel"`
	EntityDetection   bool   `json:"entity_detection"`
	FormatText        bool   `json:"format_text"`
	Punctuate         bool   `json:"punctuate"`
	RedactPii         bool   `json:"redact_pii"`
	SentimentAnalysis bool   `json:"sentiment_analysis"`
	SpeakerLabels     bool   `json:"speaker_labels"`
	TopicDetection    bool   `json:"iab_categories"`
}

package utils

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	S "github.com/AssemblyAI/assemblyai-cli/schemas"
	"github.com/gosuri/uitable"
)

func textPrintFormatted(text string, words []S.SentimentAnalysisResult) {
	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 10)
	sentences := SplitSentences(text, true)
	timestamps := GetSentenceTimestamps(sentences, words)
	for index, sentence := range sentences {
		if sentence != "" {
			stamp := ""
			if len(timestamps) > index {
				stamp = timestamps[index]
			}
			table.AddRow(stamp, sentence)
		}
	}
	fmt.Println(table)
	fmt.Println()
}

func dualChannelPrintFormatted(utterances []S.SentimentAnalysisResult) {
	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 21)
	for _, utterance := range utterances {
		start := TransformMsToTimestamp(*utterance.Start)
		speaker := fmt.Sprintf("(Channel %s)", utterance.Channel)

		sentences := SplitSentences(utterance.Text, false)
		for _, sentence := range sentences {
			table.AddRow(start, speaker, sentence)
			start = ""
			speaker = ""
		}
	}
	fmt.Println(table)
	fmt.Println()
}

func speakerLabelsPrintFormatted(utterances []S.SentimentAnalysisResult) {
	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 27)

	for _, utterance := range utterances {
		sentences := SplitSentences(utterance.Text, false)
		timestamps := GetSentenceTimestampsAndSpeaker(sentences, utterance.Words)
		for index, sentence := range sentences {
			if sentence != "" {
				info := []string{"", ""}
				if len(timestamps) > index {
					info = timestamps[index]
				}
				table.AddRow(info[0], info[1], sentence)
			}
		}
	}
	fmt.Println(table)
	fmt.Println()
}

func highlightsPrintFormatted(highlights S.AutoHighlightsResult) {
	if *highlights.Status != "success" {
		fmt.Println("Could not retrieve highlights")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.Separator = " |\t"
	table.AddRow("| count", "text")
	sort.SliceStable(highlights.Results, func(i, j int) bool {
		return int(*highlights.Results[i].Count) > int(*highlights.Results[j].Count)
	})
	for _, highlight := range highlights.Results {
		table.AddRow("| "+strconv.FormatInt(*highlight.Count, 10), highlight.Text)
	}
	fmt.Println(table)
	fmt.Println()
}

func contentSafetyPrintFormatted(labels S.ContentSafetyLabels) {
	if *labels.Status != "success" {
		fmt.Println("Could not retrieve content safety labels")
		return
	}
	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 24)
	table.Separator = " |\t"
	table.AddRow("| label", "text")
	for _, label := range labels.Results {
		var labelString string
		for _, innerLabel := range label.Labels {
			labelString = innerLabel.Label + " " + labelString
		}
		table.AddRow("| "+labelString, label.Text)
	}
	fmt.Println(table)
	fmt.Println()
}

func topicDetectionPrintFormatted(categories S.IabCategoriesResult) {
	if *categories.Status != "success" {
		fmt.Println("Could not retrieve topic detection")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 20)
	table.Separator = " |\t"
	table.AddRow("| rank", "topic")
	var ArrayCategoriesSorted []S.ArrayCategories
	for category, i := range categories.Summary {
		add := S.ArrayCategories{
			Category: category,
			Score:    i,
		}
		ArrayCategoriesSorted = append(ArrayCategoriesSorted, add)
	}
	sort.SliceStable(ArrayCategoriesSorted, func(i, j int) bool {
		return ArrayCategoriesSorted[i].Score > ArrayCategoriesSorted[j].Score
	})

	for i, category := range ArrayCategoriesSorted {
		table.AddRow(fmt.Sprintf("| %o", i+1), category.Category)
	}
	fmt.Println(table)
	fmt.Println()
}

func sentimentAnalysisPrintFormatted(sentiments []S.SentimentAnalysisResult) {
	if len(sentiments) == 0 {
		fmt.Println("Could not retrieve sentiment analysis")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 20)
	table.Separator = " |\t"
	table.AddRow("| sentiment", "text")
	for _, sentiment := range sentiments {
		sentimentStatus := sentiment.Sentiment
		table.AddRow("| "+sentimentStatus, sentiment.Text)
	}
	fmt.Println(table)
	fmt.Println()
}

func chaptersPrintFormatted(chapters []S.Chapter) {
	if len(chapters) == 0 {
		fmt.Println("Could not retrieve chapters")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 19)
	table.Separator = " |\t"
	for _, chapter := range chapters {
		start := TransformMsToTimestamp(*chapter.Start)
		end := TransformMsToTimestamp(*chapter.End)
		table.AddRow("| timestamp", fmt.Sprintf("%s-%s", start, end))
		table.AddRow("| Gist", chapter.Gist)
		table.AddRow("| Headline", chapter.Headline)
		table.AddRow("| Summary", chapter.Summary)
		table.AddRow("", "")
	}
	fmt.Println(table)
	fmt.Println()
}

func entityDetectionPrintFormatted(entities []S.Entity) {
	if len(entities) == 0 {
		fmt.Println("Could not retrieve entity detection")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 25)
	table.Separator = " |\t"
	table.AddRow("| type", "text")
	entityMap := make(map[string][]string)
	for _, entity := range entities {
		isAlreadyInMap := false
		for _, text := range entityMap[entity.EntityType] {
			if text == entity.Text {
				isAlreadyInMap = true
				break
			}
		}
		if !isAlreadyInMap {
			entityMap[entity.EntityType] = append(entityMap[entity.EntityType], entity.Text)
		}
	}
	for entityType, entityTexts := range entityMap {
		table.AddRow("| "+entityType, strings.Join(entityTexts, ", "))
	}
	fmt.Println(table)
	fmt.Println()
}

func summaryPrintFormatted(summary *string) {
	if summary == nil {
		fmt.Println("Could not retrieve summary")
		return
	}

	table := uitable.New()
	table.Wrap = true
	table.MaxColWidth = uint(width - 20)
	table.Separator = " |\t"

	table.AddRow(*summary)

	fmt.Println(table)
	fmt.Println()
}

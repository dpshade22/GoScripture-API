package similarity

import (
	"regexp"
)

func BuildVerseMap(embeddingsByVerse []Embedding) map[string]string {
	verseMap := make(map[string]string)

	// Define a regular expression pattern
	pattern := `^([\w\s]+ \d+:\d+)`
	verseNum := `^[\w\s]+ \d+:(\d+)`
	// Create a regular expression object
	regex := regexp.MustCompile(pattern)
	verseRegex := regexp.MustCompile(verseNum)

	for i, e := range embeddingsByVerse {
		matches := regex.FindStringSubmatch(embeddingsByVerse[i].Location)
		regexVerse := verseRegex.FindStringSubmatch(embeddingsByVerse[i].Location)
		if len(matches) > 0 {
			number := regexVerse[1]
			verseMap[matches[0]] = number + " " + e.Verse
		}
	}
	return verseMap
}

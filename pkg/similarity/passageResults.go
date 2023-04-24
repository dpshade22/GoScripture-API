package similarity

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func MergePassageResults(unmergedBestPassageResults []Embedding, query string, verseMap map[string]string) []Embedding {
	chapters := make(map[string][]Tuple)

	// Define a regular expression pattern
	pattern := `^([\w\s]+ \d{1,2}):(\d+)-(\d+)`

	// Create a regular expression object
	regex := regexp.MustCompile(pattern)

	for i := range unmergedBestPassageResults {
		// Test if a string matches the pattern
		matches := regex.FindAllStringSubmatch(unmergedBestPassageResults[i].Location, -1)
		if len(matches) > 0 && len(matches[0]) > 1 {
			_, ok := chapters[matches[0][1]]
			if ok {
				if len(matches[0]) > 2 {
					num, _ := strconv.Atoi(matches[0][2])
					verseSimilarity := Tuple{num, unmergedBestPassageResults[i].Similarity}
					chapters[matches[0][1]] = append(chapters[matches[0][1]], verseSimilarity)
				}
			} else {
				if len(matches[0]) > 2 {
					num, _ := strconv.Atoi(matches[0][2])
					chapters[matches[0][1]] = make([]Tuple, 0)
					verseSimilarity := Tuple{num, unmergedBestPassageResults[i].Similarity}
					chapters[matches[0][1]] = append(chapters[matches[0][1]], verseSimilarity)
				}
			}
		}
	}

	return buildPassageResults(chapters, query, verseMap)
}

func handleLocationQuery(searchBy, query string, embeddings *[]Embedding) {
	hasLoc, loc := checkIfLocation(query)
	locStringChapter := ""
	locStringVerse := ""

	if loc.Verse > 0 && searchBy == "verse" {
		locStringVerse = fmt.Sprintf("%s %d:%d", loc.Book, loc.Chapter, loc.Verse)
		fmt.Print(locStringVerse, "\n")
	} else if loc.Chapter > 0 && searchBy == "chapter" {
		locStringChapter = fmt.Sprintf("%s %d", loc.Book, loc.Chapter)
		fmt.Print(locStringChapter, "\n")
	}

	if hasLoc {
		for i, embed := range *embeddings {
			if searchBy == "chapter" && locStringChapter == embed.Location {
				(*embeddings)[i].Similarity = 0.9999
			} else if searchBy == "verse" && locStringVerse == embed.Location {
				(*embeddings)[i].Similarity = 0.9999
			}
		}
	}
}

func buildPassageResults(chapters map[string][]Tuple, query string, verseMap map[string]string) []Embedding {
	newPassages := make([]Embedding, 0)

	for k, v := range chapters {
		sort.Slice(v, func(i, j int) bool {
			return v[i].First < v[j].First
		})

		startRange := -1
		endRange := -1
		avgSim := 0.0
		runningCount := 0

		for i := 0; i < len(v); i++ {
			// loc := k + ":" + strconv.Itoa(v[i].First)

			if startRange == -1 {
				startRange = v[i].First
			}
			endRange = v[i].First

			avgSim += v[i].Second
			runningCount++

			// Check if the next verse is not consecutive or if it's the last verse in the array
			if i == len(v)-1 || (v[i+1].First != v[i].First+1 && v[i+1].First != v[i].First+2 && v[i+1].First != v[i].First+3) {
				// Rebuild the verses based on the first and last in the current consec range
				consec := ""
				for r := startRange; r <= endRange; r++ {
					loc := k + ":" + strconv.Itoa(r)
					consec += getVerse(loc, verseMap) + " "
				}

				if endRange > startRange { // Check if the passage has more than one verse
					newE := Embedding{
						Location:   k + ":" + strconv.Itoa(startRange) + "-" + strconv.Itoa(endRange),
						Verse:      consec,
						Similarity: avgSim / float64(runningCount),
					}
					newPassages = append(newPassages, newE)
				}

				// Reset avgSim, runningCount, startRange, and endRange for the next passage
				avgSim = 0.0
				runningCount = 0
				startRange = -1
				endRange = -1
			}
		}
	}

	hasLoc, loc := checkIfLocation(query)
	locStringPassage := ""
	if hasLoc {
		locStringPassage = fmt.Sprintf("%s %d:%d-%d", loc.Book, loc.Chapter, loc.Verse, loc.VerseEnd)
		fmt.Print("Has location\n")
		fmt.Print("Found passage\n")
		fmt.Print("Location: ", locStringPassage, "\n")

		newEmbed := buildPassageFromLocation(loc, verseMap)
		if strings.TrimSpace(newEmbed.Verse) != "" {
			newPassages = append(newPassages, newEmbed)

			// Sort bestSequences by Similarity in descending order.
			sort.Slice(newPassages, func(i, j int) bool {
				return newPassages[i].Similarity > newPassages[j].Similarity
			})
			return newPassages[0 : len(newPassages)-1]
		}
	}

	// Sort bestSequences by Similarity in descending order.
	sort.Slice(newPassages, func(i, j int) bool {
		return newPassages[i].Similarity > newPassages[j].Similarity
	})

	return newPassages
}

func buildPassageFromLocation(location *location, verseMap map[string]string) Embedding {
	// Create a new Embedding object
	locString := location.Book + " " + strconv.Itoa(location.Chapter) + ":" + strconv.Itoa(location.Verse)
	consecVerses := ""
	for i := location.Verse; i <= location.VerseEnd; i++ {
		verseStr := strconv.Itoa(i)
		verseStrLength := len(verseStr)
		locWithCurrentVerse := locString[:len(locString)-verseStrLength] + verseStr
		consecVerses += getVerse(locWithCurrentVerse, verseMap) + " "
		fmt.Print("i: ", i, " ")
		fmt.Print(locWithCurrentVerse, "\n")
	}
	embedding := Embedding{
		Location:   locString + "-" + strconv.Itoa(location.VerseEnd),
		Verse:      consecVerses,
		Similarity: 0.9999,
	}

	return embedding
}

func getVerse(location string, verseMap map[string]string) string {
	return verseMap[location]
}

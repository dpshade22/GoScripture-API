package similarity

import (
	"context"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"go-scripture/pkg/embeddings"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

type Embedding = embeddings.Embedding
type Tuple struct {
	First  int
	Second float64
}

func FindSimilarities(query string, embeddings []Embedding) []Embedding {
	godotenv.Load()
	apiKey := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(apiKey)
	fmt.Printf("API key: %s\n", apiKey)

	request := openai.EmbeddingRequest{
		Input: []string{query},
		Model: openai.AdaEmbeddingV2,
	}
	fmt.Println(request)

	resp, err := client.CreateEmbeddings(context.Background(), request)
	if err != nil {
		fmt.Printf("Error creating embeddings: %s", err)
		panic(err)
	}

	searchTermVector32 := resp.Data[0].Embedding
	searchTermVector := make([]float64, len(searchTermVector32))
	for i, v := range searchTermVector32 {
		searchTermVector[i] = float64(v)
	}

	for i := range embeddings {

		embeddings[i].Similarity = cosineSimilarity(embeddings[i].Embedding, searchTermVector)
	}

	sort.Slice(embeddings, func(i, j int) bool {
		return embeddings[i].Similarity > embeddings[j].Similarity
	})

	return embeddings
}

func FindBestPassages(verses []Embedding, windowSize int, numSequences int) []Embedding {

	// Sort the verses list by Location and Verse.
	sort.Slice(verses, func(i, j int) bool {
		if verses[i].Location == verses[j].Location {
			return verses[i].Verse < verses[j].Verse
		}
		return verses[i].Location < verses[j].Location
	})

	var bestSequences []Embedding
	for i := 0; i < numSequences; i++ {
		// Iterate over the verses list using a sliding window of size `windowSize`.
		bestWindow := make([]Embedding, windowSize)
		bestScore := 0.0

		for j := i; j <= len(verses)-windowSize && j >= 0; j += numSequences {
			window := verses[j : j+windowSize]
			// Calculate the average similarity score for all Embedding structs in the window.
			sumScore := 0.0
			for _, e := range window {
				sumScore += e.Similarity
			}
			avgScore := sumScore / float64(windowSize)

			// Update the best window, score, and start index if a higher score is found.
			if avgScore > bestScore {
				copy(bestWindow, window)
				bestScore = avgScore
			}
		}

		// Extract book and chapter from the Location field of the first verse in the best window.
		bookAndChapter := bestWindow[0].Location[:strings.LastIndex(bestWindow[0].Location, ":")]
		verseStart := bestWindow[0].Location[strings.LastIndex(bestWindow[0].Location, ":")+1:]
		verseEnd := bestWindow[len(bestWindow)-1].Location[strings.LastIndex(bestWindow[len(bestWindow)-1].Location, ":")+1:]
		bestLocation := fmt.Sprintf("%s:%s-%s", bookAndChapter, verseStart, verseEnd)

		// Concatenate verses in the best window.
		var bestVerse strings.Builder
		for _, e := range bestWindow {
			if bestVerse.Len() > 0 {
				bestVerse.WriteString(" ")
			}
			verseNum := e.Location[strings.LastIndex(e.Location, ":")+1:]
			verseStr := verseNum + " " + e.Verse

			bestVerse.WriteString(verseStr)
		}

		// Append the best sequence to the list of best sequences.
		bestSequences = append(bestSequences, Embedding{Location: bestLocation, Verse: bestVerse.String(), Similarity: bestScore})
	}

	// Sort bestSequences by Similarity in descending order.
	sort.Slice(bestSequences, func(i, j int) bool {
		return bestSequences[i].Similarity > bestSequences[j].Similarity
	})

	return bestSequences
}

// Functions: cosineSimilarity, findSimilarities, processPassageResults
func cosineSimilarity(a []float64, b []float64) float64 {
	if len(a) != len(b) {
		panic("vector lengths do not match")
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * float64(b[i])
		normA += a[i] * a[i]
		normB += float64(b[i]) * float64(b[i])
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func MergePassageResults(unmergedBestPassageResults []Embedding, verseMap map[string]string) []Embedding {
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

	return buildPassageEmbeddings(chapters, verseMap)
}

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

func buildPassageEmbeddings(chapters map[string][]Tuple, verseMap map[string]string) []Embedding {
	newPassages := make([]Embedding, 0)

	for k, v := range chapters {
		sort.Slice(v, func(i, j int) bool {
			return v[i].First < v[j].First
		})

		firstRange := 0
		consec := ""
		avgSim := 0.0
		runningCount := 0

		for i := 0; i < len(v); i++ {
			loc := k + ":" + strconv.Itoa(v[i].First)
			avgSim += v[i].Second
			runningCount++

			if i < len(v)-1 && (v[i].First+1 == v[i+1].First || v[i].First+2 == v[i+1].First || v[i].First+3 == v[i+1].First) {
				if runningCount == 1 {
					firstRange = v[i].First
				}
				consec += getVerse(loc, verseMap) + " "
				if v[i].First+2 == v[i+1].First {
					consec += getVerse(k+":"+strconv.Itoa(v[i].First+1), verseMap) + " "
				}
			} else {
				newE := Embedding{}
				if runningCount > 1 {
					newE = Embedding{
						Location:   k + ":" + strconv.Itoa(firstRange) + "-" + strconv.Itoa(v[i].First),
						Verse:      consec + getVerse(loc, verseMap),
						Similarity: avgSim / float64(runningCount),
					}
					newPassages = append(newPassages, newE)
				}

				// Reset consec, avgSim, and runningCount for the next passage
				consec = ""
				avgSim = 0.0
				runningCount = 0
				if i < len(v)-1 {
					firstRange = v[i+1].First
				}
			}
		}
	}

	// Sort bestSequences by Similarity in descending order.
	sort.Slice(newPassages, func(i, j int) bool {
		return newPassages[i].Similarity > newPassages[j].Similarity
	})

	return newPassages
}

func getVerse(location string, verseMap map[string]string) string {
	return verseMap[location]
}

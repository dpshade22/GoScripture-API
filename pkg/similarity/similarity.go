package similarity

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"go-scripture/pkg/embeddings"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

type Embedding = embeddings.Embedding

func FindSimilarities(query string, embeddings []Embedding, x int) []Embedding {
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

	if len(embeddings) >= x {
		return embeddings[:x]
	} else {
		return embeddings
	}
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
			bestVerse.WriteString(e.Verse)
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

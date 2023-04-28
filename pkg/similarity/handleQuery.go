package similarity

import (
	"context"
	"fmt"
	"go-scripture/pkg/embeddings"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

type Embedding = embeddings.Embedding
type LocationStruct struct {
	HasLocation bool
	Location    string
	Book        string
	Chapter     int
	Verse       int
	VerseEnd    int
}
type Tuple struct {
	First  int
	Second float64
}

func FindSimilarities(query string, embeddingsByChapter []Embedding, embeddingsByVerse []Embedding, verseMap map[string]string, searchBy string) []Embedding {
	bibleEmbeddings := embeddingsByVerse
	if searchBy == "chapter" {
		bibleEmbeddings = embeddingsByChapter
	}

	loc := checkIfLocation(query)
	if loc.HasLocation {
		query = swapQueryForPassage(query, loc, verseMap)
	}

	searchTermVector := getSearchVector(query, loc, embeddingsByChapter, embeddingsByVerse, searchBy, verseMap)
	similartyResults := calculateEmbeddingSimilarity(bibleEmbeddings, searchTermVector)
	if loc.HasLocation {
		updateExactMatchSimilarity(searchBy, loc, &similartyResults)
	}
	sort.Slice(similartyResults, func(i, j int) bool {
		return similartyResults[i].Similarity > similartyResults[j].Similarity
	})

	return similartyResults
}

func calculateEmbeddingSimilarity(embeddings []Embedding, searchTermVector []float64) []Embedding {
	numWorkers := 8
	jobs := make(chan int, len(embeddings))
	results := make(chan Embedding, len(embeddings))
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			for i := range jobs {
				embeddings[i].Similarity = cosineSimilarity(embeddings[i].Embedding, searchTermVector)
				results <- embeddings[i]
			}
			wg.Done()
		}()
	}
	for i := range embeddings {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	close(results)
	return embeddings
}

func getSearchVector(query string, loc LocationStruct, embeddingsByChapter []Embedding, embeddingsByVerse []Embedding, searchBy string, verseMap map[string]string) []float64 {
	vector := make([]float64, len(embeddingsByVerse[0].Embedding))
	foundLocalEmbedding := false
	if loc.HasLocation {
		switch searchBy {
		case "verse":
			foundLocalEmbedding, vector = getEmbeddingByLocation(loc.Book+" "+strconv.Itoa(loc.Chapter)+":"+strconv.Itoa(loc.Verse), embeddingsByVerse)
		case "passage":
			foundLocalEmbedding, vector = getEmbeddingByLocation(loc.Book+" "+strconv.Itoa(loc.Chapter)+":"+strconv.Itoa(loc.Verse)+"-"+strconv.Itoa(loc.VerseEnd), embeddingsByVerse)
		case "chapter":
			foundLocalEmbedding, vector = getEmbeddingByLocation(loc.Book+" "+strconv.Itoa(loc.Chapter), embeddingsByChapter)
		}
	}
	if !foundLocalEmbedding {
		vector = getQueryEmbedding(query)
	}
	return vector
}

func getQueryEmbedding(query string) []float64 {
	godotenv.Load()
	apiKey := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(apiKey)
	request := openai.EmbeddingRequest{
		Input: []string{query},
		Model: openai.AdaEmbeddingV2,
	}
	resp, err := client.CreateEmbeddings(context.Background(), request)
	if err != nil {
		fmt.Printf("Error creating embeddings: %s", err)
		panic(err)
	}
	embedding := make([]float64, len(resp.Data[0].Embedding))
	for i, v := range resp.Data[0].Embedding {
		embedding[i] = float64(v)
	}
	return embedding
}

func swapQueryForPassage(query string, loc LocationStruct, verseMap map[string]string) string {
	fmt.Print("User Input: " + query + "\n")
	// Check if the query is a valid Bible verse, passage, or chapter

	newVerseQuery := ""

	if loc.HasLocation {
		if loc.VerseEnd > 0 && loc.VerseEnd > loc.Verse {
			newVerseQuery = buildPassageFromLocation(loc, verseMap).Verse
			fmt.Print("New Query: " + newVerseQuery + "\n")
			return newVerseQuery
		}
	}
	return query
}

func getEmbeddingByLocation(location string, embeddings []Embedding) (bool, []float64) {
	for _, embedding := range embeddings {
		if embedding.Location == location {
			fmt.Print("FOUND EMBEDDING: " + embedding.Location + "\n")
			return true, embedding.Embedding
		}
	}
	return false, []float64{}
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

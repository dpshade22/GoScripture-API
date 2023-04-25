package similarity

import (
	"context"
	"fmt"
	"go-scripture/pkg/embeddings" // package for working with Bible verse embeddings
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"          // package for working with environment variables
	"github.com/sashabaranov/go-openai" // package for working with OpenAI API
)

type Embedding = embeddings.Embedding // define an alias for Embedding struct from embeddings package
type Tuple struct {
	First  int
	Second float64
}

func FindSimilarities(query string, embeddingsByChapter []Embedding, embeddingsByVerse []Embedding, verseMap map[string]string, searchBy string) []Embedding {
	embeddings := embeddingsByVerse
	if searchBy == "chapter" {
		embeddings = embeddingsByChapter
	}

	// create a vector representation of the query using pre-trained embeddings
	// or using the embeddings of the verse location in the query
	searchTermVector := make([]float64, len(embeddingsByVerse[0].Embedding))
	hasLoc, loc := checkIfLocation(query)
	foundLocalEmbedding := false
	if hasLoc {
		switch searchBy {
		case "verse":
			foundLocalEmbedding, searchTermVector = getEmbeddingByLocation(loc.Book+" "+strconv.Itoa(loc.Chapter)+":"+strconv.Itoa(loc.Verse), embeddingsByVerse)
		case "chapter":
			foundLocalEmbedding, searchTermVector = getEmbeddingByLocation(loc.Book+" "+strconv.Itoa(loc.Chapter), embeddingsByChapter)
		}
	}

	if !foundLocalEmbedding {
		// use OpenAI API to get the embeddings for the query
		// load API key from environment variable
		// swap query for passage if it is a valid Bible passage
		// create an embedding request using the query and pre-trained embeddings model
		// get the embeddings for the query from the OpenAI API response

		godotenv.Load()
		apiKey := os.Getenv("OPENAI_API_KEY")
		client := openai.NewClient(apiKey)
		fmt.Printf("API key: %s\n", apiKey)

		query = swapQueryForPassage(query, verseMap)

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
		for i, v := range searchTermVector32 {
			searchTermVector[i] = float64(v)
		}
	}

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

	// sort the embeddings by similarity score in descending order
	handleLocationQuery(searchBy, query, &embeddings)
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

func swapQueryForPassage(query string, verseMap map[string]string) string {
	fmt.Print("User Input: " + query + "\n")
	// Check if the query is a valid Bible verse, passage, or chapter
	hasLoc, loc := checkIfLocation(query)
	newVerseQuery := ""

	if hasLoc {
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

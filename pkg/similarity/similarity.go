package similarity

import (
	"context"
	"math"

	"github.com/sashabaranov/go-openai"
)

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

func processPassageResults(found []Embedding, x int) []Embedding {
	var passages []Embedding
	var tempPassage []Embedding

	for i := 0; i < len(found)-1; i++ {
		tempPassage = append(tempPassage, found[i])

		// Split location strings into components
		currLocParts := strings.Split(found[i].Location, ":")
		nextLocParts := strings.Split(found[i+1].Location, ":")

		currVerse, _ := strconv.Atoi(currLocParts[1])
		nextVerse, _ := strconv.Atoi(nextLocParts[1])

		// Check if the next verse is within the allowed gap range
		if nextVerse-currVerse > x+1 {
			// If the current temp passage is longer than the stored passage, replace it
			if len(tempPassage) > len(passages) {
				passages = tempPassage
			}
			// Reset temp passage
			tempPassage = []Embedding{}
		}
	}

	// Check if the last temp passage is longer than the stored passage
	if len(tempPassage) > len(passages) {
		passages = tempPassage
	}

	return passages
}


func findSimilarities(query string, embeddings []Embedding, x int) []Embedding {
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

	return embeddings[:50]
}

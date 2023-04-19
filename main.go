package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-gota/gota/dataframe"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

type Embedding struct {
	Location   string
	Verse      string
	Embedding  []float64
	Similarity float64
}

func loadEmbeddings(embeddingByChapterCSV, embeddingByVerseCSV string) ([]Embedding, []Embedding) {
	embeddingsByChapter := loadEmbeddingsFromFile(embeddingByChapterCSV, "chapter")
	embeddingsByVerse := loadEmbeddingsFromFile(embeddingByVerseCSV, "verse")

	return embeddingsByChapter, embeddingsByVerse
}

func loadEmbeddingsFromFile(file string, db string) []Embedding {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	embedCol := 5
	if db == "verse" {
		embedCol = 4
	}

	// Use gota to read the CSV file into a DataFrame
	df := dataframe.ReadCSV(f)

	var embeddings []Embedding
	for i := 0; i < df.Nrow(); i++ {
		// Extract the values for each row
		location := ""
		if db == "chapter" {
			book := df.Elem(i, 1).String()
			chapter := df.Elem(i, 2).String()
			location = fmt.Sprintf("%s %s", book, chapter)
		} else {
			location = df.Elem(i, 1).String()
		}

		verse := df.Elem(i, embedCol-2).String()
		embeddingStr := df.Elem(i, embedCol).String()

		// Parse the embedding string into a slice of float64
		embeddingStr = strings.TrimPrefix(embeddingStr, "[")
		embeddingStr = strings.TrimSuffix(embeddingStr, "]")
		embeddingValues := strings.Split(embeddingStr, ", ")

		embedding := make([]float64, len(embeddingValues))
		for j, v := range embeddingValues {
			f, _ := strconv.ParseFloat(v, 64)
			embedding[j] = f
		}

		// Append the Embedding struct to the embeddings slice
		embeddings = append(embeddings, Embedding{
			Location:  location,
			Verse:     verse,
			Embedding: embedding,
		})
	}
	return embeddings
}

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

func findSimilarities(query string, embeddings []Embedding) []Embedding {
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

// loggingMiddleware is a custom middleware function that logs incoming HTTP requests.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the details of the incoming request.
		log.Printf("Incoming request: Method=%s, Path=%s, RemoteAddr=%s, Time=%s\n",
			r.Method, r.URL.Path, r.RemoteAddr, time.Now().Format(time.RFC3339))

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

func main() {
	godotenv.Load()
	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	fmt.Println("Loading embeddings...")
	embeddingsByChapter, embeddingsByVerse := loadEmbeddings("embeddings/chapter/KJV_Bible_Embeddings_by_Chapter.csv", "embeddings/verse/KJV_Bible_Embeddings.csv")
	fmt.Println("Embeddings loaded")

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"message": "Hello World"})
	}).Methods("GET")

	router.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		searchBy := r.URL.Query().Get("search_by")
		query := r.URL.Query().Get("query")

		if searchBy != "" && query != "" {
			embeddings := embeddingsByChapter
			if searchBy == "verse" {
				embeddings = embeddingsByVerse
			}

			found := findSimilarities(query, embeddings)

			jsonArray := make([]map[string]interface{}, len(found))
			for i, e := range found {
				jsonArray[i] = map[string]interface{}{
					"index":        i,
					"location":     e.Location,
					"verse":        e.Verse,
					"similarities": e.Similarity,
				}
			}
			fmt.Println(jsonArray)
			fmt.Printf("Search by: %s, Query: %s", searchBy, query)
			json.NewEncoder(w).Encode(jsonArray)
		} else {
			http.Error(w, "Missing query parameters 'search_by' and 'query'", http.StatusBadRequest)
		}

	}).Methods("GET")

	fmt.Println("Server running on http://0.0.0.0:8080")
	http.ListenAndServe("0.0.0.0:8080", router)
}

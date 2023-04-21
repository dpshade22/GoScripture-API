package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go-scripture/pkg/api"
	"go-scripture/pkg/embeddings"
	"go-scripture/pkg/middleware"
	"go-scripture/pkg/similarity"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware)

	fmt.Println("Loading embeddings...")
	embeddingsByChapter, embeddingsByVerse := embeddings.LoadEmbeddings("embeddingsData/chapter/KJV_Bible_Embeddings_by_Chapter.csv", "embeddingsData/verse/KJV_Bible_Embeddings.csv")
	fmt.Println("Embeddings loaded")

	fmt.Printf("Building verse map...")
	verseMap := similarity.BuildVerseMap(embeddingsByVerse)
	fmt.Printf("Verse map built")
	// fmt.Println(verseMap)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"message": "Hello World"})
	}).Methods("GET")

	router.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		api.HandleSearch(w, r, embeddingsByChapter, embeddingsByVerse, verseMap)
	}).Methods("GET")

	fmt.Println("Server running on http://0.0.0.0:8080")
	http.ListenAndServe("0.0.0.0:8080", router)
}

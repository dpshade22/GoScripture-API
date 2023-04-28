package main

import (
	"encoding/json"
	"fmt"
	"go-scripture/pkg/api"
	"go-scripture/pkg/embeddings"
	"go-scripture/pkg/middleware"
	"go-scripture/pkg/similarity"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	godotenv.Load()
	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware)

	fmt.Println("Loading embeddings...")
	embeddingsByChapter, embeddingsByVerse := embeddings.LoadEmbeddings("embeddingsData/chapter/KJV_Bible_Embeddings_by_Chapter.csv", "embeddingsData/verse/KJV_Bible_Embeddings.csv")
	fmt.Println("Embeddings loaded")

	fmt.Printf("Building verse map...\n")
	verseMap := similarity.BuildVerseMap(embeddingsByVerse)
	fmt.Printf("Verse map built\n")

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"message": "Hello World"})
	}).Methods("GET")

	router.HandleFunc("/search/verse", func(w http.ResponseWriter, r *http.Request) {
		api.HandleSearchByVerse(w, r, embeddingsByVerse)
	}).Methods("GET").Queries("book", "{book}", "chapter", "{chapter}", "verse", "{verse}")
	
	router.HandleFunc("/search/chapter", func(w http.ResponseWriter, r *http.Request) {
		api.HandleSearchByChapter(w, r, embeddingsByChapter)
	}).Methods("GET").Queries("book", "{book}", "chapter", "{chapter}")
	
	router.HandleFunc("/search/passage", func(w http.ResponseWriter, r *http.Request) {
		api.HandleSearchByPassage(w, r, embeddingsByChapter, embeddingsByVerse, verseMap)
	}).Methods("GET").Queries("book", "{book}", "chapter", "{chapter}", "verse", "{verse}", "verse_end", "{verse_end}")
	
	router.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		api.HandleQuery(w, r, embeddingsByChapter, embeddingsByVerse, verseMap)
	}).Methods("GET").Queries("search_by", "{search_by}", "query", "{query}")
	

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	})

	handler := corsHandler.Handler(router)

	fmt.Println("Server running on http://0.0.0.0:8080")
	http.ListenAndServe("0.0.0.0:8080", handler)
}

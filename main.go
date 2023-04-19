package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"GoScipture-API/pkg/api"
	"GoScipture-API/pkg/embeddings"
	"GoScipture-API/pkg/middleware"
)

func main() {
	godotenv.Load()
	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware)

	fmt.Println("Loading embeddings...")
	embeddingsByChapter, embeddingsByVerse := embeddings.LoadEmbeddings("embeddings/chapter/KJV_Bible_Embeddings_by_Chapter.csv", "embeddings/verse/KJV_Bible_Embeddings.csv")
	fmt.Println("Embeddings loaded")

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"message": "Hello World"})
	}).Methods("GET")

	router.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		api.HandleSearch(w, r, embeddingsByChapter, embeddingsByVerse)
	}).Methods("GET")

	fmt.Println("Server running on http://0.0.0.0:8080")
	http.ListenAndServe("0.0.0.0:8080", router)
}

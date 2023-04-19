package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/your_username/your_project/pkg/embeddings"
	"github.com/your_username/your_project/pkg/similarity"
)

// Functions: handleSearch
func handleSearch(w http.ResponseWriter, r *http.Request, embeddingsByChapter, embeddingsByVerse []Embedding) {
	searchBy := r.URL.Query().Get("search_by")
	query := r.URL.Query().Get("query")
	xStr := r.URL.Query().Get("x")

	if searchBy != "" && query != "" {
		embeddings := embeddingsByChapter
		if searchBy == "verse" || searchBy == "passage" {
			embeddings = embeddingsByVerse
		}

		x, err := strconv.Atoi(xStr)
		if err != nil {
			x = 0
		}

		found := findSimilarities(query, embeddings, x)

		if searchBy == "passage" {
			found = processPassageResults(found, x)
		}

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
}
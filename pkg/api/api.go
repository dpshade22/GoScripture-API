package api

import (
	"encoding/json"
	"fmt"
	"go-scripture/pkg/embeddings"
	"go-scripture/pkg/similarity"
	"net/http"
	"strconv"
)

type Embedding = embeddings.Embedding

// Functions: handleSearch
func HandleSearch(w http.ResponseWriter, r *http.Request, embeddingsByChapter, embeddingsByVerse []Embedding, verseMap map[string][]Embedding) {
	searchBy := r.URL.Query().Get("search_by")
	query := r.URL.Query().Get("query")
	xStr := r.URL.Query().Get("x")

	if searchBy != "" && query != "" {
		embeddings := embeddingsByVerse
		if searchBy == "chapter" {
			embeddings = embeddingsByChapter
		}

		x, err := strconv.Atoi(xStr)
		if err != nil {
			x = 0
		}

		found := similarity.FindSimilarities(query, embeddings, x)

		if searchBy == "passage" {
			found = similarity.FindBestPassages(found, 2, 500)
			found = similarity.MergePassageResults(found)
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
		// fmt.Println(jsonArray)
		fmt.Printf("Search by: %s, Query: %s", searchBy, query)
		json.NewEncoder(w).Encode(jsonArray)
	} else {
		http.Error(w, "Missing query parameters 'search_by' and 'query'", http.StatusBadRequest)
	}
}

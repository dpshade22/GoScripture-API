package api

import (
	"encoding/json"
	"fmt"
	"go-scripture/pkg/embeddings"
	"go-scripture/pkg/similarity"
	"net/http"
)

type Embedding = embeddings.Embedding

// Functions: handleSearch
func HandleSearch(w http.ResponseWriter, r *http.Request, embeddingsByChapter, embeddingsByVerse []Embedding, verseMap map[string]string) {
	searchBy := r.URL.Query().Get("search_by")
	query := r.URL.Query().Get("query")

	if searchBy != "" && query != "" {
		embeddings := embeddingsByVerse
		if searchBy == "chapter" {
			embeddings = embeddingsByChapter
		}

		found := similarity.FindSimilarities(query, embeddings)

		if searchBy == "passage" {
			found = similarity.FindBestPassages(found, 2, 600)
			found = similarity.MergePassageResults(found, verseMap)
		} else {
			found = found[:50]
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

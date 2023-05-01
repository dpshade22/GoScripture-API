package api

import (
	"encoding/json"
	"fmt"
	"go-scripture/pkg/embeddings"
	"go-scripture/pkg/similarity"
	"net/http"

	"github.com/gorilla/mux"
)

type Embedding = embeddings.Embedding

// Functions: handleSearch

func HandleSearchByVerse(w http.ResponseWriter, r *http.Request, embeddingsByChapter []Embedding, embeddingsByVerse []Embedding, verseMap map[string]string) {
	vars := mux.Vars(r)
	book := vars["book"]
	chapter := vars["chapter"]
	verse := vars["verse"]
	locationQuery := fmt.Sprintf("%s %s:%s", book, chapter, verse)

	found := similarity.FindSimilarities(locationQuery, embeddingsByChapter, embeddingsByVerse, verseMap, "verse")

	jsonArray := make([]map[string]interface{}, len(found))
	for i, e := range found {
		jsonArray[i] = map[string]interface{}{
			"index":        i,
			"location":     e.Location,
			"verse":        e.Verse,
			"similarities": e.Similarity,
		}
	}

	fmt.Printf("Search by verse: %s", locationQuery)
	json.NewEncoder(w).Encode(jsonArray)
}

func HandleSearchByChapter(w http.ResponseWriter, r *http.Request, embeddingsByChapter []Embedding, embeddingsByVerse []Embedding, verseMap map[string]string) {
	vars := mux.Vars(r)
	book := vars["book"]
	chapter := vars["chapter"]
	locationQuery := fmt.Sprintf("%s %s", book, chapter)

	found := similarity.FindSimilarities(locationQuery, embeddingsByChapter, embeddingsByVerse, verseMap, "chapter")

	jsonArray := make([]map[string]interface{}, len(found))
	for i, e := range found {
		jsonArray[i] = map[string]interface{}{
			"index":        i,
			"location":     e.Location,
			"verse":        e.Verse,
			"similarities": e.Similarity,
		}
	}

	fmt.Printf("Search by chapter: %s", locationQuery)
	json.NewEncoder(w).Encode(jsonArray)
}

func HandleSearchByPassage(w http.ResponseWriter, r *http.Request, embeddingsByChapter []Embedding, embeddingsByVerse []Embedding, verseMap map[string]string) {
	vars := mux.Vars(r)
	book := vars["book"]
	chapter := vars["chapter"]
	verseStart := vars["verse"]
	verseEnd := vars["verse_end"]
	locationQuery := fmt.Sprintf("%s %s:%s-%s", book, chapter, verseStart, verseEnd)

	found := similarity.FindSimilarities(locationQuery, embeddingsByChapter, embeddingsByVerse, verseMap, "passage")
	found = similarity.FindBestPassages(found, 2, 200)
	found = similarity.MergePassageResults(found, locationQuery, verseMap)

	jsonArray := make([]map[string]interface{}, len(found))
	for i, e := range found {
		jsonArray[i] = map[string]interface{}{
			"index":        i,
			"location":     e.Location,
			"verse":        e.Verse,
			"similarities": e.Similarity,
		}
	}

	fmt.Printf("Search by passage: %s", locationQuery)
	json.NewEncoder(w).Encode(jsonArray)
}

func HandleQuery(w http.ResponseWriter, r *http.Request, embeddingsByChapter []Embedding, embeddingsByVerse []Embedding, verseMap map[string]string) {
	vars := mux.Vars(r)
	searchBy := vars["search_by"]
	query := vars["query"]

	if searchBy == "" || query == "" {
		http.Error(w, "Missing query parameters 'search_by' and 'query'", http.StatusBadRequest)
		return
	}

	found := similarity.FindSimilarities(query, embeddingsByChapter, embeddingsByVerse, verseMap, searchBy)

	if searchBy == "passage" {
		found = similarity.FindBestPassages(found, 2, 200)
		found = similarity.MergePassageResults(found, query, verseMap)
	} else {
		found = found[:50]
	}

	type SearchResult struct {
		Index      int     `json:"index"`
		Location   string  `json:"location"`
		Verse      string  `json:"verse"`
		Similarity float64 `json:"similarities"`
	}

	var searchResults []SearchResult
	for i, e := range found {
		searchResults = append(searchResults, SearchResult{
			Index:      i,
			Location:   e.Location,
			Verse:      e.Verse,
			Similarity: e.Similarity,
		})
	}

	// fmt.Println(jsonArray)
	fmt.Printf("Search by: %s, Query: %s", searchBy, query)
	json.NewEncoder(w).Encode(searchResults)
}

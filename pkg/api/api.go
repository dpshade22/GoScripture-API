package api

import (
	"fmt"
	"go-scripture/pkg/embeddings"
	"go-scripture/pkg/similarity"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type Embedding = embeddings.Embedding

type SearchOutput struct {
	Index        int     `json:"index"`
	Location     string  `json:"location"`
	Verse        string  `json:"verse"`
	Similarities float64 `json:"similarities"`
}

func HandleSearchByVerse(c echo.Context, embeddingsByChapter []Embedding, embeddingsByVerse []Embedding, verseMap map[string]string) error {
	book := c.QueryParam("book")
	chapter := c.QueryParam("chapter")
	verse := c.QueryParam("verse")
	locationQuery := fmt.Sprintf("%s %s:%s", book, chapter, verse)

	found := similarity.FindSimilarities(locationQuery, embeddingsByChapter, embeddingsByVerse, verseMap, "verse")

	var searchResults []SearchOutput
	for i, e := range found {
		searchResults = append(searchResults, SearchOutput{
			Index:        i,
			Location:     e.Location,
			Verse:        e.Verse,
			Similarities: e.Similarity,
		})
	}

	fmt.Printf("Search by verse: %s", locationQuery)
	return c.JSON(http.StatusOK, searchResults)
}

func HandleSearchByChapter(c echo.Context, embeddingsByChapter []Embedding, embeddingsByVerse []Embedding, verseMap map[string]string) error {
	book := c.QueryParam("book")
	chapter := c.QueryParam("chapter")
	locationQuery := fmt.Sprintf("%s %s", book, chapter)

	found := similarity.FindSimilarities(locationQuery, embeddingsByChapter, embeddingsByVerse, verseMap, "chapter")

	var searchResults []SearchOutput
	for i, e := range found {
		searchResults = append(searchResults, SearchOutput{
			Index:        i,
			Location:     e.Location,
			Verse:        e.Verse,
			Similarities: e.Similarity,
		})
	}

	fmt.Printf("Search by chapter: %s", locationQuery)
	return c.JSON(http.StatusOK, searchResults)
}

func HandleSearchByPassage(c echo.Context, embeddingsByChapter []Embedding, embeddingsByVerse []Embedding, verseMap map[string]string) error {
	book := c.QueryParam("book")
	chapter := c.QueryParam("chapter")
	verseStart := c.QueryParam("verseStart")
	verseEnd := c.QueryParam("verseEnd")
	locationQuery := fmt.Sprintf("%s %s:%s-%s", book, chapter, verseStart, verseEnd)

	found := similarity.FindSimilarities(locationQuery, embeddingsByChapter, embeddingsByVerse, verseMap, "passage")
	found = similarity.FindBestPassages(found, 2, 200)
	found = similarity.MergePassageResults(found, locationQuery, verseMap)

	var searchResults []SearchOutput
	for i, e := range found {
		searchResults = append(searchResults, SearchOutput{
			Index:        i,
			Location:     e.Location,
			Verse:        e.Verse,
			Similarities: e.Similarity,
		})
	}

	fmt.Printf("Search by passage: %s", locationQuery)
	return c.JSON(http.StatusOK, searchResults)
}

func HandleQuery(c echo.Context, embeddingsByChapter []Embedding, embeddingsByVerse []Embedding, verseMap map[string]string) error {
	searchBy := c.QueryParam("search_by")
	query := c.QueryParam("query")

	if searchBy == "" || query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing query parameters 'search_by' and 'query'")
	}

	found := similarity.FindSimilarities(query, embeddingsByChapter, embeddingsByVerse, verseMap, searchBy)

	if searchBy == "passage" {
		found = similarity.FindBestPassages(found, 2, 200)
		found = similarity.MergePassageResults(found, query, verseMap)
	} else {
		found = found[:50]
	}

	var searchResults []SearchOutput
	for i, e := range found {
		searchResults = append(searchResults, SearchOutput{
			Index:        i,
			Location:     e.Location,
			Verse:        e.Verse,
			Similarities: e.Similarity,
		})
	}

	fmt.Printf("Search by: %s, Query: %s\n", searchBy, query)
	return c.JSON(http.StatusOK, searchResults)
}

func HandleSearchAll(c echo.Context, embeddingsByChapter []Embedding, embeddingsByVerse []Embedding, verseMap map[string]string) error {
	query := c.QueryParam("query")

	var wg sync.WaitGroup
	verseFoundCh := make(chan []Embedding)
	chapterFoundCh := make(chan []Embedding)
	passageFoundCh := make(chan []Embedding)

	wg.Add(3)

	go func() {
		defer wg.Done()
		verseStart := time.Now()
		verseFound := similarity.FindSimilarities(query, embeddingsByChapter, embeddingsByVerse, verseMap, "verse")
		fmt.Printf("Search by Verse took: %s\n", time.Since(verseStart))
		verseFoundCh <- verseFound
	}()

	go func() {
		defer wg.Done()
		chapterStart := time.Now()
		chapterFound := similarity.FindSimilarities(query, embeddingsByChapter, embeddingsByVerse, verseMap, "chapter")
		fmt.Printf("Search by Chapter took: %s\n", time.Since(chapterStart))
		chapterFoundCh <- chapterFound
	}()

	go func() {
		defer wg.Done()
		passageStart := time.Now()
		passageFound := similarity.FindSimilarities(query, embeddingsByChapter, embeddingsByVerse, verseMap, "passage")
		passageFound = similarity.FindBestPassages(passageFound, 2, 200)
		passageFound = similarity.MergePassageResults(passageFound, query, verseMap)
		fmt.Printf("Search by Passage took: %s\n", time.Since(passageStart))
		passageFoundCh <- passageFound
	}()

	go func() {
		wg.Wait()
		close(verseFoundCh)
		close(chapterFoundCh)
		close(passageFoundCh)
	}()

	verseFound := <-verseFoundCh
	chapterFound := <-chapterFoundCh
	passageFound := <-passageFoundCh

	// Combine all results and sort them by similarity

	allFound := append(verseFound, append(chapterFound, passageFound...)...)
	sort.Slice(allFound, func(i, j int) bool {
		return allFound[i].Similarity > allFound[j].Similarity
	})
	allFound = allFound[:50]

	var searchResults []SearchOutput
	for i, e := range allFound {
		searchResults = append(searchResults, SearchOutput{
			Index:        i,
			Location:     e.Location,
			Verse:        e.Verse,
			Similarities: e.Similarity,
		})
	}

	fmt.Printf("Search All by: %s\n", query)
	return c.JSON(http.StatusOK, searchResults)
}

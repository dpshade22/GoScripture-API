package main

import (
	"fmt"
	"go-scripture/pkg/api"
	"go-scripture/pkg/embeddings"
	"go-scripture/pkg/similarity"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	godotenv.Load()
	e := echo.New()

	e.Use(middleware.Logger())

	fmt.Println("Loading embeddings...")
	embeddingsByChapter, embeddingsByVerse := embeddings.LoadEmbeddings("embeddingsData/chapter/KJV_Bible_Embeddings_by_Chapter.csv", "embeddingsData/verse/KJV_Bible_Embeddings.csv")
	fmt.Println("Embeddings loaded")

	fmt.Printf("Building verse map...\n")
	verseMap := similarity.BuildVerseMap(embeddingsByVerse)
	fmt.Printf("Verse map built\n")

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Hello World"})
	})

	e.GET("/search/verse", func(c echo.Context) error {
		return api.HandleSearchByVerse(c, embeddingsByChapter, embeddingsByVerse, verseMap)
	})

	e.GET("/search/chapter", func(c echo.Context) error {
		return api.HandleSearchByChapter(c, embeddingsByChapter, embeddingsByVerse, verseMap)
	})

	e.GET("/search/passage", func(c echo.Context) error {
		return api.HandleSearchByPassage(c, embeddingsByChapter, embeddingsByVerse, verseMap)
	})

	e.GET("/search", func(c echo.Context) error {
		return api.HandleQuery(c, embeddingsByChapter, embeddingsByVerse, verseMap)
	})

	e.GET("/search/all", func(c echo.Context) error {
		return api.HandleSearchAll(c, embeddingsByChapter, embeddingsByVerse, verseMap)
	})

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	e.Logger.Fatal(e.Start(":8080"))
}

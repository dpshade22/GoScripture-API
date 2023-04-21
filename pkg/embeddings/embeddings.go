package embeddings

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-gota/gota/dataframe"
)

type Embedding struct {
	Location   string
	Verse      string
	Embedding  []float64
	Similarity float64
	Index      int
}

// Functions: loadEmbeddings, loadEmbeddingsFromFile
func LoadEmbeddings(embeddingByChapterCSV, embeddingByVerseCSV string) ([]Embedding, []Embedding) {
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
			Index:     i,
		})
	}
	return embeddings
}

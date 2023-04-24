package similarity

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type location struct {
	Book     string
	Chapter  int
	Verse    int
	VerseEnd int
}

// List of valid Bible book names
var bibleBooks = []string{
	"1 Kings", "2 Kings", "1 Chronicles", "2 Chronicles", "Ezra",
	"1 Corinthians", "2 Corinthians", "Galatians", "Ephesians", "Philippians",
	"Titus", "Philemon", "Hebrews", "James", "1 Peter",
	"2 Peter", "1 John", "2 John", "3 John", "Jude",
	"Colossians", "1 Thessalonians", "2 Thessalonians", "1 Timothy", "2 Timothy",
	"Joshua", "Judges", "Ruth", "1 Samuel", "2 Samuel",
	"Genesis", "Exodus", "Leviticus", "Numbers", "Deuteronomy",
	"Nehemiah", "Esther", "Job", "Psalms", "Proverbs",
	"Ecclesiastes", "Song of Solomon", "Isaiah", "Jeremiah", "Lamentations",
	"Ezekiel", "Daniel", "Hosea", "Joel", "Amos",
	"Obadiah", "Jonah", "Micah", "Nahum", "Habakkuk",
	"Zephaniah", "Haggai", "Zechariah", "Malachi", "Matthew",
	"Mark", "Luke", "John", "Acts", "Romans",
	"Revelation",
}

// isValidBibleBook checks if the given input string contains any valid Bible book name
func isValidBibleBook(input string) (bool, string) {
	// Normalize the input string to lower case for case-insensitive comparison
	normalizedInput := strings.ToLower(input)

	// Iterate over the valid Bible book names and check if any of them match the input string
	for _, bibleBook := range bibleBooks {
		// Create a regex pattern for the current Bible book name
		// The pattern allows any number of characters between the words in the book name
		pattern := ".*" + strings.Join(strings.Fields(strings.ToLower(bibleBook)), ".*")
		// Compile the regex pattern
		re, err := regexp.Compile(pattern)
		if err != nil {
			fmt.Println("Error compiling regex:", err)
			return false, ""
		}

		// Check if the input string matches the regex pattern
		if re.MatchString(normalizedInput) {
			return true, bibleBook
		}
	}
	return false, ""
}

func checkIfLocation(query string) (bool, *location) {
	// Define a regular expression pattern
	pattern := `([\w\s]+?)(\d+)(?:[:\s-](\d+))?(?:[:\s-](\d+))?`
	// Create a regular expression object
	regex := regexp.MustCompile(pattern)

	// Test if a string matches the pattern
	matches := regex.FindStringSubmatch(query)

	if len(matches) > 0 {
		hasBook, bookName := isValidBibleBook(matches[1])

		if !hasBook {
			return false, nil
		}

		chapter, _ := strconv.Atoi(matches[2])
		verse, _ := strconv.Atoi(matches[3])
		verseEnd, _ := strconv.Atoi(matches[4])

		loc := &location{
			Book:     bookName,
			Chapter:  chapter,
			Verse:    verse,
			VerseEnd: verseEnd,
		}
		return true, loc
	}

	return false, nil
}

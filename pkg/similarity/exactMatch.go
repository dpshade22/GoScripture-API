package similarity

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

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

var alternativeBookNames = map[string]string{
	"Psalm":   "Psalms",
	"Pslam":   "Psalms",
	"Pslams":  "Psalms",
	"Gen":     "Genesis",
	"Ex":      "Exodus",
	"Lev":     "Leviticus",
	"Num":     "Numbers",
	"Deut":    "Deuteronomy",
	"Josh":    "Joshua",
	"Judg":    "Judges",
	"Ruth":    "Ruth",
	"1 Sam":   "1 Samuel",
	"2 Sam":   "2 Samuel",
	"1 Ki":    "1 Kings",
	"2 Ki":    "2 Kings",
	"1 Chr":   "1 Chronicles",
	"2 Chr":   "2 Chronicles",
	"1Sam":    "1 Samuel",
	"2Sam":    "2 Samuel",
	"1Ki":     "1 Kings",
	"2Ki":     "2 Kings",
	"1Chr":    "1 Chronicles",
	"2Chr":    "2 Chronicles",
	"Ezr":     "Ezra",
	"Neh":     "Nehemiah",
	"Est":     "Esther",
	"Prov":    "Proverbs",
	"Eccl":    "Ecclesiastes",
	"Song":    "Song of Solomon",
	"Isa":     "Isaiah",
	"Jer":     "Jeremiah",
	"Lam":     "Lamentations",
	"Ezek":    "Ezekiel",
	"Dan":     "Daniel",
	"Hos":     "Hosea",
	"Am":      "Amos",
	"Ob":      "Obadiah",
	"Jon":     "Jonah",
	"Mic":     "Micah",
	"Nah":     "Nahum",
	"Hab":     "Habakkuk",
	"Zeph":    "Zephaniah",
	"Hag":     "Haggai",
	"Zech":    "Zechariah",
	"Mal":     "Malachi",
	"Matt":    "Matthew",
	"Mk":      "Mark",
	"Lk":      "Luke",
	"Jn":      "John",
	"Rom":     "Romans",
	"1 Cor":   "1 Corinthians",
	"2 Cor":   "2 Corinthians",
	"1Cor":    "1 Corinthians",
	"2Cor":    "2 Corinthians",
	"Gal":     "Galatians",
	"Eph":     "Ephesians",
	"Phil":    "Philippians",
	"Col":     "Colossians",
	"1 Thess": "1 Thessalonians",
	"2 Thess": "2 Thessalonians",
	"1 Tim":   "1 Timothy",
	"2 Tim":   "2 Timothy",
	"1Thess":  "1 Thessalonians",
	"2Thess":  "2 Thessalonians",
	"1Tim":    "1 Timothy",
	"2Tim":    "2 Timothy",
	"Tit":     "Titus",
	"Phlm":    "Philemon",
	"Heb":     "Hebrews",
	"Jas":     "James",
	"1 Pet":   "1 Peter",
	"2 Pet":   "2 Peter",
	"1 Jn":    "1 John",
	"2 Jn":    "2 John",
	"3 Jn":    "3 John",
	"1Pet":    "1 Peter",
	"2Pet":    "2 Peter",
	"1Jn":     "1 John",
	"2Jn":     "2 John",
	"3Jn":     "3 John",
	"Rev":     "Revelation",
}

// createBookNameMap creates a map with both original and alternative book names as keys
func createBookNameMap() map[string]string {
	bookNameMap := make(map[string]string)
	for _, book := range bibleBooks {
		bookNameMap[book] = book
		bookNameMap[strings.ToLower(book)] = book
	}
	for alt, orig := range alternativeBookNames {
		bookNameMap[alt] = orig
		bookNameMap[strings.ToLower(alt)] = orig
	}
	return bookNameMap
}

var bookNameMap = createBookNameMap()

// isValidBibleBook checks if the given input string contains any valid Bible book name
func isValidBibleBook(input string) (bool, string) {
	// Normalize the input string to lower case for case-insensitive comparison
	normalizedInput := strings.ToLower(input)

	// Sort book names in descending order of length
	sortedBookNames := make([]string, 0, len(bookNameMap))
	for k := range bookNameMap {
		sortedBookNames = append(sortedBookNames, k)
	}

	sort.SliceStable(sortedBookNames, func(i, j int) bool {
		return len(sortedBookNames[i]) > len(sortedBookNames[j])
	})

	// Iterate over the valid Bible book names and check if any of them match the input string
	for _, bookName := range sortedBookNames {
		// Create a regex pattern for the current Bible book name
		// The pattern allows any number of characters between the words in the book name
		pattern := ".*" + strings.Join(strings.Fields(bookName), ".*")
		// Compile the regex pattern
		re, err := regexp.Compile(pattern)
		if err != nil {
			fmt.Println("Error compiling regex:", err)
			return false, ""
		}

		// Check if the input string matches the regex pattern
		if re.MatchString(normalizedInput) {
			return true, bookNameMap[bookName]
		}
	}
	return false, ""
}

func checkIfLocation(query string) LocationStruct {
	loc := LocationStruct{
		HasLocation:    false,
		LocationString: "",
		Book:           "",
		Chapter:        0,
		Verse:          0,
		VerseEnd:       0,
	}
	// Define a regular expression pattern
	pattern := `([\w\s]+?)(\d+)(?:.*?(\d+))?(?:.*?(\d+))?`
	// Create a regular expression object
	regex := regexp.MustCompile(pattern)

	// Test if a string matches the pattern
	matches := regex.FindStringSubmatch(query)

	if len(matches) > 0 {
		hasBook, bookName := isValidBibleBook(matches[1])

		if !hasBook {
			return loc
		}

		chapter, _ := strconv.Atoi(matches[2])
		verse, _ := strconv.Atoi(matches[3])
		verseEnd, _ := strconv.Atoi(matches[4])

		// Determine the location string based on how many matches were found
		var locationStr string
		if verseEnd != 0 {
			locationStr = fmt.Sprintf("%s %d:%d-%d", bookName, chapter, verse, verseEnd)
		} else if verse != 0 {
			locationStr = fmt.Sprintf("%s %d:%d", bookName, chapter, verse)
		} else {
			locationStr = fmt.Sprintf("%s %d", bookName, chapter)
		}

		loc := LocationStruct{
			HasLocation:    true,
			LocationString: locationStr,
			Book:           bookName,
			Chapter:        chapter,
			Verse:          verse,
			VerseEnd:       verseEnd,
		}
		return loc
	}

	return loc
}

func updateExactMatchSimilarity(searchBy string, loc LocationStruct, embeddings *[]Embedding) {
	locStringChapter := ""
	locStringVerse := ""
	locStringPassage := ""

	// Define regex patterns for identifying query types
	versePattern := regexp.MustCompile(`\w+\s+\d+:\d+$`)
	chapterPattern := regexp.MustCompile(`\w+\s+\d+$`)
	passagePattern := regexp.MustCompile(`\w+\s+\d+:\d+-\d+$`)

	// Identify the type of query based on priority
	var queryType string
	if passagePattern.MatchString(searchBy) {
		queryType = "passage"
	} else if versePattern.MatchString(searchBy) {
		queryType = "verse"
	} else if chapterPattern.MatchString(searchBy) {
		queryType = "chapter"
	} else {
		queryType = "unknown"
	}

	if loc.Verse > 0 && loc.VerseEnd > 0 {
		locStringPassage = fmt.Sprintf("%s %d:%d-%d", loc.Book, loc.Chapter, loc.Verse, loc.VerseEnd)
	} else if loc.Verse > 0 {
		locStringVerse = fmt.Sprintf("%s %d:%d", loc.Book, loc.Chapter, loc.Verse)
	} else if loc.Chapter > 0 {
		locStringChapter = fmt.Sprintf("%s %d", loc.Book, loc.Chapter)
	}

	for i, embed := range *embeddings {
		if queryType == "chapter" && locStringChapter == embed.Location {
			(*embeddings)[i].Similarity = 0.9999
		} else if queryType == "verse" && locStringVerse == embed.Location {
			(*embeddings)[i].Similarity = 0.9999
		} else if queryType == "passage" && locStringPassage == embed.Location {
			(*embeddings)[i].Similarity = 0.9999
		}
	}
}

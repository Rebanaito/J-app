package cmdoutput

import (
	"fmt"
	"japp/env"
	"japp/wordsearch"
)

func PrintResults(table env.Environment, results wordsearch.ResultEntries, query string) {
	if len(results) == 0 {
		fmt.Println("No results")
		return
	}
	for i, result := range results {
		entry := table.Dict.Entries[result.Entry.WordID]
		for i, kanji := range entry.Kanji {
			if i == 0 {
				fmt.Printf("Kanji: ")
				fmt.Printf("%v", kanji.Expression)
			} else {
				fmt.Printf(", %v", kanji.Expression)
			}
		}
		fmt.Printf("\n")
		for i, reading := range entry.Readings {
			if i == 0 {
				fmt.Printf("Readings: ")
				fmt.Printf("%v", reading.Reading)
			} else {
				fmt.Printf(", %v", reading.Reading)
			}
		}
		fmt.Printf("\n")
		for i, sense := range entry.Sense {
			if i == 0 {
				fmt.Printf("Translations: ")
			}
			for j, gloss := range sense.Glossary {
				if i == 0 && j == 0 {
					fmt.Printf("%v", gloss.Content)
				} else {
					fmt.Printf(", %v", gloss.Content)
				}
			}
		}
		fmt.Printf("\n\n")
		if i == 10 {
			break
		}
	}
}

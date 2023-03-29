package wordsearch

import (
	"japp/env"
	"japp/searchgrids"
	"regexp"
	"strings"
)

type ResultEntry struct {
	Entry searchgrids.Entry
	Score int
}

type ResultEntries []ResultEntry

func SearchQuery(table env.Environment, query string) ResultEntries {
	var words []string
	var search_results ResultEntries
	if isKanaQuery(query) {
		words = parseKana(query)
		raw_results := kanaResults(table.Kana, words)
		search_results = sortKanaResults(table, raw_results, query)
	} else if isKanjiQuery(query) {
		words = parseKanji(query)
		raw_results := kanjiResults(table.Kanji, words)
		search_results = sortKanjiResults(table, raw_results, query)
	} else {
		words = searchgrids.ParseWords(query)
		raw_results := engResults(table.English, words)
		search_results = sortEngResults(table, raw_results, query)
	}
	return search_results
}

func isKanaQuery(query string) bool {
	for _, character := range query {
		if searchgrids.IsHiragana(character) || searchgrids.IsKatakana(character) {
			return true
		} else {
			return false
		}
	}
	return false
}

func isKanjiQuery(query string) bool {
	for _, character := range query {
		if searchgrids.IsKanji(character) {
			return true
		}
	}
	return false
}

func parseKana(query string) []string {
	temp := regexp.MustCompile(`[^ぁ-ヾ ]+`).ReplaceAllString(query, "")
	parsed_words := strings.Split(temp, " ")
	return parsed_words
}

func parseKanji(query string) []string {
	temp := regexp.MustCompile(`[^ぁ-ヾ一-龯㐀-䶿 ]+`).ReplaceAllString(query, "")
	parsed_words := strings.Split(temp, " ")
	return parsed_words
}

func engResults(english *searchgrids.EngAlphabet, words []string) searchgrids.EntryList {
	var old searchgrids.EntryList
	var new searchgrids.EntryList
	for substring, word := range words {
		for position, letter := range word {
			new = engEntryList(*english, letter, position)
			if position == 0 && substring == 0 {
				old = new
			} else if substring == 0 {
				old = mergeEntryListsFirstWord(old, new)
			} else {
				old = mergeEntryLists(old, new)
			}
		}
	}
	return old
}

func kanaResults(kana *searchgrids.KanaAlphabet, words []string) searchgrids.EntryList {
	var old searchgrids.EntryList
	var new searchgrids.EntryList
	for substring, word := range words {
		for position, letter := range word {
			pos := position / 3
			new = kanaEntryList(*kana, letter, pos)
			if pos == 0 && substring == 0 {
				old = new
			} else if substring == 0 {
				old = mergeEntryListsFirstWord(old, new)
			} else {
				old = mergeEntryLists(old, new)
			}
		}
	}
	return old
}

func kanjiResults(kanji *searchgrids.KanjiAlphabet, words []string) searchgrids.EntryList {
	var old searchgrids.EntryList
	var new searchgrids.EntryList
	for substring, word := range words {
		for position, letter := range word {
			pos := position / 3
			if !searchgrids.IsKanji(letter) {
				continue
			}
			new = kanjiEntryList(*kanji, letter, pos)
			if pos == 0 && substring == 0 {
				old = new
			} else if substring == 0 {
				old = mergeEntryListsFirstWord(old, new)
			} else {
				old = mergeEntryLists(old, new)
			}
		}
	}
	return old
}

func sortEngResults(table env.Environment, raw_results searchgrids.EntryList, query string) ResultEntries {
	var results ResultEntries
	for _, entry := range raw_results {
		results = append(results, ResultEntry{entry, calculateEngScore(table, entry, query)})
	}
	quicksortResults(results, 0, len(results)-1)
	return results
}

func sortKanaResults(table env.Environment, raw_results searchgrids.EntryList, query string) ResultEntries {
	var results ResultEntries
	for _, entry := range raw_results {
		results = append(results, ResultEntry{entry, calculateKanaScore(table, entry, query)})
	}
	quicksortResults(results, 0, len(results)-1)
	return results
}

func sortKanjiResults(table env.Environment, raw_results searchgrids.EntryList, query string) ResultEntries {
	var results ResultEntries
	for _, entry := range raw_results {
		results = append(results, ResultEntry{entry, calculateKanjiScore(table, entry, query)})
	}
	quicksortResults(results, 0, len(results)-1)
	return results
}

func calculateEngScore(table env.Environment, entry searchgrids.Entry, query string) int {
	var score int
	var best int
	var content_length int

	query_length := len(query)
	for i, hash := range entry.Hash {
		score = 0
		content := hash % 100
		score += 2 - content
		hash /= 100
		gloss := hash % 100
		score -= 2 * (gloss - 1)
		hash /= 100
		score -= (hash - 1)
		content_length = len(table.Dict.Entries[entry.WordID].Sense[hash-1].Glossary[gloss-1].Content)
		score += 10 - (content_length-query_length)*(content+1)
		if i == 0 {
			best = score
		} else if score > best {
			best = score
		}
	}
	best = best * entry.Score
	return best
}

func calculateKanaScore(table env.Environment, entry searchgrids.Entry, query string) int {
	var score int
	var best int
	var reading_length int

	query_length := len(query)
	for i, index := range entry.Hash {
		score = 0
		score += 3 - index
		reading_length = len(table.Dict.Entries[entry.WordID].Readings[index].Reading)
		score *= 10 - (reading_length - query_length)
		if i == 0 {
			best = score
		} else if score > best {
			best = score
		}
	}
	best = best * entry.Score
	return best
}

func calculateKanjiScore(table env.Environment, entry searchgrids.Entry, query string) int {
	var score int
	var best int
	var kanji_length int

	query_length := len(query)
	for i, index := range entry.Hash {
		score = 0
		score += 3 - index
		kanji_length = len(table.Dict.Entries[entry.WordID].Kanji[index].Expression)
		score *= 10 - (kanji_length - query_length)
		if i == 0 {
			best = score
		} else if score > best {
			best = score
		}
	}
	best = best * entry.Score
	return best
}

func quicksortResults(results ResultEntries, low, high int) {
	if low >= high {
		return
	}
	pivot := quicksortPartition(results, low, high)
	quicksortResults(results, low, pivot-1)
	quicksortResults(results, pivot+1, high)
}

func quicksortPartition(results ResultEntries, low, high int) int {
	pivot := results[high].Score
	i := low - 1
	for j := low; j < high; j++ {
		if results[j].Score > pivot {
			i++
			results[i], results[j] = results[j], results[i]
		}
	}
	results[i+1], results[high] = results[high], results[i+1]
	return i + 1
}

// This function will be used during search. It will pull up a list of words where (letter in position) is true
func engEntryList(grid searchgrids.EngAlphabet, letter rune, position int) searchgrids.EntryList {
	var char int = int(letter) - 97
	return grid.Alphabet[char].Positions[position].List
}

func kanaEntryList(grid searchgrids.KanaAlphabet, letter rune, position int) searchgrids.EntryList {
	var char int
	if searchgrids.IsHiragana(letter) {
		char = int(letter) - 12352
	} else {
		char = int(letter) - 12448
	}
	return grid.Alphabet[char].Positions[position].List
}

func kanjiEntryList(grid searchgrids.KanjiAlphabet, letter rune, position int) searchgrids.EntryList {
	var char int
	if searchgrids.IsRegularKanji(letter) {
		char = int(letter) - 19968
	} else {
		char = int(letter) + 7600
	}
	return grid.Alphabet[char].Positions[position].List
}

// This search function narrows the list of search by a lot by merging lists of two consecutive letters and making sure that only the words that are in both lists pass
func mergeEntryListsFirstWord(old searchgrids.EntryList, new searchgrids.EntryList) searchgrids.EntryList {
	var result searchgrids.EntryList
	for _, entry := range old {
		if match, index := matchInLists(entry.WordID, new); match { // We focus on the elements of the existing list because in most cases (when we have more than a couple of letters) it will be much shorter than the new one
			var appendix searchgrids.Entry
			appendix.WordID = entry.WordID
			appendix.Score = entry.Score
			for _, hash := range entry.Hash {
				if matchHashFirstWord(new[index].Hash, hash) {
					appendix.Hash = append(appendix.Hash, hash)
				}
			}
			if len(appendix.Hash) != 0 {
				result = append(result, appendix)
			}
		}
	}
	return result
}

func mergeEntryLists(old searchgrids.EntryList, new searchgrids.EntryList) searchgrids.EntryList {
	var result searchgrids.EntryList
	for _, entry := range old {
		if match, index := matchInLists(entry.WordID, new); match { // We focus on the elements of the existing list because in most cases (when we have more than a couple of letters) it will be much shorter than the new one
			var appendix searchgrids.Entry
			appendix.WordID = entry.WordID
			appendix.Score = entry.Score
			for _, hash := range entry.Hash {
				if matchHash(new[index].Hash, hash) {
					appendix.Hash = append(appendix.Hash, hash)
				}
			}
			if len(appendix.Hash) != 0 {
				result = append(result, appendix)
			}
		}
	}
	return result
}

func matchInLists(wordID int, list searchgrids.EntryList) (match bool, index int) {
	if list == nil {
		return false, -1
	}
	first := 0
	last := len(list) - 1
	var midpoint int
	for first <= last {
		midpoint = int((first + last) / 2)
		if list[midpoint].WordID == wordID {
			return true, midpoint
		} else if list[midpoint].WordID < wordID {
			first = midpoint + 1
		} else {
			last = midpoint - 1
		}
	}
	return false, midpoint
}

func matchHashFirstWord(Hash searchgrids.Hash, hash int) bool {
	if Hash == nil {
		return false
	}
	first := 0
	last := len(Hash) - 1
	var midpoint int
	for first <= last {
		midpoint = int((first + last) / 2)
		if Hash[midpoint] == hash {
			return true
		} else if Hash[midpoint] < hash {
			first = midpoint + 1
		} else {
			last = midpoint - 1
		}
	}
	return false
}

func matchHash(Hash searchgrids.Hash, hash int) bool {
	if Hash == nil {
		return false
	}
	first := 0
	last := len(Hash) - 1
	var midpoint int
	for first <= last {
		midpoint = int((first + last) / 2)
		if Hash[midpoint] == hash || Hash[midpoint] == hash+1 {
			return true
		} else if Hash[midpoint] < hash {
			first = midpoint + 1
		} else {
			last = midpoint - 1
		}
	}
	return false
}

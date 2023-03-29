package searchgrids

import (
	"regexp"
	"strings"

	"foosoft.net/projects/jmdict"
)

type EngAlphabet struct {
	// Structure representing a slice of English alphabet characters. Each element is a struct corresponding to a letter and contains a slice of Position structs
	Alphabet []EngLetter
}

type KanaAlphabet struct {
	Alphabet []KanaLetter
}

type KanjiAlphabet struct {
	Alphabet []KanjiSymbol
}

type EngLetter struct {
	// Structure representing a letter from the English alphabet, with a slice of structures based on positions this letter occupies within different words
	Positions []Position
}

type KanaLetter struct {
	Positions []Position
}

type KanjiSymbol struct {
	Positions []Position
}

type Position struct {
	// A list of words that have this letter in the specified position (1st letter of the word, 2nd letter of the word, and so on) represented by their indash values in JMdict
	List EntryList
}

// EntryList is a slice of Entries, where each Entry is just an indash of the word in JMdict
type EntryList []Entry

type Entry struct {
	WordID int
	Score  uint16
	Hash   Hash
}

type Hash []uint16

func GenerateAlphabets(dict jmdict.Jmdict) (*EngAlphabet, *KanaAlphabet, *KanjiAlphabet) {
	var engAlphabet EngAlphabet
	var kanaAlphabet KanaAlphabet
	var kanjiAlphabet KanjiAlphabet
	fillLetters(&engAlphabet)
	fillKana(&kanaAlphabet)
	fillKanji(&kanjiAlphabet)
	for wordID, entry := range dict.Entries {
		score := scoreEntry(entry)
		engWrite(&engAlphabet, entry, wordID, score)
		kanaWrite(&kanaAlphabet, entry, wordID, score)
		kanjiWrite(&kanjiAlphabet, entry, wordID, score)
	}
	return &engAlphabet, &kanaAlphabet, &kanjiAlphabet
}

func fillLetters(alphabet *EngAlphabet) {
	for i := 0; i < 26; i++ {
		alphabet.Alphabet = append(alphabet.Alphabet, EngLetter{})
	}
}

func fillKana(alphabet *KanaAlphabet) {
	for i := 0; i < 96; i++ {
		alphabet.Alphabet = append(alphabet.Alphabet, KanaLetter{})
	}
}

func fillKanji(alphabet *KanjiAlphabet) {
	for i := 0; i < 27503; i++ {
		alphabet.Alphabet = append(alphabet.Alphabet, KanjiSymbol{})
	}
}

func scoreEntry(entry jmdict.JmdictEntry) uint16 {
	score := checkKanji(entry.Kanji) + checkContent(entry) + checkReadings(entry)
	score += 500
	return score
}

func checkKanji(kanji []jmdict.JmdictKanji) uint16 {
	var score uint16
	length := len(kanji)
	if length != 0 {
		score += uint16(length) * 10
		for i := 0; i < length; i++ {
			score = score + (uint16(length-i))*(kanjiFirst(kanji[i])-limitedKanji(kanji[i])+kanjiPriority(kanji[i]))
		}
	}
	return score
}

func kanjiFirst(kanji_entry jmdict.JmdictKanji) uint16 {
	for _, expression := range kanji_entry.Expression {
		if IsKanji(expression) {
			return 2
		} else {
			break
		}
	}
	return 0
}

func limitedKanji(kanji_entry jmdict.JmdictKanji) uint16 {
	str := strings.Join(kanji_entry.Information, "")
	var value uint16 = 0
	if str == "search-only kanji form" {
		value = 1
	} else if str == "rarely-used kanji form" {
		value = 2
	}
	return value
}

func kanjiPriority(kanji_entry jmdict.JmdictKanji) uint16 {
	var value uint16 = 0
	length := len(kanji_entry.Priorities)
	if length != 0 {
		for i := 0; i < length; i++ {
			switch kanji_entry.Priorities[i] {
			case "news1", "ichi1", "gai1", "spec1":
				value += 5
			case "news2", "ichi2", "gai2", "spec2":
				value += 2
			default:
				value += (50 - wordfreq(kanji_entry.Priorities[i])) / 10
			}
		}
	}
	return value
}

func wordfreq(priority string) uint16 {
	var value uint16 = 0
	flag := 0
	for _, char := range priority {
		if char == 'n' && flag == 0 {
			flag++
			continue
		} else if char == 'f' && flag == 1 {
			flag++
			continue
		} else if char >= 48 && char <= 57 && flag == 2 {
			value = value*10 + uint16(int(char)-48)
		} else {
			value = 0
			break
		}
	}
	return value
}

func checkContent(entry jmdict.JmdictEntry) uint16 {
	var score uint16
	senses := len(entry.Sense)
	for _, sense := range entry.Sense {
		glossaries := len(sense.Glossary)
		for _, gloss := range sense.Glossary {
			score += uint16(senses * glossaries * len(gloss.Content))
		}
	}
	return score / 10
}

func checkReadings(entry jmdict.JmdictEntry) uint16 {
	var score uint16
	score += uint16(len(entry.Readings) * 5)
	for _, reading := range entry.Readings {
		if len(reading.Information) != 0 {
			score += 2
		}
		if len(reading.Restrictions) != 0 {
			score -= 3
		}
		for _, letter := range reading.Reading {
			if IsHiragana(letter) {
				score += 2
				break
			} else if IsKatakana(letter) {
				score += 1
				break
			}
		}
	}
	return score
}

func IsHiragana(letter rune) bool {
	if letter >= 12352 && letter <= 12447 {
		return true
	}
	return false
}

func IsKatakana(letter rune) bool {
	if letter >= 12448 && letter <= 12543 {
		return true
	}
	return false
}

func IsKanji(letter rune) bool {
	if IsRegularKanji(letter) {
		return true
	} else if IsRareKanji(letter) {
		return true
	}
	return false
}

func IsRegularKanji(letter rune) bool {
	if letter >= 19968 && letter <= 40879 {
		return true
	}
	return false
}

func IsRareKanji(letter rune) bool {
	if letter >= 13312 && letter <= 19903 {
		return true
	}
	return false
}

func engWrite(alphabet *EngAlphabet, entry jmdict.JmdictEntry, wordID int, score uint16) {
	for i, sense := range entry.Sense {
		for j, gloss := range sense.Glossary {
			words := ParseWords(gloss.Content)
			if len(words) != 0 {
				for k, word := range words {
					var hash uint16 = uint16((i+1)*2000 + (j+1)*100 + k)
					writeEngWord(alphabet, word, wordID, score, hash)
				}
			}
		}
	}
}

func kanaWrite(alphabet *KanaAlphabet, entry jmdict.JmdictEntry, wordID int, score uint16) {
	for index, reading := range entry.Readings {
		writeKanaWord(alphabet, reading.Reading, wordID, score, uint16(index))
	}
}

func kanjiWrite(alphabet *KanjiAlphabet, entry jmdict.JmdictEntry, wordID int, score uint16) {
	for index, kanji := range entry.Kanji {
		writeKanjiSymbol(alphabet, kanji.Expression, wordID, score, uint16(index))
	}
}

func ParseWords(content string) []string {
	temp := regexp.MustCompile(`[^a-zA-Z ]+`).ReplaceAllString(content, "")
	temp = strings.ToLower(temp)
	parsed_words := strings.Split(temp, " ")
	return parsed_words
}

func writeEngWord(alphabet *EngAlphabet, word string, wordID int, score, hash uint16) {
	for position, letter := range word {
		var char int = int(letter) - 97 // We make use of the ASCII representation to get the indash values of slices based on the letter, which helps us save some time
		insertEngEntry(alphabet, char, position, wordID, score, hash)
	}
}

func writeKanaWord(alphabet *KanaAlphabet, word string, wordID int, score, index uint16) {
	for position, character := range word {
		pos := position / 3
		var char int
		if IsKatakana(character) {
			char = int(character) - 12448
		} else if IsHiragana(character) {
			char = int(character) - 12352
		} else {
			continue
		}
		insertKanaEntry(alphabet, char, pos, wordID, score, index)
	}
}

func writeKanjiSymbol(alphabet *KanjiAlphabet, word string, wordID int, score, index uint16) {
	for position, character := range word {
		pos := position / 3
		var char int
		if !IsKanji(character) {
			continue
		} else if IsRegularKanji(character) {
			char = int(character) - 19968
		} else {
			char = int(character) + 7600
		}
		insertKanjiEntry(alphabet, char, pos, wordID, score, index)
	}
}

// This is the main function that will be called during JMdict mapping for search.
// It takes the rune of the letter and its position within the word, the word's indash in JMdict and the alphabet struct.
func insertEngEntry(grid *EngAlphabet, char, position, wordID int, score, hash uint16) {
	length := len(grid.Alphabet[char].Positions) - 1 // We make sure that the slice for the letter has enough elements to at least match the position value
	for position > length {                          // If it doesn't, we append more elements to the slice
		grid.Alphabet[char].Positions = append(grid.Alphabet[char].Positions, Position{})
		length++
	}
	sortAndInsert(&grid.Alphabet[char].Positions[position], wordID, score, hash)
}

func insertKanaEntry(grid *KanaAlphabet, char, position, wordID int, score, index uint16) {
	length := len(grid.Alphabet[char].Positions) - 1 // We make sure that the slice for the letter has enough elements to at least match the position value
	for position > length {                          // If it doesn't, we append more elements to the slice
		grid.Alphabet[char].Positions = append(grid.Alphabet[char].Positions, Position{})
		length++
	}
	sortAndInsert(&grid.Alphabet[char].Positions[position], wordID, score, index)
}

func insertKanjiEntry(grid *KanjiAlphabet, char, position, wordID int, score, index uint16) {
	length := len(grid.Alphabet[char].Positions) - 1
	for position > length {
		grid.Alphabet[char].Positions = append(grid.Alphabet[char].Positions, Position{})
		length++
	}
	sortAndInsert(&grid.Alphabet[char].Positions[position], wordID, score, index)
}

// This functions performs the sorting (if necessary) of the entry list and inserts the new element
func sortAndInsert(position *Position, wordID int, score, hash uint16) {
	length := len(position.List)
	var entry Entry
	entry.WordID = wordID
	entry.Score = score
	entry.Hash = append(entry.Hash, hash)
	if length == 0 { // First we check for empty slices
		position.List = append(position.List, entry)
	} else {
		length -= 1
		if position.List[length].WordID != wordID {
			position.List = append(position.List, entry)
		} else {
			position.List[length].Hash = append(position.List[length].Hash, hash)
		}
	}
}

package kana

func IsKatakana(character rune) bool {
	if int(character) >= 12450 && int(character) <= 12535 {
		return true
	} else {
		return false
	}
}

func IsHiragana(character rune) bool {
	if int(character) >= 12353 && int(character) <= 12437 {
		return true
	} else {
		return false
	}
}

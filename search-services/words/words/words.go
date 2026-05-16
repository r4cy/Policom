package words

import (
	"maps"
	"slices"
	"strings"
	"unicode"

	"github.com/kljensen/snowball"
	"github.com/kljensen/snowball/english"
)

// Функция для нормализации слов, принимает строку, возвращает
// массив нормализованных слов на английском языке
func NormalizeTheWords(str string) []string {
	words := strings.FieldsFunc(str, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	unique := make(map[string]bool)
	for _, elem := range words {
		elem = strings.ToLower(elem)
		if english.IsStopWord(elem) {
			continue
		}
		stemmed, err := snowball.Stem(elem, "english", true)
		if err != nil {
			continue
		}
		if !unique[stemmed] {
			unique[stemmed] = true
		}
	}
	keys := slices.Collect(maps.Keys(unique))
	return keys
}

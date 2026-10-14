package words

import (
	_ "embed"
	"math/rand/v2"
	"strconv"
	"strings"
	"unicode"
)

//go:embed data/en.txt
var rawEN string

//go:embed data/fr.txt
var rawFR string

//go:embed data/code.txt
var rawCode string

//go:embed data/quotes.txt
var rawQuotes string

type Language int

const (
	LangEN Language = iota
	LangFR
	LangCode
)

func (l Language) String() string {
	switch l {
	case LangEN:
		return "EN"
	case LangFR:
		return "FR"
	case LangCode:
		return "Code"
	default:
		return "EN"
	}
}

type Options struct {
	Lang        Language
	Count       int
	Punctuation bool
	Numbers     bool
}

var (
	bankEN     []string
	bankFR     []string
	bankCode   []string
	bankQuotes []string

	punctEndings = []string{".", ",", "!", "?", ";", ":"}
)

func init() {
	bankEN = parseLines(rawEN)
	bankFR = parseLines(rawFR)
	bankCode = parseLines(rawCode)
	bankQuotes = parseLines(rawQuotes)
}

func parseLines(raw string) []string {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func bankFor(lang Language) []string {
	switch lang {
	case LangFR:
		return bankFR
	case LangCode:
		return bankCode
	default:
		return bankEN
	}
}

// Generate produces a slice of words based on options using zero-lock math/rand/v2.
func Generate(opts Options) []string {
	if opts.Count <= 0 {
		opts.Count = 25
	}
	bank := bankFor(opts.Lang)
	if len(bank) == 0 {
		bank = bankEN
	}

	result := make([]string, 0, opts.Count)
	capitalizeNext := opts.Punctuation
	lastIdx := -1

	for i := 0; i < opts.Count; i++ {
		if opts.Numbers && rand.IntN(10) == 0 {
			numStr := strconv.Itoa(rand.IntN(999) + 1)
			result = append(result, numStr)
			capitalizeNext = false
			continue
		}

		idx := rand.IntN(len(bank))
		if idx == lastIdx && len(bank) > 1 {
			idx = (idx + 1) % len(bank)
		}
		lastIdx = idx

		word := bank[idx]

		if opts.Punctuation && opts.Lang != LangCode {
			if capitalizeNext || rand.IntN(12) == 0 {
				word = capitalizeFirst(word)
				capitalizeNext = false
			}

			roll := rand.IntN(10)
			switch {
			case roll == 0 && i < opts.Count-1:
				p := punctEndings[rand.IntN(len(punctEndings))]
				word += p
				if p == "." || p == "!" || p == "?" {
					capitalizeNext = true
				}
			case roll == 1 && len(word) > 2:
				word = "\"" + word + "\""
			case roll == 2 && len(word) > 2:
				word = "(" + word + ")"
			}
		}

		result = append(result, word)
	}

	// Ensure last word ends with a period if punctuation mode is active
	if opts.Punctuation && opts.Lang != LangCode && len(result) > 0 {
		last := result[len(result)-1]
		lastRune := []rune(last)[len([]rune(last))-1]
		if unicode.IsLetter(lastRune) || unicode.IsDigit(lastRune) {
			result[len(result)-1] = last + "."
		}
	}

	return result
}

// RandomQuote returns a random quote split into words.
func RandomQuote() []string {
	if len(bankQuotes) == 0 {
		return []string{"Simplicity", "is", "prerequisite", "for", "reliability."}
	}
	q := bankQuotes[rand.IntN(len(bankQuotes))]
	return strings.Fields(q)
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

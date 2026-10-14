package engine

import (
	"testing"

	"ricetype/internal/words"
)

func TestEngineTypingAndDeleteWord(t *testing.T) {
	e := New(Config{
		Mode:      ModeWords,
		WordLimit: 10,
		Lang:      words.LangEN,
		WrapWidth: 60,
	})

	if len(e.TargetRunes) == 0 {
		t.Fatal("expected non-empty target runes")
	}

	// Type first 3 runes correctly
	for i := 0; i < 3 && i < len(e.TargetRunes); i++ {
		e.TypeRune(e.TargetRunes[i])
	}
	if e.CorrectChars != 3 {
		t.Fatalf("expected 3 correct chars, got %d", e.CorrectChars)
	}

	// Delete word via Ctrl+W logic
	e.DeleteWord()
	if len(e.TypedRunes) != 0 {
		t.Fatalf("expected 0 typed runes after DeleteWord, got %d", len(e.TypedRunes))
	}
	if e.CorrectChars != 0 {
		t.Fatalf("expected 0 correct chars after DeleteWord, got %d", e.CorrectChars)
	}
}

func TestSuddenDeathMode(t *testing.T) {
	e := New(Config{
		Mode:      ModeSuddenDeath,
		WordLimit: 10,
		Lang:      words.LangFR,
		WrapWidth: 60,
	})

	wrongRune := '!'
	if e.TargetRunes[0] == '!' {
		wrongRune = '?'
	}
	e.TypeRune(wrongRune)

	if !e.Finished || !e.SuddenDeath {
		t.Fatal("expected SuddenDeath mode to finish immediately on wrong key")
	}
}

func BenchmarkRewrapAndType(b *testing.B) {
	e := New(Config{
		Mode:         ModeTime,
		TimeDuration: 60,
		Lang:         words.LangFR,
		Punctuation:  true,
		Numbers:      true,
		WrapWidth:    76,
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := len(e.TypedRunes)
		if idx >= len(e.TargetRunes) {
			e.Reset()
			idx = 0
		}
		e.TypeRune(e.TargetRunes[idx])
		_ = e.CurrentLineIndex()
	}
}

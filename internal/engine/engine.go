package engine

import (
	"math"
	"strings"
	"time"

	"ricetype/internal/words"
)

type TestMode int

const (
	ModeTime TestMode = iota
	ModeWords
	ModeQuote
	ModeSuddenDeath
)

func (m TestMode) String() string {
	switch m {
	case ModeTime:
		return "temps"
	case ModeWords:
		return "mots"
	case ModeQuote:
		return "citation"
	case ModeSuddenDeath:
		return "mort subite"
	default:
		return "temps"
	}
}

// LineSpan represents a pre-calculated line [Start, End) in TargetRunes.
type LineSpan struct {
	Start int
	End   int
}

type Config struct {
	Mode         TestMode
	TimeDuration int // 15, 30, 60, 120
	WordLimit    int // 10, 25, 50, 100
	Lang         words.Language
	Punctuation  bool
	Numbers      bool
	WrapWidth    int
}

type Engine struct {
	Cfg Config

	Words       []string
	TargetRunes []rune
	TypedRunes  []rune
	Lines       []LineSpan

	// Timing
	StartTime   time.Time
	EndTime     time.Time
	TimerActive bool
	TimeLeft    int
	Finished    bool
	SuddenDeath bool

	// Incremental counters (zero-scan in View)
	CorrectChars    int
	TotalKeys       int
	WrongKeys       int
	CorrectedErrors int
	CurrentStreak   int
	MaxStreak       int
	WordsCompleted  int

	// Per-second telemetry (pre-allocated)
	WPMHistory     []float64
	RawWPMHistory  []float64
	ErrorHistory   []int
	errorsThisTick int
}

func New(cfg Config) *Engine {
	if cfg.WrapWidth <= 0 {
		cfg.WrapWidth = 70
	}
	e := &Engine{
		Cfg:           cfg,
		WPMHistory:    make([]float64, 0, 128),
		RawWPMHistory: make([]float64, 0, 128),
		ErrorHistory:  make([]int, 0, 128),
	}
	e.Reset()
	return e
}

func (e *Engine) Reset() {
	var generated []string
	switch e.Cfg.Mode {
	case ModeQuote:
		generated = words.RandomQuote()
	case ModeWords, ModeSuddenDeath:
		generated = words.Generate(words.Options{
			Lang:        e.Cfg.Lang,
			Count:       e.Cfg.WordLimit,
			Punctuation: e.Cfg.Punctuation,
			Numbers:     e.Cfg.Numbers,
		})
	default: // ModeTime
		generated = words.Generate(words.Options{
			Lang:        e.Cfg.Lang,
			Count:       80,
			Punctuation: e.Cfg.Punctuation,
			Numbers:     e.Cfg.Numbers,
		})
	}

	e.Words = generated
	joined := strings.Join(generated, " ")
	e.TargetRunes = []rune(joined)

	if cap(e.TypedRunes) < len(e.TargetRunes) {
		e.TypedRunes = make([]rune, 0, len(e.TargetRunes)+128)
	} else {
		e.TypedRunes = e.TypedRunes[:0]
	}

	e.TimerActive = false
	e.Finished = false
	e.SuddenDeath = false
	e.TimeLeft = e.Cfg.TimeDuration
	e.CorrectChars = 0
	e.TotalKeys = 0
	e.WrongKeys = 0
	e.CorrectedErrors = 0
	e.CurrentStreak = 0
	e.MaxStreak = 0
	e.WordsCompleted = 0
	e.errorsThisTick = 0

	e.WPMHistory = e.WPMHistory[:0]
	e.RawWPMHistory = e.RawWPMHistory[:0]
	e.ErrorHistory = e.ErrorHistory[:0]

	e.Rewrap(e.Cfg.WrapWidth)
}

// Rewrap pre-calculates line boundaries so View() never splits or allocates strings.
func (e *Engine) Rewrap(maxWidth int) {
	if maxWidth < 20 {
		maxWidth = 20
	}
	e.Cfg.WrapWidth = maxWidth
	e.Lines = e.Lines[:0]

	n := len(e.TargetRunes)
	if n == 0 {
		return
	}

	lineStart := 0
	lastSpace := -1

	for i := 0; i < n; i++ {
		if e.TargetRunes[i] == ' ' {
			lastSpace = i
		}
		if i-lineStart+1 > maxWidth {
			if lastSpace >= lineStart {
				// Include the trailing space on this line so the cursor flows naturally
				e.Lines = append(e.Lines, LineSpan{Start: lineStart, End: lastSpace + 1})
				lineStart = lastSpace + 1
				lastSpace = -1
			} else {
				// Single word longer than maxWidth
				e.Lines = append(e.Lines, LineSpan{Start: lineStart, End: i})
				lineStart = i
			}
		}
	}
	if lineStart < n {
		e.Lines = append(e.Lines, LineSpan{Start: lineStart, End: n})
	}
}

// appendMoreWords dynamically extends the word buffer in Time mode.
func (e *Engine) appendMoreWords(count int) {
	more := words.Generate(words.Options{
		Lang:        e.Cfg.Lang,
		Count:       count,
		Punctuation: e.Cfg.Punctuation,
		Numbers:     e.Cfg.Numbers,
	})
	e.Words = append(e.Words, more...)
	extraRunes := []rune(" " + strings.Join(more, " "))
	e.TargetRunes = append(e.TargetRunes, extraRunes...)
	e.Rewrap(e.Cfg.WrapWidth)
}

// CurrentLineIndex returns the index in e.Lines where the cursor currently sits.
func (e *Engine) CurrentLineIndex() int {
	pos := len(e.TypedRunes)
	for i, span := range e.Lines {
		if pos >= span.Start && pos < span.End {
			return i
		}
	}
	if len(e.Lines) > 0 {
		return len(e.Lines) - 1
	}
	return 0
}

// TypeRune processes a single printable character input.
func (e *Engine) TypeRune(r rune) (startedNow bool) {
	if e.Finished {
		return false
	}

	if !e.TimerActive {
		e.TimerActive = true
		e.StartTime = time.Now()
		startedNow = true
	}

	idx := len(e.TypedRunes)
	if idx >= len(e.TargetRunes) {
		return startedNow
	}

	e.TotalKeys++
	expected := e.TargetRunes[idx]
	e.TypedRunes = append(e.TypedRunes, r)

	if r == expected {
		e.CorrectChars++
		e.CurrentStreak++
		if e.CurrentStreak > e.MaxStreak {
			e.MaxStreak = e.CurrentStreak
		}
		if expected == ' ' {
			e.WordsCompleted++
		}
	} else {
		e.WrongKeys++
		e.errorsThisTick++
		e.CurrentStreak = 0
		if e.Cfg.Mode == ModeSuddenDeath {
			e.SuddenDeath = true
			e.Finish()
			return startedNow
		}
	}

	// Dynamic buffer extension in Time mode when within 120 chars of the end
	if e.Cfg.Mode == ModeTime && len(e.TargetRunes)-len(e.TypedRunes) < 120 {
		e.appendMoreWords(40)
	}

	// Completion check for fixed-length modes
	if len(e.TypedRunes) >= len(e.TargetRunes) {
		e.WordsCompleted++
		e.Finish()
	}

	return startedNow
}

// Backspace removes the last character and updates incremental counters.
func (e *Engine) Backspace() {
	n := len(e.TypedRunes)
	if n == 0 || e.Finished {
		return
	}
	lastIdx := n - 1
	typed := e.TypedRunes[lastIdx]
	expected := e.TargetRunes[lastIdx]

	if typed == expected {
		e.CorrectChars--
		if expected == ' ' && e.WordsCompleted > 0 {
			e.WordsCompleted--
		}
	} else {
		e.CorrectedErrors++
	}
	e.TypedRunes = e.TypedRunes[:lastIdx]
}

// DeleteWord removes characters back to the previous word boundary (Ctrl+W / Ctrl+Backspace).
func (e *Engine) DeleteWord() {
	if len(e.TypedRunes) == 0 || e.Finished {
		return
	}
	// If right after a space, remove trailing spaces first
	for len(e.TypedRunes) > 0 && e.TypedRunes[len(e.TypedRunes)-1] == ' ' {
		e.Backspace()
	}
	// Remove until previous space or beginning
	for len(e.TypedRunes) > 0 && e.TypedRunes[len(e.TypedRunes)-1] != ' ' {
		e.Backspace()
	}
}

// Tick records per-second telemetry and updates countdown in ModeTime.
func (e *Engine) Tick() (finished bool) {
	if !e.TimerActive || e.Finished {
		return false
	}

	e.WPMHistory = append(e.WPMHistory, e.LiveNetWPM())
	e.RawWPMHistory = append(e.RawWPMHistory, e.LiveRawWPM())
	e.ErrorHistory = append(e.ErrorHistory, e.errorsThisTick)
	e.errorsThisTick = 0

	if e.Cfg.Mode == ModeTime {
		e.TimeLeft--
		if e.TimeLeft <= 0 {
			e.Finish()
			return true
		}
	}
	return false
}

func (e *Engine) Finish() {
	if e.Finished {
		return
	}
	e.Finished = true
	e.TimerActive = false
	e.EndTime = time.Now()
	if len(e.WPMHistory) == 0 {
		e.WPMHistory = append(e.WPMHistory, e.NetWPM())
		e.RawWPMHistory = append(e.RawWPMHistory, e.RawWPM())
		e.ErrorHistory = append(e.ErrorHistory, e.WrongKeys)
	}
}

func (e *Engine) ElapsedSeconds() float64 {
	if e.StartTime.IsZero() {
		return 0
	}
	var d time.Duration
	if e.Finished && !e.EndTime.IsZero() {
		d = e.EndTime.Sub(e.StartTime)
	} else {
		d = time.Since(e.StartTime)
	}
	sec := d.Seconds()
	if e.Finished && e.Cfg.Mode == ModeTime && !e.SuddenDeath {
		return float64(e.Cfg.TimeDuration)
	}
	if sec < 0.1 {
		return 0.1
	}
	return sec
}

func (e *Engine) LiveNetWPM() float64 {
	if !e.TimerActive && !e.Finished {
		return 0
	}
	mins := e.ElapsedSeconds() / 60.0
	if mins <= 0 {
		return 0
	}
	return (float64(e.CorrectChars) / 5.0) / mins
}

func (e *Engine) LiveRawWPM() float64 {
	if !e.TimerActive && !e.Finished {
		return 0
	}
	mins := e.ElapsedSeconds() / 60.0
	if mins <= 0 {
		return 0
	}
	return (float64(e.TotalKeys) / 5.0) / mins
}

func (e *Engine) NetWPM() float64 {
	return e.LiveNetWPM()
}

func (e *Engine) RawWPM() float64 {
	return e.LiveRawWPM()
}

func (e *Engine) Accuracy() float64 {
	if e.TotalKeys == 0 {
		return 100.0
	}
	correct := e.TotalKeys - e.WrongKeys
	if correct < 0 {
		correct = 0
	}
	return (float64(correct) / float64(e.TotalKeys)) * 100.0
}

// Consistency computes 100 * (1 - stddev/mean) over WPMHistory, clamped to [0, 100].
func (e *Engine) Consistency() float64 {
	n := len(e.WPMHistory)
	if n <= 1 {
		return 100.0
	}
	sum := 0.0
	for _, v := range e.WPMHistory {
		sum += v
	}
	mean := sum / float64(n)
	if mean <= 1.0 {
		return 0.0
	}
	varianceSum := 0.0
	for _, v := range e.WPMHistory {
		diff := v - mean
		varianceSum += diff * diff
	}
	stdDev := math.Sqrt(varianceSum / float64(n))
	cv := stdDev / mean
	consistency := (1.0 - cv) * 100.0
	return math.Max(0, math.Min(100, consistency))
}

// ProgressRatio returns a [0.0, 1.0] progress value for the smooth progress bar.
func (e *Engine) ProgressRatio() float64 {
	if e.Cfg.Mode == ModeTime {
		if e.Cfg.TimeDuration <= 0 {
			return 0
		}
		done := float64(e.Cfg.TimeDuration - e.TimeLeft)
		return math.Max(0, math.Min(1, done/float64(e.Cfg.TimeDuration)))
	}
	if len(e.TargetRunes) == 0 {
		return 0
	}
	return math.Max(0, math.Min(1, float64(len(e.TypedRunes))/float64(len(e.TargetRunes))))
}

// Rank returns a title badge based on Net WPM and Accuracy.
func Rank(wpm, acc float64) string {
	if acc < 75 {
		return "Clavier Rebelle"
	}
	switch {
	case wpm >= 120:
		return "Cyber-Dieu du Clavier"
	case wpm >= 95:
		return "Maître Dactylo"
	case wpm >= 75:
		return "Véloce Éclair"
	case wpm >= 55:
		return "Dactylo Confirmé"
	case wpm >= 35:
		return "Apprenti Agile"
	default:
		return "Novice Tranquille"
	}
}

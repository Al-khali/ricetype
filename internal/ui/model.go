package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"ricetype/internal/engine"
	"ricetype/internal/storage"
	"ricetype/internal/theme"
	"ricetype/internal/words"
)

type SessionState int

const (
	StateMenu SessionState = iota
	StateTyping
	StateResults
	StateStats
)

type tickMsg time.Time
type animTickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func animTickCmd() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
		return animTickMsg(t)
	})
}

var (
	timeOptions = []int{15, 30, 60, 120}
	wordOptions = []int{10, 25, 50, 100}
)

type Model struct {
	state SessionState

	// Menu state (0: Mode, 1: Target, 2: Lang, 3: Modifiers, 4: Theme, 5: Start)
	menuRow      int
	modifierCol  int // 0: Punctuation, 1: Numbers
	zenMode      bool
	themeIdx     int
	styles       theme.Styles
	boxWidth     int
	termWidth    int
	termHeight   int

	// Core components
	eng   *engine.Engine
	store *storage.Store

	// Results celebration state
	isNewPB   bool
	animFrame int
}

func NewModel() Model {
	store := storage.Load()
	tIdx := store.ThemeIndex
	if tIdx < 0 || tIdx >= len(theme.Palettes) {
		tIdx = 0
	}

	boxW := 84
	cfg := engine.Config{
		Mode:         engine.ModeTime,
		TimeDuration: 30,
		WordLimit:    25,
		Lang:         words.LangFR,
		Punctuation:  false,
		Numbers:      false,
		WrapWidth:    boxW - 8,
	}

	return Model{
		state:    StateMenu,
		themeIdx: tIdx,
		boxWidth: boxW,
		styles:   theme.BuildStyles(theme.Palettes[tIdx], boxW),
		eng:      engine.New(cfg),
		store:    store,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) applyTheme(idx int) {
	if idx < 0 {
		idx = len(theme.Palettes) - 1
	} else if idx >= len(theme.Palettes) {
		idx = 0
	}
	m.themeIdx = idx
	m.styles = theme.BuildStyles(theme.Palettes[idx], m.boxWidth)
	m.store.ThemeIndex = idx
	_ = m.store.Save()
}

func (m *Model) startTest() {
	m.eng.Reset()
	m.isNewPB = false
	m.animFrame = 0
	m.state = StateTyping
}

func (m *Model) finishAndSave() tea.Cmd {
	m.state = StateResults
	netWpm := m.eng.NetWPM()
	rawWpm := m.eng.RawWPM()
	acc := m.eng.Accuracy()
	cons := m.eng.Consistency()
	rank := engine.Rank(netWpm, acc)

	rec := storage.Record{
		CategoryKey: m.categoryKey(),
		ModeLabel:   m.modeLabel(),
		Language:    m.eng.Cfg.Lang.String(),
		NetWPM:      netWpm,
		RawWPM:      rawWpm,
		Accuracy:    acc,
		Consistency: cons,
		Rank:        rank,
		Timestamp:   time.Now(),
	}

	m.isNewPB = m.store.RecordSession(rec, m.eng.ElapsedSeconds(), m.eng.SuddenDeath)
	if m.isNewPB {
		m.animFrame = 1
		return animTickCmd()
	}
	return nil
}

func (m *Model) categoryKey() string {
	switch m.eng.Cfg.Mode {
	case engine.ModeTime:
		return fmt.Sprintf("time_%ds_%s", m.eng.Cfg.TimeDuration, m.eng.Cfg.Lang.String())
	case engine.ModeWords:
		return fmt.Sprintf("words_%d_%s", m.eng.Cfg.WordLimit, m.eng.Cfg.Lang.String())
	case engine.ModeQuote:
		return "quote"
	case engine.ModeSuddenDeath:
		return fmt.Sprintf("suddendeath_%d_%s", m.eng.Cfg.WordLimit, m.eng.Cfg.Lang.String())
	default:
		return "default"
	}
}

func (m *Model) modeLabel() string {
	switch m.eng.Cfg.Mode {
	case engine.ModeTime:
		return fmt.Sprintf("Temps %ds", m.eng.Cfg.TimeDuration)
	case engine.ModeWords:
		return fmt.Sprintf("Mots %d", m.eng.Cfg.WordLimit)
	case engine.ModeQuote:
		return "Citation"
	case engine.ModeSuddenDeath:
		return fmt.Sprintf("Mort Subite (%d)", m.eng.Cfg.WordLimit)
	default:
		return "Test"
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		newBoxW := 84
		if msg.Width > 0 && msg.Width-6 < newBoxW {
			newBoxW = max(54, msg.Width-6)
		}
		if newBoxW != m.boxWidth {
			m.boxWidth = newBoxW
			m.styles = theme.BuildStyles(theme.Palettes[m.themeIdx], m.boxWidth)
			m.eng.Rewrap(m.boxWidth - 8)
		}
		return m, nil

	case tickMsg:
		// Only schedule next tick if still actively typing
		if m.state == StateTyping && m.eng.TimerActive && !m.eng.Finished {
			if m.eng.Tick() {
				return m, m.finishAndSave()
			}
			return m, tickCmd()
		}
		return m, nil

	case animTickMsg:
		// Bounded celebration animation (8 frames = 2 seconds, zero idle drain afterwards)
		if m.state == StateResults && m.isNewPB && m.animFrame > 0 && m.animFrame < 8 {
			m.animFrame++
			return m, animTickCmd()
		}
		return m, nil

	case tea.KeyMsg:
		// Global shortcuts
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyCtrlT:
			m.applyTheme(m.themeIdx + 1)
			return m, nil
		case tea.KeyCtrlZ:
			m.zenMode = !m.zenMode
			return m, nil
		case tea.KeyEsc:
			if m.state != StateMenu {
				m.eng.TimerActive = false
				m.state = StateMenu
				return m, nil
			}
		case tea.KeyCtrlR, tea.KeyTab:
			if m.state == StateTyping || m.state == StateResults {
				m.startTest()
				return m, nil
			}
		}

		switch m.state {
		case StateMenu:
			return m.handleMenuKey(msg)
		case StateTyping:
			return m.handleTypingKey(msg)
		case StateResults:
			return m.handleResultsKey(msg)
		case StateStats:
			return m.handleStatsKey(msg)
		}
	}

	return m, nil
}

func (m Model) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "q":
		return m, tea.Quit
	case "s", "S":
		m.state = StateStats
		return m, nil
	case "p", "P":
		m.eng.Cfg.Punctuation = !m.eng.Cfg.Punctuation
		return m, nil
	case "n", "N":
		m.eng.Cfg.Numbers = !m.eng.Cfg.Numbers
		return m, nil
	case "up", "k":
		m.menuRow--
		if m.menuRow < 0 {
			m.menuRow = 5
		}
	case "down", "j":
		m.menuRow++
		if m.menuRow > 5 {
			m.menuRow = 0
		}
	case "left", "h":
		m.adjustMenuOption(-1)
	case "right", "l":
		m.adjustMenuOption(1)
	case " ":
		if m.menuRow == 3 {
			if m.modifierCol == 0 {
				m.eng.Cfg.Punctuation = !m.eng.Cfg.Punctuation
			} else {
				m.eng.Cfg.Numbers = !m.eng.Cfg.Numbers
			}
			return m, nil
		}
		m.startTest()
		return m, nil
	case "enter":
		if m.menuRow == 3 {
			if m.modifierCol == 0 {
				m.eng.Cfg.Punctuation = !m.eng.Cfg.Punctuation
			} else {
				m.eng.Cfg.Numbers = !m.eng.Cfg.Numbers
			}
			return m, nil
		}
		m.startTest()
		return m, nil
	}
	return m, nil
}

func (m *Model) adjustMenuOption(dir int) {
	switch m.menuRow {
	case 0: // Mode
		modes := []engine.TestMode{engine.ModeTime, engine.ModeWords, engine.ModeQuote, engine.ModeSuddenDeath}
		idx := int(m.eng.Cfg.Mode) + dir
		if idx < 0 {
			idx = len(modes) - 1
		} else if idx >= len(modes) {
			idx = 0
		}
		m.eng.Cfg.Mode = modes[idx]

	case 1: // Duration or Word count
		if m.eng.Cfg.Mode == engine.ModeTime {
			idx := indexOfInt(timeOptions, m.eng.Cfg.TimeDuration) + dir
			if idx < 0 {
				idx = len(timeOptions) - 1
			} else if idx >= len(timeOptions) {
				idx = 0
			}
			m.eng.Cfg.TimeDuration = timeOptions[idx]
		} else if m.eng.Cfg.Mode == engine.ModeWords || m.eng.Cfg.Mode == engine.ModeSuddenDeath {
			idx := indexOfInt(wordOptions, m.eng.Cfg.WordLimit) + dir
			if idx < 0 {
				idx = len(wordOptions) - 1
			} else if idx >= len(wordOptions) {
				idx = 0
			}
			m.eng.Cfg.WordLimit = wordOptions[idx]
		}

	case 2: // Language
		langs := []words.Language{words.LangFR, words.LangEN, words.LangCode}
		cur := 0
		for i, l := range langs {
			if l == m.eng.Cfg.Lang {
				cur = i
				break
			}
		}
		cur += dir
		if cur < 0 {
			cur = len(langs) - 1
		} else if cur >= len(langs) {
			cur = 0
		}
		m.eng.Cfg.Lang = langs[cur]

	case 3: // Modifiers (@ ponctuation / # nombres)
		if m.modifierCol == 0 {
			m.modifierCol = 1
		} else {
			m.modifierCol = 0
		}

	case 4: // Theme
		m.applyTheme(m.themeIdx + dir)
	}
}

func (m Model) handleTypingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Word deletion shortcuts: Ctrl+W, Alt+Backspace, Ctrl+H (often sent by Ctrl+Backspace in terminals)
	if msg.Type == tea.KeyCtrlW || msg.Type == tea.KeyCtrlH || (msg.Type == tea.KeyBackspace && msg.Alt) {
		m.eng.DeleteWord()
		return m, nil
	}

	if msg.Type == tea.KeyBackspace {
		m.eng.Backspace()
		return m, nil
	}

	var charTyped rune
	var validChar bool

	if msg.Type == tea.KeySpace {
		charTyped = ' '
		validChar = true
	} else if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
		// Support multi-rune paste or single rune typing cleanly
		var cmd tea.Cmd
		for _, r := range msg.Runes {
			if m.eng.TypeRune(r) {
				cmd = tickCmd()
			}
			if m.eng.Finished {
				finishCmd := m.finishAndSave()
				return m, tea.Batch(cmd, finishCmd)
			}
		}
		return m, cmd
	}

	if validChar {
		var cmd tea.Cmd
		if m.eng.TypeRune(charTyped) {
			cmd = tickCmd()
		}
		if m.eng.Finished {
			finishCmd := m.finishAndSave()
			return m, tea.Batch(cmd, finishCmd)
		}
		return m, cmd
	}

	return m, nil
}

func (m Model) handleResultsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r", "R", "enter":
		m.startTest()
		return m, nil
	case "m", "M":
		m.state = StateMenu
		return m, nil
	case "s", "S":
		m.state = StateStats
		return m, nil
	case "q", "Q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleStatsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "m", "M", "esc", "backspace", "s", "S":
		m.state = StateMenu
		return m, nil
	case "q", "Q":
		return m, tea.Quit
	}
	return m, nil
}

func indexOfInt(slice []int, val int) int {
	for i, v := range slice {
		if v == val {
			return i
		}
	}
	return 0
}

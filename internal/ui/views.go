package ui

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ricetype/internal/engine"
	"ricetype/internal/theme"
	"ricetype/internal/words"
)

const asciiBanner = `       _          _                    
  _ __(_) ___ ___| |_ _   _ _ __   ___ 
 | '__| |/ __/ _ \ __| | | | '_ \ / _ \
 | |  | | (_|  __/ |_| |_| | |_) |  __/
 |_|  |_|\___\___|\__|\__, | .__/ \___|
                      |___/|_|         `

func (m Model) View() string {
	var content string
	switch m.state {
	case StateMenu:
		content = m.menuView()
	case StateTyping:
		content = m.typingView()
	case StateResults:
		content = m.resultsView()
	case StateStats:
		content = m.statsView()
	}

	if m.termWidth > 0 && m.termHeight > 0 {
		return lipgloss.Place(m.termWidth, m.termHeight, lipgloss.Center, lipgloss.Center, content)
	}
	return content
}

func (m Model) menuView() string {
	var b strings.Builder
	b.Grow(1024)

	b.WriteString(m.styles.Title.Render(asciiBanner) + "\n")
	sub := fmt.Sprintf("  /dev/tty speed trainer  •  rice: %s  •  sessions: %d\n\n",
		m.styles.Palette.Name, m.store.TotalTests)
	b.WriteString(m.styles.Subtitle.Render(sub))

	// Row 0: Mode pills
	modes := []struct {
		mode  engine.TestMode
		label string
	}{
		{engine.ModeTime, "⏱ temps"},
		{engine.ModeWords, "✎ mots"},
		{engine.ModeQuote, "❝ citation"},
		{engine.ModeSuddenDeath, "☠ mort subite"},
	}
	var modePills []string
	for _, item := range modes {
		if m.eng.Cfg.Mode == item.mode {
			modePills = append(modePills, m.styles.PillActive.Render(item.label))
		} else {
			modePills = append(modePills, m.styles.PillInactive.Render(item.label))
		}
	}
	b.WriteString(m.renderMenuRow(0, "Mode      ", strings.Join(modePills, " ")) + "\n\n")

	// Row 1: Target (Duration / Words / Quote info)
	var targetPills []string
	if m.eng.Cfg.Mode == engine.ModeTime {
		for _, sec := range timeOptions {
			lbl := fmt.Sprintf("%ds", sec)
			if m.eng.Cfg.TimeDuration == sec {
				targetPills = append(targetPills, m.styles.PillActive.Render(lbl))
			} else {
				targetPills = append(targetPills, m.styles.PillInactive.Render(lbl))
			}
		}
	} else if m.eng.Cfg.Mode == engine.ModeWords || m.eng.Cfg.Mode == engine.ModeSuddenDeath {
		for _, cnt := range wordOptions {
			lbl := fmt.Sprintf("%d mots", cnt)
			if m.eng.Cfg.WordLimit == cnt {
				targetPills = append(targetPills, m.styles.PillActive.Render(lbl))
			} else {
				targetPills = append(targetPills, m.styles.PillInactive.Render(lbl))
			}
		}
	} else {
		targetPills = append(targetPills, m.styles.PillActive.Render("Citation tech & littérature"))
	}
	b.WriteString(m.renderMenuRow(1, "Objectif  ", strings.Join(targetPills, " ")) + "\n\n")

	// Row 2: Language
	langs := []struct {
		lang  words.Language
		label string
	}{
		{words.LangFR, "FR Français"},
		{words.LangEN, "EN English"},
		{words.LangCode, "</> Code"},
	}
	var langPills []string
	for _, item := range langs {
		if m.eng.Cfg.Lang == item.lang {
			langPills = append(langPills, m.styles.PillActive.Render(item.label))
		} else {
			langPills = append(langPills, m.styles.PillInactive.Render(item.label))
		}
	}
	b.WriteString(m.renderMenuRow(2, "Langue    ", strings.Join(langPills, " ")) + "\n\n")

	// Row 3: Modifiers (@ ponctuation, # nombres)
	pLabel := "@ ponctuation"
	if m.menuRow == 3 && m.modifierCol == 0 {
		pLabel = "▸ @ ponctuation"
	}
	nLabel := "# nombres"
	if m.menuRow == 3 && m.modifierCol == 1 {
		nLabel = "▸ # nombres"
	}
	var modPills []string
	if m.eng.Cfg.Punctuation {
		modPills = append(modPills, m.styles.PillActive.Render(pLabel))
	} else {
		modPills = append(modPills, m.styles.PillInactive.Render(pLabel))
	}
	if m.eng.Cfg.Numbers {
		modPills = append(modPills, m.styles.PillActive.Render(nLabel))
	} else {
		modPills = append(modPills, m.styles.PillInactive.Render(nLabel))
	}
	b.WriteString(m.renderMenuRow(3, "Options   ", strings.Join(modPills, "  ")) + "\n\n")

	// Row 4: Theme selector
	themeContent := m.styles.PillActive.Render("◀ " + theme.Palettes[m.themeIdx].Name + " ▶")
	b.WriteString(m.renderMenuRow(4, "Thème     ", themeContent) + "\n\n")

	// Row 5: Start Button + Personal Best preview
	pbText := ""
	if pb, ok := m.store.PersonalBest[m.categoryKey()]; ok {
		pbText = fmt.Sprintf("   Record catégorie : %.1f WPM (%.0f%%)", pb.NetWPM, pb.Accuracy)
	}
	var startBtn string
	if m.menuRow == 5 {
		startBtn = m.styles.StartBtnFocus.Render("▶ DÉMARRER LE TEST")
	} else {
		startBtn = m.styles.StartBtn.Render("  Démarrer le Test")
	}
	b.WriteString("  " + startBtn + m.styles.Subtitle.Render(pbText) + "\n\n")

	help := "[↑/↓] Naviguer  [←/→] Choisir  [Entrée] Lancer  [P/N] Ponct./Nombres  [S] Records  [Ctrl+T] Thème  [Q] Quitter"
	b.WriteString(m.styles.Subtitle.Render(help))

	return m.styles.Box.Render(b.String())
}

func (m Model) renderMenuRow(rowIdx int, label, content string) string {
	if m.menuRow == rowIdx {
		return m.styles.RowFocused.Render("➔ "+label) + " " + content
	}
	return m.styles.RowUnfocused.Render("  "+label) + " " + content
}

func (m Model) typingView() string {
	var b strings.Builder
	b.Grow(1024)

	innerWidth := m.boxWidth - 8
	if innerWidth < 30 {
		innerWidth = 68
	}

	// 1. Smooth Progress Bar
	b.WriteString(m.renderProgressBar(innerWidth) + "\n\n")

	// 2. Live HUD (hidden in Zen Mode)
	if !m.zenMode {
		var leftHUD string
		switch m.eng.Cfg.Mode {
		case engine.ModeTime:
			leftHUD = fmt.Sprintf("⏱ %ds", m.eng.TimeLeft)
		case engine.ModeWords, engine.ModeSuddenDeath:
			leftHUD = fmt.Sprintf("✎ %d/%d mots", m.eng.WordsCompleted, len(m.eng.Words))
		case engine.ModeQuote:
			leftHUD = fmt.Sprintf("❝ %d/%d mots", m.eng.WordsCompleted, len(m.eng.Words))
		}

		wpmStr := fmt.Sprintf("%.0f WPM", m.eng.LiveNetWPM())
		accStr := fmt.Sprintf("%.0f%%", m.eng.Accuracy())
		comboStr := fmt.Sprintf("⚡ Combo x%d", m.eng.CurrentStreak)

		hudLine := fmt.Sprintf("%s   •   %s %s   •   %s %s   •   %s",
			m.styles.HUDValue.Render(leftHUD),
			m.styles.HUDLabel.Render("Vitesse:"),
			m.styles.HUDValue.Render(wpmStr),
			m.styles.HUDLabel.Render("Précision:"),
			m.styles.HUDValue.Render(accStr),
			m.styles.HUDCombo.Render(comboStr),
		)
		b.WriteString(hudLine + "\n\n")
	} else {
		b.WriteString(m.styles.Subtitle.Render("  [Mode Zen actif]") + "\n\n")
	}

	// 3. 3-Line Sliding Viewport with Run-Length ANSI Batching
	curLine := m.eng.CurrentLineIndex()
	totalLines := len(m.eng.Lines)

	startLine := curLine - 1
	if startLine < 0 {
		startLine = 0
	}
	endLine := startLine + 3
	if endLine > totalLines {
		endLine = totalLines
		if endLine-3 >= 0 {
			startLine = endLine - 3
		}
	}

	renderedLines := 0
	for lineIdx := startLine; lineIdx < endLine; lineIdx++ {
		span := m.eng.Lines[lineIdx]
		dimmed := lineIdx < curLine
		b.WriteString(m.renderLineBatched(span, dimmed) + "\n\n")
		renderedLines++
	}
	for renderedLines < 3 {
		b.WriteString("\n\n")
		renderedLines++
	}

	footer := "[Tab / Ctrl+R] Recommencer  •  [Ctrl+W] Effacer mot  •  [Ctrl+Z] Mode Zen  •  [Ctrl+T] Thème  •  [Esc] Menu"
	b.WriteString(m.styles.Subtitle.Render(footer))

	return m.styles.Box.Render(b.String())
}

// renderProgressBar builds a smooth Unicode progress line (━━━╸───────).
func (m Model) renderProgressBar(width int) string {
	ratio := m.eng.ProgressRatio()
	filled := int(math.Round(ratio * float64(width)))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	var donePart, restPart string
	if filled > 0 {
		if filled < width {
			donePart = strings.Repeat("━", filled-1) + "╸"
		} else {
			donePart = strings.Repeat("━", filled)
		}
	}
	if width-filled > 0 {
		restPart = strings.Repeat("─", width-filled)
	}
	return m.styles.ProgressDone.Render(donePart) + m.styles.ProgressRest.Render(restPart)
}

type spanKind uint8

const (
	kindCorrect spanKind = iota
	kindIncorrect
	kindCursor
	kindUntyped
)

// renderLineBatched groups adjacent characters with the same visual state into single ANSI segments.
func (m Model) renderLineBatched(span engine.LineSpan, dimmed bool) string {
	var out strings.Builder
	var runBuf strings.Builder
	out.Grow((span.End - span.Start) * 2)
	runBuf.Grow(span.End - span.Start)

	typedLen := len(m.eng.TypedRunes)
	var currentKind spanKind
	hasRun := false

	flush := func(k spanKind) {
		if runBuf.Len() == 0 {
			return
		}
		seg := runBuf.String()
		runBuf.Reset()
		switch k {
		case kindCorrect:
			if dimmed {
				out.WriteString(m.styles.DimmedChar.Render(seg))
			} else {
				out.WriteString(m.styles.CorrectChar.Render(seg))
			}
		case kindIncorrect:
			out.WriteString(m.styles.IncorrectChar.Render(seg))
		case kindCursor:
			out.WriteString(m.styles.Cursor.Render(seg))
		case kindUntyped:
			if dimmed {
				out.WriteString(m.styles.DimmedChar.Render(seg))
			} else {
				out.WriteString(m.styles.UntypedChar.Render(seg))
			}
		}
	}

	for idx := span.Start; idx < span.End; idx++ {
		targetR := m.eng.TargetRunes[idx]
		var k spanKind
		displayR := targetR

		if idx < typedLen {
			if m.eng.TypedRunes[idx] == targetR {
				k = kindCorrect
			} else {
				k = kindIncorrect
				if targetR == ' ' {
					displayR = '_'
				}
			}
		} else if idx == typedLen {
			k = kindCursor
		} else {
			k = kindUntyped
		}

		if !hasRun {
			currentKind = k
			hasRun = true
		} else if k != currentKind {
			flush(currentKind)
			currentKind = k
		}
		runBuf.WriteRune(displayR)
	}

	if hasRun {
		flush(currentKind)
	}
	return out.String()
}

func (m Model) resultsView() string {
	var b strings.Builder
	b.Grow(1536)

	netWpm := m.eng.NetWPM()
	rawWpm := m.eng.RawWPM()
	acc := m.eng.Accuracy()
	cons := m.eng.Consistency()
	rank := engine.Rank(netWpm, acc)

	// Header & Rank Badge
	title := "🏁 Session Terminée"
	if m.eng.SuddenDeath {
		title = "☠ Mort Subite — Erreur Fatale !"
	}
	headerRow := m.styles.Title.Render(title) + "   " + m.styles.Badge.Render("★ "+rank)
	b.WriteString(headerRow + "\n")

	// Animated Personal Best Celebration Banner
	if m.isNewPB {
		sparkles := "✦ ★ ✦ NOUVEAU RECORD PERSONNEL ! ✦ ★ ✦"
		if m.animFrame%2 == 1 {
			b.WriteString("\n" + m.styles.PBBanner.Render(sparkles) + "\n")
		} else {
			b.WriteString("\n" + m.styles.PBBannerAlt.Render("★ ✦ ★ NOUVEAU RECORD PERSONNEL ! ★ ✦ ★") + "\n")
		}
	}
	b.WriteString("\n")

	// 3 Stat Cards side by side
	card1 := m.styles.CardBox.Render(
		m.styles.StatTitle.Render("VITESSE NETTE") + "\n" +
			m.styles.StatBigVal.Render(fmt.Sprintf("%.1f WPM", netWpm)) + "\n" +
			m.styles.StatSubVal.Render(fmt.Sprintf("Brut: %.1f WPM", rawWpm)),
	)

	card2 := m.styles.CardBox.Render(
		m.styles.StatTitle.Render("PRÉCISION & RÉGUL.") + "\n" +
			m.styles.StatBigVal.Render(fmt.Sprintf("%.1f%%", acc)) + "\n" +
			m.styles.StatSubVal.Render(fmt.Sprintf("Consistance: %.0f%%", cons)),
	)

	card3 := m.styles.CardBox.Render(
		m.styles.StatTitle.Render("FRAPPES & COMBO") + "\n" +
			m.styles.StatBigVal.Render(fmt.Sprintf("%d / %d / %d", m.eng.CorrectChars, m.eng.WrongKeys, m.eng.CorrectedErrors)) + "\n" +
			m.styles.StatSubVal.Render(fmt.Sprintf("Combo Max: x%d (%.1fs)", m.eng.MaxStreak, m.eng.ElapsedSeconds())),
	)

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, card1, "  ", card2, "  ", card3) + "\n\n")

	// Graduated WPM Graph + Error timeline
	b.WriteString(m.styles.Title.Render("Courbe de Vitesse (WPM) & Erreurs par seconde :") + "\n")
	b.WriteString(m.renderGraduatedGraph(6, m.boxWidth-18) + "\n")

	footer := "[Entrée / R] Rejouer  •  [Tab] Nouveau texte  •  [S] Records & Stats  •  [M / Esc] Menu  •  [Q] Quitter"
	b.WriteString(m.styles.Subtitle.Render(footer))

	return m.styles.Box.Render(b.String())
}

func (m Model) renderGraduatedGraph(height, maxCols int) string {
	history := m.eng.WPMHistory
	errHist := m.eng.ErrorHistory
	if len(history) == 0 {
		return m.styles.Subtitle.Render("  Pas assez de données pour afficher le graphique.\n")
	}

	// Downsample if history exceeds maxCols
	if maxCols < 20 {
		maxCols = 40
	}
	sampledWPM := history
	sampledErr := errHist
	if len(history) > maxCols {
		sampledWPM = make([]float64, maxCols)
		sampledErr = make([]int, maxCols)
		step := float64(len(history)) / float64(maxCols)
		for i := 0; i < maxCols; i++ {
			srcIdx := int(float64(i) * step)
			if srcIdx >= len(history) {
				srcIdx = len(history) - 1
			}
			sampledWPM[i] = history[srcIdx]
			if srcIdx < len(errHist) {
				sampledErr[i] = errHist[srcIdx]
			}
		}
	}

	maxVal := 10.0
	for _, v := range sampledWPM {
		if v > maxVal {
			maxVal = v
		}
	}

	bars := []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	var out strings.Builder

	for r := height - 1; r >= 0; r-- {
		// Y-axis label on top, middle, bottom rows
		var yLabel string
		switch r {
		case height - 1:
			yLabel = fmt.Sprintf("%4.0f │ ", maxVal)
		case height / 2:
			yLabel = fmt.Sprintf("%4.0f │ ", maxVal/2)
		case 0:
			yLabel = "   0 │ "
		default:
			yLabel = "     │ "
		}
		out.WriteString(m.styles.AxisLabel.Render(yLabel))

		var rowBars strings.Builder
		for _, v := range sampledWPM {
			scaled := (v / maxVal) * float64(height)
			rem := scaled - float64(r)
			if rem <= 0 {
				rowBars.WriteString(" ")
			} else if rem >= 1 {
				rowBars.WriteString("█")
			} else {
				idx := int(rem * 8)
				if idx < 1 {
					idx = 1
				} else if idx > 8 {
					idx = 8
				}
				rowBars.WriteString(bars[idx])
			}
		}
		out.WriteString(m.styles.GraphBar.Render(rowBars.String()) + "\n")
	}

	// X-axis separator and error markers
	out.WriteString(m.styles.AxisLabel.Render("     └─" + strings.Repeat("─", len(sampledWPM))) + "\n")
	out.WriteString(m.styles.AxisLabel.Render("  Err: "))
	for i := 0; i < len(sampledWPM); i++ {
		if i < len(sampledErr) && sampledErr[i] > 0 {
			out.WriteString(m.styles.GraphErr.Render("×"))
		} else {
			out.WriteString(m.styles.AxisLabel.Render("·"))
		}
	}
	out.WriteString("\n")

	return out.String()
}

func (m Model) statsView() string {
	var b strings.Builder
	b.Grow(1536)

	b.WriteString(m.styles.Title.Render("🏆 Records Personnels & Historique") + "\n\n")

	// Global summary row
	summary := fmt.Sprintf("Tests complétés : %d   •   Temps total : %.0fs   •   Record absolu : %.1f WPM   •   Moyenne récente : %.1f WPM\n\n",
		m.store.TotalTests, m.store.TotalSeconds, m.store.BestOverallWPM(), m.store.AverageWPM())
	b.WriteString(m.styles.HUDValue.Render(summary))

	// Personal Bests by category
	b.WriteString(m.styles.RowFocused.Render("Meilleurs Scores par Catégorie :") + "\n")
	if len(m.store.PersonalBest) == 0 {
		b.WriteString(m.styles.Subtitle.Render("  Aucun record enregistré pour le moment. Lancez une partie !\n\n"))
	} else {
		keys := make([]string, 0, len(m.store.PersonalBest))
		for k := range m.store.PersonalBest {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			rec := m.store.PersonalBest[k]
			line := fmt.Sprintf("  %-18s [%-4s]  ➔  %6.1f WPM   (Précision: %5.1f%%  |  Consistance: %3.0f%%  |  %s)",
				rec.ModeLabel, rec.Language, rec.NetWPM, rec.Accuracy, rec.Consistency, rec.Rank)
			b.WriteString(m.styles.StatSubVal.Render(line) + "\n")
		}
		b.WriteString("\n")
	}

	// Recent history (up to 8 entries)
	b.WriteString(m.styles.RowFocused.Render("Dernières Sessions :") + "\n")
	if len(m.store.History) == 0 {
		b.WriteString(m.styles.Subtitle.Render("  Aucun historique disponible.\n"))
	} else {
		limit := min(8, len(m.store.History))
		for i := 0; i < limit; i++ {
			rec := m.store.History[i]
			dateStr := rec.Timestamp.Format("02/01 15:04")
			line := fmt.Sprintf("  %s  •  %-16s [%-4s]  •  %5.1f WPM  •  %5.1f%%  •  %s",
				dateStr, rec.ModeLabel, rec.Language, rec.NetWPM, rec.Accuracy, rec.Rank)
			b.WriteString(m.styles.Subtitle.Render(line) + "\n")
		}
	}

	b.WriteString("\n" + m.styles.Subtitle.Render("[Esc / M / S] Retour au Menu  •  [Ctrl+T] Changer Thème  •  [Q] Quitter"))
	return m.styles.Box.Render(b.String())
}

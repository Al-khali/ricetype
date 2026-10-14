package theme

import "github.com/charmbracelet/lipgloss"

type Palette struct {
	Name      string
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Error     lipgloss.Color
	Muted     lipgloss.Color
	Text      lipgloss.Color
	CursorFg  lipgloss.Color
}

var Palettes = []Palette{
	{
		Name:      "Cyberpunk",
		Primary:   lipgloss.Color("#00f5ff"),
		Secondary: lipgloss.Color("#e2b714"),
		Error:     lipgloss.Color("#ff0055"),
		Muted:     lipgloss.Color("#526075"),
		Text:      lipgloss.Color("#e2e8f0"),
		CursorFg:  lipgloss.Color("#000000"),
	},
	{
		Name:      "Catppuccin Mocha",
		Primary:   lipgloss.Color("#cba6f7"),
		Secondary: lipgloss.Color("#fab387"),
		Error:     lipgloss.Color("#f38ba8"),
		Muted:     lipgloss.Color("#6c7086"),
		Text:      lipgloss.Color("#cdd6f4"),
		CursorFg:  lipgloss.Color("#1e1e2e"),
	},
	{
		Name:      "Dracula",
		Primary:   lipgloss.Color("#bd93f9"),
		Secondary: lipgloss.Color("#ff79c6"),
		Error:     lipgloss.Color("#ff5555"),
		Muted:     lipgloss.Color("#6272a4"),
		Text:      lipgloss.Color("#f8f8f2"),
		CursorFg:  lipgloss.Color("#282a36"),
	},
	{
		Name:      "Tokyo Night",
		Primary:   lipgloss.Color("#7aa2f7"),
		Secondary: lipgloss.Color("#ff9e64"),
		Error:     lipgloss.Color("#f7768e"),
		Muted:     lipgloss.Color("#565f89"),
		Text:      lipgloss.Color("#c0caf5"),
		CursorFg:  lipgloss.Color("#1a1b26"),
	},
	{
		Name:      "Nord",
		Primary:   lipgloss.Color("#88c0d0"),
		Secondary: lipgloss.Color("#ebcb8b"),
		Error:     lipgloss.Color("#bf616a"),
		Muted:     lipgloss.Color("#616e88"),
		Text:      lipgloss.Color("#eceff4"),
		CursorFg:  lipgloss.Color("#2e3440"),
	},
	{
		Name:      "Gruvbox Dark",
		Primary:   lipgloss.Color("#8ec07c"),
		Secondary: lipgloss.Color("#fabd2f"),
		Error:     lipgloss.Color("#fb4934"),
		Muted:     lipgloss.Color("#928374"),
		Text:      lipgloss.Color("#ebdbb2"),
		CursorFg:  lipgloss.Color("#282828"),
	},
}

// Styles holds pre-allocated Lipgloss styles for zero-allocation rendering in View().
type Styles struct {
	Palette Palette

	Title         lipgloss.Style
	Subtitle      lipgloss.Style
	Box           lipgloss.Style
	PillActive    lipgloss.Style
	PillInactive  lipgloss.Style
	RowFocused    lipgloss.Style
	RowUnfocused  lipgloss.Style
	StartBtn      lipgloss.Style
	StartBtnFocus lipgloss.Style

	// Typing styles
	CorrectChar   lipgloss.Style
	IncorrectChar lipgloss.Style
	UntypedChar   lipgloss.Style
	DimmedChar    lipgloss.Style
	Cursor        lipgloss.Style

	// HUD & Progress
	HUDLabel     lipgloss.Style
	HUDValue     lipgloss.Style
	HUDCombo     lipgloss.Style
	ProgressDone lipgloss.Style
	ProgressRest lipgloss.Style

	// Results & Stats
	CardBox     lipgloss.Style
	StatTitle   lipgloss.Style
	StatBigVal  lipgloss.Style
	StatSubVal  lipgloss.Style
	Badge       lipgloss.Style
	PBBanner    lipgloss.Style
	PBBannerAlt lipgloss.Style
	GraphBar    lipgloss.Style
	GraphErr    lipgloss.Style
	AxisLabel   lipgloss.Style
}

func BuildStyles(p Palette, boxWidth int) Styles {
	if boxWidth < 50 {
		boxWidth = 82
	}
	return Styles{
		Palette: p,
		Title: lipgloss.NewStyle().
			Foreground(p.Primary).
			Bold(true),
		Subtitle: lipgloss.NewStyle().
			Foreground(p.Muted).
			Italic(true),
		Box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.Muted).
			Padding(1, 3).
			Width(boxWidth),
		PillActive: lipgloss.NewStyle().
			Foreground(p.CursorFg).
			Background(p.Primary).
			Bold(true).
			Padding(0, 1),
		PillInactive: lipgloss.NewStyle().
			Foreground(p.Muted).
			Padding(0, 1),
		RowFocused: lipgloss.NewStyle().
			Foreground(p.Secondary).
			Bold(true),
		RowUnfocused: lipgloss.NewStyle().
			Foreground(p.Muted),
		StartBtn: lipgloss.NewStyle().
			Foreground(p.Text).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.Muted).
			Padding(0, 3),
		StartBtnFocus: lipgloss.NewStyle().
			Foreground(p.CursorFg).
			Background(p.Secondary).
			Bold(true).
			Padding(0, 4),

		CorrectChar: lipgloss.NewStyle().
			Foreground(p.Primary),
		IncorrectChar: lipgloss.NewStyle().
			Foreground(p.Error).
			Underline(true).
			Bold(true),
		UntypedChar: lipgloss.NewStyle().
			Foreground(p.Muted),
		DimmedChar: lipgloss.NewStyle().
			Foreground(p.Muted).
			Faint(true),
		Cursor: lipgloss.NewStyle().
			Background(p.Secondary).
			Foreground(p.CursorFg).
			Bold(true),

		HUDLabel: lipgloss.NewStyle().
			Foreground(p.Muted),
		HUDValue: lipgloss.NewStyle().
			Foreground(p.Secondary).
			Bold(true),
		HUDCombo: lipgloss.NewStyle().
			Foreground(p.Primary).
			Bold(true),
		ProgressDone: lipgloss.NewStyle().
			Foreground(p.Primary),
		ProgressRest: lipgloss.NewStyle().
			Foreground(p.Muted),

		CardBox: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(p.Muted).
			Padding(0, 2).
			Width(22),
		StatTitle: lipgloss.NewStyle().
			Foreground(p.Muted),
		StatBigVal: lipgloss.NewStyle().
			Foreground(p.Secondary).
			Bold(true),
		StatSubVal: lipgloss.NewStyle().
			Foreground(p.Text),
		Badge: lipgloss.NewStyle().
			Foreground(p.CursorFg).
			Background(p.Primary).
			Bold(true).
			Padding(0, 2),
		PBBanner: lipgloss.NewStyle().
			Foreground(p.CursorFg).
			Background(p.Secondary).
			Bold(true).
			Padding(0, 2),
		PBBannerAlt: lipgloss.NewStyle().
			Foreground(p.Secondary).
			Bold(true).
			Padding(0, 2),
		GraphBar: lipgloss.NewStyle().
			Foreground(p.Primary),
		GraphErr: lipgloss.NewStyle().
			Foreground(p.Error).
			Bold(true),
		AxisLabel: lipgloss.NewStyle().
			Foreground(p.Muted),
	}
}

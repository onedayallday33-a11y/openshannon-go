package cli

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Core Colors
	orangePrimary   = lipgloss.Color("#D97706") // Sunset Orange
	orangeSecondary = lipgloss.Color("#F59E0B") // Amber
	slatePrimary    = lipgloss.Color("#1E293B") // Dark Slate
	slateSecondary  = lipgloss.Color("#475569") // Muted Slate
	bgDeep          = lipgloss.Color("#0F172A") // Midnight Background
	textMain        = lipgloss.Color("#E2E8F0") // Off-white text

	// Component Styles
	bannerStyle = lipgloss.NewStyle().
			Foreground(orangePrimary).
			Bold(true).
			MarginLeft(2).
			MarginTop(1)

	taglineStyle = lipgloss.NewStyle().
			Foreground(slateSecondary).
			Italic(true).
			MarginLeft(4).
			MarginBottom(1)

	// Sticky Header Info
	infoBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(slateSecondary).
			Padding(1, 2).
			Margin(0, 2)

	labelStyle = lipgloss.NewStyle().
			Foreground(orangeSecondary).
			Bold(true).
			Width(10)

	valueStyle = lipgloss.NewStyle().
			Foreground(textMain)

	// Chat Blocks
	roleUserStyle = lipgloss.NewStyle().
			Foreground(orangePrimary).
			Background(lipgloss.Color("#2D1D05")). // Subtle orange tint
			Bold(true).
			Padding(0, 1)

	roleAssistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#1E3A8A")). // Subtle blue/slate background
			Bold(true).
			Padding(0, 1)

	// Viewport Layout
	viewportStyle = lipgloss.NewStyle().
			Padding(0, 2)

	// Footer / Input Area
	footerBoxStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(slateSecondary)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(orangePrimary).
			Padding(0, 1)

	// Status & Indicators
	statusStyle = lipgloss.NewStyle().
			Foreground(slateSecondary).
			Italic(true)

	toolCallStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")). // Green icon
			Bold(true).
			MarginLeft(2)

	toolNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94A3B8")).
			Italic(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true).
			Padding(0, 1)

	// Shortcuts & Hints
	shortcutStyle = lipgloss.NewStyle().
			Foreground(slateSecondary).
			MarginLeft(2)

	commandHintStyle = lipgloss.NewStyle().
			Foreground(orangeSecondary).
			Bold(true)

	// Selection Box (Slash Commands)
	suggestionBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(slateSecondary).
			Background(bgDeep).
			Padding(0, 1).
			Width(65)

	selectedCommandStyle = lipgloss.NewStyle().
			Background(orangePrimary).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1)

	unselectedCommandStyle = lipgloss.NewStyle().
			Foreground(slateSecondary).
			Padding(0, 1)

	dividerStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(slateSecondary).
			Margin(0, 2)

	// Scrollbar Styles
	scrollbarStyle = lipgloss.NewStyle().
			Foreground(orangeSecondary)

	scrollbarTrackStyle = lipgloss.NewStyle().
			Foreground(slateSecondary)
)

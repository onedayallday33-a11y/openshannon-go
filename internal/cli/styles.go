package cli

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Brand Colors from OpenClaude Reference
	shannonOrange = lipgloss.Color("#D97706") // Sunset Orange
	shannonSlate  = lipgloss.Color("#475569") // Muted Slate
	shannonDark   = lipgloss.Color("#0F172A") // Deep Background
	shannonGreen  = lipgloss.Color("#10B981") // Success Green (Muted)
	
	// ASCII Banner Style
	bannerStyle = lipgloss.NewStyle().
			Foreground(shannonOrange).
			Bold(true).
			MarginLeft(2).
			MarginTop(1)
			
	taglineStyle = lipgloss.NewStyle().
			Foreground(shannonSlate).
			Italic(true).
			MarginLeft(4).
			MarginBottom(1)
			
	// Info Box Style (Rounded like reference)
	infoBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(shannonSlate).
			Padding(0, 1).
			MarginLeft(2).
			MarginBottom(1)
			
	labelStyle = lipgloss.NewStyle().
			Foreground(shannonOrange).
			Width(10)
			
	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E2E8F0"))
			
	// Text Styles
	roleUserStyle = lipgloss.NewStyle().
			Foreground(shannonOrange).
			Bold(true).
			Padding(0, 1)
			
	roleAssistantStyle = lipgloss.NewStyle().
			Foreground(shannonOrange).
			Bold(true).
			Padding(0, 1)
			
	statusStyle = lipgloss.NewStyle().
			Foreground(shannonSlate).
			Italic(true)
			
	readyIndicatorStyle = lipgloss.NewStyle().
			Foreground(shannonSlate).
			MarginLeft(2)
			
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)
			
	dividerStyle = lipgloss.NewStyle().
			Foreground(shannonSlate).
			Margin(1, 0)
			
	// Suggestion Box Styles
	suggestionBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(shannonSlate).
			Padding(0, 1).
			Background(shannonDark).
			Width(60)

	selectedCommandStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#FFFFFF")).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1)

	unselectedCommandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94A3B8")).
			Padding(0, 1)

	// Layout
	headerStyle = lipgloss.NewStyle().
			Foreground(shannonSlate).
			Bold(true).
			Padding(0, 1).
			MarginBottom(1)
			
	helpStyle = lipgloss.NewStyle().
			Foreground(shannonSlate).
			MarginLeft(2)

	shortcutHintStyle = lipgloss.NewStyle().
			Foreground(shannonSlate).
			MarginLeft(2).
			MarginBottom(1)
)

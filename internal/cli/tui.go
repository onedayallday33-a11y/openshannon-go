package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/onedayallday33-a11y/openshannon-go/internal/agent"
	"github.com/onedayallday33-a11y/openshannon-go/internal/config"
	"github.com/onedayallday33-a11y/openshannon-go/internal/types"
)

const (
	banner = `
 ██████╗ ██████╗ ███████╗███╗   ██╗                            
██╔═══██╗██╔══██╗██╔════╝████╗  ██║                            
██║   ██║██████╔╝█████╗  ██╔██╗ ██║                            
██║   ██║██╔═══╝ ██╔══╝  ██║╚██╗██║                            
╚██████╔╝██║     ███████╗██║ ╚████║                            
 ╚═════╝ ╚═╝     ╚══════╝╚═╝  ╚═══╝                            
                                                               
███████╗██╗  ██╗ █████╗ ███╗   ██╗███╗   ██╗ ██████╗ ███╗   ██╗
██╔════╝██║  ██║██╔══██╗████╗  ██║████╗  ██║██╔═══██╗████╗  ██║
███████╗███████║███████║██╔██╗ ██║██╔██╗ ██║██║   ██║██╔██╗ ██║
╚════██║██╔══██║██╔══██║██║╚██╗██║██║╚██╗██║██║   ██║██║╚██╗██║
███████║██║  ██║██║  ██║██║ ╚████║██║ ╚████║╚██████╔╝██║ ╚████║
╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝╚═╝  ╚═══╝ ╚═════╝ ╚═╝  ╚═══╝                                                               
`
	tagline = "✦ Based on Go. Responsive. Without Limits. ✦"
)

type state int

const (
	stateInput state = iota
	stateThinking
)

type Message struct {
	Role    string
	Content string
}

type Model struct {
	agent           *agent.Agent
	viewport        viewport.Model
	textarea        textarea.Model
	spinner         spinner.Model
	renderer        *glamour.TermRenderer
	state           state
	history         []Message
	currResponse    string
	width           int
	height          int
	err             error
	showSuggestions bool
	suggestions     []agent.SlashCommand
	suggestionIdx   int
	executingTool   string
	isDragging      bool
	dragStartY      int
	
	// Channels for streaming
	eventChan chan types.AgentEvent
	errChan   chan error

	// Cache
	renderedHistory string
	cachedHeader    string
}

func NewModel(a *agent.Agent) Model {
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 10000
	ta.SetHeight(1)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(orangePrimary)

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)

	vp := viewport.New(80, 20)

	return Model{
		agent:    a,
		textarea: ta,
		viewport: vp,
		spinner:  s,
		renderer: renderer,
		state:    stateInput,
		width:    80, // Default fallback
		height:   24, // Default fallback
	}
}

type initMsg struct{}
type eventMsg types.AgentEvent
type finishMsg string

func (m Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, func() tea.Msg { return initMsg{} })
}

type responseMsg string
type errMsg struct{ err error }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		spCmd tea.Cmd
	)

	// Intercept keys if suggestions are showing
	if m.showSuggestions {
		if kmsg, ok := msg.(tea.KeyMsg); ok {
			switch kmsg.Type {
			case tea.KeyUp:
				m.suggestionIdx--
				if m.suggestionIdx < 0 {
					m.suggestionIdx = len(m.suggestions) - 1
				}
				return m, nil
			case tea.KeyDown:
				m.suggestionIdx++
				if m.suggestionIdx >= len(m.suggestions) {
					m.suggestionIdx = 0
				}
				return m, nil
			case tea.KeyEnter, tea.KeyTab:
				selected := m.suggestions[m.suggestionIdx]
				m.textarea.SetValue("/" + selected.Name() + " ")
				m.showSuggestions = false
				m.textarea.CursorEnd()
				return m, nil
			case tea.KeyEsc:
				m.showSuggestions = false
				return m, nil
			}
		}
	}

	// Message sending and interception
	if kmsg, ok := msg.(tea.KeyMsg); ok {
		switch kmsg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			if kmsg.Alt {
				m.textarea.InsertString("\n")
				return m, nil
			}
			if m.state == stateInput && strings.TrimSpace(m.textarea.Value()) != "" {
				userPrompt := strings.TrimSpace(m.textarea.Value())
				if userPrompt == "/clear" {
					m.history = nil
				} else {
					m.history = append(m.history, Message{Role: "user", Content: userPrompt})
				}
				m.textarea.Reset()
				m.state = stateThinking
				m.currResponse = ""
				m.executingTool = ""
				m.updateViewport()

				m.eventChan = make(chan types.AgentEvent)
				m.errChan = make(chan error, 1)

				return m, tea.Batch(
					m.spinner.Tick,
					m.runAgent(userPrompt),
				)
			}
			return m, nil
		case tea.KeyCtrlJ: // Newline shortcut (often Ctrl+Enter in terminals)
			m.textarea.InsertString("\n")
			return m, nil
		case tea.KeyCtrlV:
			raw, _ := clipboard.ReadAll()
			m.textarea.SetValue(m.textarea.Value() + raw)
			return m, nil
		case tea.KeyCtrlK:
			// Copy last assistant response
			for i := len(m.history) - 1; i >= 0; i-- {
				if m.history[i].Role == "assistant" {
					_ = clipboard.WriteAll(m.history[i].Content)
					break
				}
			}
			return m, nil
		}

		// Trigger suggestions dynamically
		val := m.textarea.Value()
		if strings.HasPrefix(val, "/") && !strings.Contains(val, " ") {
			prefix := val[1:]
			allCmds := agent.GetDispatcher().GetRegisteredCommands()
			var filtered []agent.SlashCommand
			for _, cmd := range allCmds {
				if strings.HasPrefix(cmd.Name(), prefix) {
					filtered = append(filtered, cmd)
				}
			}

			if len(filtered) > 0 {
				m.suggestions = filtered
				m.showSuggestions = true
				if m.suggestionIdx >= len(filtered) {
					m.suggestionIdx = 0
				}
			} else {
				m.showSuggestions = false
			}
		} else {
			m.showSuggestions = false
		}
	}

	// Handle Mouse Events for Scrollbar Dragging
	if mmsg, ok := msg.(tea.MouseMsg); ok {
		if mmsg.Type == tea.MouseLeft {
			if mmsg.Action == tea.MouseActionPress && mmsg.X >= m.viewport.Width && mmsg.Y < m.viewport.Height {
				m.isDragging = true
				return m, nil
			}
			if mmsg.Action == tea.MouseActionRelease {
				m.isDragging = false
			}
		}

		if m.isDragging && mmsg.Action == tea.MouseActionMotion {
			if m.viewport.TotalLineCount() > m.viewport.Height {
				percent := float64(mmsg.Y) / float64(m.viewport.Height)
				targetY := int(percent * float64(m.viewport.TotalLineCount()))
				m.viewport.SetYOffset(targetY)
				return m, nil
			}
		}
	}

	oldHeight := m.textarea.Height()
	m.textarea, tiCmd = m.textarea.Update(msg)

	// Dynamic height adjustment (min 1, max 5)
	newHeight := m.textarea.LineCount()
	if newHeight < 1 {
		newHeight = 1
	}
	if newHeight > 5 {
		newHeight = 5
	}

	if newHeight != oldHeight {
		m.textarea.SetHeight(newHeight)
		m.recalculateLayout()
	}

	m.viewport, vpCmd = m.viewport.Update(msg)
	m.spinner, spCmd = m.spinner.Update(msg)

	switch msg := msg.(type) {
	case initMsg:
		m.updateViewport()
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalculateLayout()
		m.updateViewport()
		return m, nil

	case eventMsg:
		switch msg.Type {
		case types.EventTextDelta:
			m.currResponse += msg.Text
		case types.EventToolStart:
			m.executingTool = msg.Tool.Name
		case types.EventToolEnd:
			m.executingTool = ""
		case types.EventThinkingStart:
			m.executingTool = ""
			m.currResponse = ""
		}
		m.updateViewport()
		return m, m.waitForStream(m.eventChan, m.errChan)

	case finishMsg:
		m.renderedHistory = "" // Invalidate cache on new message
		if string(msg) != "Conversation history cleared." {
			m.history = append(m.history, Message{Role: "assistant", Content: string(msg)})
		}
		m.state = stateInput
		m.currResponse = ""
		m.executingTool = ""
		m.updateViewport()
		return m, nil

	case errMsg:
		m.err = msg.err
		m.state = stateInput
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd, spCmd)
}

func (m *Model) recalculateLayout() {
	footerHeight := lipgloss.Height(m.renderFooter())

	m.viewport.Width = m.width - 2 // Leave space for scrollbar and padding
	m.viewport.Height = m.height - footerHeight - 1
	if m.viewport.Height < 1 {
		m.viewport.Height = 1
	}
	m.textarea.SetWidth(m.width - 6) // Internal padding
}

func (m *Model) updateViewport() {
	var sb strings.Builder

	// 1. Header (Cached)
	if m.cachedHeader == "" {
		m.cachedHeader = m.renderHeader()
	}
	sb.WriteString(m.cachedHeader)
	sb.WriteString("\n")

	// 2. Welcome or History
	if len(m.history) == 0 {
		sb.WriteString(m.renderWelcome())
	} else {
		// Optimization: Check if history was already rendered
		// In a production app, we'd only render the newest items.
		// For now, let's at least avoid glamour re-runs on large chunks if possible.
		// (Simplified incremental rendering)
		if m.renderedHistory == "" || m.state == stateInput {
			var histSb strings.Builder
			for _, msg := range m.history {
				histSb.WriteString(m.renderMessage(msg))
				histSb.WriteString("\n")
			}
			m.renderedHistory = histSb.String()
		}
		sb.WriteString(m.renderedHistory)
	}

	// 3. Current Stream (LLM Thinking/Streaming)
	if m.state == stateThinking {
		sb.WriteString(roleAssistantStyle.Render("SHANNON"))
		sb.WriteString("\n")
		if m.executingTool != "" {
			sb.WriteString(toolCallStyle.Render("⚙ RUNNING "))
			sb.WriteString(toolNameStyle.Render(m.executingTool))
			sb.WriteString(m.spinner.View())
		} else if m.currResponse != "" {
			// Note: currResponse is only the *current* turn delta
			rendered, _ := m.renderer.Render(m.currResponse)
			sb.WriteString(rendered)
		} else {
			sb.WriteString(m.spinner.View() + statusStyle.Render(" Thinking..."))
		}
	}

	m.viewport.SetContent(viewportStyle.Render(sb.String()))
	if m.viewport.Height > 0 {
		m.viewport.GotoBottom()
	}
}

func (m Model) renderMessage(msg Message) string {
	var sb strings.Builder
	if msg.Role == "user" {
		// Add vertical space before each user message turn
		sb.WriteString("\n")
		sb.WriteString(roleUserStyle.Render("USER"))
		sb.WriteString("\n")
		sb.WriteString(lipgloss.NewStyle().PaddingLeft(2).PaddingBottom(1).Render(msg.Content))
	} else {
		sb.WriteString(roleAssistantStyle.Render("SHANNON"))
		sb.WriteString("\n")
		rendered, _ := m.renderer.Render(msg.Content)
		sb.WriteString(rendered)
		// Add space below assistant response
		sb.WriteString("\n")
	}
	return sb.String()
}

func (m Model) renderHeader() string {
	bannerTxt := bannerStyle.Render(banner)
	tag := taglineStyle.Render(tagline)

	info := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, labelStyle.Render("Provider"), valueStyle.Render("OpenAI (Compatible)")),
		lipgloss.JoinHorizontal(lipgloss.Left, labelStyle.Render("Model"), valueStyle.Render(config.OpenAIModel())),
		lipgloss.JoinHorizontal(lipgloss.Left, labelStyle.Render("Endpoint"), valueStyle.Render(config.OpenAIBaseURL())),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		bannerTxt,
		tag,
		infoBoxStyle.Render(info),
	)
}

func (m Model) renderWelcome() string {
	res := "\n"
	res += lipgloss.NewStyle().
		Foreground(orangeSecondary).
		Bold(true).
		MarginLeft(2).
		Render("● shannon Ready - type /help to begin") + "\n"
	res += lipgloss.NewStyle().
		Foreground(slateSecondary).
		MarginLeft(2).
		Render("openshannon v0.1.0 (Go Native)") + "\n"
	return res
}

func (m Model) renderFooter() string {
	input := inputStyle.Width(m.width - 4).Render(m.textarea.View())

	hints := lipgloss.JoinHorizontal(lipgloss.Top,
		shortcutStyle.Render("Enter "), commandHintStyle.Render("send"),
		shortcutStyle.Render(" / "), commandHintStyle.Render("commands"),
		shortcutStyle.Render(" Ctrl+J "), commandHintStyle.Render("newline"),
		shortcutStyle.Render(" Ctrl+K "), commandHintStyle.Render("copy last"),
	)

	return footerBoxStyle.Width(m.width).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			input,
			hints,
		),
	)
}

func (m Model) renderSuggestions() string {
	if !m.showSuggestions || len(m.suggestions) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, cmd := range m.suggestions {
		name := "/" + cmd.Name()
		desc := cmd.Description()

		line := fmt.Sprintf("%-15s %s", name, desc)
		if i == m.suggestionIdx {
			sb.WriteString(selectedCommandStyle.Render(line))
		} else {
			sb.WriteString(unselectedCommandStyle.Render(line))
		}
		sb.WriteString("\n")
	}
	return suggestionBoxStyle.Render(sb.String())
}

func (m Model) renderScrollbar() string {
	if m.viewport.TotalLineCount() <= m.viewport.Height {
		return ""
	}

	trackHeight := m.viewport.Height
	if trackHeight <= 0 {
		return ""
	}

	// Calculate thumb height and position
	totalContentHeight := m.viewport.TotalLineCount()
	visibleHeight := m.viewport.Height
	scrollOffset := m.viewport.YOffset

	thumbHeight := int(float64(visibleHeight) * float64(visibleHeight) / float64(totalContentHeight))
	if thumbHeight < 1 {
		thumbHeight = 1
	}

	maxScroll := totalContentHeight - visibleHeight
	thumbOffset := 0
	if maxScroll > 0 {
		thumbOffset = int(float64(scrollOffset) * float64(visibleHeight-thumbHeight) / float64(maxScroll))
	}

	var sb strings.Builder
	for i := 0; i < trackHeight; i++ {
		if i >= thumbOffset && i < thumbOffset+thumbHeight {
			sb.WriteString(scrollbarStyle.Render("█"))
		} else {
			sb.WriteString(scrollbarTrackStyle.Render("│"))
		}
		if i < trackHeight-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func (m Model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	viewport := m.viewport.View()
	scrollbar := m.renderScrollbar()
	suggestions := m.renderSuggestions()
	footer := m.renderFooter()

	content := viewport
	if scrollbar != "" {
		content = lipgloss.JoinHorizontal(lipgloss.Top, viewport, scrollbar)
	}

	if suggestions != "" {
		return lipgloss.JoinVertical(lipgloss.Left,
			content,
			suggestions,
			footer,
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		content,
		footer,
	)
}

func (m Model) runAgent(prompt string) tea.Cmd {
	return func() tea.Msg {
		if m.errChan == nil || m.eventChan == nil {
			return errMsg{fmt.Errorf("channels not initialized")}
		}

		ctx := context.Background()
		go func() {
			output, err := m.agent.Run(ctx, prompt, func(ev types.AgentEvent) {
				m.eventChan <- ev
			})
			if err != nil {
				m.errChan <- err
				return
			}
			m.eventChan <- types.AgentEvent{Type: "FINISH", Text: output}
		}()

		return m.waitForStream(m.eventChan, m.errChan)()
	}
}

func (m Model) waitForStream(ch chan types.AgentEvent, errCh chan error) tea.Cmd {
	return func() tea.Msg {
		select {
		case err := <-errCh:
			return errMsg{err}
		case ev, ok := <-ch:
			if !ok {
				return nil
			}
			if ev.Type == "FINISH" {
				return finishMsg(ev.Text)
			}
			return eventMsg(ev)
		}
	}
}

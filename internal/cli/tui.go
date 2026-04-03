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
}

func NewModel(a *agent.Agent) Model {
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 10000
	ta.SetHeight(3)

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
				// Handle slash commands locally if needed, but Agent.Run does it.
				m.history = append(m.history, Message{Role: "user", Content: userPrompt})
				m.textarea.Reset()
				m.state = stateThinking
				m.currResponse = ""
				m.executingTool = ""
				m.updateViewport()
				return m, tea.Batch(
					m.spinner.Tick,
					m.runAgent(userPrompt),
				)
			}
			return m, nil // Block Enter from adding newline normally
		case tea.KeyCtrlV:
			raw, _ := clipboard.ReadAll()
			m.textarea.SetValue(m.textarea.Value() + raw)
			return m, nil
		}

		// Trigger suggestions
		val := m.textarea.Value()
		if val == "/" {
			m.showSuggestions = true
			m.suggestions = agent.GetDispatcher().GetRegisteredCommands()
			m.suggestionIdx = 0
		} else if !strings.HasPrefix(val, "/") {
			m.showSuggestions = false
		}
	}

	m.textarea, tiCmd = m.textarea.Update(msg)
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
		return m, nil

	case finishMsg:
		m.history = append(m.history, Message{Role: "assistant", Content: string(msg)})
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

	m.viewport.Width = m.width
	m.viewport.Height = m.height - footerHeight - 1
	if m.viewport.Height < 1 {
		m.viewport.Height = 1
	}
	m.textarea.SetWidth(m.width - 6) // Internal padding
}

func (m *Model) updateViewport() {
	var sb strings.Builder

	// Show the branding and session info at the very top of the scrollback
	sb.WriteString(m.renderHeader())
	sb.WriteString("\n")

	if len(m.history) == 0 {
		sb.WriteString(m.renderWelcome())
	} else {
		for _, msg := range m.history {
			sb.WriteString(m.renderMessage(msg))
			sb.WriteString("\n")
		}
	}

	if m.state == stateThinking {
		sb.WriteString(roleAssistantStyle.Render("SHANNON"))
		sb.WriteString("\n")
		if m.executingTool != "" {
			sb.WriteString(toolCallStyle.Render("⚙ RUNNING "))
			sb.WriteString(toolNameStyle.Render(m.executingTool))
			sb.WriteString(m.spinner.View())
		} else if m.currResponse != "" {
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
		shortcutStyle.Render(" Alt+Enter "), commandHintStyle.Render("newline"),
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

func (m Model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	viewport := m.viewport.View()
	suggestions := m.renderSuggestions()
	footer := m.renderFooter()

	if suggestions != "" {
		return lipgloss.JoinVertical(lipgloss.Left,
			viewport,
			suggestions,
			footer,
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		viewport,
		footer,
	)
}

func (m Model) runAgent(prompt string) tea.Cmd {
	return func() tea.Msg {
		ch := make(chan types.AgentEvent)
		errCh := make(chan error, 1)

		ctx := context.Background()
		go func() {
			output, err := m.agent.Run(ctx, prompt, func(ev types.AgentEvent) {
				ch <- ev
			})
			if err != nil {
				errCh <- err
				return
			}
			ch <- types.AgentEvent{Type: "FINISH", Text: output}
		}()

		return m.waitForStream(ch, errCh)()
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
			return tea.Sequence(func() tea.Msg { return eventMsg(ev) }, m.waitForStream(ch, errCh))()
		}
	}
}

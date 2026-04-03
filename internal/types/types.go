package types

import "github.com/onedayallday33-a11y/openshannon-go/internal/toolapi"

// Role constants
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Message represents a single message in the conversation history
type Message struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

// ContentBlock represents a generic block of content (text, image, tool use/result)
type ContentBlock struct {
	Type       string       `json:"type"` // "text", "image", "tool_use", "tool_result"
	Text       string       `json:"text,omitempty"`
	Image      *ImageSource `json:"image,omitempty"`
	ToolUse    *ToolUse     `json:"tool_use,omitempty"`
	ToolResult *ToolResult  `json:"tool_result,omitempty"`
}

// ImageSource contains base64 image data
type ImageSource struct {
	Type      string `json:"type"` // "base64"
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// ToolUse represents a request from the LLM to use a tool
type ToolUse struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// ToolResult represents the output of a tool execution sent back to the LLM
type ToolResult struct {
	ToolUseID string      `json:"tool_use_id"`
	Content   interface{} `json:"content"`
}

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusInProgress TaskStatus = "in_progress"
	StatusCompleted  TaskStatus = "completed"
	StatusBlocked    TaskStatus = "blocked"
)

// Task represents a single trackable item
type Task struct {
	ID          int        `json:"id"`
	Subject     string     `json:"subject"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
}

// AgentConfig for initializing an agent
type AgentConfig struct {
	Model      string
	System     string
	MaxTurns   int
	Tools      []toolapi.Tool
}

// AgentEventType defines the kind of progress event
type AgentEventType string

const (
	EventTextDelta     AgentEventType = "text"
	EventToolStart     AgentEventType = "tool_start"
	EventToolEnd       AgentEventType = "tool_end"
	EventThinkingStart AgentEventType = "thinking_start"
)

// AgentEvent is used for real-time progress reporting
type AgentEvent struct {
	Type       AgentEventType
	Text       string
	Tool       *ToolUse
	ToolResult string
}

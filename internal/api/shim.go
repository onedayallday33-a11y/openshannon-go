package api

import (
	"fmt"
	"github.com/onedayallday33-a11y/openshannon-go/internal/types"
)

// ConvertRequest translates Anthropic format to OpenAI chat completions format
func ConvertRequest(req *AnthropicMessageRequest) *OpenAIChatRequest {
	openaiReq := &OpenAIChatRequest{
		Model:                req.Model,
		Stream:               req.Stream,
		MaxCompletionTokens:  req.MaxTokens,
		MaxTokens:            req.MaxTokens,
		Temperature:          req.Temperature,
		TopP:                 req.TopP,
	}

	if req.Stream {
		openaiReq.StreamOptions = &StreamOption{IncludeUsage: true}
	}

	// Translate System
	systemMsg := ""
	if req.System != nil {
		switch s := req.System.(type) {
		case string:
			systemMsg = s
		}
	}
	if systemMsg != "" {
		openaiReq.Messages = append(openaiReq.Messages, OpenAIMessage{
			Role:    "system",
			Content: systemMsg,
		})
	}

	// Translate Messages
	for _, m := range req.Messages {
		var openaiContent interface{}
		var toolCalls []OpenAIToolCall
		var toolCallID string

		switch content := m.Content.(type) {
		case string:
			openaiContent = content
		case []types.ContentBlock:
			// If it's just one text block, many APIs prefer a string
			if len(content) == 1 && content[0].Type == "text" {
				openaiContent = content[0].Text
			} else {
				var blocks []interface{}
				for _, b := range content {
					switch b.Type {
					case "text":
						blocks = append(blocks, map[string]interface{}{
							"type": "text",
							"text": b.Text,
						})
					case "tool_use":
						// OpenAI handles tool use via ToolCalls field, not content
						toolCalls = append(toolCalls, OpenAIToolCall{
							ID:   b.ToolUse.ID,
							Type: "function",
							Function: OpenAIToolFunction{
								Name:      b.ToolUse.Name,
								Arguments: "{}", // Will be filled if needed, but usually we send results
							},
						})
					case "tool_result":
						// OpenAI handles tool results as a separate message role "tool"
						// But here we are still inside the message loop.
						// We'll handle "tool" role below.
						toolCallID = b.ToolResult.ToolUseID
						openaiContent = fmt.Sprintf("%v", b.ToolResult.Content)
					}
				}
				if len(blocks) > 0 {
					openaiContent = blocks
				}
			}
		}

		role := m.Role
		if m.Role == "user" && toolCallID != "" {
			role = "tool"
		}

		msg := OpenAIMessage{
			Role:    role,
			Content: openaiContent,
		}
		if len(toolCalls) > 0 {
			msg.ToolCalls = toolCalls
		}
		if toolCallID != "" {
			msg.ToolCallID = toolCallID
		}

		openaiReq.Messages = append(openaiReq.Messages, msg)
	}

	// Translate Tools
	for _, t := range req.Tools {
		tool := OpenAITool{
			Type: "function",
		}
		tool.Function.Name = t.Name
		tool.Function.Description = t.Description
		tool.Function.Parameters = t.InputSchema
		openaiReq.Tools = append(openaiReq.Tools, tool)
	}

	return openaiReq
}

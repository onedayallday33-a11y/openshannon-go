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
		switch content := m.Content.(type) {
		case string:
			openaiReq.Messages = append(openaiReq.Messages, OpenAIMessage{
				Role:    m.Role,
				Content: content,
			})
		case []types.ContentBlock:
			var currentBlocks []interface{}
			var toolCalls []OpenAIToolCall

			for _, b := range content {
				switch b.Type {
				case "text":
					currentBlocks = append(currentBlocks, map[string]interface{}{
						"type": "text",
						"text": b.Text,
					})
				case "tool_use":
					argsJSON, _ := json.Marshal(b.ToolUse.Input)
					toolCalls = append(toolCalls, OpenAIToolCall{
						ID:   b.ToolUse.ID,
						Type: "function",
						Function: OpenAIToolFunction{
							Name:      b.ToolUse.Name,
							Arguments: string(argsJSON),
						},
					})
				case "tool_result":
					// If we have accumulated text/tool_use, flush them first
					if len(currentBlocks) > 0 || len(toolCalls) > 0 {
						var content interface{}
						if len(currentBlocks) == 1 && currentBlocks[0].(map[string]interface{})["type"] == "text" {
							content = currentBlocks[0].(map[string]interface{})["text"]
						} else if len(currentBlocks) > 0 {
							content = currentBlocks
						} else if m.Role == "assistant" && len(toolCalls) > 0 {
							// For Assistant role, if we have tool calls, content can be null or empty
							content = ""
						} else {
							content = ""
						}
						
						openaiReq.Messages = append(openaiReq.Messages, OpenAIMessage{
							Role:      m.Role,
							Content:   content,
							ToolCalls: toolCalls,
						})
						currentBlocks = nil
						toolCalls = nil
					}
					
					// Now add the tool result as a separate message
					openaiReq.Messages = append(openaiReq.Messages, OpenAIMessage{
						Role:       "tool",
						Content:    fmt.Sprintf("%v", b.ToolResult.Content),
						ToolCallID: b.ToolResult.ToolUseID,
					})
				}
			}

			// Final flush for remaining blocks
			if len(currentBlocks) > 0 || len(toolCalls) > 0 {
				var content interface{}
				if len(currentBlocks) == 1 && currentBlocks[0].(map[string]interface{})["type"] == "text" {
					content = currentBlocks[0].(map[string]interface{})["text"]
				} else if len(currentBlocks) > 0 {
					content = currentBlocks
				} else if m.Role == "assistant" && len(toolCalls) > 0 {
					content = ""
				} else {
					content = ""
				}

				openaiReq.Messages = append(openaiReq.Messages, OpenAIMessage{
					Role:      m.Role,
					Content:   content,
					ToolCalls: toolCalls,
				})
			}
		}
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

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/onedayallday33-a11y/openshannon-go/internal/api"
	"github.com/onedayallday33-a11y/openshannon-go/internal/types"
)

// Run executes the agent loop for a given prompt
func (a *Agent) Run(ctx context.Context, prompt string, onEvent func(types.AgentEvent)) (string, error) {
	// 0. Check for Slash Commands
	res, err := GetDispatcher().Dispatch(ctx, a, prompt)
	if err != nil {
		return "", err
	}
	if res.IsHandled {
		if res.DirectOutput != "" {
			return res.DirectOutput, nil
		}
		if res.PromptText != "" {
			prompt = res.PromptText
		}
	}

	// 1. Initial Prompt
	a.AddMessage(types.Message{
		Role: types.RoleUser,
		Content: []types.ContentBlock{
			{Type: "text", Text: prompt},
		},
	})

	client := api.NewClient()

	for turn := 0; turn < a.Config.MaxTurns; turn++ {
		// Emit thinking event
		if onEvent != nil {
			onEvent(types.AgentEvent{Type: types.EventThinkingStart})
		}

		// 2. Prepare API Request (Cached)
		if len(a.anthropicToolsCache) == 0 {
			for _, t := range a.Config.Tools {
				a.anthropicToolsCache = append(a.anthropicToolsCache, api.AnthropicTool{
					Name:        t.Name(),
					Description: t.Description(),
					InputSchema: t.InputSchema(),
				})
			}
		}

		// 3. Call LLM (Streaming)
		req := &api.AnthropicMessageRequest{
			Model:    a.Config.Model,
			System:   a.Config.System,
			Messages: a.toAnthropicMessages(),
			Tools:    a.anthropicToolsCache,
			Stream:   true,
		}

		resp, err := client.DoRequest(ctx, req)
		if err != nil {
			return "", err
		}

		// 4. Handle non-200 status codes (Must check before starting stream)
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
		}

		// 5. Start streaming events (Ownership of resp.Body moves to StreamEvents)
		events, errCh := api.StreamEvents(resp)
		
		assistantMessage := types.Message{Role: types.RoleAssistant, Content: []types.ContentBlock{}}
		var currentText strings.Builder
		var currentTools = make(map[int]*types.ToolUse)
		var currentToolJSON = make(map[int]*strings.Builder)
		var usage *types.Usage
		
		loop:
		for {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case err := <-errCh:
				if err != nil {
					return "", err
				}
				break loop
			case event, ok := <-events:
				if !ok {
					break loop
				}
				
				switch event.Type {
				case "content_block_delta":
					deltaType, _ := event.Delta["type"].(string)
					
					if deltaType == "text_delta" {
						text, _ := event.Delta["text"].(string)
						currentText.WriteString(text)
						if onEvent != nil {
							onEvent(types.AgentEvent{Type: types.EventTextDelta, Text: text})
						}
					} else if deltaType == "input_json_delta" {
						choiceIdx := event.Index
						
						if cb := event.ContentBlk; cb != nil {
							if cb["type"] == "tool_use" {
								if _, exists := currentTools[choiceIdx]; !exists {
									currentTools[choiceIdx] = &types.ToolUse{
										ID:   cb["id"].(string),
										Name: cb["name"].(string),
									}
									currentToolJSON[choiceIdx] = &strings.Builder{}
								}
							}
						}
						
						if partialJSON, ok := event.Delta["partial_json"].(string); ok {
							if sb, ok := currentToolJSON[choiceIdx]; ok {
								sb.WriteString(partialJSON)
							}
						}
					}
				case "message_delta":
					if event.Usage != nil {
						usage = &types.Usage{
							InputTokens:  event.Usage["input_tokens"],
							OutputTokens: event.Usage["output_tokens"],
						}
					}
				}
			}
		}

		// 4. Finalize text and tool calls for this turn
		finalText := currentText.String()
		if finalText != "" {
			assistantMessage.Content = append(assistantMessage.Content, types.ContentBlock{
				Type: "text",
				Text: finalText,
			})
		}
		if usage != nil {
			assistantMessage.Usage = usage
		}

		// Parse accumulated JSON for tools
		for idx, tool := range currentTools {
			var input map[string]interface{}
			if sb, ok := currentToolJSON[idx]; ok {
				if err := json.Unmarshal([]byte(sb.String()), &input); err != nil {
					// Fallback: use an empty map if JSON is invalid, but log the error
					fmt.Printf("Warning: failed to parse tool input: %v\n", err)
					input = make(map[string]interface{})
				}
			}
			tool.Input = input
			assistantMessage.Content = append(assistantMessage.Content, types.ContentBlock{
				Type:    "tool_use",
				ToolUse: tool,
			})
		}

		a.AddMessage(assistantMessage)

		// 5. Execute tools if any
		hasToolUse := len(currentTools) > 0
		if !hasToolUse {
			if a.OnTurnEnd != nil {
				a.OnTurnEnd(a)
			}
			return finalText, nil
		}

		for _, block := range assistantMessage.Content {
			if block.Type == "tool_use" {
				if onEvent != nil {
					onEvent(types.AgentEvent{Type: types.EventToolStart, Tool: block.ToolUse})
				}

				// Take snapshot before tool use (simplified)
				// For now, let's track common files mentioned in input
				if filePath, ok := block.ToolUse.Input["path"].(string); ok {
					_ = a.FileHistory.TrackFile(filePath)
				}
				if filePath, ok := block.ToolUse.Input["filePath"].(string); ok {
					_ = a.FileHistory.TrackFile(filePath)
				}

				res, err := a.HandleToolUse(ctx, block.ToolUse.ID, block.ToolUse.Name, block.ToolUse.Input)
				if err != nil {
					res = fmt.Sprintf("Error: %v", err)
				}
				
				if onEvent != nil {
					onEvent(types.AgentEvent{Type: types.EventToolEnd, Tool: block.ToolUse, ToolResult: res})
				}

				a.AddMessage(types.Message{
					Role: types.RoleUser,
					Content: []types.ContentBlock{
						{
							Type: "tool_result",
							ToolResult: &types.ToolResult{
								ToolUseID: block.ToolUse.ID,
								Content:   res,
							},
						},
					},
				})
			}
		}
		if a.OnTurnEnd != nil {
			a.OnTurnEnd(a)
		}
	}

	return "", fmt.Errorf("reached max turns (%d) without final answer", a.Config.MaxTurns)
}

func (a *Agent) toAnthropicMessages() []api.AnthropicMessage {
	if len(a.History) == 0 {
		return nil
	}
	msgs := make([]api.AnthropicMessage, len(a.History))
	for i := range a.History {
		m := &a.History[i]
		msgs[i] = api.AnthropicMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}
	return msgs
}

// HandleToolUse is a helper to execute tool and append result to history
func (a *Agent) HandleToolUse(ctx context.Context, toolID, toolName string, input map[string]interface{}) (string, error) {
	t, ok := a.Tools[toolName]
	if !ok {
		return "", fmt.Errorf("tool %s not found", toolName)
	}

	result, err := t.Execute(ctx, input)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil // Return error as content so AI can fix it
	}

	// Format result (simplified)
	return fmt.Sprintf("%v", result), nil
}

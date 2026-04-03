package api

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
)

// StreamEvents reads SSE from OpenAI and converts to Anthropic stream events on the fly
func StreamEvents(resp *http.Response) (<-chan AnthropicStreamEvent, <-chan error) {
	events := make(chan AnthropicStreamEvent)
	errCh := make(chan error, 1)

	go func() {
		defer resp.Body.Close()
		defer close(events)
		defer close(errCh)

		reader := bufio.NewReader(resp.Body)
		
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					errCh <- err
				}
				break
			}

			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}

			if !bytes.HasPrefix(line, []byte("data: ")) {
				continue
			}

			data := bytes.TrimPrefix(line, []byte("data: "))
			if bytes.Equal(data, []byte("[DONE]")) {
				break
			}

			var chunk OpenAIStreamChunk
			if err := json.Unmarshal(data, &chunk); err != nil {
				continue
			}

			for _, choice := range chunk.Choices {
				// Handle Text deltas
				if choice.Delta.Content != nil {
					events <- AnthropicStreamEvent{
						Type:  "content_block_delta",
						Index: choice.Index,
						Delta: map[string]interface{}{
							"type": "text_delta",
							"text": *choice.Delta.Content,
						},
					}
				}

				// Handle Tool Call deltas
				for _, tc := range choice.Delta.ToolCalls {
					events <- AnthropicStreamEvent{
						Type:  "content_block_delta",
						Index: choice.Index,
						Delta: map[string]interface{}{
							"type":         "input_json_delta",
							"partial_json": tc.Function.Arguments,
						},
						// We also need to signal tool use start if ID or Name is present
						ContentBlk: map[string]interface{}{
							"type": "tool_use",
							"id":   tc.ID,
							"name": tc.Function.Name,
						},
					}
				}
			}

			// Handle Usage
			if chunk.Usage != nil {
				events <- AnthropicStreamEvent{
					Type: "message_delta",
					Usage: map[string]int{
						"input_tokens":  chunk.Usage.PromptTokens,
						"output_tokens": chunk.Usage.CompletionTokens,
					},
				}
			}
		}
	}()

	return events, errCh
}

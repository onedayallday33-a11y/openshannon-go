package api

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
)

// var json is already declared in client.go

// StreamEvents reads SSE from OpenAI and converts to Anthropic stream events on the fly
func StreamEvents(resp *http.Response) (<-chan AnthropicStreamEvent, <-chan error) {
	events := make(chan AnthropicStreamEvent, 64) // Buffered channel for better throughput
	errCh := make(chan error, 1)

	go func() {
		defer resp.Body.Close()
		defer close(events)
		defer close(errCh)

		reader := bufio.NewReader(resp.Body)
		dataPrefix := []byte("data: ")
		doneMsg := []byte("[DONE]")
		
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

			if !bytes.HasPrefix(line, dataPrefix) {
				continue
			}

			data := bytes.TrimPrefix(line, dataPrefix)
			if bytes.Equal(data, doneMsg) {
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
					event := AnthropicStreamEvent{
						Type:  "content_block_delta",
						Index: tc.Index, // Use the tool call index, not choice index
						Delta: map[string]interface{}{
							"type":         "input_json_delta",
							"partial_json": tc.Function.Arguments,
						},
					}

					// Only include ContentBlk (id/name) if they are present in this chunk
					if tc.ID != "" || tc.Function.Name != "" {
						event.ContentBlk = map[string]interface{}{
							"type": "tool_use",
							"id":   tc.ID,
							"name": tc.Function.Name,
						}
					}
					events <- event
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

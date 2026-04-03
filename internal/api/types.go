package api

// Anthropic types

type AnthropicMessageRequest struct {
	Model       string                 `json:"model"`
	Messages    []AnthropicMessage     `json:"messages"`
	System      interface{}            `json:"system,omitempty"`
	Tools       []AnthropicTool        `json:"tools,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float32                `json:"temperature,omitempty"`
	TopP        float32                `json:"top_p,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
}

type AnthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` 
}

type AnthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema,omitempty"`
}

type AnthropicUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

type AnthropicStreamEvent struct {
	Type       string                 `json:"type"`
	Message    map[string]interface{} `json:"message,omitempty"`
	Index      int                    `json:"index,omitempty"`
	ContentBlk map[string]interface{} `json:"content_block,omitempty"`
	Delta      map[string]interface{} `json:"delta,omitempty"`
	Usage      map[string]int         `json:"usage,omitempty"`
}

// OpenAI types

type OpenAIChatRequest struct {
	Model                string          `json:"model"`
	Messages             []OpenAIMessage `json:"messages"`
	Stream               bool            `json:"stream"`
	StreamOptions        *StreamOption   `json:"stream_options,omitempty"`
	MaxCompletionTokens  int             `json:"max_completion_tokens,omitempty"`
	MaxTokens            int             `json:"max_tokens,omitempty"`
	Temperature          float32         `json:"temperature,omitempty"`
	TopP                 float32         `json:"top_p,omitempty"`
	Tools                []OpenAITool    `json:"tools,omitempty"`
}

type StreamOption struct {
	IncludeUsage bool `json:"include_usage"`
}

type OpenAIMessage struct {
	Role         string           `json:"role"`
	Content      interface{}      `json:"content,omitempty"`
	ToolCalls    []OpenAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID   string           `json:"tool_call_id,omitempty"`
	Name         string           `json:"name,omitempty"`
}

type OpenAIToolCall struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Function     OpenAIToolFunction     `json:"function"`
	ExtraContent map[string]interface{} `json:"extra_content,omitempty"`
}

type OpenAIToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type OpenAITool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Parameters  map[string]interface{} `json:"parameters"`
		Strict      bool                   `json:"strict,omitempty"`
	} `json:"function"`
}

type OpenAIStreamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role      string `json:"role,omitempty"`
			Content   *string `json:"content,omitempty"`
			ToolCalls []struct {
				Index    int    `json:"index"`
				ID       string `json:"id,omitempty"`
				Type     string `json:"type,omitempty"`
				Function struct {
					Name      string `json:"name,omitempty"`
					Arguments string `json:"arguments,omitempty"`
				} `json:"function,omitempty"`
				ExtraContent map[string]interface{} `json:"extra_content,omitempty"`
			} `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

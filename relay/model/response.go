package model

// Response API - https://platform.openai.com/docs/api-reference/responses

type ResponseRequest struct {
	Input              any                   `json:"input,omitempty"` // string, array, or object
	Model              string                `json:"model"`
	Instructions       string                `json:"instructions,omitempty"`
	PreviousResponseID string                `json:"previous_response_id,omitempty"`
	Parameters         *ResponseParameters   `json:"parameters,omitempty"`
	Tools              []ResponseTool        `json:"tools,omitempty"`
	ToolChoice         any                   `json:"tool_choice,omitempty"` // "auto", "none", "required", or object
	Include            []string              `json:"include,omitempty"`
	Metadata           map[string]string     `json:"metadata,omitempty"`
	Temperature        *float64              `json:"temperature,omitempty"`
	MaxOutputTokens    *int                  `json:"max_output_tokens,omitempty"`
	TopP               *float64              `json:"top_p,omitempty"`
	ParallelToolCalls  *bool                 `json:"parallel_tool_calls,omitempty"`
	Store              *bool                 `json:"store,omitempty"`
	Stream             bool                  `json:"stream,omitempty"`
	StreamOptions      *StreamOptions        `json:"stream_options,omitempty"`
	Truncation         string                `json:"truncation,omitempty"` // "auto" or "disabled"
	User               string                `json:"user,omitempty"`
}

type ResponseParameters struct {
	IncludeOutput     []string `json:"include_output,omitempty"`
	MaxToolRoundtrips int      `json:"max_tool_roundtrips,omitempty"`
}

type ResponseTool struct {
	Type     string `json:"type"` // "function"
	Function *Tool  `json:"function,omitempty"`
}

type ResponseOutput struct {
	Type      string          `json:"type"` // "message", "function_call", etc.
	ID        string          `json:"id,omitempty"`
	Status    string          `json:"status,omitempty"` // "completed", "in_progress"
	Role      string          `json:"role,omitempty"`   // "assistant", "user"
	Content   []ResponseContent `json:"content,omitempty"`
}

type ResponseContent struct {
	Type        string `json:"type"` // "output_text", "input_text", etc.
	Text        string `json:"text,omitempty"`
	Annotations []any  `json:"annotations,omitempty"`
}

type ResponseObject struct {
	ID                  string             `json:"id"`
	Object              string             `json:"object"` // "response"
	CreatedAt           int64              `json:"created_at"`
	Status              string             `json:"status"` // "completed", "in_progress", "failed", etc.
	Error               *Error             `json:"error,omitempty"`
	IncompleteDetails   any                `json:"incomplete_details,omitempty"`
	Instructions        string             `json:"instructions,omitempty"`
	MaxOutputTokens     *int               `json:"max_output_tokens,omitempty"`
	Model               string             `json:"model"`
	Output              []ResponseOutput   `json:"output"`
	ParallelToolCalls   *bool              `json:"parallel_tool_calls,omitempty"`
	PreviousResponseID  string             `json:"previous_response_id,omitempty"`
	Reasoning           *ResponseReasoning `json:"reasoning,omitempty"`
	Store               *bool              `json:"store,omitempty"`
	Temperature         *float64           `json:"temperature,omitempty"`
	Text                *ResponseText      `json:"text,omitempty"`
	ToolChoice          any                `json:"tool_choice,omitempty"`
	Tools               []ResponseTool     `json:"tools,omitempty"`
	TopP                *float64           `json:"top_p,omitempty"`
	Truncation          string             `json:"truncation,omitempty"`
	Usage               *ResponseUsage     `json:"usage,omitempty"`
	User                string             `json:"user,omitempty"`
	Metadata            map[string]string  `json:"metadata,omitempty"`
}

type ResponseReasoning struct {
	Effort  string `json:"effort,omitempty"`
	Summary string `json:"summary,omitempty"`
}

type ResponseText struct {
	Format *ResponseFormat `json:"format,omitempty"`
}

type ResponseUsage struct {
	InputTokens         int                       `json:"input_tokens"`
	InputTokensDetails  *ResponseInputDetails     `json:"input_tokens_details,omitempty"`
	OutputTokens        int                       `json:"output_tokens"`
	OutputTokensDetails *ResponseOutputDetails    `json:"output_tokens_details,omitempty"`
	TotalTokens         int                       `json:"total_tokens"`
}

type ResponseInputDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

type ResponseOutputDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

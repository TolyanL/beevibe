package dto

// LLMRequest — запрос чат-комплишена.
type LLMRequest struct {
	System      string  `json:"system"`
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
}

// LLMUsage — расход токенов по ответу модели.
type LLMUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

// LLMResponse — ответ чат-комплишена.
type LLMResponse struct {
	Text  string   `json:"text"`
	Usage LLMUsage `json:"usage"`
}

// TotalTokens — суммарное число токенов запроса и ответа.
func (r LLMResponse) TotalTokens() int {
	return r.Usage.PromptTokens + r.Usage.CompletionTokens
}

package llm

import "context"

// LLM — порт чат-комплишена: всё, что связано с обращением к модели.
type LLM interface {
	Complete(ctx context.Context, req Request) (Response, error)
	Prompt(name string) (string, error)
	PromptVersion() string
}

// Request — запрос чат-комплишена.
type Request struct {
	System      string  `json:"system"`
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
}

// Usage — расход токенов по ответу модели.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

// Response — ответ чат-комплишена.
type Response struct {
	Text  string `json:"text"`
	Usage Usage  `json:"usage"`
}

// TotalTokens — суммарное число токенов запроса и ответа.
func (r Response) TotalTokens() int {
	return r.Usage.PromptTokens + r.Usage.CompletionTokens
}

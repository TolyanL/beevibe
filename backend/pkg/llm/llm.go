package llm

import "context"

// LLM — порт чат-комплишена: всё, что связано с обращением к модели.
type LLM interface {
	Complete(ctx context.Context, req Request) (Response, error)
	Prompt(name string) (string, error)
	PromptVersion() string
}

// Effort — уровень рассуждения модели; "" и "default" означают «не отправлять поле».
type Effort string

const (
	EffortDefault Effort = ""
	EffortNone    Effort = "none"
	EffortMinimal Effort = "minimal"
	EffortLow     Effort = "low"
	EffortMedium  Effort = "medium"
	EffortHigh    Effort = "high"
	EffortXHigh   Effort = "xhigh"
	EffortMax     Effort = "max"

	effortDefaultLiteral Effort = "default"
)

// Valid — значение из allow-list провайдеров.
func (e Effort) Valid() bool {
	switch e {
	case EffortDefault, effortDefaultLiteral, EffortNone, EffortMinimal,
		EffortLow, EffortMedium, EffortHigh, EffortXHigh, EffortMax:
		return true
	default:
		return false
	}
}

func (e Effort) value() string {
	if e == effortDefaultLiteral {
		return ""
	}

	return string(e)
}

// Request — запрос чат-комплишена.
// Длина ответа не ограничивается: её задаёт уровень рассуждения модели,
// а не лимит токенов в запросе.
type Request struct {
	System      string  `json:"system"`
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature"`
	Reasoning   Effort  `json:"reasoning"`
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

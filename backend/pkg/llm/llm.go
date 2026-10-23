package llm

import "context"

type LLM interface {
	Complete(ctx context.Context, req Request) (Response, error)
	Prompt(name string) (string, error)
	PromptVersion() string
}

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

type Request struct {
	System      string  `json:"system"`
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature"`
	Reasoning   Effort  `json:"reasoning"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type Response struct {
	Text  string `json:"text"`
	Usage Usage  `json:"usage"`
}

func (r Response) TotalTokens() int {
	return r.Usage.PromptTokens + r.Usage.CompletionTokens
}

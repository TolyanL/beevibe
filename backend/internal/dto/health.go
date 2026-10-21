package dto

// HealthOutput — ответ служебной ручки.
type HealthOutput struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

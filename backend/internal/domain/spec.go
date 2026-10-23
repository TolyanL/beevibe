package domain

type GameSpec struct {
	Genre         string   `json:"genre"`
	Mechanic      string   `json:"mechanic"`
	Controls      string   `json:"controls"`
	WinCondition  string   `json:"win_condition"`
	LoseCondition string   `json:"lose_condition"`
	Palette       []string `json:"palette"`
	MapSize       MapSize  `json:"map_size"`
	Style         Style    `json:"style"`
	Audience      string   `json:"audience"`
	Assets        []string `json:"assets"`
}

package model

type Message struct {
	Type   string `json:"Type"`
	Data   string `json:"Data"`
	Target int    `json:"Target"`
}

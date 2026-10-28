package net

type PlayerState struct {
	Name string  `json:"name"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	HP   int     `json:"hp"`
}

type Message struct {
	Type    string        `json:"type"` // "state"
	Player  *PlayerState  `json:"player,omitempty"`
	Players []PlayerState `json:"players,omitempty"`
}

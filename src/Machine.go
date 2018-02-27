package main

// Machine represents one of the alti-worker.
type Machine struct {
	Name     string   `json:"name"`
	Nickname string   `json:"nickname"`
	Type     string   `json:"type"`
	Ping     JSONTime `json:"ping"`
	Pong     JSONTime `json:"pong"`
	Extra    string   `json:"extra"`
}

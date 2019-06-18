package main

// Cmd represents rm or mv cmd.
type Cmd struct {
	Ops  string   `json:"ops"`
	Args []string `json:"args"`
	ID   string   `json:"id"`   // id of this cmd / job
	Done string   `json:"done"` // done queue to publish
}

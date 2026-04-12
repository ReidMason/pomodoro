package models

type RequestType int

const (
	SetTask RequestType = iota
	Start
	TogglePaused
)

type Request struct {
	Type RequestType `json:"type"`
}

type SetTaskCommand struct {
	Type RequestType `json:"type"`
	Task string      `json:"task"`
}

package models

type CommandType int

const (
	SetTask CommandType = iota
	Start
	TogglePaused
)

type Command struct {
	Type CommandType `json:"type"`
}

type SetTaskCommand struct {
	Type CommandType `json:"type"`
	Task string      `json:"task"`
}

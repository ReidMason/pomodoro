package models

type CommandType int

const (
	UpdateTask CommandType = iota
	Start
)

type Command struct {
	Type CommandType `json:"type"`
}

type UpdateTaskCommand struct {
	Task string `json:"task"`
}

package models

type RequestKind int

const (
	SetTask RequestKind = iota
	Start
)

type Request struct {
	Kind RequestKind `json:"type"`
}

type SetTaskRequest struct {
	Kind RequestKind `json:"type"`
	Task string      `json:"task"`
}

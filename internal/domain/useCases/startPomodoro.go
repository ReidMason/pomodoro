package usecases

import (
	"github.com/ReidMason/pomodoro/internal/domain/models/pomodoro"
)

type StartPomodoro struct {
	pomodoro *pomodoro.Pomodoro
}

func NewStartPomodoro(pomodoro *pomodoro.Pomodoro) *StartPomodoro {
	return &StartPomodoro{
		pomodoro: pomodoro,
	}
}

func (sp *StartPomodoro) Handle() {
	sp.pomodoro.Start()
}

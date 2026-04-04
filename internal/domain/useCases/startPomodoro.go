package usecases

import (
	"github.com/ReidMason/pomodoro/internal/domain/models"
)

type StartPomodoro struct {
	pomodoro *models.Pomodoro
}

func NewStartPomodoro(pomodoro *models.Pomodoro) *StartPomodoro {
	return &StartPomodoro{
		pomodoro: pomodoro,
	}
}

func (sp *StartPomodoro) Handle() {
	sp.pomodoro.Start()
}

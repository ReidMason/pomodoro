package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/ReidMason/pomodoro/internal/domain/models"
	usecases "github.com/ReidMason/pomodoro/internal/domain/useCases"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	eventHandler := EventHandler{WG: &wg}

	interval := 5 * time.Second
	pomodoro := models.NewPomodoro(interval)
	pomodoro.AddSubscriber(eventHandler.HandlePomodoroEvent)

	startPomodoro := usecases.NewStartPomodoro(pomodoro)
	startPomodoro.Handle()

	wg.Wait()
}

type EventHandler struct {
	WG *sync.WaitGroup
}

func (eh *EventHandler) HandlePomodoroEvent(event string, pomodoro *models.Pomodoro) {
	fmt.Println(pomodoro.GetTimeRemaining())

	if event == "Done" {
		eh.WG.Done()
	}
}

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

	pomodoroDuration := 1 * time.Second   // Should be 25 minutes
	shortBreakDuration := 1 * time.Second // Shouldbe 5 minutes
	longBreakDuration := 1 * time.Second  // Shouldbe 20 minutes
	pomodoro := models.NewPomodoro(pomodoroDuration, shortBreakDuration, longBreakDuration)
	pomodoro.AddSubscriber(eventHandler.HandlePomodoroEvent)

	startPomodoro := usecases.NewStartPomodoro(*pomodoro)
	startPomodoro.Handle()

	wg.Wait()
}

type EventHandler struct {
	WG *sync.WaitGroup
}

func (eh *EventHandler) HandlePomodoroEvent(event string, pomodoro *models.Pomodoro) {
	fmt.Println(event, pomodoro.GetTimeRemaining())

	if event == "LongBreak.Done" {
		eh.WG.Done()
	}
}

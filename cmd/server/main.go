package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ReidMason/pomodoro/internal/domain/models"
	usecases "github.com/ReidMason/pomodoro/internal/domain/useCases"
)

func main() {
	hub := newHub()
	go hub.run()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Testing"))
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Serving WS")
		serveWs(hub, w, r)
	})

	startPom(hub)

	log.Println("Starting server")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func startPom(hub *Hub) {
	eventHandler := EventHandler{hub: hub}
	task := "testing"
	pomodoroDuration := 1 * time.Second   // Should be 25 minutes
	shortBreakDuration := 1 * time.Second // Shouldbe 5 minutes
	longBreakDuration := 1 * time.Second  // Shouldbe 20 minutes
	pomodoro := models.NewPomodoro(pomodoroDuration, shortBreakDuration, longBreakDuration, task)
	pomodoro.AddSubscriber(eventHandler.HandlePomodoroEvent)

	hub.Pomodoro = pomodoro

	startPomodoro := usecases.NewStartPomodoro(*pomodoro)
	startPomodoro.Handle()
}

type EventHandler struct {
	hub *Hub
}

func (eh *EventHandler) HandlePomodoroEvent(event models.PomodoroEvent, pomodoro models.Pomodoro) {
	fmt.Println(event, pomodoro.GetTimeRemaining())

	body, err := json.Marshal(pomodoro)
	if err != nil {
		log.Println("failed to marshal response")
		return
	}

	eh.hub.broadcast <- body
}

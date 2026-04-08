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

type pomodoroEventMessage struct {
	Event    models.PomodoroEvent
	Pomodoro models.Pomodoro
}

func startPom(hub *Hub) {
	task := "testing"
	pomodoroDuration := 5 * time.Second   // Should be 25 minutes
	shortBreakDuration := 3 * time.Second // Shouldbe 5 minutes
	longBreakDuration := 2 * time.Second  // Shouldbe 20 minutes
	pomodoro := models.NewPomodoro(pomodoroDuration, shortBreakDuration, longBreakDuration, task)

	ch := make(chan pomodoroEventMessage, 32)
	go func() {
		for msg := range ch {
			fmt.Println(msg.Event, msg.Pomodoro.GetTimeRemaining())

			body, err := json.Marshal(msg.Pomodoro)
			if err != nil {
				log.Println("failed to marshal response")
				return
			}

			hub.broadcast <- body
		}
	}()

	pomodoro.AddSubscriber(func(e models.PomodoroEvent, p models.Pomodoro) {
		ch <- pomodoroEventMessage{Event: e, Pomodoro: p}
	})

	hub.Pomodoro = pomodoro

	startPomodoro := usecases.NewStartPomodoro(*pomodoro)
	startPomodoro.Handle()
}

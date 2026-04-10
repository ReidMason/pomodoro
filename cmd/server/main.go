package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

type pomodoroEventMessage struct {
	Event    models.PomodoroEvent
	Pomodoro models.Pomodoro
}

func convertSecondsDuration(t int) time.Duration {
	return time.Second * time.Duration(t)
}

func loadEnvVarInt(envVar string, defaultValue int) int {
	value := os.Getenv(envVar)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Fatal("Missing environment variable", envVar)
	}

	return intValue
}

func startPom(hub *Hub) {
	pomodoroDuration := loadEnvVarInt("POMODORO_DURATION", 1200)
	shortBreakDuration := loadEnvVarInt("SHORT_BREAK_DURATION", 300)
	longBreakDuration := loadEnvVarInt("LONG_BREAK_DURATION", 900)

	loop := false
	loopVar := os.Getenv("LOOP")
	if loopVar == "true" {
		loop = true
	}

	task := "testing"
	pomodoro := models.NewPomodoro(convertSecondsDuration(pomodoroDuration), convertSecondsDuration(shortBreakDuration), convertSecondsDuration(longBreakDuration), task, loop)

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

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ReidMason/pomodoro/internal/domain/models/pomodoro"
)

func sendPomodoroUpdate(hub *Hub) {
	body, err := json.Marshal(hub.Pomodoro.ToDto())
	if err != nil {
		log.Println("failed to marshal response")
		return
	}

	hub.broadcast <- body
}

func main() {
	hub := newHub()
	pomodoro := createPomodoro(hub)
	hub.Pomodoro = pomodoro

	go hub.run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Serving WS")
		serveWs(hub, w, r)

		sendPomodoroUpdate(hub)
	})

	log.Println("Starting server")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Listening on: ", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

type pomodoroEventMessage struct {
	Event    pomodoro.PomodoroEvent
	Pomodoro pomodoro.Pomodoro
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

func createPomodoro(hub *Hub) *pomodoro.Pomodoro {
	pomodoroDuration := loadEnvVarInt("POMODORO_DURATION", 1200)
	shortBreakDuration := loadEnvVarInt("SHORT_BREAK_DURATION", 300)
	longBreakDuration := loadEnvVarInt("LONG_BREAK_DURATION", 900)

	loop := false
	loopVar := os.Getenv("LOOP")
	if loopVar == "true" {
		loop = true
	}

	p := pomodoro.New(convertSecondsDuration(pomodoroDuration), convertSecondsDuration(shortBreakDuration), convertSecondsDuration(longBreakDuration), loop)

	ch := make(chan pomodoroEventMessage, 32)
	go func() {
		for msg := range ch {
			fmt.Println(msg.Event, msg.Pomodoro.ToDto().TimeRemaining)
			sendPomodoroUpdate(hub)
		}
	}()

	p.AddSubscriber(func(event pomodoro.PomodoroEvent, pomodoro pomodoro.Pomodoro) {
		ch <- pomodoroEventMessage{Event: event, Pomodoro: pomodoro}
	})

	return p
}

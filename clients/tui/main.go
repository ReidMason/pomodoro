package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/ReidMason/pomodoro/internal/domain/models"
	"github.com/gorilla/websocket"

	tea "charm.land/bubbletea/v2"
)

type model struct {
	text     string
	pomodoro models.Pomodoro
}

func startWsClient(program *tea.Program) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	addr := "localhost:8080"
	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			var pom models.Pomodoro
			err = json.Unmarshal(message, &pom)
			if err != nil {
				log.Println("Failed to unmarshal pomodoro data", err)
			}
			program.Send(newPomodoroData(pom))
		}
	}()

	//		ticker := time.NewTicker(time.Second)
	//		defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		// case t := <-ticker.C:
		// 	err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
		// 	if err != nil {
		// 		log.Println("Write:", err)
		// 	}
		case <-interrupt:
			log.Println("interrupt")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}

			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func main() {
	pom := models.Pomodoro{}
	m := initModel(pom)

	p := tea.NewProgram(m)
	go startWsClient(p)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error starting program: %v", err)
	}
}

func initModel(pomodoro models.Pomodoro) model {
	return model{
		text:     "Connecting...",
		pomodoro: pomodoro,
	}
}

func (m model) Init() tea.Cmd {
	return scheduleTick()
}

type tickMsg time.Time

func scheduleTick() tea.Cmd {
	return tea.Every(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type newPomodoroData models.Pomodoro

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.text = time.Time(msg).Format("01/02 03:04:05PM")
		return m, scheduleTick()
	case newPomodoroData:
		m.pomodoro = models.Pomodoro(msg)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	s := m.text
	s += "\n" + m.pomodoro.CycleStage.String()
	s += formatTime(m.pomodoro.TimeRemaining)

	return tea.NewView(s)
}

func formatTime(duration time.Duration) string {
	d := max(duration, 0)
	m := int(d / time.Minute)
	s := int((d % time.Minute) / time.Second)
	return fmt.Sprintf("\n%d:%02d", m, s)
}

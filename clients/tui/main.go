package main

import (
	"encoding/json"
	"flag"
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
	time     time.Time
	pomodoro models.Pomodoro
	status   string
	host     string
}

func dial(u url.URL) (*websocket.Conn, error) {
	var err error
	for range 30 {
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		return c, nil
	}

	return nil, err
}

type connectionStatusUpdate string

func startWsClient(program *tea.Program) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	addr := "localhost:8080"
	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}

	c, err := dial(u)
	if err != nil {
		log.Println("Failed to connect", err)
		return
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			switch err.(type) {
			case *websocket.CloseError:
				program.Send(connectionStatusUpdate("Connection lost, reconnecting..."))
				c, err = dial(u)
				if err != nil {
					program.Send(connectionStatusUpdate("Connection lost"))
					return
				}
				continue
			default:
				if err != nil {
					program.Send(connectionStatusUpdate("Connection unstable, unable to read messages"))
					return
				}
			}

			var pom models.Pomodoro
			err = json.Unmarshal(message, &pom)
			if err != nil {
				program.Send(connectionStatusUpdate("Connection unstable, receiving bad data"))
				continue
			}
			program.Send(newPomodoroData(pom))
			program.Send(connectionStatusUpdate("Connected"))
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
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
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
	host := flag.String("s", "http://locahost:8080", "Server host address")
	flag.Parse()
	if host == nil {
		log.Fatal("Provide a server url")
	}

	pom := models.Pomodoro{}
	m := initModel(pom, *host)

	p := tea.NewProgram(m)
	go startWsClient(p)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error starting program: %v", err)
	}
}

func initModel(pomodoro models.Pomodoro, host string) model {
	return model{
		time:     getTime(),
		pomodoro: pomodoro,
		host:     host,
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
		m.time = getTime()
		return m, scheduleTick()
	case newPomodoroData:
		m.pomodoro = models.Pomodoro(msg)
	case connectionStatusUpdate:
		m.status = string(msg)
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func getTime() time.Time {
	return time.Now()
}

func formatTimestamp(t time.Time) string {
	return time.Time(t).Format("01/02 03:04:05PM")
}

func (m model) View() tea.View {
	s := formatTimestamp(m.time) + "\n"
	s += "Task: " + m.pomodoro.Task
	s += "\n" + m.pomodoro.CycleStage.String()
	s += formatTimeDuration(m.pomodoro.TimeRemaining)
	s += "\n\n" + m.status

	return tea.NewView(s)
}

func formatTimeDuration(duration time.Duration) string {
	d := max(duration, 0)
	m := int(d / time.Minute)
	s := int((d % time.Minute) / time.Second)
	return fmt.Sprintf("\n%d:%02d", m, s)
}

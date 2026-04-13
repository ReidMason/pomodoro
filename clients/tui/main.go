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
	"github.com/ReidMason/pomodoro/internal/domain/models/pomodoro"
	"github.com/gorilla/websocket"

	tea "charm.land/bubbletea/v2"
)

type model struct {
	time      time.Time
	pomodoro  pomodoro.PomodoroDto
	status    string
	websocket *websocket.Conn
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
type websocketClientConnectedEvent *websocket.Conn

func startWsClient(program *tea.Program, host string) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}

	c, err := dial(u)
	if err != nil {
		log.Println("Failed to connect", err)
		return
	}
	defer c.Close()
	program.Send(websocketClientConnectedEvent(c))

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
				program.Send(websocketClientConnectedEvent(c))
				continue
			default:
				if err != nil {
					program.Send(connectionStatusUpdate("Connection unstable, unable to read messages"))
					return
				}
			}

			var pom pomodoro.PomodoroDto
			err = json.Unmarshal(message, &pom)
			if err != nil {
				program.Send(connectionStatusUpdate("Connection unstable, receiving bad data"))
				continue
			}
			program.Send(newPomodoroData(pom))
			program.Send(connectionStatusUpdate("Connected"))
		}
	}()

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

	pom := pomodoro.PomodoroDto{}
	m := initModel(pom)
	p := tea.NewProgram(m)

	go startWsClient(p, *host)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error starting program: %v", err)
	}
}

func initModel(pomodoro pomodoro.PomodoroDto) model {
	return model{
		time:     getTime(),
		pomodoro: pomodoro,
		status:   "Connecting...",
	}
}

func (m model) Init() tea.Cmd {
	return scheduleTick()
}

type tickMsg time.Time

func scheduleTick() tea.Cmd {
	tickSpeed := time.Second

	return tea.Every(tickSpeed, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type newPomodoroData pomodoro.PomodoroDto

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.time = getTime()
		return m, scheduleTick()
	case newPomodoroData:
		m.pomodoro = pomodoro.PomodoroDto(msg)
	case websocketClientConnectedEvent:
		m.websocket = msg
	case connectionStatusUpdate:
		m.status = string(msg)
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "t":
			setTaskCommand := models.SetTaskRequest{
				Kind: models.SetTask,
				Task: formatTimestamp(time.Now()),
			}
			payload, err := json.Marshal(setTaskCommand)
			if err != nil {
				return m, nil
			}
			m.websocket.WriteMessage(websocket.TextMessage, payload)
			return m, nil
		case "s":
			payload, err := json.Marshal(models.Request{
				Kind: models.Start,
			})
			if err != nil {
				return m, nil
			}
			m.websocket.WriteMessage(websocket.TextMessage, payload)
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

	remaining := time.Until(m.pomodoro.PhaseEndsAt)
	s += formatTimeDuration(remaining)
	s += "\n\n" + m.status

	return tea.NewView(s)
}

func formatTimeDuration(duration time.Duration) string {
	d := max(duration, 0)
	m := int(d / time.Minute)
	s := int((d % time.Minute) / time.Second)
	ms := int((d % time.Second) / time.Millisecond)
	return fmt.Sprintf("\n%d:%02d.%03d", m, s, ms)
}

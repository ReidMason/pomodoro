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

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type connectionStatus int

const (
	connecting connectionStatus = iota
	connectionLostReconnecting
	connectionLost
	connected
	reconnecting
)

type model struct {
	time             time.Time
	pomodoro         pomodoro.PomodoroDto
	connectionStatus connectionStatus
	websocket        *websocket.Conn

	spinner     spinner.Model
	settingTask bool
	textInput   textinput.Model
}

func dial(u url.URL) (*websocket.Conn, error) {
	for {
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		return c, nil
	}
}

type connectionStatusUpdate connectionStatus
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
				program.Send(connectionStatusUpdate(connectionLostReconnecting))
				c, err = dial(u)
				if err != nil {
					program.Send(connectionStatusUpdate(connectionLost))
					return
				}
				program.Send(websocketClientConnectedEvent(c))
				continue
			default:
				if err != nil {
					// program.Send(connectionStatusUpdate("Connection unstable, unable to read messages"))
					return
				}
			}

			var pom pomodoro.PomodoroDto
			err = json.Unmarshal(message, &pom)
			if err != nil {
				// program.Send(connectionStatusUpdate("Connection unstable, receiving bad data"))
				continue
			}
			program.Send(newPomodoroData(pom))
			program.Send(connectionStatusUpdate(connected))
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
	ti := textinput.New()
	ti.Placeholder = "Task name"
	ti.SetVirtualCursor(true)

	return model{
		time:             getTime(),
		pomodoro:         pomodoro,
		connectionStatus: connecting,
		spinner:          spinner.New(),
		settingTask:      false,
		textInput:        ti,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(scheduleTick(), m.spinner.Tick, textinput.Blink)
}

type tickMsg time.Time

func scheduleTick() tea.Cmd {
	tickSpeed := time.Second

	return tea.Every(tickSpeed, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type newPomodoroData pomodoro.PomodoroDto
type startSettingTask struct{}
type stopSettingTask struct{}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.time = getTime()
		return m, scheduleTick()
	case startSettingTask:
		if m.settingTask {
			return m, nil
		}

		m.settingTask = true
		m.textInput.Focus()
		return m, nil
	case stopSettingTask:
		m.settingTask = false
		m.textInput.SetValue("")
		return m, nil
	case newPomodoroData:
		m.pomodoro = pomodoro.PomodoroDto(msg)
	case websocketClientConnectedEvent:
		m.websocket = msg
	case connectionStatusUpdate:
		m.connectionStatus = connectionStatus(msg)
		return m, nil
	case tea.KeyPressMsg:
		return handleKeypress(m, msg)
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

var submitBinding = key.NewBinding(key.WithKeys("enter"))
var escBinding = key.NewBinding(key.WithKeys("esc", "escape"))

func handleKeypress(m model, msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	}

	if m.textInput.Focused() && key.Matches(msg, escBinding) {
		return m, func() tea.Msg { return stopSettingTask{} }
	}

	if m.textInput.Focused() && key.Matches(msg, submitBinding) {
		submitted := m.textInput.Value()

		setTaskCommand := models.SetTaskRequest{
			Kind: models.SetTask,
			Task: submitted,
		}
		payload, err := json.Marshal(setTaskCommand)
		if err != nil {
			return m, nil
		}
		m.websocket.WriteMessage(websocket.TextMessage, payload)
		return m, func() tea.Msg { return stopSettingTask{} }
	}

	if m.settingTask && m.textInput.Focused() {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "t":
		return m, func() tea.Msg { return startSettingTask{} }
	case "s":
		payload, err := json.Marshal(models.Request{
			Kind: models.Start,
		})
		if err != nil {
			return m, nil
		}
		m.websocket.WriteMessage(websocket.TextMessage, payload)
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
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder())

	s := formatTimestamp(m.time) + "\n"
	s += "Task: " + m.pomodoro.Task
	s += fmt.Sprintf("\nPomodori: %d/4", m.pomodoro.PomodoriCompleted%4+1)
	s += "\n" + m.pomodoro.CycleStage.String()

	remaining := time.Until(m.pomodoro.PhaseEndsAt)
	s += formatTimeDuration(remaining)

	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	statusString := "unknown"
	switch m.connectionStatus {
	case connecting:
		statusString = m.spinner.View() + " Connecting..."
	case connected:
		statusStyle = statusStyle.Foreground(lipgloss.Green)
		statusString = "Connected"
	}

	if m.settingTask {
		statusString = m.textInput.View()
	}

	s += fmt.Sprintf("\n\n%s", statusStyle.Render(statusString))

	// s += "\n\n" + helpStyle(m.connectionStatus)

	return tea.NewView(style.Render(s))
}

func formatTimeDuration(duration time.Duration) string {
	d := max(duration, 0)
	m := int(d / time.Minute)
	s := int((d % time.Minute) / time.Second)
	ms := int((d % time.Second) / time.Millisecond)
	return fmt.Sprintf("\n%d:%02d.%03d", m, s, ms)
}

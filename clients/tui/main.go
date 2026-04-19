package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gen2brain/beeep"
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
)

type model struct {
	time             time.Time
	pomodoro         pomodoro.PomodoroDto
	connectionStatus connectionStatus
	websocket        *websocket.Conn

	width, height int

	spinner       spinner.Model
	settingTask   bool
	textInput     textinput.Model
	tryingToStart bool
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

func alert(message string) {
	beeep.Alert("Pomodoro", message, "")
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
		tryingToStart:    false,
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
type startPomodoriTask struct{}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textInput.SetWidth(msg.Width - 5)
		return m, nil
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
	case startPomodoriTask:
		payload, err := json.Marshal(models.Request{
			Kind: models.Start,
		})
		if err != nil {
			return m, nil
		}
		m.websocket.WriteMessage(websocket.TextMessage, payload)
		return m, nil
	case newPomodoroData:
		newData := pomodoro.PomodoroDto(msg)
		handleAlerts(newData, m.pomodoro)
		m.pomodoro = newData
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
var quitBinding = key.NewBinding(key.WithKeys("ctrl+c", "q"))
var setTaskBinding = key.NewBinding(key.WithKeys("t"))
var startPomodoroBinding = key.NewBinding(key.WithKeys("s"))

func handleKeypress(m model, msg tea.KeyPressMsg) (model, tea.Cmd) {
	switch {
	case key.Matches(msg, quitBinding):
		return m, tea.Quit
	case m.textInput.Focused() && key.Matches(msg, escBinding):
		return m, func() tea.Msg { return stopSettingTask{} }
	case m.textInput.Focused() && key.Matches(msg, submitBinding):
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

		var cmds []tea.Cmd
		cmds = append(cmds, func() tea.Msg { return stopSettingTask{} })

		if m.tryingToStart {
			m.tryingToStart = false
			cmds = append(cmds, func() tea.Msg { return startPomodoriTask{} })
		}

		return m, tea.Sequence(cmds...)
	case m.settingTask && m.textInput.Focused():
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	case key.Matches(msg, setTaskBinding):
		return m, func() tea.Msg { return startSettingTask{} }
	case key.Matches(msg, startPomodoroBinding):
		if m.pomodoro.Task == "" {
			m.tryingToStart = true
			return m, func() tea.Msg { return startSettingTask{} }
		}
		return m, func() tea.Msg { return startPomodoriTask{} }
	}

	return m, nil
}

func handleAlerts(newData, oldData pomodoro.PomodoroDto) {
	if oldData.CycleStage == pomodoro.Idle || newData.CycleStage == oldData.CycleStage {
		return
	}

	switch oldData.CycleStage {
	case pomodoro.PomodoroInProgress:
		alert("Pomodori finished! Take a break")
	case pomodoro.ShortBreakInProgress:
		alert("Short break finished! Time to start a new pomodori")
	case pomodoro.LongBreakInProgress:
		alert("Long break finished! Time to start a new pomodori")
	}
}

func getTime() time.Time {
	return time.Now()
}

func formatTimestamp(t time.Time) string {
	return time.Time(t).Format("01/02 03:04:05PM")
}

func valueOrDefault(text, defaultValue string) string {
	if text == "" {
		return defaultValue
	}

	return text
}

func applyMutedStyling(style lipgloss.Style) lipgloss.Style {
	return style.Foreground(lipgloss.Color("#626262"))
}

var helpStyle = applyMutedStyling(lipgloss.NewStyle())

func (m model) View() tea.View {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Width(m.width)

	generalStyle := lipgloss.NewStyle()

	s := "Task: " + valueOrDefault(m.pomodoro.Task, helpStyle.Render("No task set"))
	s += fmt.Sprintf("\nPomodori: %d/4\n", m.pomodoro.PomodoriCompleted%4+1)

	s += fmt.Sprintf("\n%s", m.pomodoro.CycleStage.String())

	timeRemaining := time.Until(m.pomodoro.PhaseEndsAt)
	timeRemainingDisplay := formatTimeDuration(timeRemaining)
	timeRemainingStyle := lipgloss.NewStyle()

	if m.pomodoro.CycleStage == pomodoro.Idle {
		timeRemainingStyle = applyMutedStyling(timeRemainingStyle)
	}

	s += timeRemainingStyle.Render(timeRemainingDisplay)

	if m.connectionStatus != connected {
		generalStyle = applyMutedStyling(generalStyle)
	}

	s = generalStyle.Render(s)

	statusStyle := lipgloss.NewStyle()
	statusString := statusStyle.Render("unknown")
	switch m.connectionStatus {
	case connecting:
		statusStyle = statusStyle.Foreground(lipgloss.Blue)
		statusString = m.spinner.View() + " Connecting..."
	case connectionLostReconnecting:
		statusStyle = statusStyle.Foreground(lipgloss.Yellow)
		statusString = m.spinner.View() + " Connection lost, reconnecting..."
	case connectionLost:
		statusStyle = statusStyle.Foreground(lipgloss.Red)
		statusString = "Connection lost"
	case connected:
		statusStyle = statusStyle.Foreground(lipgloss.Green)
		statusString = "Connected"
	}

	statusString = statusStyle.Render(statusString)

	if m.settingTask {
		s += "\n\nSet task: \n" + m.textInput.View()
	}

	bottomBar := statusRow(m.width-2, statusString, helpStyle.Render(formatTimestamp(m.time)))
	s += "\n\n"
	s += bottomBar

	return tea.NewView(style.Render(s))
}

func statusRow(totalWidth int, left, right string) string {
	leftBlock := lipgloss.NewStyle().Render(left)
	leftW := lipgloss.Width(leftBlock)
	rightW := max(0, totalWidth-leftW)
	rightBlock := lipgloss.NewStyle().
		Width(rightW).
		Align(lipgloss.Right).
		Render(right)
	return lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, rightBlock)
}

func formatTimeDuration(duration time.Duration) string {
	d := max(duration, 0)
	m := int(d / time.Minute)
	s := int((d % time.Minute) / time.Second)
	ms := int((d % time.Second) / time.Millisecond)
	return fmt.Sprintf("\n%d:%02d.%03d", m, s, ms)
}

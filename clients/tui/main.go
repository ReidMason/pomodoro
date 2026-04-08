package main

import (
	"log"
	"time"

	tea "charm.land/bubbletea/v2"
)

type model struct {
	text string
}

func main() {
	p := tea.NewProgram(initModel())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error starting program: %v", err)
	}
}

func initModel() model {
	return model{
		text: "Testing!",
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.text = time.Time(msg).Format("01/02 03:04:05PM")
		return m, scheduleTick()
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

	return tea.NewView(s)
}

package pomodoro

import (
	"time"
)

type SubscriberFunc func(event PomodoroEvent, p Pomodoro)

type PomodoroDto struct {
	CycleStage        State         `json:"cycleStage"`
	Task              string        `json:"task"`
	TimeRemaining     time.Duration `json:"timeRemaining"`
	PomodoriCompleted int           `json:"pomodorosCompleted"`
}

type Pomodoro struct {
	state              State
	task               string
	phaseEndsAt        time.Time
	subscribers        []SubscriberFunc
	pomodoriCompleted  int
	pomodoroDuration   time.Duration
	shortBreakDuration time.Duration
	longBreakDuration  time.Duration
	loop               bool
}

func New(pomodoroDuration, shortBreakDuration, longBreakDuration time.Duration, loop bool) *Pomodoro {
	return &Pomodoro{
		state:              Idle,
		task:               "",
		pomodoriCompleted:  0,
		pomodoroDuration:   pomodoroDuration,
		shortBreakDuration: shortBreakDuration,
		longBreakDuration:  longBreakDuration,
		loop:               loop,
	}
}

func (p Pomodoro) ToDto() PomodoroDto {
	return PomodoroDto{
		CycleStage:        p.state,
		Task:              p.task,
		TimeRemaining:     time.Until(p.phaseEndsAt),
		PomodoriCompleted: p.pomodoriCompleted,
	}
}

func (p *Pomodoro) AddSubscriber(subscriberFunc SubscriberFunc) {
	p.subscribers = append(p.subscribers, subscriberFunc)
}

func (p *Pomodoro) HandleCommand(command Command) {
	state, event := p.state.HandleCommand(p, command)
	p.state = state

	if event != None {
		p.notifySubscribers(event)
	}
}

func (p Pomodoro) notifySubscribers(event PomodoroEvent) {
	for _, subscriber := range p.subscribers {
		subscriber(event, p)
	}
}

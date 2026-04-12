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
	timeRemaining      time.Duration
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
		timeRemaining:      0,
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
		TimeRemaining:     p.timeRemaining,
		PomodoriCompleted: p.pomodoriCompleted,
	}
}

func (p *Pomodoro) SetTask(task string) {
	if p.state == PomodoroStage {
		return
	}

	p.task = task
	p.notifySubscribers(TaskUpdated)
}

func (p *Pomodoro) AddSubscriber(subscriberFunc SubscriberFunc) {
	p.subscribers = append(p.subscribers, subscriberFunc)
}

func (p *Pomodoro) HandleCommand(command Command) {
	p.state = p.state.HandleCommand(p, command)
}

func (p *Pomodoro) Start() {
	if p.task == "" {
		return
	}

	// go run(p, runPomodoro)
}

func (p Pomodoro) GetTimeRemaining() time.Duration {
	return p.timeRemaining
}

func (p Pomodoro) notifySubscribers(event PomodoroEvent) {
	for _, subscriber := range p.subscribers {
		subscriber(event, p)
	}
}

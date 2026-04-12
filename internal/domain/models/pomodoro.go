package models

import (
	"time"
)

type SubscriberFunc func(event PomodoroEvent, p Pomodoro)
type State func(pomodoro *Pomodoro) State

type PomodoroDto struct {
	CycleStage        CycleStage    `json:"cycleStage"`
	Task              string        `json:"task"`
	TimeRemaining     time.Duration `json:"timeRemaining"`
	PomodoriCompleted int           `json:"pomodorosCompleted"`
}

type Pomodoro struct {
	cycleStage         CycleStage
	task               string
	timeRemaining      time.Duration
	subscribers        []SubscriberFunc
	pomodoriCompleted  int
	pomodoroDuration   time.Duration
	shortBreakDuration time.Duration
	longBreakDuration  time.Duration
	loop               bool
}

func NewPomodoro(pomodoroDuration, shortBreakDuration, longBreakDuration time.Duration, loop bool) *Pomodoro {
	return &Pomodoro{
		cycleStage:         Idle,
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
		CycleStage:        p.cycleStage,
		Task:              p.task,
		TimeRemaining:     p.timeRemaining,
		PomodoriCompleted: p.pomodoriCompleted,
	}
}

func (p *Pomodoro) SetTask(task string) {
	if p.cycleStage == PomodoroStage {
		return
	}

	p.task = task
	p.notifySubscribers(TaskUpdated)
}

func (p *Pomodoro) AddSubscriber(subscriberFunc SubscriberFunc) {
	p.subscribers = append(p.subscribers, subscriberFunc)
}

func (p *Pomodoro) Start() {
	if p.task == "" {
		return
	}

	go run(p, runPomodoro)
}

func (p Pomodoro) GetTimeRemaining() time.Duration {
	return p.timeRemaining
}

func runPomodoro(pomodoro *Pomodoro) State {
	pomodoro.cycleStage = PomodoroStage
	pomodoro.timeRemaining = pomodoro.pomodoroDuration
	for pomodoro.timeRemaining > 0 {
		pomodoro.notifySubscribers(PomodoroSecondElapsed)
		pomodoro.timeRemaining -= time.Second
		time.Sleep(time.Second)
	}

	pomodoro.notifySubscribers(PomodoroDone)
	time.Sleep(time.Second)
	pomodoro.pomodoriCompleted++
	if pomodoro.pomodoriCompleted >= 4 {
		return runLongBreak
	}

	return runShortBreak
}

func runShortBreak(pomodoro *Pomodoro) State {
	pomodoro.cycleStage = ShortBreakStage
	pomodoro.timeRemaining = pomodoro.shortBreakDuration
	for pomodoro.timeRemaining > 0 {
		pomodoro.notifySubscribers(ShortBreakSecondElapsed)
		pomodoro.timeRemaining -= time.Second
		time.Sleep(time.Second)
	}

	pomodoro.notifySubscribers(ShortBreakDone)
	time.Sleep(time.Second)

	return runPomodoro
}

func runLongBreak(pomodoro *Pomodoro) State {
	pomodoro.pomodoriCompleted = 0
	pomodoro.cycleStage = LongBreakStage
	pomodoro.timeRemaining = pomodoro.longBreakDuration
	for pomodoro.timeRemaining > 0 {
		pomodoro.notifySubscribers(LongBreakSecondElapsed)
		pomodoro.timeRemaining -= time.Second
		time.Sleep(time.Second)
	}

	pomodoro.notifySubscribers(LongBreakDone)
	time.Sleep(time.Second)

	if pomodoro.loop {
		return runPomodoro
	}

	return nil
}

func run(pomodoro *Pomodoro, start State) {
	current := start
	for {
		current = current(pomodoro)
		if current == nil {
			return
		}
	}
}

func (p Pomodoro) notifySubscribers(event PomodoroEvent) {
	for _, subscriber := range p.subscribers {
		subscriber(event, p)
	}
}

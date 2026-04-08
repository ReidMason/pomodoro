package models

import (
	"time"
)

type SubscriberFunc func(event PomodoroEvent, p Pomodoro)
type State func(pomodoro Pomodoro) (Pomodoro, State)

type Stage int

const (
	PomodoroStage = iota
	ShortBreakStage
	LongBreakStage
)

type Pomodoro struct {
	Stage              Stage         `json:"stage"`
	Task               string        `json:"task"`
	TimeRemaining      time.Duration `json:"timeRemaining"`
	subscribers        []SubscriberFunc
	PomodorosCompleted int `json:"pomodorosCompleted"`
	pomodoroDuration   time.Duration
	shortBreakDuration time.Duration
	longBreakDuration  time.Duration
}

func NewPomodoro(pomodoroDuration, shortBreakDuration, longBreakDuration time.Duration, task string) *Pomodoro {
	return &Pomodoro{
		Stage:              PomodoroStage,
		Task:               task,
		TimeRemaining:      0,
		PomodorosCompleted: 0,
		pomodoroDuration:   pomodoroDuration,
		shortBreakDuration: shortBreakDuration,
		longBreakDuration:  longBreakDuration,
	}
}

func (p *Pomodoro) AddSubscriber(subscriberFunc SubscriberFunc) {
	p.subscribers = append(p.subscribers, subscriberFunc)
}

func (p Pomodoro) Start() {
	go run(p, runPomodoro)
}

func (p Pomodoro) GetTimeRemaining() time.Duration {
	return p.TimeRemaining
}

func runPomodoro(pomodoro Pomodoro) (Pomodoro, State) {
	pomodoro.Stage = PomodoroStage
	pomodoro.TimeRemaining = pomodoro.pomodoroDuration
	for pomodoro.TimeRemaining > 0 {
		pomodoro.notifySubscribers(PomodoroSecondElapsed)
		pomodoro.TimeRemaining -= time.Second
		time.Sleep(time.Second)
	}

	pomodoro.notifySubscribers(PomodoroDone)
	time.Sleep(time.Second)
	pomodoro.PomodorosCompleted++
	if pomodoro.PomodorosCompleted >= 4 {
		return pomodoro, runLongBreak
	}
	return pomodoro, runShortBreak
}

func runShortBreak(pomodoro Pomodoro) (Pomodoro, State) {
	pomodoro.Stage = ShortBreakStage
	pomodoro.TimeRemaining = pomodoro.shortBreakDuration
	for pomodoro.TimeRemaining > 0 {
		pomodoro.notifySubscribers(ShortBreakSecondElapsed)
		pomodoro.TimeRemaining -= time.Second
		time.Sleep(time.Second)
	}

	pomodoro.notifySubscribers(ShortBreakDone)
	time.Sleep(time.Second)

	return pomodoro, runPomodoro
}

func runLongBreak(pomodoro Pomodoro) (Pomodoro, State) {
	pomodoro.PomodorosCompleted = 0
	pomodoro.Stage = LongBreakStage
	pomodoro.TimeRemaining = pomodoro.longBreakDuration
	for pomodoro.TimeRemaining > 0 {
		pomodoro.notifySubscribers(LongBreakSecondElapsed)
		pomodoro.TimeRemaining -= time.Second
		time.Sleep(time.Second)
	}

	pomodoro.notifySubscribers(LongBreakDone)
	time.Sleep(time.Second)

	// TODO: remove this infinite loop
	// return pomodoro, nil
	return pomodoro, runPomodoro
}

func run(pomodoro Pomodoro, start State) Pomodoro {
	current := start
	for {
		pomodoro, current = current(pomodoro)
		if current == nil {
			return pomodoro
		}
	}
}

func (p Pomodoro) notifySubscribers(event PomodoroEvent) {
	for _, subscriber := range p.subscribers {
		subscriber(event, p)
	}
}

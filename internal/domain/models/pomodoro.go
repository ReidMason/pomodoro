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
	stage              Stage
	task               string
	timeRemaining      time.Duration
	subscribers        []SubscriberFunc
	pomodorosCompleted int
	pomodoroDuration   time.Duration
	shortBreakDuration time.Duration
	longBreakDuration  time.Duration
}

const oneSecond = 1 * time.Second

func NewPomodoro(pomodoroDuration, shortBreakDuration, longBreakDuration time.Duration, task string) *Pomodoro {
	return &Pomodoro{
		stage:              PomodoroStage,
		task:               task,
		timeRemaining:      0,
		pomodorosCompleted: 0,
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
	return p.timeRemaining
}

func runPomodoro(pomodoro Pomodoro) (Pomodoro, State) {
	pomodoro.stage = PomodoroStage
	pomodoro.timeRemaining = pomodoro.pomodoroDuration
	for pomodoro.timeRemaining > 0 {
		pomodoro.notifySubscribers(PomodoroSecondElapsed)
		pomodoro.timeRemaining -= oneSecond
		time.Sleep(oneSecond)
	}

	pomodoro.notifySubscribers(PomodoroDone)
	pomodoro.pomodorosCompleted++
	if pomodoro.pomodorosCompleted >= 4 {
		return pomodoro, runLongBreak
	}
	return pomodoro, runShortBreak
}

func runShortBreak(pomodoro Pomodoro) (Pomodoro, State) {
	pomodoro.stage = ShortBreakStage
	pomodoro.timeRemaining = pomodoro.shortBreakDuration
	for pomodoro.timeRemaining > 0 {
		pomodoro.notifySubscribers(ShortBreakSecondElapsed)
		pomodoro.timeRemaining -= oneSecond
		time.Sleep(oneSecond)
	}

	pomodoro.notifySubscribers(ShortBreakDone)

	return pomodoro, runPomodoro
}

func runLongBreak(pomodoro Pomodoro) (Pomodoro, State) {
	pomodoro.pomodorosCompleted = 0
	pomodoro.stage = LongBreakStage
	pomodoro.timeRemaining = pomodoro.longBreakDuration
	for pomodoro.timeRemaining > 0 {
		pomodoro.notifySubscribers(LongBreakSecondElapsed)
		pomodoro.timeRemaining -= oneSecond
		time.Sleep(oneSecond)
	}

	pomodoro.notifySubscribers(LongBreakDone)
	return pomodoro, nil
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

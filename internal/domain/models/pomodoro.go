package models

import (
	"time"
)

type SubscriberFunc func(event string, p *Pomodoro)
type State func(pomodoro Pomodoro) (Pomodoro, State)

type Pomodoro struct {
	status             string
	timeRemaining      time.Duration
	subscribers        []SubscriberFunc
	pomodorosCompleted int
	pomodoroDuration   time.Duration
	shortBreakDuration time.Duration
	longBreakDuration  time.Duration
}

const oneSecond = 1 * time.Second

func NewPomodoro(pomodoroDuration, shortBreakDuration, longBreakDuration time.Duration) *Pomodoro {
	return &Pomodoro{
		status:             "",
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
	pomodoro.timeRemaining = pomodoro.pomodoroDuration
	for pomodoro.timeRemaining > 0 {
		pomodoro.notifySubscribers("Pomodoro.SecondElapsed")
		pomodoro.timeRemaining -= oneSecond
		time.Sleep(oneSecond)
	}

	pomodoro.notifySubscribers("Pomodoro.Done")
	pomodoro.pomodorosCompleted++
	if pomodoro.pomodorosCompleted >= 4 {
		return pomodoro, runLongBreak
	}
	return pomodoro, runShortBreak
}

func runShortBreak(pomodoro Pomodoro) (Pomodoro, State) {
	pomodoro.timeRemaining = pomodoro.shortBreakDuration
	for pomodoro.timeRemaining > 0 {
		pomodoro.notifySubscribers("ShortBreak.SecondElapsed")
		pomodoro.timeRemaining -= oneSecond
		time.Sleep(oneSecond)
	}

	pomodoro.notifySubscribers("ShortBreak.Done")

	return pomodoro, runPomodoro
}

func runLongBreak(pomodoro Pomodoro) (Pomodoro, State) {
	pomodoro.pomodorosCompleted = 0
	pomodoro.timeRemaining = pomodoro.longBreakDuration
	for pomodoro.timeRemaining > 0 {
		pomodoro.notifySubscribers("LongBreak.SecondElapsed")
		pomodoro.timeRemaining -= oneSecond
		time.Sleep(oneSecond)
	}

	pomodoro.notifySubscribers("LongBreak.Done")
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

func (p *Pomodoro) notifySubscribers(event string) {
	for _, subscriber := range p.subscribers {
		subscriber(event, p)
	}
}

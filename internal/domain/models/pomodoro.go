package models

import (
	"time"
)

type SubscriberFunc func(event PomodoroEvent, p Pomodoro)
type State func(pomodoro *Pomodoro) State

type Pomodoro struct {
	CycleStage         CycleStage    `json:"cycleStage"`
	Task               string        `json:"task"`
	TimeRemaining      time.Duration `json:"timeRemaining"`
	subscribers        []SubscriberFunc
	PomodorosCompleted int `json:"pomodorosCompleted"`
	pomodoroDuration   time.Duration
	shortBreakDuration time.Duration
	longBreakDuration  time.Duration
	loop               bool
}

func NewPomodoro(pomodoroDuration, shortBreakDuration, longBreakDuration time.Duration, task string, loop bool) *Pomodoro {
	return &Pomodoro{
		CycleStage:         Idle,
		Task:               task,
		TimeRemaining:      0,
		PomodorosCompleted: 0,
		pomodoroDuration:   pomodoroDuration,
		shortBreakDuration: shortBreakDuration,
		longBreakDuration:  longBreakDuration,
		loop:               loop,
	}
}

func (p *Pomodoro) SetTask(task string) {
	if p.CycleStage == PomodoroStage {
		return
	}

	p.Task = task
	p.notifySubscribers(TaskUpdated)
}

func (p *Pomodoro) AddSubscriber(subscriberFunc SubscriberFunc) {
	p.subscribers = append(p.subscribers, subscriberFunc)
}

func (p *Pomodoro) Start() {
	go run(p, runPomodoro)
}

func (p Pomodoro) GetTimeRemaining() time.Duration {
	return p.TimeRemaining
}

func runPomodoro(pomodoro *Pomodoro) State {
	pomodoro.CycleStage = PomodoroStage
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
		return runLongBreak
	}

	return runShortBreak
}

func runShortBreak(pomodoro *Pomodoro) State {
	pomodoro.CycleStage = ShortBreakStage
	pomodoro.TimeRemaining = pomodoro.shortBreakDuration
	for pomodoro.TimeRemaining > 0 {
		pomodoro.notifySubscribers(ShortBreakSecondElapsed)
		pomodoro.TimeRemaining -= time.Second
		time.Sleep(time.Second)
	}

	pomodoro.notifySubscribers(ShortBreakDone)
	time.Sleep(time.Second)

	return runPomodoro
}

func runLongBreak(pomodoro *Pomodoro) State {
	pomodoro.PomodorosCompleted = 0
	pomodoro.CycleStage = LongBreakStage
	pomodoro.TimeRemaining = pomodoro.longBreakDuration
	for pomodoro.TimeRemaining > 0 {
		pomodoro.notifySubscribers(LongBreakSecondElapsed)
		pomodoro.TimeRemaining -= time.Second
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

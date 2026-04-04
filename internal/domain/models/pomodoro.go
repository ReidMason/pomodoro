package models

import "time"

type SubscriberFunc func(event string, p *Pomodoro)

type Pomodoro struct {
	status        string
	timeRemaining time.Duration
	subscribers   []SubscriberFunc
	interval      time.Duration
}

const oneSecond = 1 * time.Second

func NewPomodoro(interval time.Duration) *Pomodoro {
	return &Pomodoro{
		status:        "",
		timeRemaining: 0,
		interval:      interval,
	}
}

func (p *Pomodoro) AddSubscriber(subscriberFunc SubscriberFunc) {
	p.subscribers = append(p.subscribers, subscriberFunc)
}

func (p *Pomodoro) Start() {
	p.timeRemaining = p.interval
	go p.runPomodoro()
}

func (p *Pomodoro) GetTimeRemaining() time.Duration {
	return p.timeRemaining
}

func (p *Pomodoro) runPomodoro() {
	for p.timeRemaining > 0 {
		p.notifySubscribers("SecondElapsed")
		p.timeRemaining -= oneSecond
		time.Sleep(oneSecond)
	}

	p.notifySubscribers("Done")
}

func (p *Pomodoro) notifySubscribers(event string) {
	for _, subscriber := range p.subscribers {
		subscriber(event, p)
	}
}

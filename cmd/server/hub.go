package main

import (
	"slices"
	"time"

	"github.com/ReidMason/pomodoro/internal/domain/models/pomodoro"
)

const TICK_INTERVAL = time.Millisecond * 100

type Hub struct {
	// Registered clients.
	clients map[*Client]struct{}

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	deregister chan *Client

	Pomodoro   *pomodoro.Pomodoro
	ticker     *time.Ticker
	tickerSync chan struct{}
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		deregister: make(chan *Client),
		clients:    make(map[*Client]struct{}),
		tickerSync: make(chan struct{}, 1),
	}
}

func (h *Hub) syncTicker() {
	pom := h.Pomodoro.ToDto()
	statesThatNeedTicks := []pomodoro.State{
		pomodoro.PomodoroInProgress,
		pomodoro.ShortBreakInProgress,
		pomodoro.LongBreakInProgress,
	}

	if pom.CycleStage == pomodoro.Idle && h.ticker != nil {
		h.ticker.Stop()
		h.ticker = nil
	} else if slices.Contains(statesThatNeedTicks, pom.CycleStage) && h.ticker == nil {
		h.ticker = time.NewTicker(TICK_INTERVAL)
	}
}

func (h *Hub) run() {
	h.Pomodoro.AddSubscriber(func(event pomodoro.PomodoroEvent, p pomodoro.Pomodoro) {
		h.tickerSync <- struct{}{}
	})

	for {
		var tickCh <-chan time.Time
		if h.ticker != nil {
			tickCh = h.ticker.C
		}

		select {
		case <-tickCh:
			h.Pomodoro.HandleCommand(pomodoro.Command{
				Kind: pomodoro.Tick,
			})
		case <-h.tickerSync:
			h.syncTicker()
		case client := <-h.register:
			h.clients[client] = struct{}{}
		case client := <-h.deregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

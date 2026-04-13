package pomodoro

import "time"

type State int

const (
	Idle State = iota
	PomodoroInProgress
	ShortBreakInProgress
	LongBreakInProgress
)

var stageNameMapping = map[State]string{
	Idle:                 "Idle",
	PomodoroInProgress:   "Pomodoro",
	ShortBreakInProgress: "Short break",
	LongBreakInProgress:  "Long break",
}

func (s State) String() string {
	return stageNameMapping[s]
}

var stateHandlers = map[State]commandHandler{
	Idle:                 HandleCommandIdle,
	PomodoroInProgress:   HandleCommandPomodoroInProgress,
	ShortBreakInProgress: HandleCommandShortBreakInProgress,
	LongBreakInProgress:  HandleCommandLongBreakInProgress,
}

func (s State) HandleCommand(pomodoro *Pomodoro, command Command) (State, PomodoroEvent) {
	return stateHandlers[s](pomodoro, command)
}

func HandleCommandIdle(pomodoro *Pomodoro, command Command) (State, PomodoroEvent) {
	switch command.Kind {
	case SetTask:
		pomodoro.task = command.Task
		return Idle, TaskUpdated
	case Start:
		pomodoro.phaseEndsAt = time.Now().Add(pomodoro.pomodoroDuration)
		return PomodoroInProgress, PomodoroStarted
	}

	return Idle, None
}

func HandleCommandPomodoroInProgress(pomodoro *Pomodoro, command Command) (State, PomodoroEvent) {
	switch command.Kind {
	case Tick:
		remaining := time.Until(pomodoro.phaseEndsAt)
		if remaining <= 0 {
			pomodoro.phaseEndsAt = time.Now().Add(pomodoro.shortBreakDuration)
			return ShortBreakInProgress, ShortBreakStarted
		}

		return PomodoroInProgress, PomodoroSecondElapsed
	}

	return PomodoroInProgress, None
}

func HandleCommandShortBreakInProgress(pomodoro *Pomodoro, command Command) (State, PomodoroEvent) {
	switch command.Kind {
	case Tick:
		remaining := time.Until(pomodoro.phaseEndsAt)
		if remaining <= 0 {
			pomodoro.phaseEndsAt = time.Now().Add(pomodoro.pomodoroDuration)
			return PomodoroInProgress, PomodoroStarted
		}

		return ShortBreakInProgress, ShortBreakSecondElapsed
	}

	return ShortBreakInProgress, None
}

func HandleCommandLongBreakInProgress(pomodoro *Pomodoro, command Command) (State, PomodoroEvent) {
	switch command.Kind {
	case SetTask:
		pomodoro.task = command.Task
		return Idle, TaskUpdated
	}

	return LongBreakInProgress, None
}

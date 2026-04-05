package models

type PomodoroEvent int

const (
	PomodoroSecondElapsed PomodoroEvent = iota
	PomodoroDone
	ShortBreakSecondElapsed
	ShortBreakDone
	LongBreakSecondElapsed
	LongBreakDone
)

var eventNameMapping = map[PomodoroEvent]string{
	PomodoroSecondElapsed:   "Pomodoro.SecondElapsed",
	PomodoroDone:            "Pomodoro.Done",
	ShortBreakSecondElapsed: "ShortBreak.SecondElapsed",
	ShortBreakDone:          "ShortBreak.Done",
	LongBreakSecondElapsed:  "LongBreak.SecondElapsed",
	LongBreakDone:           "LongBreak.Done",
}

func (pe PomodoroEvent) String() string {
	return eventNameMapping[pe]
}

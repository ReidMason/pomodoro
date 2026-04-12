package pomodoro

type PomodoroEvent int

const (
	None PomodoroEvent = iota
	PomodoroSecondElapsed
	PomodoroDone
	ShortBreakSecondElapsed
	ShortBreakDone
	LongBreakSecondElapsed
	LongBreakDone
	TaskUpdated
)

var eventNameMapping = map[PomodoroEvent]string{
	PomodoroSecondElapsed:   "Pomodoro.SecondElapsed",
	PomodoroDone:            "Pomodoro.Done",
	ShortBreakSecondElapsed: "ShortBreak.SecondElapsed",
	ShortBreakDone:          "ShortBreak.Done",
	LongBreakSecondElapsed:  "LongBreak.SecondElapsed",
	LongBreakDone:           "LongBreak.Done",
	TaskUpdated:             "TaskUpdated",
}

func (pe PomodoroEvent) String() string {
	return eventNameMapping[pe]
}

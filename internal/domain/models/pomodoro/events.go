package pomodoro

type PomodoroEvent int

const (
	None PomodoroEvent = iota

	PomodoroStarted
	PomodoroSecondElapsed

	ShortBreakStarted
	ShortBreakSecondElapsed
	ShortBreakDone

	LongBreakStarted
	LongBreakSecondElapsed
	LongBreakDone

	TaskUpdated
)

var eventNameMapping = map[PomodoroEvent]string{
	PomodoroStarted:         "Pomodoro.Started",
	PomodoroSecondElapsed:   "Pomodoro.SecondElapsed",
	ShortBreakSecondElapsed: "ShortBreak.SecondElapsed",
	ShortBreakDone:          "ShortBreak.Done",
	LongBreakSecondElapsed:  "LongBreak.SecondElapsed",
	LongBreakDone:           "LongBreak.Done",
	TaskUpdated:             "TaskUpdated",
}

func (pe PomodoroEvent) String() string {
	return eventNameMapping[pe]
}

package pomodoro

type PomodoroEvent int

const (
	None PomodoroEvent = iota

	PomodoroStarted

	ShortBreakStarted
	ShortBreakDone

	LongBreakStarted
	LongBreakDone

	SecondElapsed
	TaskUpdated
)

var eventNameMapping = map[PomodoroEvent]string{
	PomodoroStarted: "Pomodoro.Started",
	ShortBreakDone:  "ShortBreak.Done",
	LongBreakDone:   "LongBreak.Done",
	TaskUpdated:     "TaskUpdated",
}

func (pe PomodoroEvent) String() string {
	return eventNameMapping[pe]
}

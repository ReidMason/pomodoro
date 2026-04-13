package pomodoro

type PomodoroEvent int

const (
	None PomodoroEvent = iota

	PomodoroStarted

	ShortBreakStarted
	ShortBreakDone

	LongBreakStarted
	LongBreakDone

	TaskUpdated
)

var eventNameMapping = map[PomodoroEvent]string{
	PomodoroStarted: "Pomodoro.Started",

	ShortBreakStarted: "ShortBreak.Started",
	ShortBreakDone:    "ShortBreak.Done",

	LongBreakStarted: "LongBreak.Started",
	LongBreakDone:    "LongBreak.Done",

	TaskUpdated: "TaskUpdated",
}

func (pe PomodoroEvent) String() string {
	return eventNameMapping[pe]
}

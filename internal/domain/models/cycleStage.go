package models

type CycleStage int

const (
	PomodoroStage = iota
	ShortBreakStage
	LongBreakStage
)

var stageNameMapping = map[CycleStage]string{
	PomodoroStage:   "Pomodoro",
	ShortBreakStage: "Short break",
	LongBreakStage:  "Long break",
}

func (s CycleStage) String() string {
	return stageNameMapping[s]
}

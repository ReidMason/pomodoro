package models

type CycleStage int

const (
	Idle = iota
	PomodoroStage
	ShortBreakStage
	LongBreakStage
)

var stageNameMapping = map[CycleStage]string{
	Idle:            "Idle",
	PomodoroStage:   "Pomodoro",
	ShortBreakStage: "Short break",
	LongBreakStage:  "Long break",
}

func (s CycleStage) String() string {
	return stageNameMapping[s]
}

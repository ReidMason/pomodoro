package pomodoro

type State int

const (
	Idle State = iota
	PomodoroStage
	ShortBreakStage
	LongBreakStage
)

var stageNameMapping = map[State]string{
	Idle:            "Idle",
	PomodoroStage:   "Pomodoro",
	ShortBreakStage: "Short break",
	LongBreakStage:  "Long break",
}

func (s State) String() string {
	return stageNameMapping[s]
}

var stateHandlers = map[State]commandHandler{
	Idle: HandleCommandIdle,
}

func (s State) HandleCommand(pomodoro *Pomodoro, command Command) State {
	return stateHandlers[s](pomodoro, command)
}

func HandleCommandIdle(pomodoro *Pomodoro, command Command) State {
	return Idle
}

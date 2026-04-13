package pomodoro

type Command struct {
	Kind CommandKind
	Task string
}

type CommandKind int

const (
	Start CommandKind = iota
	SetTask
	Tick
)

type commandHandler func(pomodoro *Pomodoro, command Command) (State, PomodoroEvent)

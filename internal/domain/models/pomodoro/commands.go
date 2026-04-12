package pomodoro

type Command int

const (
	Start Command = iota
)

type commandHandler func(pomodoro *Pomodoro, command Command) State

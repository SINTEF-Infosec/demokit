package core

type ActionCondition func(event *Event) bool

type Action struct {
	Name        string
	Do          EventHandler
	DoCondition ActionCondition
	DoDelay		int
	Then        *Action
}

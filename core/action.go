package core

type Action struct {
	Name        string
	Do          EventHandler
	DoCondition func() bool
	Then        *Action
}

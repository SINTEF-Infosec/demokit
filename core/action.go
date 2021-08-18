package core

type Action struct {
	Name        string
	Do          func()
	DoCondition func() bool
	Then        *Action
}

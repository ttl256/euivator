package cmd

import "fmt"

type AtInputPositionError struct {
	Position int
	Err      error
}

func (e AtInputPositionError) Error() string {
	baseMsg := "at position %d"
	if e.Err != nil {
		return fmt.Sprintf(baseMsg+": %v", e.Position, e.Err)
	}
	return fmt.Sprintf(baseMsg, e.Position)
}

func (e AtInputPositionError) Unwrap() error {
	return e.Err
}

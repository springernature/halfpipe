package linterrors

import "fmt"

type TriggerError struct {
	TriggerName string
}

func NewTriggerError(triggerName string) TriggerError {
	return TriggerError{triggerName}
}

func (e TriggerError) Error() string {
	return fmt.Sprintf("Invalid trigger '%s': you are only allowed one of these in a pipeline", e.TriggerName)
}

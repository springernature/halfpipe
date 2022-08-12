package linterrors

import "fmt"

type DeprecatedField struct {
	Name   string
	Reason string
}

func NewDeprecatedField(name string, reason string) DeprecatedField {
	return DeprecatedField{name, reason}
}

func (e DeprecatedField) Error() string {
	return fmt.Sprintf("deprecated field '%s': %s", e.Name, e.Reason)
}

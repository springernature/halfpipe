package linterrors

import "fmt"

type UnsupportedField struct {
	Text string
}

func NewUnsupportedField(text string) UnsupportedField {
	return UnsupportedField{text}
}

func (e UnsupportedField) Error() string {
	return fmt.Sprintf("unsupported item: %s", e.Text)
}

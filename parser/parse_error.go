package parser

import "fmt"

type ParseError struct {
	Message string
}

func NewParseError(message string) ParseError {
	return ParseError{message}
}

func (e ParseError) Error() string {
	return fmt.Sprintf("Error parsing manifest: %s", e.Message)
}

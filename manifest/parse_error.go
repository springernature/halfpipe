package manifest

type ParseError struct {
	Message string
}

func NewParseError(message string) ParseError {
	return ParseError{message}
}

func (e ParseError) Error() string {
	return e.Message
}

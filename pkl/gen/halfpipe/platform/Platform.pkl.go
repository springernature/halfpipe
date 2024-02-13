// Code generated from Pkl module `halfpipe`. DO NOT EDIT.
package platform

import (
	"encoding"
	"fmt"
)

type Platform string

const (
	Concourse Platform = "concourse"
	Actions   Platform = "actions"
)

// String returns the string representation of Platform
func (rcv Platform) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(Platform)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for Platform.
func (rcv *Platform) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "concourse":
		*rcv = Concourse
	case "actions":
		*rcv = Actions
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid Platform`, str)
	}
	return nil
}

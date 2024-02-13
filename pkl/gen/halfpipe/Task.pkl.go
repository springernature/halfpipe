// Code generated from Pkl module `halfpipe`. DO NOT EDIT.
package halfpipe

import "github.com/apple/pkl-go/pkl"

type Task interface {
	GetType() string

	GetName() *string

	GetTimeout() *pkl.Duration

	GetRetries() int
}

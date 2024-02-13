// Code generated from Pkl module `halfpipe`. DO NOT EDIT.
package halfpipe

import "github.com/apple/pkl-go/pkl"

type Run interface {
	Task

	GetScript() string
}

var _ Run = (*RunImpl)(nil)

type RunImpl struct {
	Type string `pkl:"type"`

	Script string `pkl:"script"`

	Name *string `pkl:"name"`

	Timeout *pkl.Duration `pkl:"timeout"`

	Retries int `pkl:"retries"`
}

func (rcv *RunImpl) GetType() string {
	return rcv.Type
}

func (rcv *RunImpl) GetScript() string {
	return rcv.Script
}

func (rcv *RunImpl) GetName() *string {
	return rcv.Name
}

func (rcv *RunImpl) GetTimeout() *pkl.Duration {
	return rcv.Timeout
}

func (rcv *RunImpl) GetRetries() int {
	return rcv.Retries
}

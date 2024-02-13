// Code generated from Pkl module `halfpipe`. DO NOT EDIT.
package halfpipe

import "github.com/apple/pkl-go/pkl"

type Sequence interface {
	Task

	GetTasks() []Task
}

var _ Sequence = (*SequenceImpl)(nil)

type SequenceImpl struct {
	Type string `pkl:"type"`

	Tasks []Task `pkl:"tasks"`

	Name *string `pkl:"name"`

	Timeout *pkl.Duration `pkl:"timeout"`

	Retries int `pkl:"retries"`
}

func (rcv *SequenceImpl) GetType() string {
	return rcv.Type
}

func (rcv *SequenceImpl) GetTasks() []Task {
	return rcv.Tasks
}

func (rcv *SequenceImpl) GetName() *string {
	return rcv.Name
}

func (rcv *SequenceImpl) GetTimeout() *pkl.Duration {
	return rcv.Timeout
}

func (rcv *SequenceImpl) GetRetries() int {
	return rcv.Retries
}

// Code generated from Pkl module `halfpipe`. DO NOT EDIT.
package halfpipe

import "github.com/apple/pkl-go/pkl"

type Parallel interface {
	Task

	GetTasks() []Task
}

var _ Parallel = (*ParallelImpl)(nil)

type ParallelImpl struct {
	Type string `pkl:"type"`

	Tasks []Task `pkl:"tasks"`

	Name *string `pkl:"name"`

	Timeout *pkl.Duration `pkl:"timeout"`

	Retries int `pkl:"retries"`
}

func (rcv *ParallelImpl) GetType() string {
	return rcv.Type
}

func (rcv *ParallelImpl) GetTasks() []Task {
	return rcv.Tasks
}

func (rcv *ParallelImpl) GetName() *string {
	return rcv.Name
}

func (rcv *ParallelImpl) GetTimeout() *pkl.Duration {
	return rcv.Timeout
}

func (rcv *ParallelImpl) GetRetries() int {
	return rcv.Retries
}

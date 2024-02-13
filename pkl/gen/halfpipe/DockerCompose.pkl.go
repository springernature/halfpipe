// Code generated from Pkl module `halfpipe`. DO NOT EDIT.
package halfpipe

import "github.com/apple/pkl-go/pkl"

type DockerCompose interface {
	Task

	GetPath() string

	GetCommand() *string
}

var _ DockerCompose = (*DockerComposeImpl)(nil)

type DockerComposeImpl struct {
	Type string `pkl:"type"`

	Path string `pkl:"path"`

	Command *string `pkl:"command"`

	Name *string `pkl:"name"`

	Timeout *pkl.Duration `pkl:"timeout"`

	Retries int `pkl:"retries"`
}

func (rcv *DockerComposeImpl) GetType() string {
	return rcv.Type
}

func (rcv *DockerComposeImpl) GetPath() string {
	return rcv.Path
}

func (rcv *DockerComposeImpl) GetCommand() *string {
	return rcv.Command
}

func (rcv *DockerComposeImpl) GetName() *string {
	return rcv.Name
}

func (rcv *DockerComposeImpl) GetTimeout() *pkl.Duration {
	return rcv.Timeout
}

func (rcv *DockerComposeImpl) GetRetries() int {
	return rcv.Retries
}

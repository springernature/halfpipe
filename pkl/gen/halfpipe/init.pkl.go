// Code generated from Pkl module `halfpipe`. DO NOT EDIT.
package halfpipe

import "github.com/apple/pkl-go/pkl"

func init() {
	pkl.RegisterMapping("halfpipe", Halfpipe{})
	pkl.RegisterMapping("halfpipe#Run", RunImpl{})
	pkl.RegisterMapping("halfpipe#DockerCompose", DockerComposeImpl{})
	pkl.RegisterMapping("halfpipe#Parallel", ParallelImpl{})
	pkl.RegisterMapping("halfpipe#Sequence", SequenceImpl{})
}

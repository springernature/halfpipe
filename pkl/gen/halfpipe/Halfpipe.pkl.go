// Code generated from Pkl module `halfpipe`. DO NOT EDIT.
package halfpipe

import (
	"context"

	"github.com/apple/pkl-go/pkl"
	"github.com/springernature/halfpipe/pkl/gen/halfpipe/platform"
)

type Halfpipe struct {
	Team string `pkl:"team"`

	Pipeline string `pkl:"pipeline"`

	// platform is the target CI system
	Platform platform.Platform `pkl:"platform"`

	Tasks []Task `pkl:"tasks"`

	Output any `pkl:"output"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Halfpipe
func LoadFromPath(ctx context.Context, path string) (ret *Halfpipe, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return nil, err
	}
	defer func() {
		cerr := evaluator.Close()
		if err == nil {
			err = cerr
		}
	}()
	ret, err = Load(ctx, evaluator, pkl.FileSource(path))
	return ret, err
}

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Halfpipe
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (*Halfpipe, error) {
	var ret Halfpipe
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

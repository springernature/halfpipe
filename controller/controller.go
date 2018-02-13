package controller

import (
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/model"
	"github.com/springernature/halfpipe/parser"
)

type Controller struct {
	Fs      afero.Afero
	Linters []linters.Linter
}

func (c Controller) Process() (string, []error) {

	//1. setup
	exists, err := c.Fs.Exists(".halfpipe.io")
	if err != nil {
		return "", []error{err}
	}
	if !exists {
		return "", []error{errors.New(".halfpipe.io does not exist")}
	}

	var blah model.Manifest

	//2. linters
	var errs []error
	for _, linter := range c.Linters {
		errs = append(errs, linter.Lint(blah)...)
	}

	if len(errs) > 0 {
		return "", errs
	}

	//interate + accumulate errors

	//3. generate pipeline

	return "pipeline!", nil

	//exists, err := fs.Exists(".halfpipe.io")
	//if err != nil {
	//	return err
	//}
	//if !exists {
	//	return errors.New("missing .halfpipe.io")
	//}
	//
	//_, err = ParseManifest(fs)
	//if err != nil {
	//	return err
	//}
	//
	//return nil
}

func ParseManifest(fs afero.Afero) (model.Manifest, error) {
	content, err := fs.ReadFile(".halfpipe.io")
	if err != nil {
		return model.Manifest{}, nil
	}

	if len(content) == 0 {
		return model.Manifest{}, errors.New(".halfpipe.io must not be empty")
	}

	stringContent := string(content)
	man, errs := parser.Parse(stringContent)
	if len(errs) != 0 {
		return model.Manifest{}, errs[0]
	}

	return man, nil
}

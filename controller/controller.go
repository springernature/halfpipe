package controller

import (
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

func Process(fs afero.Afero) error {

	exists, err := fs.Exists(".halfpipe.io")
	if err != nil {
		return err
	}
	if exists {
		return nil
	} else {
		return errors.New("missing .halfpipe.io")
	}

}

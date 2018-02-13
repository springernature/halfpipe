package config

import (
	"io"

	"github.com/robwhitby/halfpipe-cli/model"
	"github.com/spf13/afero"
)

const (
	DocumentationRootUrl = "http://docs.halfpipe.io"
	ManifestFilename     = ".halfpipe.io"
)

type Config struct {
	FileSystem    afero.Fs
	Options       Options
	OutputWriter  io.Writer
	ErrorWriter   io.Writer
	SecretChecker model.SecretChecker
	Version       string
}

type Options struct {
	ShowVersion bool `short:"v" long:"version" description:"Display version"`
	Args        Args `positional-args:"true"`
}

type Args struct {
	Dir string `positional-arg-name:"directory" description:"Path to process. Defaults to pwd."`
}

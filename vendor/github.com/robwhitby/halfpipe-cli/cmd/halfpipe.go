package main

import (
	"syscall"

	"os"

	"github.com/jessevdk/go-flags"
	"github.com/robwhitby/halfpipe-cli/config"
	"github.com/robwhitby/halfpipe-cli/controller"
	"github.com/spf13/afero"
)

var version = "0.0" // should be set by build

func main() {
	conf := config.Config{
		FileSystem:    afero.NewOsFs(),
		OutputWriter:  os.Stdout,
		ErrorWriter:   os.Stderr,
		SecretChecker: func(s string) bool { return false },
		Version:       version,
	}
	flags.Parse(&conf.Options)

	if ok := controller.Process(conf); !ok {
		syscall.Exit(1)
	}
}

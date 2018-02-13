package controller

import (
	"testing"

	"bytes"

	"os"

	"github.com/robwhitby/halfpipe-cli/config"
	. "github.com/robwhitby/halfpipe-cli/model"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

const root = "/root/"

var opts = config.Options{
	ShowVersion: false,
	Args: config.Args{
		Dir: root,
	},
}

func setup() (config.Config, *bytes.Buffer, *bytes.Buffer) {
	stdOut := bytes.NewBufferString("")
	stdErr := bytes.NewBufferString("")

	//only 'valid.secret' exists
	secretChecker := func(s string) bool { return s == "valid.secret" }

	conf := config.Config{
		FileSystem:    afero.NewMemMapFs(),
		Options:       opts,
		OutputWriter:  stdOut,
		ErrorWriter:   stdErr,
		SecretChecker: secretChecker,
		Version:       "0.1",
	}

	conf.FileSystem.Mkdir(root, 0777)
	return conf, stdOut, stdErr
}

func TestNoManifest(t *testing.T) {
	conf, _, stdErr := setup()
	ok := Process(conf)

	assert.False(t, ok)
	expectedError := NewFileError(config.ManifestFilename, "does not exist")
	assert.Contains(t, stdErr.String(), expectedError.Error())
}

func TestManifestParseError(t *testing.T) {
	conf, _, stdErr := setup()
	afero.WriteFile(conf.FileSystem, root+config.ManifestFilename, []byte("^&*(^&*"), 0777)
	ok := Process(conf)

	assert.False(t, ok)
	expectedError := NewParseError("")
	assert.Contains(t, stdErr.String(), expectedError.Error())
}

func TestManifestLintError(t *testing.T) {
	conf, _, stdErr := setup()
	afero.WriteFile(conf.FileSystem, root+config.ManifestFilename, []byte("foo: bar"), 0777)
	ok := Process(conf)

	assert.False(t, ok)
	expectedError := NewMissingField("team")
	assert.Contains(t, stdErr.String(), expectedError.Error())
}

func TestManifestRequiredFileError(t *testing.T) {
	conf, _, stdErr := setup()
	yaml := `
team: foo
repo: 
  uri: git@github.com/foo/bar.git
tasks:
- name: run
  script: ./build.sh
  image: bar
`
	afero.WriteFile(conf.FileSystem, root+config.ManifestFilename, []byte(yaml), 0777)
	ok := Process(conf)

	assert.False(t, ok)
	expectedError := NewFileError("./build.sh", "does not exist")
	assert.Contains(t, stdErr.String(), expectedError.Error())
}

func TestManifestRequiredSecretError(t *testing.T) {
	conf, _, stdErr := setup()
	yaml := `
team: foo
repo: 
  uri: git@github.com/foo/bar.git
tasks:
- name: run
  script: build.sh
  image: bar
  vars:
    badsecret: ((path.to.key))
    goodsecret: ((valid.secret))
`
	afero.WriteFile(conf.FileSystem, root+config.ManifestFilename, []byte(yaml), 0777)
	afero.WriteFile(conf.FileSystem, root+"build.sh", []byte("x"), 0777)
	ok := Process(conf)

	assert.False(t, ok)
	expectedError := NewMissingSecret("path.to.key")
	assert.Contains(t, stdErr.String(), expectedError.Error())

	unexpected := NewMissingSecret("valid.secret")
	assert.NotContains(t, stdErr.String(), unexpected.Error())
}

func TestValidManifest(t *testing.T) {
	conf, stdOut, stdErr := setup()

	yaml := `
team: foo
repo: 
  uri: git@github.com/foo/bar.git
tasks:
- name: run
  script: build.sh
  image: bar
  vars:
    secret: ((valid.secret))
`
	afero.WriteFile(conf.FileSystem, root+config.ManifestFilename, []byte(yaml), 0777)
	afero.WriteFile(conf.FileSystem, "/root/build.sh", []byte("x"), 0777)
	ok := Process(conf)

	assert.True(t, ok)
	assert.Empty(t, stdErr.String())
	assert.Contains(t, stdOut.String(), "Good job")
}

func TestController_ChecksRootDir(t *testing.T) {
	conf, _, stdErr := setup()
	conf.Options.Args.Dir = "/invalid/root"
	ok := Process(conf)

	assert.False(t, ok)
	expectedError := NewFileError("/invalid/root", "is not a directory")
	assert.Contains(t, stdErr.String(), expectedError.Error())
}

func TestAbsDirectory_Abs(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	fs.MkdirAll("/some/dir", 0777)

	dir, _ := absDir("/some/dir/", fs)
	assert.Equal(t, "/some/dir", dir)

	dir, _ = absDir("/some/dir/../dir", fs)
	assert.Equal(t, "/some/dir", dir)
}

func TestAbsDirectory_Relative(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	pwd, _ := os.Getwd()
	fs.MkdirAll(pwd, 0777)

	dir, _ := absDir(".", fs)
	assert.Equal(t, pwd, dir)

	dir, _ = absDir("", fs)
	assert.Equal(t, pwd, dir)

	fs.MkdirAll(pwd+"/foo", 0777)

	dir, _ = absDir("foo", fs)
	assert.Equal(t, pwd+"/foo", dir)

	dir, _ = absDir("./foo/", fs)
	assert.Equal(t, pwd+"/foo", dir)
}

func TestAbsDirectory_Errors(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	pwd, _ := os.Getwd()
	fs.MkdirAll(pwd, 0777)

	fileError := NewFileError("missing", "is not a directory")

	_, err := absDir("missing", fs)
	assert.Equal(t, fileError, err)

	fs.WriteFile("/file", []byte{}, 0777)
	_, err = absDir("/file", fs)
	assert.IsType(t, fileError, err)
}

func TestOption_Version(t *testing.T) {
	conf, stdOut, _ := setup()
	conf.Options.ShowVersion = true
	ok := Process(conf)

	assert.True(t, ok)
	assert.Equal(t, versionMessage(conf.Version)+"\n", stdOut.String())
}

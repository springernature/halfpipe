package retrigger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/concourse/atc"
	"os"
	"os/exec"
	"strings"
)

type Build atc.Build

func (build Build) String() string {
	return fmt.Sprintf("%s/%s", build.PipelineName, build.JobName)
}

func (build Build) Retrigger() (err error) {
	flyPath, err := exec.LookPath("fly")
	if err != nil {
		return
	}

	cmd := exec.Cmd{
		Path:   flyPath,
		Args:   []string{"fly", "-t", build.TeamName, "trigger-job", "-j", fmt.Sprintf(`"%s/%s"`, build.PipelineName, build.JobName)},
		Stderr: os.Stderr,
		Stdout: os.Stdout,
	}
	fmt.Printf("$ %s\n", strings.Join(cmd.Args, " "))

	err = cmd.Run()
	return
}

type Builds []Build

func (builds Builds) GetErrored() (erroredBuilds Builds) {
	for _, build := range builds {
		if build.Status == "errored" {
			erroredBuilds = append(erroredBuilds, build)
		}
	}
	return
}

func (builds Builds) IsLatest(build Build) bool {
	highestIdForBuild := 0
	for _, b := range builds {
		if b.PipelineName == build.PipelineName && b.JobName == build.JobName {
			if b.ID > highestIdForBuild {
				highestIdForBuild = b.ID
			}
		}
	}

	return build.ID >= highestIdForBuild
}

func GetBuilds(team string, count string) (builds Builds, err error) {
	flyPath, err := exec.LookPath("fly")
	if err != nil {
		return
	}

	stdoutBuffer := bytes.Buffer{}
	cmd := exec.Cmd{
		Path:   flyPath,
		Args:   []string{"fly", "-t", team, "builds", "-t", team, "-c", count, "--json"},
		Stderr: os.Stderr,
		Stdout: &stdoutBuffer,
	}
	fmt.Printf("$ %s\n", strings.Join(cmd.Args, " "))

	err = cmd.Run()
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(stdoutBuffer.String()), &builds)
	return
}

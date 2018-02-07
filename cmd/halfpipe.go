package main

import (
	"fmt"
	"reflect"
)

var(
	// This field will be populated in Concourse from the version resource
	// go build -ldflags "-X main.version`cat version/version`"
	version string
)

func getVersion() string {
	if version == "" {
		return "dev"
	}
	return version
}

func main() {
	fmt.Println("Hello World")
	fmt.Println("Current version is: " + getVersion())
}

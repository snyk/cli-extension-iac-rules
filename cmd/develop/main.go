package main

import (
	"log"

	"github.com/snyk/cli-extension-cloud/cloud"
	"github.com/snyk/go-application-framework/pkg/devtools"
)

func main() {
	cmd, err := devtools.Cmd(cloud.Init)
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// This file has been created by "go generate" as initial code. go generate will never update it, EXCEPT if you remove it.

// So, update it for your need.
package main

import (
	"os"

	"github.com/forj-oss/forjj/utils"
	"gopkg.in/alecthomas/kingpin.v2"
)

var cliApp JenkinsApp

func main() {
	cliApp.init()

	switch kingpin.MustParse(cliApp.App.Parse(os.Args[1:])) {
	case "service start":
		*cliApp.params.template_dir, _ = utils.Abs(*cliApp.params.template_dir)
		cliApp.start_server()
	default:
		kingpin.Usage()
	}
}

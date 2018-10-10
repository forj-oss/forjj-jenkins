// This file has been created by "go generate" as initial code. go generate will never update it, EXCEPT if you remove it.

// So, update it for your need.
package main

import (
	"log"

	"github.com/forj-oss/forjj/utils"
	"github.com/forj-oss/goforjj"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
	"fmt"
)

type JenkinsApp struct {
	App                 *kingpin.Application
	params              Params
	socket              string
	templateDefaultPath string // Combined with templateDirDefault constant, set the absolute path at plugin startup.
	Yaml                goforjj.YamlPlugin
}

const templateDirDefault = "templates"

type Params struct {
	socket_file  *string
	socket_path  *string
	template_dir *string
	daemon       *bool // Currently not used - Lot of concerns with daemonize in go... Stay in foreground
}

var (
	build_branch string
	build_commit string
	build_date string
	build_tag string
)

func (a *JenkinsApp) init() {
	a.loadPluginDef()

	appName := "forjj-jenkins"
	a.App = kingpin.New(appName, "CI jenkins plugin for FORJJ.")
	var version string
	if PRERELEASE {
		version = appName + " pre-release V" + VERSION
	} else {
		version = appName + " V" + VERSION
	}

	if build_branch != "master" {
		version += fmt.Sprintf(" branch %s", build_branch)
	}
	if build_tag == "false" {
		version += fmt.Sprintf(" patched - %s - %s", build_date, build_commit)
	}

	if version != "" {
		a.App.Version(version).Author("Christophe Larsonneur <clarsonneur@gmail.com>")
	}

	// true to create the Infra
	daemon := a.App.Command("service", "jenkins REST API service")
	daemon.Command("start", "start jenkins REST API service")
	a.params.socket_file = daemon.Flag("socket-file", "Socket file to use").Default(a.Yaml.Runtime.Service.Socket).String()
	a.params.socket_path = daemon.Flag("socket-path", "Socket file path to use").Default("/tmp/forjj-socks").String()
	a.params.daemon = daemon.Flag("daemon", "Start process in background like a daemon").Short('d').Bool()
	a.params.template_dir = daemon.Flag("templates", "Path to templates files.").Default(templateDirDefault).String()

}

// loadPluginDef Load application defaults
func (a *JenkinsApp) loadPluginDef() {
	yaml.Unmarshal([]byte(YamlDesc), &a.Yaml)
	if a.Yaml.Runtime.Service.Socket == "" {
		a.Yaml.Runtime.Service.Socket = "jenkins.sock"
		log.Printf("Set default socket file: %s", a.Yaml.Runtime.Service.Socket)
	}

	a.templateDefaultPath, _ = utils.Abs(templateDirDefault)
}

// This file has been created by "go generate" as initial code. go generate will never update it, EXCEPT if you remove it.
package main

import (
	"forjj-jenkins/reportlogs"
	"log"
	"net/http"

	"github.com/forj-oss/forjj-modules/trace"

	"github.com/forj-oss/goforjj"
)

// DoCreate Do creating plugin task
// req_data contains the request data posted by forjj. Structure generated from 'jenkins.yaml'.
// ret_data contains the response structure to return back to forjj.
//
func DoCreate(r *http.Request, req *CreateReq, ret *goforjj.PluginData) (httpCode int) {
	p, code := req.check_source_existence(ret)
	if p == nil {
		return code
	}

	reportlogs.SetLogsFunc(map[string]func(string, ...interface{}){
		"reportLog": func(format string, parameters ...interface{}) {
			log.Printf(ret.StatusAdd(format, parameters...))
		},
		"reportError": func(format string, parameters ...interface{}) {
			log.Printf(ret.Errorf(format, parameters...))
		},
		"pluginLog": func(format string, parameters ...interface{}) {
			log.Printf(format, parameters...)
		},
	})

	if p.initialize_from(req, ret) != nil {
		return
	}

	if p.create_jenkins_sources(ret) != nil {
		return
	}

	if p.saveYaml(goforjj.FilesSource, jenkins_file, &p.yamlPlugin, ret, nil) != nil {
		return
	}

	if p.saveYaml(goforjj.FilesDeploy, jenkinsDeployFile, &p.yaml, ret, nil) != nil {
		return
	}

	p.saveRunYaml(ret, nil)

	ret.CommitMessage = "Creating initial jenkins source files as defined by the Forjfile."

	return
}

// DoUpdate is the update plugin task
// req_data contains the request data posted by forjj. Structure generated from 'jenkins.yaml'.
// ret_data contains the response structure to return back to forjj.
// forjj-jenkins.yaml is loaded by default.
//
func DoUpdate(r *http.Request, req *UpdateReq, ret *goforjj.PluginData) (_ int) {
	p, ok := req.checkSourceExistence(ret)
	if !ok {
		return
	}

	reportlogs.SetLogsFunc(map[string]func(string, ...interface{}){
		"reportLog": func(format string, parameters ...interface{}) {
			log.Printf(ret.StatusAdd(format, parameters...))
		},
		"reportError": func(format string, parameters ...interface{}) {
			gotrace.Error(ret.Errorf(format, parameters...))
		},
		"pluginLog": func(format string, parameters ...interface{}) {
			log.Printf(format, parameters...)
		},
	})

	if !p.loadYaml(goforjj.FilesSource, jenkins_file, &p.yamlPlugin, ret, false) {
		return
	}
	if !p.loadYaml(goforjj.FilesDeploy, jenkinsDeployFile, &p.yaml, ret, true) {
		return
	}

	// TODO: Use the GithubStruct.UpdateFrom(...)
	var updated bool
	if p.update_from(req, ret, &updated) != nil {
		return
	}
	if p.update_projects(req, ret, &updated) != nil {
		return
	}
	if p.update_jenkins_sources(ret, &updated) != nil {
		return
	}

	if p.runBuildDeploy(req, &updated) != nil {
		return
	}

	if p.addBuiltFiles(ret, &updated) != nil {
		return
	}
	if p.saveYaml(goforjj.FilesSource, jenkins_file, &p.yamlPlugin, ret, &updated) != nil {
		return
	}
	if p.saveYaml(goforjj.FilesDeploy, jenkinsDeployFile, &p.yaml, ret, &updated) != nil {
		return
	}
	if p.saveRunYaml(ret, &updated) != nil {
		return
	}

	if updated {
		ret.CommitMessage = "Updating jenkins source files requested by Forjfile."
	} else {
		log.Print(ret.StatusAdd("No update detected. Jenkins source files hasn't been updated."))
	}
	return
}

// DoMaintain Do maintaining plugin task
// req_data contains the request data posted by forjj. Structure generated from 'jenkins.yaml'.
// ret_data contains the response structure to return back to forjj.
//
func DoMaintain(r *http.Request, req *MaintainReq, ret *goforjj.PluginData) (httpCode int) {
	if !req.checkSourceExistence(ret) {
		return
	}

	reportlogs.SetLogsFunc(map[string]func(string, ...interface{}){
		"reportLog": func(format string, parameters ...interface{}) {
			log.Printf(ret.StatusAdd(format, parameters...))
		},
		"reportError": func(format string, parameters ...interface{}) {
			log.Printf(ret.Errorf(format, parameters...))
		},
		"pluginLog": func(format string, parameters ...interface{}) {
			log.Printf(format, parameters...)
		},
	})

	// loop on list of jenkins instances defined by a collection of */jenkins.yaml
	if !req.Instantiate(req, ret) {
		return
	}
	return
}

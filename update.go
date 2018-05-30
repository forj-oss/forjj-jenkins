// This file has been created by "go generate" as initial code. go generate will never update it, EXCEPT if you remove it.

// So, update it for your need.
package main

import (
	"strings"
	"log"
	"os"
	"path"

	"github.com/forj-oss/goforjj"
)

// Return ok if the jenkins instance sources and template sources exists
func (r *UpdateReq) checkSourceExistence(ret *goforjj.PluginData) (p *JenkinsPlugin, status bool) {
	log.Print("Checking Jenkins source code existence.")
	if _, err := os.Stat(r.Forj.ForjjSourceMount); err != nil {
		ret.Errorf("Unable to update jenkins instances. '%s' is inexistent or innacessible. %s", r.Forj.ForjjSourceMount, err)
		return
	}

	srcPath := path.Join(r.Forj.ForjjSourceMount, r.Forj.ForjjInstanceName)

	log.Print("Checking Jenkins deploy path.")
	if _, err := os.Stat(path.Join(r.Forj.ForjjDeployMount, r.Forj.ForjjDeploymentEnv)); err != nil {
		ret.Errorf("Unable to update jenkins instances. '%s'/'%s' is inexistent or innacessible. %s", r.Forj.ForjjDeployMount, r.Forj.ForjjDeploymentEnv, err)
		return
	}
	deployPath := path.Join(r.Forj.ForjjDeployMount, r.Forj.ForjjDeploymentEnv, r.Forj.ForjjInstanceName)

	p = newPlugin(srcPath, deployPath)

	ret.StatusAdd("environment checked.")
	return p, true
}

func (r *JenkinsPlugin) update_jenkins_sources(ret *goforjj.PluginData, updated *bool) (err error) {
	if err = r.DefineSources(); err != nil {
		log.Printf(ret.Errorf("%s", err))
		return
	}

	log.Print("Start copying source files...")
	if err = r.copy_source_files(ret, updated); err != nil {
		return
	}

	log.Print("Start Generating source files...")
	if err = r.generate_source_files(ret, updated); err != nil {
		return
	}

	if err = r.generate_jobsdsl(ret, updated); err != nil {
		return
	}

	return
}

func IsUpdated(updated *bool) {
	if updated != nil {
		*updated = true
	}
}

// Function which adds maintain options as part of the plugin answer in create/update phase.
// forjj won't add any driver name because 'maintain' phase read the list of drivers to use from forjj-maintain.yml
// So --git-us is not available for forjj maintain.
func (r *UpdateArgReq) SaveMaintainOptions(ret *goforjj.PluginData) {
	if ret.Options == nil {
		ret.Options = make(map[string]goforjj.PluginOption)
	}
}

func addMaintainOptionValue(options map[string]goforjj.PluginOption, option, value, defaultv, help string) goforjj.PluginOption {
	opt, ok := options[option]
	if ok && value != "" {
		opt.Value = value
		return opt
	}
	if !ok {
		opt = goforjj.PluginOption{Help: help}
		if value == "" {
			opt.Value = defaultv
		} else {
			opt.Value = value
		}
	}
	return opt
}

// update_projects add project data in the jenkins.yaml file
func (jp *JenkinsPlugin) update_projects(req *UpdateReq, ret *goforjj.PluginData, status *bool) error {
	projects := ProjectInfo{}
	projects.set_project_info(req.Forj.ForjCommonStruct)
	instanceData := req.Objects.App[req.Forj.ForjjInstanceName]
	projects.set_infra_remote(instanceData.SeedJobRepo)
	projects.setIsProDeploy(strings.ToLower(instanceData.ProDeployment) == "true")

	return projects.set_projects_to(req.Objects.Projects, jp, ret, status, req.Forj.ForjjInfra)
}

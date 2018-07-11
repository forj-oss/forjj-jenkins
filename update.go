// This file has been created by "go generate" as initial code. go generate will never update it, EXCEPT if you remove it.

// So, update it for your need.
package main

import (
	log "forjj-jenkins/reportlogs"
	"strings"

	"github.com/forj-oss/goforjj"
)

func (r *JenkinsPlugin) update_jenkins_sources(ret *goforjj.PluginData, updated *bool) (err error) {
	if err = r.DefineSources(); err != nil {
		log.Printf(ret.Errorf("%s", err))
		return
	}

	log.Printf("Start copying source files...")
	if err = r.copy_source_files(ret, updated); err != nil {
		return
	}

	log.Printf("Start Generating source files...")
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
	projects.setDslInfo(instanceData.SeedJobStruct)
	// TODO: Information not used. To clean it up.
	projects.setIsProDeploy(strings.ToLower(instanceData.ProDeployment) == "true")

	return projects.set_projects_to(req.Objects.Projects, jp, ret, status, req.Forj.ForjjInfra)
}

func (jp *JenkinsPlugin) runBuildDeploy(creds map[string]string) (err error) {
	run, found := jp.templates_def.Build[jp.yaml.Deploy.Deployment.To]
	if !found {
		log.Printf("No run_build section defined for deploy-to=%s. No build processed. If you need one, create run_build/%s: in templates.yaml", jp.deployEnv, jp.deployEnv)
		return
	}

	model := jp.Model()
	model.loadCreds(jp.InstanceName, creds)

	if err = run.run(jp.InstanceName, jp.deployPath, model, jp.auths); err != nil {
		log.Errorf("Unable to build to %s. %s", jp.yaml.Deploy.Deployment.To, err)
		return
	}
	return
}

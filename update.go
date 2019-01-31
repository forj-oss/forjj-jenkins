// This file has been created by "go generate" as initial code. go generate will never update it, EXCEPT if you remove it.

// So, update it for your need.
package main

import (
	"fmt"
	log "forjj-jenkins/reportlogs"
	"strings"

	"github.com/forj-oss/goforjj"
)

func (jp *JenkinsPlugin) update_jenkins_sources(ret *goforjj.PluginData, updated *bool) (err error) {
	if err = jp.DefineSources(); err != nil {
		log.Printf(ret.Errorf("%s", err))
		return
	}

	log.Printf("Start copying source files...")
	if err = jp.copy_source_files(ret, updated); err != nil {
		return
	}

	log.Printf("Start Generating source files...")
	if err = jp.generate_source_files(ret, updated); err != nil {
		return
	}

	if err = jp.generate_jobsdsl(ret, updated); err != nil {
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

	return projects.set_projects_to(req.Objects.Projects, jp, ret, status, req.Forj.ForjjInfra, instanceData.JenkinsfilePath)
}

func (jp *JenkinsPlugin) runBuildDeploy(username string, creds map[string]string, createSteps bool) (err error) {
	deployTo := jp.yaml.Deploy.Deployment.To
	run, found := jp.templates_def.Build[deployTo]
	if !found {
		log.Printf("No run_build section defined for deploy-to=%s. No build processed. If you need one, create run_build/%s: in templates.yaml", jp.deployEnv, jp.deployEnv)
		return
	}

	model := jp.Model()
	model.loadCreds(username, jp.InstanceName, creds)

	var runNormalSteps, runFailureSteps []string
	if createSteps {
		runNormalSteps = run.Steps.WhenCreate
		runFailureSteps = run.Steps.WhenCreateFailed
	} else {
		runNormalSteps = run.Steps.WhenUpdate
		runFailureSteps = run.Steps.WhenUpdateFailed
	}

	if jp.runTasks != nil && len(jp.runTasks) > 0 {
		runNormalSteps = jp.runTasks
		runFailureSteps = []string{}
	}

	if runNormalSteps == nil || len(runNormalSteps) == 0 {
		if run.RunCommand == "" {
			log.Printf("yaml:/run_build/%s/run is depreciated. Use yaml:/run_build/%s/steps and yaml:/run_build/%s/tasks", deployTo, deployTo, deployTo)
		}
		if err = run.run(jp.InstanceName, jp.source_path, jp.deployPath, model, jp.auths); err != nil {
			log.Errorf("Unable to build to %s. %s", jp.yaml.Deploy.Deployment.To, err)
		}
		return
	}

	defer func() {
		if err == nil {
			return
		}

		jp.runSteps(deployTo, runFailureSteps, run.Tasks, model)
	}()

	err = jp.runSteps(deployTo, runNormalSteps, run.Tasks, model)

	return
}

func (jp *JenkinsPlugin) runSteps(deployTo string, steps []string, tasks map[string]RunStruct, model *JenkinsPluginModel) (err error) {
	tasksList := "None"
	if len(tasks) > 0 {
		for name, task := range tasks {
			tasksList += "- " + name
			if task.Description != "" {
				tasksList += " - " + task.Description
			} else {
				tasksList += " - (description empty)"
			}
			tasksList += "\n"
		}
	}
	for _, stepName := range steps {
		step, found := tasks[stepName]
		if !found {
			err = fmt.Errorf("Cannot run '%s' Step. '%s' was not defined in the `templates.yaml` under `run_build/%s`. Possible tasks are:\n%s",
				stepName, stepName, deployTo, tasksList)
			return
		}

		if err = step.run(jp.InstanceName, jp.source_path, jp.deployPath, model, jp.auths); err != nil {
			err = fmt.Errorf("Unable to build %s to %s. %s", stepName, jp.yaml.Deploy.Deployment.To, err)
			return
		}
	}
	return
}

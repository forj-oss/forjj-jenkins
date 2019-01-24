package main

import (
	"fmt"
	log "forjj-jenkins/reportlogs"
	"os"
	"path"

	"github.com/forj-oss/goforjj"
)

// Return ok if the jenkins instance exist
func (r *MaintainReq) checkSourceExistence(ret *goforjj.PluginData) (status bool) {
	log.Printf("Checking Jenkins source code path existence.")

	srcPath := path.Join(r.Forj.ForjjDeployMount, r.Forj.ForjjDeploymentEnv, r.Forj.ForjjInstanceName)
	if _, err := os.Stat(path.Join(srcPath, maintain_cmds_file)); err != nil {
		log.Printf(ret.Errorf("Unable to maintain instance name '%s' without deploy code.\n"+
			"Use update to update it, commit, push and retry. %s.", srcPath, err))
		return
	}

	ret.StatusAdd("environment checked.")
	status = true
	return
}

// TODO: Need to define where to deploy (dev/itg/pro/local/other) - Is it still needed?

// Instantiate Instance given by the request.
func (r *MaintainReq) Instantiate(req *MaintainReq, ret *goforjj.PluginData) (_ bool) {
	instance := r.Forj.ForjjInstanceName
	mount := r.Forj.ForjjDeployMount
	src := path.Join(mount, r.Forj.ForjjDeploymentEnv, instance)
	if _, err := os.Stat(path.Join(src, maintain_cmds_file)); err != nil {
		log.Reportf("'%s' is not a forjj plugin source code model. No '%s' found. ignored.", src, jenkins_file)
		return true
	}
	p := newPlugin("", mount)
	p.setEnv(req.Forj.ForjjDeploymentEnv, req.Forj.ForjjInstanceName)

	p.auths = NewDockerAuths(r.Objects.App[instance].RegistryAuth)

	// Load deploy configuration
	if !p.loadYaml(goforjj.FilesDeploy, jenkinsDeployFile, &p.yaml, ret, true) {
		return
	}

	// Load templates.yml to get the list of deployment commands.
	if !p.loadRunYaml(ret) {
		return
	}

	if !p.GetMaintainData(req, ret) {
		return
	}
	log.Reportf("Maintaining '%s'", p.InstanceName)
	if err := os.Chdir(src); err != nil {
		log.Errorf("Unable to enter in '%s'. %s", src, err)
		return
	}
	// start a command as described by the source code.

	if p.run.Steps.WhenDeploy == nil || len(p.run.Steps.WhenDeploy) == 0 {
		if p.run.RunCommand == "" {
			log.Printf("yaml:/run is depreciated. Use yaml:/steps and yaml:/tasks")
		}
		if err := p.run.run(instance, p.source_path, p.deployPath, p.Model(), p.auths); err != nil {
			log.Errorf("Unable to instantiate to %s. %s", p.yaml.Deploy.Deployment.To, err)
			return
		}
	}
	var err error

	defer func(){
		if err == nil {
			return
		}
		if err = r.runSteps(p.run.Steps.WhenDeployFailed, p.run.Tasks, p.Model(), p) ; err != nil {
			log.Errorf("%s", err)
			return
		}
	}()

	if err = r.runSteps(p.run.Steps.WhenDeploy, p.run.Tasks, p.Model(), p) ; err != nil {
		log.Errorf("%s", err)
		return
	}

	return true
}

func (r *MaintainReq) runSteps(steps []string, tasks map[string]RunStruct, model *JenkinsPluginModel, jp *JenkinsPlugin) (err error) {
	instance := r.Forj.ForjjInstanceName
	for _, stepName := range steps {
		step, found := jp.run.Tasks[stepName]
		if !found {
			err = fmt.Errorf("Cannot run '%s' Step. '%s' from %s unfound", stepName, stepName, "run_deploy")
			return
		}

		if err = step.run(instance, jp.source_path, jp.deployPath, jp.Model(), jp.auths); err != nil {
			err = fmt.Errorf("Unable to build %s to %s. %s", stepName, jp.yaml.Deploy.Deployment.To, err)
			return
		}
	}
	return
}

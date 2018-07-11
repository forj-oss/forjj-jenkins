package main

import (
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
	auths := NewDockerAuths(r.Objects.App[instance].RegistryAuth)

	src := path.Join(mount, r.Forj.ForjjDeploymentEnv, instance)
	if _, err := os.Stat(path.Join(src, maintain_cmds_file)); err != nil {
		log.Reportf("'%s' is not a forjj plugin source code model. No '%s' found. ignored.", src, jenkins_file)
		return true
	}
	p := newPlugin("", mount)
	p.setEnv(req.Forj.ForjjDeploymentEnv, req.Forj.ForjjInstanceName)

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
	if err := p.run.run(instance, p.deployPath, p.Model(), auths) ; err != nil {
		log.Errorf("Unable to instantiate to %s.", p.yaml.Deploy.Deployment.To, err)
		return
	}
	return true
}


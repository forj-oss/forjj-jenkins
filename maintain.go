package main

import (
	"log"
	"os"
	"path"

	"github.com/forj-oss/goforjj"
)

// Return ok if the jenkins instance exist
func (r *MaintainReq) checkSourceExistence(ret *goforjj.PluginData) (status bool) {
	log.Print("Checking Jenkins source code path existence.")

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
		log.Printf("'%s' is not a forjj plugin source code model. No '%s' found. ignored.", src, jenkins_file)
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
	ret.StatusAdd("Maintaining '%s'", p.InstanceName)
	if err := os.Chdir(src); err != nil {
		ret.Errorf("Unable to enter in '%s'. %s", src, err)
		return
	}
	if !p.instantiateInstance(instance, auths, ret) {
		return
	}
	return true
}

func (p *JenkinsPlugin) instantiateInstance(instance string, auths *DockerAuths, ret *goforjj.PluginData) (status bool) {
	// start a command as described by the source code.
	if p.run.RunCommand == "" {
		log.Printf(ret.Errorf("Unable to instantiate to %s. Deploy Command is empty.", p.yaml.Deploy.Deployment.To))
		return
	}

	for server := range auths.Auths {
		if err := auths.authenticate(server); err != nil {
			log.Printf(ret.Errorf("Unable to instantiate. %s", err))
			return
		}
	}

	log.Printf(ret.StatusAdd("Running '%s'", p.run.RunCommand))

	env := os.Environ()
	if v := os.Getenv("DOOD_SRC"); v != "" {
		env = append(env, "SRC="+path.Join(v, instance)+"/")
		log.Printf("DOOD_SRC detected. Env added : 'SRC' = '%s'", path.Join(v, instance)+"/")
	}
	if v := os.Getenv("DOOD_DEPLOY"); v != "" {
		deployPath := v
		env = append(env, "DEPLOY="+deployPath)
		log.Printf("DOOD_DEPLOY detected. Env added : 'DEPLOY' = '%s'", deployPath)
	}

	model := p.Model()
	for key, env_to_set := range p.run.Env {
		if env_to_set.If != "" {
			// check if If evaluation return something or not. if not, the environment key won't be created.
			if v, err := Evaluate(env_to_set.If, model); err != nil {
				ret.Errorf("Deployment '%s'. Error in evaluating '%s'. %s", p.yaml.Deploy.Deployment.To, key, err)
			} else {
				if v == "" {
					continue
				}
			}
		}
		if v, err := Evaluate(env_to_set.Value, model); err != nil {
			ret.Errorf("Deployment '%s'. Error in evaluating '%s'. %s", p.yaml.Deploy.Deployment.To, key, err)
		} else {
			env = append(env, key+"="+v)
			log.Printf("Env added : '%s' = '%s'", key, v)
		}
	}

	err := runFlowCmd("/bin/sh", env,
		func(line string) {
			log.Printf(ret.StatusAdd(line))
		},
		func(line string) {
			log.Printf(ret.StatusAdd(line))

		}, "-c", p.run.RunCommand)
	if err != nil {
		curDir, _ := os.Getwd()
		log.Printf(ret.Errorf("%s (pwd: %s)", err, curDir))
	}

	return true
}

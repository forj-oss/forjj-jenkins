package main

import (
	"log"
	"os"
	"path"

	"github.com/forj-oss/goforjj"
)

// Return ok if the jenkins instance exist
func (r *MaintainReq) check_source_existence(ret *goforjj.PluginData) (status bool) {
	log.Print("Checking Jenkins source code path existence.")

	src_path := path.Join(r.Forj.ForjjDeployMount, r.Forj.ForjjDeploymentEnv, r.Forj.ForjjInstanceName)
	if _, err := os.Stat(path.Join(src_path, maintain_cmds_file)); err != nil {
		log.Printf(ret.Errorf("Unable to maintain instance name '%s' without deploy code.\n"+
			"Use update to update it, commit, push and retry. %s.", src_path, err))
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
	p := newPlugin("", src)

	p.setEnv(req.Forj.ForjjDeploymentEnv, req.Forj.ForjjInstanceName)

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

	s, err := run_cmd("/bin/sh", env, "-c", p.run.RunCommand)
	log.Printf(ret.StatusAdd(string(s)))
	if err != nil {
		curDir, _ := os.Getwd()
		log.Printf(ret.Errorf("%s (pwd: %s)", err, curDir))
	}

	return true
}

package main

import (
	"fmt"
	log "forjj-jenkins/reportlogs"
	"os"
	"path"
)

type RunStruct struct {
	RunCommand string               `yaml:"run"`
	Env        map[string]EnvStruct `yaml:"environment"`
	Files      RunFilesStruct
}

// run execute the defined command with the context model and docker authentication given
//
// Use reportLog or Errorf to report to the end users.
// But errors detected is not interrupting the program immediatelly in order to report all issues to
// fix
func (r RunStruct) run(instance, deployPath string, model *JenkinsPluginModel, auths *DockerAuths) (err error) {
	// start a command as described by the source code.

	// Errors detected if true.
	log.Printf("Checking command to run and context")

	if r.RunCommand == "" {
		log.Errorf("Unable to run a command: Command is empty")
	}

	for server := range auths.Auths {
		if err = auths.authenticate(server); err != nil {
			return
		}
	}

	env := os.Environ()
	if v := os.Getenv("DOOD_SRC"); v != "" {
		log.Printf("DOOD detected.")
		srcPath := path.Join(v, instance) + "/"
		env = append(env, "SRC="+srcPath)
		log.Printf("Env added : 'SRC' = '%s'", srcPath)
	} else {
		env = append(env, "SRC=/deploy/")
	}

	if v := os.Getenv("DOOD_DEPLOY"); v != "" {
		deployPath := path.Join(v, instance) + "/"
		env = append(env, "DEPLOY="+deployPath)
		log.Printf("Env added : 'DEPLOY' = '%s'", deployPath)
	} else {
		env = append(env, "DEPLOY="+deployPath)
		log.Printf("Env added : 'DEPLOY' = '%s'", deployPath)
	}
	if v := os.Getenv("DOCKER_DOOD"); v != "" {
		env = append(env, "DOCKER_DOOD="+v)
		log.Printf("Env added : 'DOCKER_DOOD' = '%s'", v)
	}

	if v := os.Getenv("DOCKER_DOOD_BECOME"); v != "" {
		env = append(env, "DOCKER_DOOD_BECOME="+v)
		log.Printf("Env added : 'DOCKER_DOOD_BECOME' = '%s'", v)
	}

	for key, envToSet := range r.Env {
		if envToSet.If != "" {
			// check if If evaluation return something or not. if not, the environment key won't be created.
			if v, err := Evaluate(envToSet.If, model); err != nil {
				log.Errorf("Error in evaluating '%s'. %s", key, err)
			} else {
				if v == "" {
					continue
				}
			}
		}
		if v, err := Evaluate(envToSet.Value, model); err != nil {
			log.Errorf("Error in evaluating '%s'. %s", key, err)
		} else {
			env = append(env, key+"="+v)
			log.Printf("Env added : '%s' = '%s'", key, v)
		}
	}

	if err = r.Files.createFiles(model, deployPath); err != nil {
		return
	}

	if log.HasReportedErrors() {
		return fmt.Errorf("Unable to run command. Errors detected")
	}

	if err = os.Chdir(deployPath); err != nil {
		return fmt.Errorf("Unable to move to '%s'. %s", deployPath, err)
	}

	log.Reportf("Running '%s' from '%s'", r.RunCommand, deployPath)

	err = runFlowCmd("/bin/sh", env,
		func(line string) {
			log.Reportf(line)
		},
		func(line string) {
			log.Reportf(line)

		}, "-c", r.RunCommand)

	r.Files.deleteFiles()

	return
}

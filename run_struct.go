package main

import (
	"fmt"
	"path"
	"os"
	log "forjj-jenkins/reportlogs"
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
		env = append(env, "SRC="+path.Join(v, instance)+"/")
		log.Printf("DOOD_SRC detected. Env added : 'SRC' = '%s'", path.Join(v, instance)+"/")
	}
	if v := os.Getenv("DOOD_DEPLOY"); v != "" {
		deployPath := v
		env = append(env, "DEPLOY="+deployPath)
		log.Printf("DOOD_DEPLOY detected. Env added : 'DEPLOY' = '%s'", deployPath)
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

	log.Reportf("Running '%s'", r.RunCommand)

	err = runFlowCmd("/bin/sh", env,
		func(line string) {
			log.Reportf(line)
		},
		func(line string) {
			log.Reportf(line)

		}, "-c", r.RunCommand)

	r.Files.deleteFiles()

	if err != nil {
		curDir, _ := os.Getwd()
		return fmt.Errorf("%s (pwd: %s)", err, curDir)
	}
	return nil
}

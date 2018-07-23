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
func (r RunStruct) run(instance, sourcePath, deployPath string, model *JenkinsPluginModel, auths *DockerAuths) (err error) {
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

	env := r.defineEnv(instance, sourcePath, deployPath, model)

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
			log.Reportf("%s", line)
		},
		func(line string) {
			log.Reportf("%s", line)

		}, "-c", r.RunCommand)

	r.Files.deleteFiles()

	return
}

type envMapFunc struct {
	key  string
	more func(par, value string) (string, []string, bool)
}

type envMapFuncs []envMapFunc

func (r RunStruct) setEnv(parameters envMapFuncs) (env []string) {
	env = []string{}
	for _, envFunc := range parameters {
		v := os.Getenv(envFunc.key)
		value, envToAdd, ignore := envFunc.more(envFunc.key, v)
		if envToAdd != nil {
			env = append(env, envToAdd...)
			for _, theEnv := range envToAdd {
				log.Printf("Env added : '%s'", theEnv)
			}
		}
		if ignore || (v == "" && value == "") {
			continue
		}
		if value != "" {
			v = value
		}
		env = append(env, envFunc.key+"="+v)
		log.Printf("Env added : '%s' = '%s'", envFunc.key, v)
	}
	return
}

func (r RunStruct) noEnvFunc(_, _ string) (_ string, _ []string, _ bool) {
	return
}

func (r RunStruct) defineEnv(instance, sourcePath, deployPath string, model *JenkinsPluginModel) (env []string) {
	dood := false
	doodBecome := false

	env = r.setEnv([]envMapFunc{
		envMapFunc{"DOCKER_DOOD", func(par, value string) (ret string, _ []string, ignore bool) {
			if value != "" {
				log.Printf("DOOD detected.")
				dood = true
				return
			}
			ignore = true
			return
		},
		},
		envMapFunc{"DOOD_SRC", func(par, value string) (ret string, env []string, ignore bool) {
			if value == "" || !dood {
				env = []string{"SRC=" + sourcePath}
				ignore = !dood
				return
			}
			ret = path.Join(value, instance) + "/"
			env = []string{"SRC=" + ret}
			return
		},
		},
		envMapFunc{"DOOD_DEPLOY", func(par, value string) (ret string, env []string, ignore bool) {
			if value == "" || !dood {
				env = []string{"DEPLOY=" + deployPath}
				ignore = !dood
				return
			}
			ret = path.Join(value, instance) + "/"
			env = []string{"DEPLOY=" + ret}
			return
		},
		},
		envMapFunc{"DOCKER_DOOD_BECOME", func(par, value string) (ret string, _ []string, ignore bool) {
			if value != "" {
				log.Printf("DOOD_BECOME detected.")
				doodBecome = true
				return
			}
			ignore = true
			return
		},
		},
		envMapFunc{"GID", func(par, value string) (ret string, _ []string, ignore bool) {
			ignore = !doodBecome
			return
		},
		},
		envMapFunc{"UID", func(par, value string) (ret string, _ []string, ignore bool) {
			ignore = !doodBecome
			return
		},
		},
		envMapFunc{"LOGNAME", r.noEnvFunc},
		envMapFunc{"PATH", r.noEnvFunc},
		envMapFunc{"TERM", r.noEnvFunc},
		envMapFunc{"HOSTNAME", r.noEnvFunc},
	})

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
	return
}

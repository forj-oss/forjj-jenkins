package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/forj-oss/goforjj"
)

// RunFilesStruct List of files
type RunFilesStruct map[string]RunFileStruct

// RunFileStruct detailled a File content and actions
type RunFileStruct struct {
	RemoveWhenDone bool   `yaml:"remove-when-done"`
	Value          string `yaml:"content"`
	If             string
}

// createFiles loop on files to create them if needed.
func (fs RunFilesStruct) createFiles(model *JenkinsPluginModel, deployPath string, ret *goforjj.PluginData) error {
	for key, env_to_set := range fs {
		if env_to_set.If != "" {
			// check if If evaluation return something or not. if not, the environment key won't be created.
			if v, err := Evaluate(env_to_set.If, model); err != nil {
				ret.Errorf("Error in evaluating '%s'. %s", key, err)
			} else {
				if v == "" {
					continue
				}
			}
		}
		if v, err := Evaluate(env_to_set.Value, model); err != nil {
			ret.Errorf("Error in evaluating '%s'. %s", key, err)
		} else {
			if err := env_to_set.createFile(deployPath, key, v); err != nil {
				return fmt.Errorf(ret.Errorf("%s", err))
			}
			fd, err := os.Create(key)
			if err != nil {
				return fmt.Errorf("Unable to create %s. %s", key, err)
			}

			fd.WriteString(v)
		}
	}
	return nil
}

// createFiles loop on files to create them if needed.
func (fs RunFilesStruct) deleteFiles() error {
	for name, fileState := range fs {
		if fileState.RemoveWhenDone {
			// remove the file created.
			os.Remove(name)
		}
	}
	return nil
}

func (f RunFileStruct) createFile(deployPath, name, value string) error {
	fd, err := os.Create(path.Join(deployPath, name))
	if err != nil {
		return fmt.Errorf("Unable to create %s. %s", name, err)
	}
	defer fd.Close()

	fd.WriteString(value)
	log.Printf("%s created", name)
	return nil
}

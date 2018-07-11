package main

import (
	log "forjj-jenkins/reportlogs"
	"os"
)

// RunFilesStruct List of files
type RunFilesStruct map[string]RunFileStruct

// createFiles loop on files to create them if needed.
func (fs RunFilesStruct) createFiles(model *JenkinsPluginModel, deployPath string) (err error) {
	for key, envToSet := range fs {
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
			if err := envToSet.createFile(deployPath, key, v); err != nil {
				log.Errorf("%s", err)
				continue
			}
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

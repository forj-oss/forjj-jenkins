package main

import (
	"fmt"
	log "forjj-jenkins/reportlogs"
	"os"
	"path"
)

// RunFileStruct detailled a File content and actions
type RunFileStruct struct {
	RemoveWhenDone bool   `yaml:"remove-when-done"`
	Value          string `yaml:"content"`
	If             string
	CreateSubDir   bool `yaml:"create-subdir"`
}

func (f RunFileStruct) createFile(deployPath, name, value string) error {
	filePath := path.Dir(path.Join(deployPath, name))

	if _, err := os.Stat(filePath); err != nil && os.IsNotExist(err) && f.CreateSubDir {
		if err = os.MkdirAll(filePath, 0755); err != nil {
			return fmt.Errorf("Unable to create '%s'. %s", filePath, err)
		} else {
			log.Printf("Path '%s' created", filePath)
		}
	}

	fd, err := os.Create(path.Join(deployPath, name))
	if err != nil {
		return fmt.Errorf("Unable to create %s. %s", name, err)
	}
	defer fd.Close()

	fd.WriteString(value)
	log.Printf("%s created", name)
	return nil
}

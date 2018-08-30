package main

type Project struct {
	Name            string
	SourceType      string
	Github          GithubStruct `yaml:",omitempty"`
	Git             GitStruct    `yaml:",omitempty"`
	InfraRepo       bool         `yaml:",omitempty"`
	Role            string       `yaml:"role,omitempty"`
	JenkinsfilePath string       `yaml:"jenkinsfile-path,omitempty"`
	all             *Projects
}

func (p *Project) Remove() bool {
	return true
}

func (p *Project) Model(jp *JenkinsPlugin) (ret *ProjectModel) {
	ret = new(ProjectModel)
	ret.Project = *p
	ret.Source = jp.yaml
	return
}

func (p *Project) Add() error {
	return nil
}

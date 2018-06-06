package main

// ProjectModel define structure used by projects.Generate to generate <jobs>.groovy in <jenkins>/jobs-dsl/
type ProjectModel struct {
	Project Project
	Source  YamlJenkins
}
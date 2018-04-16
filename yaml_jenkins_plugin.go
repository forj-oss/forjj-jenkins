package main

// Used for the jenkins yaml source and generate template data.
type YamlJenkinsPlugin struct {
	TemplatePath string `yaml:"template-path,omitempty"`
	JenkinsPluginVersion string `yaml:"version"`
}
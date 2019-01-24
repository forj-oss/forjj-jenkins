package main

// BuildStepsStruct defines yaml structure under yaml:/run_build/<deployTo>/steps
//
type BuildStepsStruct struct {
	WhenCreate       []string `yaml:"when-create,omitempty"`
	WhenCreateFailed []string `yaml:"when-create-failed,omitempty"`

	WhenUpdate       []string `yaml:"when-update,omitempty"`
	WhenUpdateFailed []string `yaml:"when-update-failed,omitempty"`

}

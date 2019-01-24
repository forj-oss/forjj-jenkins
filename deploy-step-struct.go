package main

// DeployStepsStruct defines yaml structure under templates.yaml-yaml:/run_deploy/<deployTo>/steps
// or forjj-deploy.yaml--yaml:/steps
//
type DeployStepsStruct struct {
	WhenDeploy       []string `yaml:"when-deploy,omitempty"`
	WhenDeployFailed []string `yaml:"when-update-failed,omitempty"`
}

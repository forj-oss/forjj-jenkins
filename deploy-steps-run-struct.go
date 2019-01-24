package main

// DeployStepsRunStruct define a collection to steps to execute
// We can define some Tasks command which won't be executed during a build (create/update) or a deploy (maintain).
// RunCommand is still supported but ignored if steps are define
// Env and Files can be defined as global run use case
type DeployStepsRunStruct struct {
	Steps     DeployStepsStruct    `yaml:",omitempty"`
	Tasks     map[string]RunStruct `yaml:",omitempty"`
	RunStruct                      // Kept for compatibility. if Steps/Tasks are defined, RunStruct will be ignored.
	// This struct is ignored from `forjj update` side as well.
}

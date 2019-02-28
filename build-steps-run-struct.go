package main

import "fmt"

// BuildStepsRunStruct define a collection to steps to execute
// We can define some Tasks command which won't be executed during a build (create/update) or a deploy (maintain).
// RunCommand is still supported but ignored if steps are define
// Env and Files can be defined as global run use case
type BuildStepsRunStruct struct {
	Steps     *BuildStepsStruct     `yaml:",omitempty"`
	Tasks     map[string]RunStruct `yaml:",omitempty"`
	RunStruct `yaml:",inline"`     // Kept for compatibility. if Steps/Tasks are defined, RunStruct will be ignored.
	// This struct is ignored from `forjj update` side as well.
}

// Validate tasks/steps defined.
func (bsr *BuildStepsRunStruct) Validate() (err error) {
	if bsr.Steps == nil && bsr.Tasks == nil {
		return
	}

	// check if steps defines valid tasks
	analyze := newBuildSteps()
	analyze.add("when-create", bsr.Steps.WhenCreate)
	analyze.add("when-create-failed", bsr.Steps.WhenCreateFailed)
	analyze.add("when-update", bsr.Steps.WhenUpdate)
	analyze.add("when-update-failed", bsr.Steps.WhenUpdateFailed)

	err = analyze.foreach(func(event, task string) (err error) {
		if _, found := bsr.Tasks[task]; !found {
			err = fmt.Errorf("On %s: task '%s' is not defined in the list of tasks", event, task)
		}
		return
	})

	// check tasks dependencies
	loopDetect := newLoopDetect(func(element string) (elements []string) {
		run, found := bsr.Tasks[element]
		if !found {
			return
		}
		return run.DependsOn
	})
	for task:= range bsr.Tasks {
		if err = loopDetect.check(task) ; err != nil {
			return fmt.Errorf("In %s", err)
		}
	}
	return
}

// GetTask return the task name
func (bsr *BuildStepsRunStruct) GetTask(name string) (task RunStruct, found bool) {
	task, found = bsr.Tasks[name]
	return
}
package main

type buildSteps map[string][]string

func newBuildSteps() (ret buildSteps) {
	ret = make(buildSteps)
	return
}

func (bs buildSteps) add(event string, tasks []string) {
	bs[event] = tasks
}

func (bs buildSteps) foreach(validate func(string, string) error) (err error) {
	for event, tasks := range bs {
		for _, task := range tasks {
			if err = validate(event, task) ; err != nil {
				return err
			}
		}
	}
	return
}
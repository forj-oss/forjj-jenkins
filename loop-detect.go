package main

import "fmt"

type loopDetect struct {
	list    map[string]bool
	subList func(element string) (elments []string)
}

func newLoopDetect(subList func(element string) (elments []string)) (ret *loopDetect) {
	ret = new(loopDetect)
	ret.list = make(map[string]bool)
	ret.subList = subList
	return
}

func (ld *loopDetect) check(element string) (_ error) {
	if ld == nil {
		return
	}

	if _, found := ld.list[element]; !found {
		ld.list[element] = true
	} else {
		return fmt.Errorf("%s - loop detected. Please fix", element)
	}

	for _, value := range ld.subList(element) {
		if err := ld.check(value) ; err != nil {
			return fmt.Errorf("%s/%s", element, value)
		}
	}
	return
}

func (ld *loopDetect) run(element string, runTask func(string) error) (_ error) {
	if ld == nil {
		return
	}

	for _, value := range ld.subList(element) {
		if err := ld.run(value, runTask) ; err != nil {
			return fmt.Errorf("%s/%s", element, err)
		}
	}
	if err := runTask(element) ; err != nil {
		return fmt.Errorf("%s: %s", element, err)
	}

	return
}

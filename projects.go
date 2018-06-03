package main

import (
	"fmt"
	"net/url"

	"github.com/forj-oss/forjj-modules/trace"
	"github.com/forj-oss/goforjj"
)

type ProjectInfo struct {
	ForjCommonStruct
	dslRepo         string
	dslPath         string
	isDslDefault    bool
	IsProDeployment bool `yaml:"production-deployment,omitempty"`
}

func (pi *ProjectInfo) set_project_info(forj ForjCommonStruct) {
	pi.ForjCommonStruct = forj
}

func (pi *ProjectInfo) setDslInfo(seedJob SeedJobStruct) {
	if seedJob.Path == "" && seedJob.Repo == "" {
		pi.isDslDefault = true
		pi.dslRepo = seedJob.DefaultRepo
		pi.dslPath = seedJob.DefaultPath
	} else {
		pi.dslRepo = seedJob.Repo
		pi.dslPath = seedJob.Path
	}
}

func (pi *ProjectInfo) setIsProDeploy(isProDeployment bool) {
	pi.IsProDeployment = isProDeployment
}

func (pi *ProjectInfo) set_projects_to(projects map[string]ProjectsInstanceStruct, r *JenkinsPlugin,
	ret *goforjj.PluginData, status *bool, InfraName string) (_ error) {
	if pi.ForjjInfraUpstream == "" {
		ret.StatusAdd("Unable to add a new project without a remote GIT repository. Jenkins JobDSL requirement. " +
			"To enable this feature, add a remote GIT to your infra --infra-upstream or define the JobDSL Repository to clone.")
		IsUpdated(status)
		return
	}

	if v, err := url.Parse(pi.dslRepo); err != nil {
		ret.Errorf("JobDSL: Infra remote URL issue. Check `seed-job-repo` parameter given by Forjj. %s", err)
		return err
	} else {
		if v.Scheme == "" {
			err = fmt.Errorf("JobDSL: Invalid Remote repository Url '%s'. Check `seed-job-repo` parameter given by Forjj. A scheme must exist", pi.dslRepo)
			ret.Errorf("%s", err)
			return err
		}
	}
	// Initialize JobDSL structure
	r.yaml.Projects = NewProjects(pi.ForjjInstanceName, pi.dslRepo, pi.dslPath, pi.isDslDefault)

	// Retrieve list of Repository (projects) to manage
	for name, prj := range projects {
		if prj.RepoDeployHosted != "true" {
			gotrace.Trace("Project %s ignored, because not deploying in production an '%s' repo role.", name, prj.RepoRole)
			continue
		}
		switch prj.RemoteType {
		case "github":
			r.yaml.Projects.AddGithub(name, &prj.GithubStruct)
		case "git":
			r.yaml.Projects.AddGit(name, &prj.GitStruct)
		}
	}
	IsUpdated(status)
	return
}

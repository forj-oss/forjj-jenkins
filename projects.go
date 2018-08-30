package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/forj-oss/goforjj"
)

type Projects struct {
	DslRepo    string `yaml:"dsl-repo,omitempty"`
	DslPath    string `yaml:"dsl-path,omitempty"`
	DslDefault bool   `yaml:"dsl-default,omitempty"`
	infra_name string
	List       map[string]Project
}

// NewProjects creates a list of projects, empty, with DSL path setup by default.
func NewProjects(InstanceName, repo, Dslpath string, dslDefault bool) *Projects {
	p := new(Projects)
	p.DslPath = Dslpath
	p.DslRepo = repo

	p.DslDefault = dslDefault
	p.List = make(map[string]Project)
	return p
}

// AddGithub add the github Jenkins project configuration to the list of projects to manage
func (p *Projects) AddGithub(name string, d *GithubStruct, repoRole, jenkinsfilePath string) bool {
	data := new(GithubStruct)
	data.SetFrom(d)
	p.List[name] = Project{Name: name, SourceType: "github", Github: *data, all: p, Role: repoRole, JenkinsfilePath: jenkinsfilePath}
	return true
}

// AddGit add the git Jenkins project configuration to the list of projects to manage
func (p *Projects) AddGit(name string, d *GitStruct, repoRole, jenkinsfilePath string) bool {
	data := new(GitStruct)
	data.SetFrom(d)
	p.List[name] = Project{Name: name, SourceType: "git", Git: *data, all: p, Role: repoRole, JenkinsfilePath: jenkinsfilePath}
	return true
}

// Generates Jobs-dsl files in the given checked-out GIT repository.
func (p *Projects) Generates(jp *JenkinsPlugin, instanceName string, ret *goforjj.PluginData, status *bool) (_ error) {
	if !p.DslDefault {
		log.Printf(ret.StatusAdd("Deploy: JobDSL groovy files ignored. To generate them, unset seed-job-repo and seed-job-path."))
		return
	}

	templateDir := jp.template_dir
	repoPath := path.Join(jp.deploysParentPath, jp.deployEnv)

	if f, err := os.Stat(repoPath); err != nil {
		return err
	} else {
		if !f.IsDir() {
			return fmt.Errorf(ret.Errorf("Repo cloned path '%s' is not a directory.", repoPath))
		}
	}

	jobsDSLpath := path.Join(repoPath, p.DslPath)
	if f, err := os.Stat(jobsDSLpath); err != nil {
		if err := os.MkdirAll(jobsDSLpath, 0755); err != nil {
			return err
		}
	} else {
		if !f.IsDir() {
			return fmt.Errorf(ret.Errorf("path '%s' is not a directory.", jobsDSLpath))
		}
	}

	tmpl := new(TmplSource)
	tmpl.Template = "jobs-dsl/job-dsl.groovy"
	tmpl.Chmod = 0644

	for name, prj := range p.List {
		name = strings.Replace(name, "-", "_", -1)
		if u, err := tmpl.Generate(prj.Model(jp), templateDir, jobsDSLpath, name+".groovy"); err != nil {
			log.Printf("Deploy: Unable to generate '%s'. %s",
				path.Join(jobsDSLpath, name+".groovy"), ret.Errorf("%s", err))
			return err
		} else if u {
			IsUpdated(status)
			ret.AddFile(goforjj.FilesDeploy, path.Join(p.DslPath, name+".groovy"))
			log.Printf(ret.StatusAdd("Deploy: Project '%s' (%s) generated", name, path.Join(p.DslPath, name+".groovy")))
		}
	}
	return nil
}

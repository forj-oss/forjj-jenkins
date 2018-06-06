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

func NewProjects(InstanceName, repo, Dslpath string, dslDefault bool) *Projects {
	p := new(Projects)
	p.DslPath = Dslpath
	p.DslRepo = repo

	p.DslDefault = dslDefault
	p.List = make(map[string]Project)
	return p
}

func (p *Projects) AddGithub(name string, d *GithubStruct) bool {
	data := new(GithubStruct)
	data.SetFrom(d)
	p.List[name] = Project{Name: name, SourceType: "github", Github: *data, all: p}
	return true
}

func (p *Projects) AddGit(name string, d *GitStruct) bool {
	data := new(GitStruct)
	data.SetFrom(d)
	p.List[name] = Project{Name: name, SourceType: "git", Git: *data, all: p}
	return true
}

// Generates Jobs-dsl files in the given checked-out GIT repository.
func (p *Projects) Generates(jp *JenkinsPlugin, instance_name string, ret *goforjj.PluginData, status *bool) (_ error) {
	if !p.DslDefault {
		log.Printf(ret.StatusAdd("Deploy: JobDSL groovy files ignored. To generate them, unset seed-job-repo and seed-job-path."))
		return
	}

	template_dir := jp.template_dir
	repo_path := jp.deployPath

	if f, err := os.Stat(repo_path); err != nil {
		return err
	} else {
		if !f.IsDir() {
			return fmt.Errorf(ret.Errorf("Repo cloned path '%s' is not a directory.", repo_path))
		}
	}

	jobs_dsl_path := path.Join(repo_path, p.DslPath)
	if f, err := os.Stat(jobs_dsl_path); err != nil {
		if err := os.MkdirAll(jobs_dsl_path, 0755); err != nil {
			return err
		}
	} else {
		if !f.IsDir() {
			return fmt.Errorf(ret.Errorf("path '%s' is not a directory.", repo_path))
		}
	}

	tmpl := new(TmplSource)
	tmpl.Template = "jobs-dsl/job-dsl.groovy"
	tmpl.Chmod = 0644

	for name, prj := range p.List {
		name = strings.Replace(name, "-", "_", -1)
		if u, err := tmpl.Generate(prj.Model(jp), template_dir, jobs_dsl_path, name+".groovy"); err != nil {
			log.Printf("Deploy: Unable to generate '%s'. %s",
				path.Join(jobs_dsl_path, name+".groovy"), ret.Errorf("%s", err))
			return err
		} else if u {
			IsUpdated(status)
			ret.AddFile(goforjj.FilesDeploy, path.Join(instance_name, p.DslPath, name+".groovy"))
			log.Printf(ret.StatusAdd("Deploy: Project '%s' (%s) generated", name, path.Join(p.DslPath, name+".groovy")))
		}
	}
	return nil
}

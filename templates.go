package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

const template_file = "templates.yaml"

// Contains functions to manage source code from templates

type YamlTemplates struct {
	Defaults DefaultsStruct
	Features TmplFeatures
	Sources  TmplSources
	Run      map[string]DeployStepsRunStruct `yaml:"run_deploy"`
	Build    map[string]BuildStepsRunStruct  `yaml:"run_build"`
	Variants map[string]string
}

type DefaultsStruct struct {
	Dockerfile   DockerfileStruct
	JenkinsImage FinalImageStruct
}

type TmplFeatures struct {
	Common TmplFeaturesStruct
	Deploy map[string]TmplFeaturesStruct
}

type TmplFeaturesStruct []string

type TmplSources struct {
	Common TmplSourcesStruct
	Deploy map[string]TmplSourcesStruct
}

type TmplSourcesStruct map[string]TmplSource

type EnvStruct struct {
	Value string
	If    string
}

// SourceModel creates the Model data used by gotemplates in create/update context.
// The model is not updated until call to CleanModel()
func (p *JenkinsPlugin) SourceModel() *JenkinsPluginSourceModel {
	if JPS_Model != nil {
		return JPS_Model
	}
	JPS_Model = new(JenkinsPluginSourceModel)
	JPS_Model.Source = p.yaml
	return JPS_Model
}

func (p *JenkinsPlugin) CleanSourceModel() {
	JPS_Model = nil
}

func Evaluate(value string, data interface{}) (string, error) {
	var doc bytes.Buffer
	tmpl := template.New("jenkins_plugin_data")

	if !strings.Contains(value, "{{") {
		return value, nil
	}
	if _, err := tmpl.Parse(value); err != nil {
		return "", err
	}
	if err := tmpl.Execute(&doc, data); err != nil {
		return "", err
	}
	ret := doc.String()
	log.Printf("'%s' were interpreted to '%s'", value, ret)
	return ret, nil
}

//Load templates definition file from template dir.
func (p *JenkinsPlugin) LoadTemplatesDef() error {
	templatef := path.Join(p.template_dir, template_file)
	if _, err := os.Stat(templatef); err != nil {
		return fmt.Errorf("Unable to find templates definition file '%s'. %s.", templatef, err)
	}

	if d, err := ioutil.ReadFile(templatef); err != nil {
		return fmt.Errorf("Unable to load '%s'. %s.", templatef, err)
	} else {
		if err := yaml.Unmarshal(d, &p.templates_def); err != nil {
			return fmt.Errorf("Unable to load yaml file format '%s'. %s.", templatef, err)
		}
	}
	return nil
}

// Load list of files to copy and files to generate
func (p *JenkinsPlugin) DefineSources() error {
	// load all features
	p.yaml.Features = make([]string, 0, 5)
	for _, f := range p.templates_def.Features.Common {
		if v, err := Evaluate(f, p.SourceModel()); err != nil {
			return fmt.Errorf("Unable to evaluate '%s'. %s", f, err)
		} else {
			if v == "" {
				log.Printf("INFO! No feature defined with '%s'.", f)
				continue
			}
			f = v
		}
		if f != "" {
			p.yaml.Features = append(p.yaml.Features, f)
		}
	}

	if deploy_features, ok := p.templates_def.Features.Deploy[p.yaml.Deploy.Deployment.To]; ok {
		for _, f := range deploy_features {
			if f != "" {
				p.yaml.Features = append(p.yaml.Features, f)
			}
		}
	}

	// TODO: Load additionnal features from maintainer source path or file. This will permit adding more features and let the plugin manage generated path from update task.

	// Load all sources
	p.sources = make(map[string]TmplSource)
	p.templates = make(map[string]TmplSource)
	p.built = make(map[string]TmplSource)
	p.generated = make(map[string]TmplSource)

	choose_file := func(file string, f TmplSource) error {
		if file == "" {
			return nil
		}
		if f.If != "" {
			if v, err := Evaluate(f.If, p.SourceModel()); err != nil {
				return fmt.Errorf("Unable to evaluate the '%s' condition '%s'. %s", file, f.If, err)
			} else {
				if v == "" || strings.ToLower(v) == "false" {
					log.Printf("Condition '%s' negative (false or empty). '%s' ignored.", f.If, file)
					return nil
				}
			}
		}
		if f.Source != "" {
			p.sources[file] = f
			log.Printf("SRC : selected: %s", file)
		} else if f.Template != "" {
			p.templates[file] = f
			log.Printf("TMPL: selected: %s", file)
		} else if f.Built != "" {
			p.built[file] = f
			log.Printf("BUILT: selected: %s", file)
		} else if f.Generated != "" {
			p.generated[file] = f
			log.Printf("GENERATED: selected: %s", file)
			f.storeMD5(path.Join(p.source_path, f.Generated))
		}
		return nil
	}

	for file, f := range p.templates_def.Sources.Common {
		if err := choose_file(file, f); err != nil {
			return err
		}
	}

	if deploy_sources, ok := p.templates_def.Sources.Deploy[p.yaml.Deploy.Deployment.To]; ok {
		for file, f := range deploy_sources {
			if err := choose_file(file, f); err != nil {
				return err
			}
		}
	}

	p.CleanSourceModel()

	return nil
}

func lookup(m map[string]interface{}, key string) (interface{}, error) {
	val, ok := m[key]
	if !ok {
		return nil, errors.New("missing key " + key)
	}
	return val, nil
}

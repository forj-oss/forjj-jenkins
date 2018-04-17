package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/forj-oss/forjj/utils"
	"github.com/forj-oss/goforjj"
	"gopkg.in/yaml.v2"
)

type JenkinsPluginSourceModel struct {
	Source YamlJenkins
}

type JenkinsPluginModel struct {
	Creds  map[string]string
}

var JPS_Model *JenkinsPluginSourceModel
var JP_Model *JenkinsPluginModel

type JenkinsPlugin struct {
	yaml          YamlJenkins       // jenkins.<env>.yaml source file per environment.
	yamlPlugin    YamlJenkinsPlugin // jenkins.yaml source file
	source_path   string            // Source Path
	deployPath    string            // Deployment Path
	deployEnv     string            // Deployment environment where files have to be generated.
	InstanceName  string            // Instance name where files have to be generated.
	template_dir  string
	template_file string
	templates_def YamlTemplates // See templates.go. templates.yaml structure.
	run           RunStruct
	sources       map[string]TmplSource
	templates     map[string]TmplSource
}

type DeployApp struct {
	Deployment DeployStruct
	// Those 2 different parameters are defined at create time and can be updated with change.
	// There are default deployment task and name. This can be changed at maintain time
	// to reflect the maintain deployment task to execute.
	Ssl YamlSSLStruct
}

type YamlSSLStruct struct {
	CaCertificate string `json:"ca-certificate"` // CA root certificate which certify your jenkins instance.
	Certificate   string `json:"certificate"`    // SSL Certificate file to certify your jenkins instance.
	key           string // key for the SSL certificate.
}

const jenkins_file = "forjj-jenkins.yaml"
const maintain_cmds_file = "maintain-cmd.yaml"

func newPlugin(src, deploy string) (p *JenkinsPlugin) {
	p = new(JenkinsPlugin)

	p.source_path = src
	p.deployPath = deploy
	if v, err := utils.Abs(*cliApp.params.template_dir); err != nil {
		log.Printf("Unable to set template dir with '%s'. %s", *cliApp.params.template_dir, err)
	} else {
		p.template_dir = v
	}
	return
}

// GetMaintainData prepare the Model with maintain data (usually credentials)
func (p *JenkinsPlugin) GetMaintainData(req *MaintainReq, ret *goforjj.PluginData) (_ bool) {
	// TODO:
	// Get the list of exclusive maintain data in Creds to facilitate the copy to Model
	// The MaintainData is already centralized...

	model := p.Model()
	if v, found := req.Objects.App[p.InstanceName]; !found {
		ret.Errorf("Request issue. App instance '%s' is missing in list of object.", p.InstanceName)
		return
	} else {
		if p.yaml.Deploy.Ssl.Certificate == "" && v.SslPrivateKey != "" {
			ret.Errorf("A private key is given, but there is no Certificate data.")
			return
		}
		if model.Creds == nil {
			model.Creds = make(map[string]string)
		}

		model.Creds["SslPrivateKey"] = v.SslPrivateKey

		if v.AdminPwd != "" {
			model.Creds["AdminPwd"] = v.AdminPwd
		}

		model.Creds["GithubUserPassword"] = v.GithubUserPassword
	}
	return true
}

// At create time: create jenkins source from req
func (p *JenkinsPlugin) initialize_from(r *CreateReq, ret *goforjj.PluginData) (err error) {
	if _, found := r.Objects.App[p.InstanceName]; !found {
		err = fmt.Errorf("Request format issue. Unable to find the jenkins instance '%s'", p.InstanceName)
		ret.Errorf("%s", err)
		return
	}
	jenkins_instance := r.Objects.App[p.InstanceName]

	p.yaml.Deploy.Deployment.SetFrom(&jenkins_instance.DeployStruct)
	// Initialize deployment data and set default values
	if p.yaml.Deploy.Deployment.To == "" {
		p.yaml.Deploy.Deployment.To = "docker"
		ret.StatusAdd("Default to 'docker' Deployment.")
	}

	if p.yaml.Deploy.Deployment.ServiceAddr == "" {
		p.yaml.Deploy.Deployment.ServiceAddr = "localhost"
		ret.StatusAdd("Default to 'localhost' deployment service name.")
	}

	if p.yaml.Deploy.Deployment.ServicePort == "" {
		p.yaml.Deploy.Deployment.ServicePort = "8080"
		ret.StatusAdd("Default to '8080' deployment service port.")
	}

	// Set SSL data
	p.yaml.Deploy.Ssl.SetFrom(&jenkins_instance.SslStruct)

	// Forjj predefined settings (instance/organization) are set at create time only.
	// I do not recommend to update them, manually by hand in the `forjj-jenkins.yaml`.
	// Updating the instance name could be possible but not for now.
	// As well Moving an instance to another organization could be possible, but I do not see a real use case.
	// So, they are fixed and saved at create time. Update/maintain won't never update them later.
	if err = p.DefineDeployCommand(); err != nil {
		ret.Errorf("Unable to define the default deployment command. %s", err)
		return
	}

	// Initialize Dockerfile data and set default values
	log.Printf("CreateReq : %#v\n", r)
	p.yaml.Dockerfile.SetFrom(&jenkins_instance.DockerfileStruct)
	log.Printf("p.yaml.Dockerfile : %#v\n", p.yaml.Dockerfile)

	// Initialize Jenkins Image data and set default values
	p.yaml.JenkinsImage.SetFrom(&jenkins_instance.FinalImageStruct)

	if err = p.add_projects(r, ret); err != nil {
		return
	}

	if p.yaml.GithubUser.SetFrom(&jenkins_instance.GithubUserStruct) {
		ret.StatusAdd("github-user defined")
		log.Printf("github-user defined with '%s'", p.yaml.GithubUser.Name)
	}

	return
}

func (p *JenkinsPlugin) DefineDeployCommand() error {
	if err := p.LoadTemplatesDef(); err != nil {
		return fmt.Errorf("%s", err)
	}

	if _, ok := p.templates_def.Run[p.yaml.Deploy.Deployment.To]; !ok {
		list := make([]string, 0, len(p.templates_def.Run))
		for element := range p.templates_def.Run {
			list = append(list, element)
		}
		return fmt.Errorf("'%s' deploy type is unknown (templates.yaml). Valid are %s", p.yaml.Deploy.Deployment.To, list)
	}

	return nil
}

// TODO: Detect if the commands was manually updated to avoid updating it if end user did it alone.

// At update time: Update jenkins source from req or forjj-jenkins.yaml input.
func (p *JenkinsPlugin) update_from(r *UpdateReq, ret *goforjj.PluginData, status *bool) error {
	instance := r.Forj.ForjjInstanceName
	instance_data := r.Objects.App[instance]

	var deploy DeployStruct = p.yaml.Deploy.Deployment
	if ok := deploy.UpdateFrom(&instance_data.DeployStruct); ok {
		ret.StatusAdd("Deployment to '%s' updated.", instance_data.To)
		IsUpdated(status)
	}
	p.yaml.Deploy.Deployment = deploy

	var Ssl YamlSSLStruct = p.yaml.Deploy.Ssl
	if ok := Ssl.UpdateFrom(&instance_data.SslStruct); ok {
		ret.StatusAdd("Deployment to '%s' updated.", instance_data.To)
		IsUpdated(status)
	}
	p.yaml.Deploy.Ssl = Ssl

	if err := p.DefineDeployCommand(); err != nil {
		ret.Errorf("Unable to update the deployement command. %s", err)
		return err
	}

	if p.yaml.Dockerfile.UpdateFrom(&instance_data.DockerfileStruct) {
		ret.StatusAdd("Dockerfile updated.")
		IsUpdated(status)
	}
	// Org used only if no set anymore.
	if p.yaml.JenkinsImage.UpdateFrom(&instance_data.FinalImageStruct) {
		ret.StatusAdd("Jenkins master docker image data updated.")
		IsUpdated(status)
	}

	if p.yaml.GithubUser.UpdateFrom(&instance_data.GithubUserStruct) {
		ret.StatusAdd("Jenkins github-user credential updated.")
		IsUpdated(status)
	}

	return nil
}

func (p *JenkinsPlugin) saveYaml(where, fileName string, data interface{}, ret *goforjj.PluginData, status *bool) (_ error) {
	destPath := p.source_path
	if where == goforjj.FilesDeploy {
		destPath = p.deployPath
	}
	file := path.Join(destPath, fileName)

	if f, err := os.Stat(destPath); err != nil {
		if err := os.MkdirAll(destPath, 0755); err != nil {
			return err
		}
	} else {
		if !f.IsDir() {
			return fmt.Errorf(ret.Errorf("path '%s' is not a directory.", destPath))
		}
	}

	origMD5, _ := md5sum(file)
	d, err := yaml.Marshal(data)
	if err != nil {
		ret.Errorf("Unable to encode forjj-jenkins configuration data in yaml. %s", err)
		return err
	}
	finalMD5 := md5.New().Sum(d)

	if bytes.Equal(origMD5, finalMD5) {
		return
	}

	if err = ioutil.WriteFile(file, d, 0644); err != nil {
		ret.Errorf("Unable to save '%s'. %s", file, err)
		return err
	}
	// Be careful to not introduce the local mount which in containers can be totally different (due to docker -v)
	ret.AddFile(where, path.Join(p.InstanceName, fileName))
	ret.StatusAdd("%s: '%s' instance saved (%s).", where, p.InstanceName, path.Join(p.InstanceName, fileName))
	log.Printf("%s: '%s' instance saved.", where, file)
	IsUpdated(status)
	return
}

func (p *JenkinsPlugin) loadYaml(where, fileName string, data interface{}, ret *goforjj.PluginData) (status bool) {
	destPath := p.source_path
	if where == goforjj.FilesDeploy {
		destPath = p.deployPath
	}
	file := path.Join(destPath, fileName)

	log.Printf("Loading '%s'...", file)
	if d, err := ioutil.ReadFile(file); err != nil {
		ret.Errorf("Unable to read '%s'. %s", file, err)
		return
	} else {
		if err = yaml.Unmarshal(d, data); err != nil {
			ret.Errorf("Unable to decode '%s' configuration data from yaml. %s", fileName, err)
			return
		}
	}
	log.Printf("'%s' instance loaded.", file)
	return true
}

func (p *JenkinsPlugin) saveRunYaml(ret *goforjj.PluginData, status *bool) (_ error) {
	run, found := p.templates_def.Run[p.yaml.Deploy.Deployment.To]
	if !found {
		ret.Errorf("Deployment '%s' command not found.", p.yaml.Deploy.Deployment.To)
		return
	}

	return p.saveYaml(goforjj.FilesDeploy, maintain_cmds_file, &run, ret, status)
}

func (p *JenkinsPlugin) loadRunYaml(ret *goforjj.PluginData) (_ bool) {
	return p.loadYaml(goforjj.FilesDeploy, maintain_cmds_file, &p.run, ret)
}

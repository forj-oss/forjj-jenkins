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

var JPS_Model *JenkinsPluginSourceModel

type JenkinsPlugin struct {
	yaml              YamlJenkins       // jenkins.yaml deploy file.
	yamlPlugin        YamlJenkinsPlugin // jenkins.yaml source file
	source_path       string            // Source Path (sourceMountPath + InstanceName)
	deploysParentPath string            // Deployment repository path
	deployPath        string            // Deployment Path (deployRepoPath + deployEnv + InstanceName)
	deployEnv         string            // Deployment environment where files have to be generated.
	InstanceName      string            // Instance name where files have to be generated.
	template_dir      string
	template_file     string
	templates_def     YamlTemplates // See templates.go. templates.yaml structure.
	run               RunStruct
	sources           map[string]TmplSource
	templates         map[string]TmplSource
	built             map[string]TmplSource
	auths             *DockerAuths
}

type YamlSSLStruct struct {
	CaCertificate string `yaml:"ca-certificate,omitempty"` // CA root certificate which certify your jenkins instance.
	Certificate   string `yaml:"certificate,omitempty"`    // SSL Certificate file to certify your jenkins instance.
	key           string // key for the SSL certificate.
	Method        string // SSL Method used in templates (none, selfsigned, manual, ...)
}

const jenkins_file = "forjj-jenkins.yaml"
const jenkinsDeployFile = "forjj-deploy.yaml"
const maintain_cmds_file = "maintain-cmd.yaml"

func newPlugin(srcRepoPath, deploysMountPath string) (p *JenkinsPlugin) {
	p = new(JenkinsPlugin)

	p.source_path = srcRepoPath
	p.deploysParentPath = deploysMountPath
	return
}

// setEnv set the deployment environment where deploy files have to be generated.
func (p *JenkinsPlugin) setEnv(deployEnv, instanceName string) {
	p.deployPath = path.Join(p.deploysParentPath, deployEnv, instanceName)
	p.deployEnv = deployEnv
	p.InstanceName = instanceName
}

func (p *JenkinsPlugin) defineTemplateDir(jenkins_instance AppInstanceStruct) error {
	// Analyze Forjfile input.
	if jenkins_instance.SourceTemplates != "" {
		if v, err := utils.Abs(path.Join(p.source_path, "templates", jenkins_instance.SourceTemplates)); err != nil {
			return fmt.Errorf("Unable to define template directory. %s", err)
		} else {
			p.yamlPlugin.TemplatePath = v
		}
		if _, err := os.Stat(p.yamlPlugin.TemplatePath); err != nil {
			return fmt.Errorf("Unable to define template directory. Template path '%s' is inexistent or inaccessible. %s", p.yamlPlugin.TemplatePath, err)
		}
	}

	// Set template_dir

	if p.yamlPlugin.TemplatePath != "" {
		p.SetTemplateDir(p.yamlPlugin.TemplatePath)
		log.Printf("Using custom templates: %s", p.yamlPlugin.TemplatePath)
	} else if *cliApp.params.template_dir != templateDirDefault {
		p.SetTemplateDir(*cliApp.params.template_dir)
		log.Printf("Using templates defined by the cli: %s", *cliApp.params.template_dir)
	} else {
		p.SetTemplateDir(cliApp.templateDefaultPath)
		log.Printf("Using default templates: %s", cliApp.templateDefaultPath)
	}
	return nil
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
		if p.yaml.Deploy.Ssl.Method == "manual" {
			if p.yaml.Deploy.Ssl.Certificate == "" {
				ret.Errorf("SSL - manual method: Certificate data missing. Update your Forjfile, then do `forjj update`")
				return
			}
			if v.SslPrivateKey == "" {
				ret.Errorf("SSL - manual method: A RSA SSL private key missing. Update your forjj credential data, then do `forjj update`")
				return
			}
		}
		if model.Creds == nil {
			model.Creds = make(map[string]string)
		}
		if model.Env == nil {
			model.Env = make(map[string]string)
		}

		model.loadCreds(req.Forj.ForjjUsername, p.InstanceName, req.Creds)
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

	if v := jenkins_instance.SourceTemplates; v != "" {
		p.yamlPlugin.TemplatePath = v
	}

	if err = p.defineTemplateDir(jenkins_instance); err != nil {
		err = fmt.Errorf("Unable to define your template source path. %s", err)
		ret.Errorf("Unable to define your template source path. %s", err)
		return
	}

	p.yaml.AppExtent = jenkins_instance.Extent

	p.yaml.Deploy.Name = r.Forj.ForjjDeploymentEnv
	p.yaml.Deploy.Type = r.Forj.ForjjDeploymentType

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

	// Define default public url if not set.
	p.yaml.Deploy.DefineDefaultPublicURL()

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
	p.yaml.Dockerfile.SetFrom(&jenkins_instance.DockerfileStruct)

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

	if v := instance_data.SourceTemplates; v != "" {
		p.yamlPlugin.TemplatePath = v
	}

	if err := p.defineTemplateDir(instance_data); err != nil {
		err = fmt.Errorf("Unable to define your template source path. %s", err)
		ret.Errorf("Unable to define your template source path. %s", err)
		return err
	}

	p.yaml.AppExtent = instance_data.Extent

	p.yaml.Deploy.Name = r.Forj.ForjjDeploymentEnv
	p.yaml.Deploy.Type = r.Forj.ForjjDeploymentType

	var deploy DeployStruct = p.yaml.Deploy.Deployment
	if ok := deploy.UpdateFrom(&instance_data.DeployStruct); ok {
		ret.StatusAdd("Deployment to '%s' updated.", instance_data.To)
		IsUpdated(status)
	}
	p.yaml.Deploy.Deployment = deploy

	var Ssl YamlSSLStruct = p.yaml.Deploy.Ssl
	if ok := Ssl.UpdateFrom(&instance_data.SslStruct); ok {
		ret.StatusAdd("'%s' SSL configuration updated.", instance_data.To)
		IsUpdated(status)
	}
	p.yaml.Deploy.Ssl = Ssl

	// Define default public url if not set.
	p.yaml.Deploy.DefineDefaultPublicURL()

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

func (p *JenkinsPlugin) loadYaml(where, fileName string, data interface{}, ret *goforjj.PluginData, ignored bool) (status bool) {
	destPath := p.source_path
	if where == goforjj.FilesDeploy {
		destPath = p.deployPath
	}
	file := path.Join(destPath, fileName)

	log.Printf("Loading '%s'...", file)
	if d, err := ioutil.ReadFile(file); err != nil {
		if ignored {
			log.Printf("Unable to read '%s'. %s", file, err)
			return true
		}
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
	return p.loadYaml(goforjj.FilesDeploy, maintain_cmds_file, &p.run, ret, false)
}

func (p *JenkinsPlugin) SetTemplateDir(templatePath string) (err error) {
	p.template_dir, err = utils.Abs(templatePath)
	return
}

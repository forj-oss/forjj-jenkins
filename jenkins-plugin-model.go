package main

// JenkinsPluginModel provides the structure to evaluate with template before running commands.
type JenkinsPluginModel struct {
	Env    map[string]string
	Creds  map[string]string
	Config YamlJenkins
}

var JP_Model *JenkinsPluginModel

// Model creates the Model data used by gotemplates in maintain context.
func (p *JenkinsPlugin) Model() (model *JenkinsPluginModel) {
	if JP_Model != nil {
		return JP_Model
	}
	JP_Model = new(JenkinsPluginModel)
	JP_Model.Creds = make(map[string]string)
	JP_Model.Env = make(map[string]string)
	JP_Model.Config = p.yaml
	return JP_Model
}

func (jpm *JenkinsPluginModel) loadCreds(instanceName string, creds map[string]string) {
	if jpm == nil || jpm.Creds == nil {
		return
	}

	// Predefine credentials from jenkins.yaml
	appPrefix := "app-" + instanceName + "-"
	credList := map[string]string{
		"SslPrivateKey":      appPrefix + "ssl-private-key",
		"AdminPwd":           appPrefix + "admin-pwd",
		"GithubUserPassword": appPrefix + "github-user-password",
	}

	for credName, credValue := range credList {
		if v, found := creds[credValue]; found && v != "" {
			jpm.Creds[credName] = v
		}
	}

	// Extended Credentials
	for credName, credValue := range creds {
		jpm.Creds[credName] = credValue
	}
}
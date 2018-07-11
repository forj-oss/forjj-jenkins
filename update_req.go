package main

import (
	log "forjj-jenkins/reportlogs"
	"os"
	"path"

	"github.com/forj-oss/goforjj"
)

// Return ok if the jenkins instance sources and template sources exists
func (r *UpdateReq) checkSourceExistence(ret *goforjj.PluginData) (p *JenkinsPlugin, status bool) {
	log.Printf("Checking Jenkins source code existence.")
	if _, err := os.Stat(r.Forj.ForjjSourceMount); err != nil {
		ret.Errorf("Unable to update jenkins instances. '%s' is inexistent or innacessible. %s", r.Forj.ForjjSourceMount, err)
		return
	}

	srcPath := path.Join(r.Forj.ForjjSourceMount, r.Forj.ForjjInstanceName)

	log.Printf("Checking Jenkins deploy path.")
	if _, err := os.Stat(path.Join(r.Forj.ForjjDeployMount, r.Forj.ForjjDeploymentEnv)); err != nil {
		ret.Errorf("Unable to update jenkins instances. '%s'/'%s' is inexistent or innacessible. %s", r.Forj.ForjjDeployMount, r.Forj.ForjjDeploymentEnv, err)
		return
	}

	p = newPlugin(srcPath, r.Forj.ForjjDeployMount)
	p.setEnv(r.Forj.ForjjDeploymentEnv, r.Forj.ForjjInstanceName)

	ret.StatusAdd("environment checked.")
	return p, true
}

package main

import (
	"net/url"

	"github.com/forj-oss/forjj-modules/trace"
)

type DeployApp struct {
	Deployment DeployStruct
	Name       string
	Type       string
	// Those 2 different parameters are defined at create time and can be updated with change.
	// There are default deployment task and name. This can be changed at maintain time
	// to reflect the maintain deployment task to execute.
	Ssl YamlSSLStruct
}

// DefineDefaultPublicURL validate and define a valid URL from Ssl and Deployment information
// if the URL setup is invalid, an error is generated but not exited. The default URL is then set.
//
func (d *DeployApp) DefineDefaultPublicURL() {
	if d == nil {
		return
	}

	if d.Deployment.PublicServiceUrl != "" {
		if _, err := url.ParseRequestURI(d.Deployment.PublicServiceUrl); err != nil {
			gotrace.Error("Public Service URL '%s' is invalid. %s. Using default setup.", d.Deployment.PublicServiceUrl, err)
		}
	}

	proto := "http"
	port := ""
	if d.Ssl.Method != "none" {
		proto = "https"
		if d.Deployment.ServicePort != "443" {
			port = ":" + d.Deployment.ServicePort
		}
	} else {
		if d.Deployment.ServicePort != "80" {
			port = ":" + d.Deployment.ServicePort
		}
	}

	d.Deployment.PublicServiceUrl = proto + "://" + d.Deployment.ServiceAddr + port
	gotrace.Trace("Public service Url is set to: %s", d.Deployment.PublicServiceUrl)
}

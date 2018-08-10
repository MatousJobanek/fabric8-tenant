package template

import (
	"github.com/fabric8-services/fabric8-tenant/utils"
	"fmt"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	"github.com/fabric8-services/fabric8-notification/template"
)

//var templates = map[string][]utils.DownloadFileFunction{
//	"run":   templateLocations(""),
//	"stage": templateLocations(""),
//	"che": templateLocations(
//		"http://central.maven.org/maven2/io/fabric8/tenant/packages/fabric8-tenant-che-mt/2.0.85/fabric8-tenant-che-mt-2.0.85-openshift.yml",
//		"http://central.maven.org/maven2/io/fabric8/tenant/packages/fabric8-tenant-che-quotas-oso/2.0.85/fabric8-tenant-che-quotas-oso-2.0.85-openshift.yml",
//		//"http://central.maven.org/maven2/io/fabric8/tenant/packages/fabric8-tenant-team/2.0.11/fabric8-tenant-team-2.0.11-openshift.yml"
//	),
//	"jenkins": templateLocations("http://central.maven.org/maven2/io/fabric8/tenant/packages/fabric8-tenant-jenkins/4.0.93/fabric8-tenant-jenkins-4.0.93-openshift.yml",
//		"http://central.maven.org/maven2/io/fabric8/tenant/packages/fabric8-tenant-jenkins-quotas-oso/4.0.93/fabric8-tenant-jenkins-quotas-oso-4.0.93-openshift.yml"),
//}

const (
	rawFileFabric8TenantServiceURL = "https://github.com/fabric8-services/fabric8-tenant/blob/%s/%s"
	templatesDirectory             = "template/templates/"
)

var templateNames = map[string]string{
	"run":     "",
	"stage":   "",
	"che":     "fabric8-tenant-che-openshift.yml",
	"che-mt":  "fabric8-tenant-che-mt-openshift.yml",
	"jenkins": "fabric8-tenant-jenkins-openshift.yml",
}

func RetrieveTemplatesObjects(namespace *string, username string, config *configuration.Data) (Objects, error) {
	tmpls := make([]string, 0)
	if utils.IsEmpty(namespace) {
		for _, tmplName := range templateNames {
			retrieveTemplate(&tmpls, tmplName, config)
		}
	} else {
		retrieveTemplate(&tmpls, templateNames[*namespace], config)
	}

	objects, err := ProcessTemplates(username, config, tmpls...)
	if err != nil {
		return objects, err
	}
	return objects, nil
}

func retrieveTemplate(templates *[]string, tmplName string, config *configuration.Data) error {
	pathToTemplate := templatesDirectory + tmplName
	var (
		content []byte
		err     error
	)
	if config.GetFabric8TenantServiceRepoSha() != "" {
		content, err = utils.DownloadFile(fmt.Sprintf(rawFileFabric8TenantServiceURL, config.GetFabric8TenantServiceRepoSha(), pathToTemplate))
	} else {
		content, err = template.Asset(pathToTemplate)
	}
	if err != nil {
		return err
	}
	*templates = append(*templates, string(content))

}

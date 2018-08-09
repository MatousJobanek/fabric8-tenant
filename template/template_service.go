package template

import (
	"github.com/fabric8-services/fabric8-tenant/utils"
	"fmt"
	"github.com/fabric8-services/fabric8-tenant/configuration"
)

var templates = map[string][]utils.DownloadFileFunction{
	"run":   templateLocations(""),
	"stage": templateLocations(""),
	"che": templateLocations(
		"http://central.maven.org/maven2/io/fabric8/tenant/packages/fabric8-tenant-che-mt/2.0.85/fabric8-tenant-che-mt-2.0.85-openshift.yml",
		"http://central.maven.org/maven2/io/fabric8/tenant/packages/fabric8-tenant-che-quotas-oso/2.0.85/fabric8-tenant-che-quotas-oso-2.0.85-openshift.yml",
		//"http://central.maven.org/maven2/io/fabric8/tenant/packages/fabric8-tenant-team/2.0.11/fabric8-tenant-team-2.0.11-openshift.yml"
	),
	"jenkins": templateLocations("http://central.maven.org/maven2/io/fabric8/tenant/packages/fabric8-tenant-jenkins/4.0.93/fabric8-tenant-jenkins-4.0.93-openshift.yml",
		"http://central.maven.org/maven2/io/fabric8/tenant/packages/fabric8-tenant-jenkins-quotas-oso/4.0.93/fabric8-tenant-jenkins-quotas-oso-4.0.93-openshift.yml"),
}

func RetrieveTemplatesObjects(namespace *string, username string, config *configuration.Data) (Objects, error) {
	tmpls := make([]string, 0)
	if utils.IsEmpty(namespace) {
		for _, sources := range templates {
			downloadTemplate(&tmpls, sources)
		}
	} else {
		downloadTemplate(&tmpls, templates[*namespace])
	}

	objects, err := ProcessTemplates(username, config, tmpls...)
	if err != nil {
		return objects, err
	}
	return objects, nil
}

func templateLocations(urls ...string) []utils.DownloadFileFunction {
	downloadFunctions := make([]utils.DownloadFileFunction, 0, len(urls))
	for _, url := range urls {
		downloadFunctions = append(downloadFunctions, utils.NewDownloadFileFunction(url))
	}
	return downloadFunctions
}

func downloadTemplate(templates *[]string, sources []utils.DownloadFileFunction) {
	for _, source := range sources {
		template, err := source()
		if err != nil {
			fmt.Println(err)
		}
		*templates = append(*templates, string(template))
	}
}

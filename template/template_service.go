package template

import (
	"github.com/fabric8-services/fabric8-tenant/utils"
	"fmt"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	"github.com/fabric8-services/fabric8-tenant/template/generated"
)

//go:generate go-bindata -prefix "./assets/" -pkg assets -o ./generated/templates.go ./assets/...

const (
	rawFileFabric8TenantServiceURL = "https://github.com/fabric8-services/fabric8-tenant/blob/%s/%s"
	templatesDirectory             = "template/assets/"
)

type template struct {
	filename        string
	namespaceSuffix string
	defaultParams   map[string]string
}

var (
	stageParams = map[string]string{"ENV_TYPE": "stage"}
	runParams   = map[string]string{"ENV_TYPE": "run"}
	noParams    map[string]string
)

func newTemplate(filename, namespaceSuffix string, defaultParams map[string]string) template {
	return template{
		filename:        filename,
		defaultParams:   defaultParams,
		namespaceSuffix: namespaceSuffix,
	}
}

var templateNames = map[string]template{
	"run":     newTemplate("fabric8-tenant-environment.yml", "-run", runParams),
	"stage":   newTemplate("fabric8-tenant-environment.yml", "-stage", stageParams),
	"che":     newTemplate("fabric8-tenant-che.yml", "-che", noParams),
	"che-mt":  newTemplate("fabric8-tenant-che-mt.yml", "-che", noParams),
	"jenkins": newTemplate("fabric8-tenant-jenkins.yml", "-jenkins", noParams),
}

func RetrieveTemplatesObjects(namespace *string, username string, config *configuration.Data) (Objects, error) {
	templates := make([]string, 0)

	if utils.IsEmpty(namespace) {
		for _, template := range templateNames {
			content, err := retrieveTemplate(template, config)
			if err != nil {
				return nil, err
			}
			templates = append(templates, content)
		}

	} else {
		content, err := retrieveTemplate(templateNames[*namespace], config)
		if err != nil {
			return nil, err
		}
		templates = append(templates, content)
	}

	objects, err := ProcessTemplates(username, *namespace, config, templates...)
	if err != nil {
		return objects, err
	}
	return objects, nil
}

func retrieveTemplate(template template, config *configuration.Data) (string, error) {
	var (
		content []byte
		err     error
	)
	if config.GetFabric8TenantServiceRepoSha() != "" {
		pathToTemplate := templatesDirectory + template.filename
		content, err = utils.DownloadFile(fmt.Sprintf(rawFileFabric8TenantServiceURL, config.GetFabric8TenantServiceRepoSha(), pathToTemplate))
	} else {
		content, err = assets.Asset(template.filename)
	}
	if err != nil {
		return "", err
	}
	return string(content), nil
}

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
	templatesDirectory             = "templateDef/assets/"
)

type Template struct {
	filename      string
	defaultParams map[string]string
	content       string
}

var (
	stageParams = map[string]string{"ENV_TYPE": "stage"}
	runParams   = map[string]string{"ENV_TYPE": "run"}
	noParams    map[string]string
)

func newTemplateDef(filename string, defaultParams map[string]string) Template {
	return Template{
		filename:      filename,
		defaultParams: defaultParams,
	}
}

var templateNames = map[string]Template{
	"run":     newTemplateDef("fabric8-tenant-environment.yml", runParams),
	"stage":   newTemplateDef("fabric8-tenant-environment.yml", stageParams),
	"che":     newTemplateDef("fabric8-tenant-che.yml", noParams),
	"che-mt":  newTemplateDef("fabric8-tenant-che-mt.yml", noParams),
	"jenkins": newTemplateDef("fabric8-tenant-jenkins.yml", noParams),
}

func RetrieveTemplates(namespace *string, config *configuration.Data) ([]Template, error) {
	templates := make([]Template, 0)

	if utils.IsEmpty(namespace) {
		for _, template := range templateNames {
			err := retrieveTemplate(&template, config)
			if err != nil {
				return nil, err
			}
			templates = append(templates, template)
		}

	} else {
		template := templateNames[*namespace]
		err := retrieveTemplate(&template, config)
		if err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}

	return templates, nil
}

func retrieveTemplate(template *Template, config *configuration.Data) error {
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
	template.content = string(content)
	return err
}

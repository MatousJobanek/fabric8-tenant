package environment

import (
	"github.com/fabric8-services/fabric8-tenant/utils"
	"fmt"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	"github.com/satori/go.uuid"
	"github.com/arquillian/ike-prow-plugins/pkg/assets/generated"
	"time"
)

//go:generate go-bindata -prefix "./templates/" -pkg templates -o ./generated/templates.go ./templates/...

const (
	rawFileFabric8TenantServiceURL = "https://github.com/fabric8-services/fabric8-tenant/blob/%s/%s"
	templatesDirectory             = "templateDef/assets/"
)

var DefaultNamespaces = []string{"run", "stage", "che", "jenkins"}

var templateNames = map[string]*Template{
	"run":   newTemplate("fabric8-tenant-environment.yml", runParams),
	"stage": newTemplate("fabric8-tenant-environment.yml", stageParams),
	"che":   newTemplate("fabric8-tenant-che-mt.yml", noParams),
	//"che":  newTemplate("fabric8-tenant-che.yml", noParams),
	"jenkins": newTemplate("fabric8-tenant-jenkins.yml", noParams),
}

type Service struct {
	config *configuration.Data
}

func NewService(config *configuration.Data) *Service {
	return &Service{config: config}
}

type EnvData struct {
	NsType   string
	NamespaceName string
	Name string
	Template Template
	ExpiresAt *time.Time
}

func (s *Service) GetEnvData(space *uuid.UUID, nsType string) (*EnvData, error) {
	template := templateNames[nsType]
	err := retrieveTemplate(template, s.config)
	if err != nil {
		return nil, err
	}

	return &EnvData{
		Template: *template,
		NsType:   nsType,
	}, nil

}

//func RetrieveTemplates(namespace *string, config *configuration.Data) ([]Template, error) {
//	templates := make([]Template, 0)
//
//	if utils.IsEmpty(namespace) {
//		for _, template := range templateNames {
//			err := retrieveTemplate(&template, config)
//			if err != nil {
//				return nil, err
//			}
//			templates = append(templates, template)
//		}
//
//	} else {
//
//	}
//
//	return templates, nil
//}

func retrieveTemplate(template *Template, config *configuration.Data) error {
	var (
		content []byte
		err     error
	)
	if config.GetFabric8TenantServiceRepoSha() != "" {
		pathToTemplate := templatesDirectory + template.Filename
		content, err = utils.DownloadFile(fmt.Sprintf(rawFileFabric8TenantServiceURL, config.GetFabric8TenantServiceRepoSha(), pathToTemplate))
	} else {
		content, err = assets.Asset(template.Filename)
	}
	template.Content = string(content)
	return err
}

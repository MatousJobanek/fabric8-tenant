package environment

import (
	"gopkg.in/yaml.v2"
	"regexp"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	"strings"
	"fmt"
	"github.com/fabric8-services/fabric8-tenant/keycloak"
	authClient "github.com/fabric8-services/fabric8-tenant/auth/client"
)

const (
	FieldKind            = "kind"
	FieldAPIVersion      = "apiVersion"
	FieldObjects         = "objects"
	FieldSpec            = "spec"
	FieldTemplate        = "templateDef"
	FieldItems           = "items"
	FieldMetadata        = "metadata"
	FieldLabels          = "labels"
	FieldReplicas        = "replicas"
	FieldVersion         = "version"
	FieldNamespace       = "namespace"
	FieldName            = "name"
	FieldStatus          = "status"
	FieldResourceVersion = "resourceVersion"

	ValKindTemplate               = "Template"
	ValKindNamespace              = "Namespace"
	ValKindConfigMap              = "ConfigMap"
	ValKindLimitRange             = "LimitRange"
	ValKindProject                = "Project"
	ValKindProjectRequest         = "ProjectRequest"
	ValKindPersistenceVolumeClaim = "PersistentVolumeClaim"
	ValKindService                = "Service"
	ValKindSecret                 = "Secret"
	ValKindServiceAccount         = "ServiceAccount"
	ValKindRoleBindingRestriction = "RoleBindingRestriction"
	ValKindRoleBinding            = "RoleBinding"
	ValKindRoute                  = "Route"
	ValKindJob                    = "Job"
	ValKindList                   = "List"
	ValKindDeployment             = "Deployment"
	ValKindDeploymentConfig       = "DeploymentConfig"
	ValKindResourceQuota          = "ResourceQuota"

	varUserName              = "USER_NAME"
	varProjectUser           = "PROJECT_USER"
	varProjectRequestingUser = "PROJECT_REQUESTING_USER"
	varProjectAdminUser      = "PROJECT_ADMIN_USER"
	varNamespacePrefix       = "NAMESPACE_PREFIX"
	varKeycloakURL           = "KEYCLOAK_URL"
)

var sortOrder = map[string]int{
	"Namespace":              1,
	"ProjectRequest":         1,
	"RoleBindingRestriction": 2,
	"LimitRange":             3,
	"ResourceQuota":          4,
	"Secret":                 5,
	"ServiceAccount":         6,
	"Service":                7,
	"RoleBinding":            8,
	"PersistentVolumeClaim":  9,
	"ConfigMap":              10,
	"DeploymentConfig":       11,
	"Route":                  12,
	"Job":                    13,
}

type Objects []map[string]interface{}
type Object map[string]interface{}

type Template struct {
	Filename      string
	defaultParams map[string]string
	content       string
}

var (
	stageParams = map[string]string{"ENV_TYPE": "stage"}
	runParams   = map[string]string{"ENV_TYPE": "run"}
	noParams    map[string]string
)

func newTemplate(filename string, defaultParams map[string]string) Template {
	return Template{
		Filename:      filename,
		defaultParams: defaultParams,
	}
}

func (t Template) Process(vars map[string]string) (Objects, error) {

	var objects Objects
	templateVars := merge(vars, t.defaultParams)
	pt, err := t.process(templateVars)
	if err != nil {
		return objects, err
	}
	return ParseObjects(pt)
}

// Process takes a K8/Openshift Template as input and resolves the variable expresions
func (t Template)  process(variables map[string]string) (string, error) {
	reg := regexp.MustCompile(`\${([A-Z_0-9]+)}`)
	return string(reg.ReplaceAllFunc([]byte(t.content), func(found []byte) []byte {
		variableName := toVariableName(string(found))
		if variable, ok := variables[variableName]; ok {
			return []byte(variable)
		}
		return found
	})), nil
}

func CollectVars(user string, config *configuration.Data) map[string]string {
	userName := RetrieveUserName(user)

	vars := map[string]string{
		varUserName:              userName,
		varProjectUser:           user,
		varProjectRequestingUser: user,
		varNamespacePrefix:       userName,
		//varProjectAdminUser:      config.MasterUser,
	}

	return merge(vars, getVariables(config))
}

func merge(target, second map[string]string) map[string]string {
	for k, v := range second {
		if _, exist := target[k]; !exist {
			target[k] = v
		}
	}
	return target
}

// RetrieveUserName returns a safe namespace basename based on a username
func RetrieveUserName(username string) string {
	return regexp.MustCompile("[^a-z0-9]").ReplaceAllString(strings.Split(username, "@")[0], "-")
}

func getVariables(config *configuration.Data) map[string]string {
	keycloakConfig := keycloak.Config{
		BaseURL: config.GetKeycloakURL(),
		Realm:   config.GetKeycloakRealm(),
		Broker:  config.GetKeycloakOpenshiftBroker(),
	}

	templateVars, err := config.GetTemplateValues()
	if err != nil {
		panic(err)
	}
	templateVars["COMMIT"] = "123abc"
	templateVars["KEYCLOAK_URL"] = ""
	templateVars["KEYCLOAK_OSO_ENDPOINT"] = keycloakConfig.CustomBrokerTokenURL("openshift-v3")
	templateVars["KEYCLOAK_GITHUB_ENDPOINT"] = fmt.Sprintf("%s%s?for=https://github.com", config.GetAuthURL(), authClient.RetrieveTokenPath())

	return templateVars
}

func toVariableName(exp string) string {
	return exp[:len(exp)-1][2:]
}

// ParseObjects return a string yaml and return a array of the objects/items from a Template/List kind
func ParseObjects(source string) (Objects, error) {
	var template Object

	err := yaml.Unmarshal([]byte(source), &template)
	if err != nil {
		return nil, err
	}

	if GetKind(template) == ValKindTemplate || GetKind(template) == ValKindList {
		var ts []interface{}
		if GetKind(template) == ValKindTemplate {
			ts = template[FieldObjects].([]interface{})
		} else if GetKind(template) == ValKindList {
			ts = template[FieldItems].([]interface{})
		}
		var objs Objects
		for _, obj := range ts {
			parsedObj := obj.(map[interface{}]interface{})
			stringKeys := make(Object, len(parsedObj))
			for key, value := range parsedObj {
				stringKeys[key.(string)] = value
			}
			objs = append(objs, stringKeys)
		}
		return objs, nil
	}

	return Objects{template}, nil
}

func GetName(obj Object) string {
	if meta, metaFound := obj[FieldMetadata].(Object); metaFound {
		if name, nameFound := meta[FieldName].(string); nameFound {
			return name
		}
	}
	return ""
}

func GetNamespace(obj Object) string {
	if meta, metaFound := obj[FieldMetadata].(Object); metaFound {
		if name, nameFound := meta[FieldNamespace].(string); nameFound {
			return name
		}
	}
	return ""
}

func GetKind(obj Object) string {
	if kind, kindFound := obj[FieldKind].(string); kindFound {
		return kind
	}
	return ""
}

// ByKind represents a list of Openshift objects sortable by Kind
type ByKind Objects

func (a ByKind) Len() int      { return len(a) }
func (a ByKind) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByKind) Less(i, j int) bool {
	iO := 30
	jO := 30

	if val, ok := sortOrder[GetKind(a[i])]; ok {
		iO = val
	}
	if val, ok := sortOrder[GetKind(a[j])]; ok {
		jO = val
	}
	return iO < jO
}

package template

import (
	"gopkg.in/yaml.v2"
	"regexp"
	"sort"
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
	FieldTemplate        = "template"
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

	varProjectName           = "PROJECT_NAME"
	varProjectTemplateName   = "PROJECT_TEMPLATE_NAME"
	varProjectDisplayName    = "PROJECT_DISPLAYNAME"
	varProjectDescription    = "PROJECT_DESCRIPTION"
	varProjectUser           = "PROJECT_USER"
	varProjectRequestingUser = "PROJECT_REQUESTING_USER"
	varProjectAdminUser      = "PROJECT_ADMIN_USER"
	varProjectNamespace      = "PROJECT_NAMESPACE"
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

type Objects []map[interface{}]interface{}
type Object map[interface{}]interface{}

func ProcessTemplates(username string, config *configuration.Data, tmpls ...string) (Objects, error) {
	namespace := CreateNSName(username)

	vars := map[string]string{
		varProjectName:           namespace,
		varProjectTemplateName:   namespace,
		varProjectDisplayName:    namespace,
		varProjectDescription:    namespace,
		varProjectUser:           username,
		varProjectRequestingUser: username,
		varProjectNamespace:      namespace,
		//varProjectAdminUser:      config.MasterUser,
	}

	for k, v := range getVariables(config) {
		if _, exist := vars[k]; !exist {
			vars[k] = v
		}
	}

	var objects []map[interface{}]interface{}
	for _, template := range tmpls {
		pt, err := Process(template, vars)
		if err != nil {
			return objects, err
		}
		objs, err := ParseObjects(pt, namespace)
		if err != nil {
			return objects, err
		}
		objects = append(objects, objs...)
	}

	sort.Sort(ByKind(objects))
	return objects, nil
}

// CreateNSName returns a safe namespace basename based on a username
func CreateNSName(username string) string {
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

	templateVars["KEYCLOAK_URL"] = ""
	templateVars["KEYCLOAK_OSO_ENDPOINT"] = keycloakConfig.CustomBrokerTokenURL("openshift-v3")
	templateVars["KEYCLOAK_GITHUB_ENDPOINT"] = fmt.Sprintf("%s%s?for=https://github.com", config.GetAuthURL(), authClient.RetrieveTokenPath())

	return templateVars
}

// Process takes a K8/Openshift Template as input and resolves the variable expresions
func Process(template string, variables map[string]string) (string, error) {
	reg := regexp.MustCompile(`\${([A-Z_0-9]+)}`)
	return string(reg.ReplaceAllFunc([]byte(template), func(found []byte) []byte {
		variableName := toVariableName(string(found))
		if variable, ok := variables[variableName]; ok {
			return []byte(variable)
		}
		return found
	})), nil
}

func toVariableName(exp string) string {
	return exp[:len(exp)-1][2:]
}

func replaceTemplateExpression(template string) string {
	reg := regexp.MustCompile(`\${([A-Z_]+)}`)
	return reg.ReplaceAllString(template, "{{.$1}}")
}

// ParseObjects return a string yaml and return a array of the objects/items from a Template/List kind
func ParseObjects(source string, namespace string) ([]map[interface{}]interface{}, error) {
	var template map[interface{}]interface{}

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
		var objs []map[interface{}]interface{}
		for _, obj := range ts {
			objs = append(objs, obj.(map[interface{}]interface{}))
		}
		if namespace != "" {
			for _, obj := range objs {
				kind := GetKind(obj)
				if val, ok := obj[FieldMetadata].(map[interface{}]interface{}); ok && kind != ValKindProjectRequest && kind != ValKindNamespace {
					if _, ok := val[FieldNamespace]; !ok {
						val[FieldNamespace] = namespace
					}
				}
			}
		}
		return objs, nil
	}

	return []map[interface{}]interface{}{template}, nil
}

func GetName(obj map[interface{}]interface{}) string {
	if meta, metaFound := obj[FieldMetadata].(map[interface{}]interface{}); metaFound {
		if name, nameFound := meta[FieldName].(string); nameFound {
			return name
		}
	}
	return ""
}

func GetNamespace(obj map[interface{}]interface{}) string {
	if meta, metaFound := obj[FieldMetadata].(map[interface{}]interface{}); metaFound {
		if name, nameFound := meta[FieldNamespace].(string); nameFound {
			return name
		}
	}
	return ""
}

func GetKind(obj map[interface{}]interface{}) string {
	if kind, kindFound := obj[FieldKind].(string); kindFound {
		return kind
	}
	return ""
}

// ByKind represents a list of Openshift objects sortable by Kind
type ByKind []map[interface{}]interface{}

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

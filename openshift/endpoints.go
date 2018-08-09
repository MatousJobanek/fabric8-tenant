package openshift

import (
	"fmt"
	"github.com/fabric8-services/fabric8-tenant/template"
)

type ObjectEndpoints struct {
	methods map[string]*MethodDefinition
}

var (
	objectEndpoints = map[string]*ObjectEndpoints{
		template.ValKindNamespace: endpoints(
			endpoint(`/api/v1/namespaces`, POST(IgnoreConflicts)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "name"}}`, PUT(), PATCH(), GET(), DELETE())),

		template.ValKindProject: endpoints(
			endpoint(`/oapi/v1/projects`, POST(WhenConflictThenDeleteAndRetry)),
			endpoint(`/oapi/v1/projects/{{ index . "metadata" "name"}}`, PUT(), PATCH(), GET(), DELETE())),

		template.ValKindProjectRequest: endpoints(
			endpoint(`/oapi/v1/projectrequests`, POST(IgnoreConflicts), GET())),

		template.ValKindRoleBinding: endpoints(
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/rolebindings`, POST(WhenConflictThenDeleteAndRetry)),
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/rolebindings/{{ index . "metadata" "name"}}`, PUT(), PATCH(), GET(), DELETE())),

		template.ValKindRoleBindingRestriction: endpoints(
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/rolebindingrestrictions`, POST(WhenConflictThenDeleteAndRetry)),
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/rolebindingrestrictions/{{ index . "metadata" "name"}}`, PUT(), PATCH(), GET(), DELETE())),

		template.ValKindRoute: endpoints(
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/routes`, POST(WhenConflictThenDeleteAndRetry)),
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/routes/{{ index . "metadata" "name"}}`, PUT(), PATCH(), GET(), DELETE())),

		template.ValKindDeployment: endpoints(
			endpoint(`/apis/extensions/v1beta1/namespaces/{{ index . "metadata" "namespace"}}/deployments`, POST(WhenConflictThenDeleteAndRetry)),
			endpoint(`/apis/extensions/v1beta1/namespaces/{{ index . "metadata" "namespace"}}/deployments/{{ index . "metadata" "name"}}`, PUT(), PATCH(), GET(), DELETE())),

		template.ValKindDeploymentConfig: endpoints(
			endpoint(`/apis/apps.openshift.io/v1/namespaces/{{ index . "metadata" "namespace"}}/deploymentconfigs`, POST(WhenConflictThenDeleteAndRetry)),
			endpoint(`/apis/apps.openshift.io/v1/namespaces/{{ index . "metadata" "namespace"}}/deploymentconfigs/{{ index . "metadata" "name"}}`, PUT(), PATCH(), GET(), DELETE())),

		template.ValKindPersistenceVolumeClaim: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/persistentvolumeclaims`, POST(IgnoreConflicts)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/persistentvolumeclaims/{{ index . "metadata" "name"}}`, PUT(), PATCH(), GET(), DELETE())),

		template.ValKindService: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/services`, POST(WhenConflictThenDeleteAndRetry)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/services/{{ index . "metadata" "name"}}`, PUT(), PATCH(), GET(), DELETE())),

		template.ValKindSecret: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/secrets`, POST(WhenConflictThenDeleteAndRetry)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/secrets/{{ index . "metadata" "name"}}`, PUT(), PATCH(), GET(), DELETE())),

		template.ValKindServiceAccount: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/serviceaccounts`, POST(IgnoreConflicts)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/serviceaccounts/{{ index . "metadata" "name"}}`, PUT(), PATCH(), GET(), DELETE())),

		template.ValKindConfigMap: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/configmaps`, POST(WhenConflictThenDeleteAndRetry)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/configmaps/{{ index . "metadata" "name"}}`, PUT(), PATCH(), GET(), DELETE())),

		template.ValKindResourceQuota: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/resourcequotas`, POST(WhenConflictThenDeleteAndRetry)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/resourcequotas/{{ index . "metadata" "name"}}`, PUT(), PATCH(), GET(), DELETE())),

		template.ValKindLimitRange: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/limitranges`, POST(WhenConflictThenDeleteAndRetry)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/limitranges/{{ index . "metadata" "name"}}`, PUT(), PATCH(), GET(), DELETE())),

		template.ValKindJob: endpoints(
			endpoint(`/apis/batch/v1/namespaces/{{ index . "metadata" "namespace"}}/jobs`, POST(WhenConflictThenDeleteAndRetry)),
			endpoint(`/apis/batch/v1/namespaces/{{ index . "metadata" "namespace"}}/jobs/{{ index . "metadata" "name"}}`, PUT(), PATCH(), GET(), DELETE())),
	}
	deleteOptions = `apiVersion: v1
kind: DeleteOptions
gracePeriodSeconds: 0
orphanDependents: false`

	adminRole = `apiVersion: v1
kind: RoleBinding
metadata:
  name: admin`
)

func endpoint(endpoint string, methodsDefCreators ...methodDefCreator) func(methods map[string]*MethodDefinition) {
	return func(methods map[string]*MethodDefinition) {
		for _, methodDefCreator := range methodsDefCreators {
			methodDef := methodDefCreator(endpoint)
			methods[methodDef.action] = methodDef
		}
	}
}

func endpoints(endpoints ...func(methods map[string]*MethodDefinition)) *ObjectEndpoints {
	methods := make(map[string]*MethodDefinition)
	for _, endpoint := range endpoints {
		endpoint(methods)
	}
	return &ObjectEndpoints{methods: methods}
}

func (e *ObjectEndpoints) ApplyWithMethodCallback(client *Client, object template.Object, action string) error {
	method, err := e.getMethodDefinition(action, object)
	if err != nil {
		return err
	}
	response, err := client.Do(method.requestCreator, object)
	LogRequestInfo(client, object, e, method, response)
	if err != nil {
		return err
	}
	if len(method.callbacks) == 0 {
		return checkHTTPCode(response, err)
	} else {
		for _, callback := range method.callbacks {
			err := callback(client, object, e, method, response)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func (e *ObjectEndpoints) getMethodDefinition(method string, object map[interface{}]interface{}) (*MethodDefinition, error) {
	methodDef, found := e.methods[method]
	if !found {
		return nil, fmt.Errorf("method definition %s for %s not supported", method, template.GetKind(object))
	}
	return methodDef, nil
}

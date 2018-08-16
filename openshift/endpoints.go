package openshift

import (
	"fmt"
	"github.com/fabric8-services/fabric8-tenant/template"
	"github.com/fabric8-services/fabric8-tenant/utils"
	"net/http"
)

type ObjectEndpoints struct {
	methods map[string]*MethodDefinition
}

var (
	objectEndpoints = map[string]*ObjectEndpoints{
		template.ValKindNamespace: endpoints(
			endpoint(`/api/v1/namespaces`, POST(IgnoreConflicts)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		template.ValKindProject: endpoints(
			endpoint(`/oapi/v1/projects`, POST(WhenConflictThenPatch)),
			endpoint(`/oapi/v1/projects/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		template.ValKindProjectRequest: endpoints(
			endpoint(`/oapi/v1/projectrequests`, POST(IgnoreConflicts)),
			endpoint(`/oapi/v1/projects/{{ index . "metadata" "name"}}`, GET(), DELETE())),

		template.ValKindRoleBinding: endpoints(
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/rolebindings`, POST(WhenConflictThenPatch)),
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/rolebindings/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		template.ValKindRoleBindingRestriction: endpoints(
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/rolebindingrestrictions`, POST(WhenConflictThenPatch)),
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/rolebindingrestrictions/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		template.ValKindRoute: endpoints(
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/routes`, POST(WhenConflictThenPatch)),
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/routes/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		template.ValKindDeployment: endpoints(
			endpoint(`/apis/extensions/v1beta1/namespaces/{{ index . "metadata" "namespace"}}/deployments`, POST(WhenConflictThenPatch)),
			endpoint(`/apis/extensions/v1beta1/namespaces/{{ index . "metadata" "namespace"}}/deployments/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		template.ValKindDeploymentConfig: endpoints(
			endpoint(`/apis/apps.openshift.io/v1/namespaces/{{ index . "metadata" "namespace"}}/deploymentconfigs`, POST(WhenConflictThenPatch)),
			endpoint(`/apis/apps.openshift.io/v1/namespaces/{{ index . "metadata" "namespace"}}/deploymentconfigs/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		template.ValKindPersistenceVolumeClaim: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/persistentvolumeclaims`, POST(IgnoreConflicts)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/persistentvolumeclaims/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		template.ValKindService: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/services`, POST(WhenConflictThenPatch)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/services/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		template.ValKindSecret: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/secrets`, POST(WhenConflictThenPatch)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/secrets/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		template.ValKindServiceAccount: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/serviceaccounts`, POST(IgnoreConflicts)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/serviceaccounts/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		template.ValKindConfigMap: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/configmaps`, POST(WhenConflictThenPatch)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/configmaps/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		template.ValKindResourceQuota: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/resourcequotas`, POST(WhenConflictThenPatch)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/resourcequotas/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		template.ValKindLimitRange: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/limitranges`, POST(WhenConflictThenPatch)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/limitranges/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		template.ValKindJob: endpoints(
			endpoint(`/apis/batch/v1/namespaces/{{ index . "metadata" "namespace"}}/jobs`, POST(WhenConflictThenPatch)),
			endpoint(`/apis/batch/v1/namespaces/{{ index . "metadata" "namespace"}}/jobs/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),
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

type Result struct {
	response *http.Response
	body       []byte
}

func newResult(response *http.Response, err error) (*Result, error) {
	return &Result{
		response: response,
		body:     utils.ReadBody(response),
	}, err
}

func (e *ObjectEndpoints) Apply(client *Client, object template.Object, action string) (*Result, error) {
	method, err := e.getMethodDefinition(action, object)
	if err != nil {
		return nil, nil
	}
	var (
		reqBody []byte
		result  *Result
	)
	if len(method.beforeCallbacks) != 0 {
		for _, beforeCallback := range method.beforeCallbacks {
			method, reqBody, err = beforeCallback(client, object, e, method)
			if err != nil {
				return nil, err
			}
		}
		result, err = newResult(client.Do(method.requestCreator, object, reqBody))
	} else {
		result, err = newResult(client.MarshalAndDo(method.requestCreator, object))
	}

	LogRequestInfo(client, object, method, result)
	if err != nil {
		return result, err
	}

	if len(method.afterCallbacks) == 0 {
		return result, checkHTTPCode(result, err)
	} else {
		for _, afterCallback := range method.afterCallbacks {
			err := afterCallback(client, object, e, method, result)
			if err != nil {
				return result, err
			}
		}
		return result, nil
	}
}

func (e *ObjectEndpoints) getMethodDefinition(method string, object template.Object) (*MethodDefinition, error) {
	methodDef, found := e.methods[method]
	if !found {
		return nil, fmt.Errorf("method definition %s for %s not supported", method, template.GetKind(object))
	}
	return methodDef, nil
}

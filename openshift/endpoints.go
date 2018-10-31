package openshift

import (
	"fmt"
	"github.com/fabric8-services/fabric8-tenant/environment"
	"gopkg.in/yaml.v2"
	"net/http"
)

type ObjectEndpoints struct {
	methods map[string]*MethodDefinition
}

var (
	objectEndpoints = map[string]*ObjectEndpoints{
		environment.ValKindNamespace: endpoints(
			endpoint(`/api/v1/namespaces`, POST(IgnoreConflicts, GetObject)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		environment.ValKindProject: endpoints(
			endpoint(`/oapi/v1/projects`, POST(WhenConflictThenDeleteAndRedo, GetObject)),
			endpoint(`/oapi/v1/projects/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		environment.ValKindProjectRequest: endpoints(
			endpoint(`/oapi/v1/projectrequests`, POST(IgnoreConflicts, GetObject)),
			endpoint(`/oapi/v1/projects/{{ index . "metadata" "name"}}`, PATCH().WithModifier(Skip), GET(), DELETE())),

		environment.ValKindRole: endpoints(
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/roles`, POST(WhenConflictThenDeleteAndRedo)),
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/roles/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		environment.ValKindRoleBinding: endpoints(
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/rolebindings`, POST(WhenConflictThenDeleteAndRedo)),
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/rolebindings/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		environment.ValKindRoleBindingRestriction: endpoints(
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/rolebindingrestrictions`, POST(WhenConflictThenDeleteAndRedo).WithModifier(NeedMasterToken)),
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/rolebindingrestrictions/{{ index . "metadata" "name"}}`,
				PATCH().WithModifier(NeedMasterToken), GET().WithModifier(NeedMasterToken), DELETE().WithModifier(NeedMasterToken))),

		environment.ValKindRoute: endpoints(
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/routes`, POST(WhenConflictThenDeleteAndRedo)),
			endpoint(`/oapi/v1/namespaces/{{ index . "metadata" "namespace"}}/routes/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		environment.ValKindDeployment: endpoints(
			endpoint(`/apis/extensions/v1beta1/namespaces/{{ index . "metadata" "namespace"}}/deployments`, POST(WhenConflictThenDeleteAndRedo)),
			endpoint(`/apis/extensions/v1beta1/namespaces/{{ index . "metadata" "namespace"}}/deployments/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		environment.ValKindDeploymentConfig: endpoints(
			endpoint(`/apis/apps.openshift.io/v1/namespaces/{{ index . "metadata" "namespace"}}/deploymentconfigs`, POST(WhenConflictThenDeleteAndRedo)),
			endpoint(`/apis/apps.openshift.io/v1/namespaces/{{ index . "metadata" "namespace"}}/deploymentconfigs/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		environment.ValKindPersistenceVolumeClaim: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/persistentvolumeclaims`, POST(IgnoreConflicts)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/persistentvolumeclaims/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		environment.ValKindService: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/services`, POST(WhenConflictThenDeleteAndRedo)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/services/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		environment.ValKindSecret: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/secrets`, POST(WhenConflictThenDeleteAndRedo)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/secrets/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		environment.ValKindServiceAccount: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/serviceaccounts`, POST(IgnoreConflicts)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/serviceaccounts/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		environment.ValKindConfigMap: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/configmaps`, POST(WhenConflictThenDeleteAndRedo)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/configmaps/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		environment.ValKindResourceQuota: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/resourcequotas`, POST(WhenConflictThenDeleteAndRedo, GetObject)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/resourcequotas/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		environment.ValKindLimitRange: endpoints(
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/limitranges`, POST(WhenConflictThenDeleteAndRedo)),
			endpoint(`/api/v1/namespaces/{{ index . "metadata" "namespace"}}/limitranges/{{ index . "metadata" "name"}}`, PATCH(), GET(), DELETE())),

		environment.ValKindJob: endpoints(
			endpoint(`/apis/batch/v1/namespaces/{{ index . "metadata" "namespace"}}/jobs`, POST(WhenConflictThenDeleteAndRedo)),
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
	body     []byte
}

func (r *Result) bodyToObject() (environment.Object, error) {
	var obj environment.Object
	err := yaml.Unmarshal(r.body, &obj)
	return obj, err
}

func (e *ObjectEndpoints) Apply(client *Client, object environment.Object, action string) (*Result, error) {
	method, err := e.getMethodDefinition(action, object)
	if err != nil {
		return nil, nil
	}
	var (
		reqBody []byte
		result  *Result
	)
	if len(method.beforeDoCallbacks) != 0 {
		for _, beforeCallback := range method.beforeDoCallbacks {
			method, reqBody, err = beforeCallback(client, object, e, method)
			if err != nil {
				return nil, err
			}
		}
	} else {
		reqBody, err = yaml.Marshal(object)
		if err != nil {
			return nil, err
		}
	}

	if method.requestCreator.skip {
		return nil, nil
	}

	// do the request
	result, err = client.Do(method.requestCreator, object, reqBody)

	// if error occurred and no response was retrieved (probably error before doing a request)
	if err != nil && result == nil {
		LogRequest(object, method, result).WithFields(map[string]interface{}{
			"err": err,
		}).Error("unable request resource")
		return result, err
	}
	//LogRequest(object, method, result).Info("resource requested")

	if len(method.afterDoCallbacks) == 0 {
		return result, checkHTTPCode(result, err)
	} else {
		for _, afterCallback := range method.afterDoCallbacks {
			err := afterCallback(client, object, e, method, result)
			if err != nil {
				return result, err
			}
		}
		return result, nil
	}
}

func (e *ObjectEndpoints) getMethodDefinition(method string, object environment.Object) (*MethodDefinition, error) {
	methodDef, found := e.methods[method]
	if !found {
		return nil, fmt.Errorf("method definition %s for %s not supported", method, environment.GetKind(object))
	}
	return methodDef, nil
}

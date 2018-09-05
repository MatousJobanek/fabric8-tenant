package openshift

import (
	"github.com/fabric8-services/fabric8-tenant/environment"
	"net/http"
	"fmt"
	"gopkg.in/yaml.v2"
	"github.com/fabric8-services/fabric8-tenant/utils"
	log "github.com/sirupsen/logrus"
)

type BeforeCallback func(client *Client, object environment.Object, objEndpoints *ObjectEndpoints, method *MethodDefinition) (*MethodDefinition, []byte, error)
type AfterCallback func(client *Client, object environment.Object, objEndpoints *ObjectEndpoints, method *MethodDefinition, result *Result) error

// Before callbacks

func GetObjectAndMerge(client *Client, object environment.Object, objEndpoints *ObjectEndpoints, method *MethodDefinition) (*MethodDefinition, []byte, error) {
	result, err := objEndpoints.Apply(client, object, http.MethodGet)
	if err != nil {
		if result.response.StatusCode == http.StatusNotFound {
			return getMethodAndMarshalObject(objEndpoints, http.MethodPost, object)
		}
		return nil, nil, err
	}
	modifiedJson, err := utils.MarshalYAMLToJSON(object)
	if err != nil {
		return nil, nil, err
	}
	return method, modifiedJson, nil
}

func getMethodAndMarshalObject(objEndpoints *ObjectEndpoints, method string, object environment.Object) (*MethodDefinition, []byte, error) {
	post, err := objEndpoints.getMethodDefinition(method, object)
	if err != nil {
		return nil, nil, err
	}
	bytes, err := yaml.Marshal(object)
	if err != nil {
		return nil, nil, err
	}
	return post, bytes, nil
}

// After callbacks

func WhenConflictThenPatch(client *Client, object environment.Object, objEndpoints *ObjectEndpoints, method *MethodDefinition, result *Result) error {
	if result.response.StatusCode == http.StatusConflict {
		return checkHTTPCode(objEndpoints.Apply(client, object, http.MethodPatch))
	}
	return checkHTTPCode(result, nil)
}

func IgnoreConflicts(client *Client, object environment.Object, objEndpoints *ObjectEndpoints, method *MethodDefinition, result *Result) error {
	if result.response.StatusCode == http.StatusConflict {
		return nil
	}
	return checkHTTPCode(result, nil)
}

func GetObject(client *Client, object environment.Object, objEndpoints *ObjectEndpoints, method *MethodDefinition, result *Result) error {
	// todo - shouldn't we check the response codes here as well?
	_, err := objEndpoints.Apply(client, object, http.MethodGet)
	return err
}

func GetObjectExpects404(client *Client, object environment.Object, endpoint *ObjectEndpoints, method *MethodDefinition, result *Result) error {
	if result.response.StatusCode == http.StatusNotFound {
		return nil
	}
	err := checkHTTPCode(result, nil)
	if err != nil {
		return err
	}
	get, err := endpoint.getMethodDefinition(http.MethodGet, object)
	if err != nil {
		return err
	}
	getResponse, err := client.MarshalAndDo(get.requestCreator, object)
	if getResponse.StatusCode != http.StatusNotFound {
		err = checkHTTPCode(newResult(getResponse, err))
		if err == nil {
			return fmt.Errorf("object %s wasn't removed", object)
		}
	}
	return err
}

func checkHTTPCode(result *Result, e error) error {
	if e == nil && result.response != nil && (result.response.StatusCode < 200 || result.response.StatusCode >= 300) {
		return fmt.Errorf("server responded with status: %d for the request %s %s with the body %s",
			result.response.StatusCode, result.response.Request.Method, result.response.Request.URL, result.body)
	}
	return e
}

func LogRequestInfo(client *Client, object environment.Object, method *MethodDefinition, result *Result) error {
	log.WithFields(map[string]interface{}{
		"status":      result.response.StatusCode,
		"method":      method.action,
		"cluster_url": result.response.Request.URL,
		"namespace":   environment.GetNamespace(object),
		"name":        environment.GetName(object),
		"kind":        environment.GetKind(object),
		"request":     yamlString(object),
		"response":    result,
	}).Info("resource requested")
	return nil
}

func yamlString(data environment.Object) string {
	b, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Sprintf("Could not marshal yaml %v", data)
	}
	return string(b)
}

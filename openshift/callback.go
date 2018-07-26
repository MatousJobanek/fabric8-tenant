package openshift

import (
	"github.com/fabric8-services/fabric8-tenant/template"
	"net/http"
	"fmt"
	"gopkg.in/yaml.v2"
	"github.com/fabric8-services/fabric8-tenant/utils"
)

type Callback func(client *Client, object template.Object, endpoints *ObjectEndpoints, method *MethodDefinition, response *http.Response) error

func WhenConflictThenDeleteAndRetry(client *Client, object template.Object, endpoints *ObjectEndpoints, method *MethodDefinition, response *http.Response) error {
	if response.StatusCode == http.StatusConflict {
		err := endpoints.ApplyWithMethodCallback(client, object, http.MethodDelete)
		if err != nil {
			return err
		}
		return checkHTTPCode(client.Do(method.requestCreator, object))
	}
	return checkHTTPCode(response, nil)
}

func IgnoreConflicts(client *Client, object template.Object, endpoints *ObjectEndpoints, method *MethodDefinition, response *http.Response) error {
	if response.StatusCode == http.StatusConflict {
		return nil
	}
	return checkHTTPCode(response, nil)
}

func GetObject(client *Client, object template.Object, endpoints *ObjectEndpoints, method *MethodDefinition, response *http.Response) error {
	// todo - shouldn't we check the response codes here as well?
	return endpoints.ApplyWithMethodCallback(client, object, http.MethodGet)
}

func GetObjectExpects404(client *Client, object template.Object, endpoint *ObjectEndpoints, method *MethodDefinition, response *http.Response) error {
	err := checkHTTPCode(response, nil)
	if err != nil {
		return err
	}
	get, err := endpoint.getMethodDefinition(http.MethodGet, object)
	if err != nil {
		return err
	}
	getResponse, err := client.Do(get.requestCreator, object)
	if getResponse.StatusCode != http.StatusNotFound {
		err = checkHTTPCode(getResponse, err)
		if err == nil {
			return fmt.Errorf("obbject %s wasn't removed", object)
		}
	}
	return err
}

func checkHTTPCode(response *http.Response, e error) error {
	if e == nil && response != nil && (response.StatusCode < 200 || response.StatusCode >= 300) {
		return fmt.Errorf("server responded with status: %s", response.StatusCode)
	}
	return e
}

func LogRequestInfo(client *Client, object template.Object, endpoints *ObjectEndpoints, method *MethodDefinition, response *http.Response) error {
	client.Log.WithFields(map[string]interface{}{
		"status":      response.StatusCode,
		"method":      method.action,
		"cluster_url": response.Request.URL,
		"namespace":   template.GetNamespace(object),
		"name":        template.GetName(object),
		"kind":        template.GetKind(object),
		"request":     yamlString(object),
		"response":    utils.ReadBody(response),
	}).Info("resource requested")
	return nil
}

func yamlString(data map[interface{}]interface{}) string {
	b, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Sprintf("Could not marshal yaml %v", data)
	}
	return string(b)
}
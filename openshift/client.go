package openshift

import (
	"net/http"
	"bytes"
	"net/http/httputil"
	"fmt"
	"gopkg.in/yaml.v2"
	"crypto/tls"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	"github.com/fabric8-services/fabric8-tenant/template"
	"strings"

	tmpl "html/template"
	"github.com/fabric8-services/fabric8-tenant/utils"
	"github.com/fabric8-services/fabric8-tenant/log"
	"github.com/fabric8-services/fabric8-tenant/controller"
)

type Client struct {
	client        *http.Client
	MasterURL     string
	Token         string
	HTTPTransport http.RoundTripper
	Log           log.Logger
}

type WithClientBuilder struct {
	Client *Client
}

type ClientWithObjectsBuilder struct {
	client  *Client
	objects template.Objects
}

func NewClient(log log.Logger, cluster controller.Cluster, config *configuration.Data) *WithClientBuilder {
	httpTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.APIServerInsecureSkipTLSVerify(),
		},
	}

	return NewClientWithTransport(log, cluster.APIURL, cluster.Token, config, httpTransport)
}

func NewClientWithTransport(log log.Logger, clusterURL, token string, config *configuration.Data, httpTransport http.RoundTripper) *WithClientBuilder {
	return &WithClientBuilder{Client: &Client{
		client:        createHTTPClient(httpTransport),
		MasterURL:     clusterURL,
		Token:         token,
		HTTPTransport: httpTransport,
		Log:           log,
	}}
}

// CreateHTTPClient returns an HTTP client with the options settings,
// or a default HTTP client if nothing was specified
func createHTTPClient(HTTPTransport http.RoundTripper) *http.Client {
	if HTTPTransport != nil {
		return &http.Client{
			Transport: HTTPTransport,
		}
	}
	return http.DefaultClient
}

func (b *WithClientBuilder) ApplyAll(objects template.Objects) *ClientWithObjectsBuilder {
	return &ClientWithObjectsBuilder{
		client:  b.Client,
		objects: objects,
	}
}

func (b *ClientWithObjectsBuilder) WithPostMethod() error {
	return applyAll(b.client, http.MethodPost, b.objects)
}

func (b *ClientWithObjectsBuilder) WithPatchMethod() error {
	return applyAll(b.client, http.MethodPatch, b.objects)
}

func (b *ClientWithObjectsBuilder) WithPutMethod() error {
	return applyAll(b.client, http.MethodPut, b.objects)
}

func (b *ClientWithObjectsBuilder) WithGetMethod() error {
	return applyAll(b.client, http.MethodGet, b.objects)
}

func (b *ClientWithObjectsBuilder) WithDeleteMethod() error {
	return applyAll(b.client, http.MethodDelete, b.objects)
}

func applyAll(client *Client, action string, objects []map[interface{}]interface{}) error {
	for _, object := range objects {
		objectEndpoint, found := objectEndpoints[template.GetKind(object)]
		if !found {
			return fmt.Errorf("there is no supported endpoint for the object %s", template.GetKind(object))
		}
		err := objectEndpoint.ApplyWithMethodCallback(client, object, action)
		if err != nil {
			return err
		}
	}
	return nil
}

type urlCreator func(urlTemplate string) func() (URL string, err error)

type RequestCreator struct {
	creator func(urlCreator urlCreator, body []byte) (*http.Request, error)
}

func (c *Client) Do(requestCreator RequestCreator, object template.Object) (*http.Response, error) {
	req, err := requestCreator.createRequestFor(c.MasterURL, object)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	// for debug only
	if false {
		rb, _ := httputil.DumpRequest(req, true)
		fmt.Println("-----------------")
		fmt.Println(string(rb))
		fmt.Println(object)
		fmt.Println("================")
		fmt.Println(utils.ReadBody(resp))
		fmt.Println("-----------------")
		fmt.Println(resp.StatusCode)
	}
	return resp, err
}

func (c *RequestCreator) createRequestFor(masterURL string, object map[interface{}]interface{}) (*http.Request, error) {
	body, err := yaml.Marshal(object)
	if err != nil {
		return nil, err
	}

	urlCreator := func(urlTemplate string) func() (string, error) {
		return func() (string, error) {
			return createURL(masterURL, urlTemplate, object)
		}
	}

	return c.creator(urlCreator, body)
}

func createURL(hostURL, urlTemplate string, object map[interface{}]interface{}) (string, error) {
	target, err := tmpl.New("url").Parse(urlTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = target.Execute(&buf, object)
	if err != nil {
		return "", err
	}
	str := buf.String()
	if strings.HasSuffix(hostURL, "/") {
		hostURL = hostURL[0 : len(hostURL)-1]
	}

	return hostURL + str, nil
}

func newDefaultRequest(action string, createURL func() (string, error), body []byte) (*http.Request, error) {
	url, err := createURL()
	//fmt.Println(action)
	//fmt.Println(url)
	if url == "" {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(action, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/yaml")
	req.Header.Set("Content-Type", "application/yaml")
	return req, nil
}

package openshift

import (
	"github.com/fabric8-services/fabric8-tenant/environment"
	"net/http"
	"fmt"
	"github.com/fabric8-services/fabric8-tenant/log"
	"net/http/httputil"
	"github.com/fabric8-services/fabric8-tenant/utils"
	"bytes"
	"strings"

	tmpl "html/template"

	"gopkg.in/yaml.v2"
)

type Client struct {
	client    *http.Client
	Log       log.Logger
	MasterURL string
	Token     string
}

func newClient(log log.Logger, httpTransport http.RoundTripper, masterURL string, token string) *Client {
	return &Client{
		client:    createHTTPClient(httpTransport),
		Log:       log,
		MasterURL: masterURL,
		Token:     token,
	}
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

type urlCreator func(urlTemplate string) func() (URL string, err error)

type RequestCreator struct {
	creator func(urlCreator urlCreator, body []byte) (*http.Request, error)
}

func (c *Client) MarshalAndDo(requestCreator RequestCreator, object environment.Object) (*http.Response, error) {
	body, err := yaml.Marshal(object)
	if err != nil {
		return nil, err
	}
	return c.Do(requestCreator, object, body)
}

func (c *Client) Do(requestCreator RequestCreator, object environment.Object, body []byte) (*http.Response, error) {
	req, err := requestCreator.createRequestFor(c.MasterURL, object, body)
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

func (c *RequestCreator) createRequestFor(masterURL string, object environment.Object, body []byte) (*http.Request, error) {
	urlCreator := func(urlTemplate string) func() (string, error) {
		return func() (string, error) {
			return createURL(masterURL, urlTemplate, object)
		}
	}

	return c.creator(urlCreator, body)
}

func createURL(hostURL, urlTemplate string, object environment.Object) (string, error) {
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

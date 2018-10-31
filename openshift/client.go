package openshift

import (
	"bytes"
	"fmt"
	"github.com/fabric8-services/fabric8-tenant/utils"
	"net/http"
	"net/http/httputil"
	"strings"

	tmpl "html/template"

	"github.com/fabric8-services/fabric8-tenant/environment"
	"gopkg.in/yaml.v2"
)

type Client struct {
	client        *http.Client
	MasterURL     string
	TokenProducer TokenProducer
}
type TokenProducer func(forceMasterToken bool) string

func newClient(httpTransport http.RoundTripper, masterURL string, TokenProducer TokenProducer) *Client {
	return &Client{
		client:        createHTTPClient(httpTransport),
		MasterURL:     masterURL,
		TokenProducer: TokenProducer,
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
	creator         func(urlCreator urlCreator, body []byte) (*http.Request, error)
	needMasterToken bool
	skip            bool
}

func (c *Client) MarshalAndDo(requestCreator RequestCreator, object environment.Object) (*Result, error) {
	body, err := yaml.Marshal(object)
	if err != nil {
		return nil, err
	}
	return c.Do(requestCreator, object, body)
}

func (c *Client) Do(requestCreator RequestCreator, object environment.Object, body []byte) (*Result, error) {

	req, err := requestCreator.createRequestFor(c.MasterURL, object, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.TokenProducer(requestCreator.needMasterToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	readBody, toDefer := utils.ReadBody(resp)
	defer toDefer()

	// for debug only
	if false {
		rb, _ := httputil.DumpRequest(req, true)
		fmt.Println("-----------------")
		fmt.Println(string(rb))
		fmt.Println(object)
		fmt.Println("================")
		fmt.Println(readBody())
		fmt.Println("-----------------")
		fmt.Println(resp.StatusCode)
	}

	return newResult(resp, readBody(), err)
}

func newResult(response *http.Response, body []byte, err error) (*Result, error) {
	return &Result{
		response: response,
		body:     body,
	}, err
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

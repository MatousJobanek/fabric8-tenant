package openshift

import (
	"net/http"
)

type MethodDefinition struct {
	action         string
	callbacks      []Callback
	requestCreator RequestCreator
}

func NewMethodDefinition(action string, callbacks []Callback, requestCreator RequestCreator) *MethodDefinition {
	return &MethodDefinition{
		action:         action,
		callbacks:      callbacks,
		requestCreator: requestCreator,
	}
}

type methodDefCreator func(endpoint string) *MethodDefinition

func POST(callbacks ...Callback) methodDefCreator {
	return func(urlTemplate string) *MethodDefinition {
		return NewMethodDefinition(
			http.MethodPost,
			append(callbacks, GetObject),
			RequestCreator{creator: func(urlCreator urlCreator, body []byte) (*http.Request, error) {
				return newDefaultRequest(http.MethodPost, urlCreator(urlTemplate), body)
			}})
	}
}

func PUT(callbacks ...Callback) methodDefCreator {
	return func(urlTemplate string) *MethodDefinition {
		return NewMethodDefinition(
			http.MethodPut,
			callbacks,
			RequestCreator{creator: func(urlCreator urlCreator, body []byte) (*http.Request, error) {
				return newDefaultRequest(http.MethodPut, urlCreator(urlTemplate), body)
			}})
	}
}
func PATCH(callbacks ...Callback) methodDefCreator {
	return func(urlTemplate string) *MethodDefinition {
		return NewMethodDefinition(
			http.MethodPatch,
			callbacks,
			RequestCreator{creator: func(urlCreator urlCreator, body []byte) (*http.Request, error) {
				req, err := newDefaultRequest(http.MethodPatch, urlCreator(urlTemplate), body)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", "application/merge-patch+json")
				return req, err
			}})
	}
}
func GET(callbacks ...Callback) methodDefCreator {
	return func(urlTemplate string) *MethodDefinition {
		return NewMethodDefinition(
			http.MethodGet,
			callbacks,
			RequestCreator{creator: func(urlCreator urlCreator, body []byte) (*http.Request, error) {
				return newDefaultRequest(http.MethodGet, urlCreator(urlTemplate), body)
			}})
	}
}

func DELETE(callbacks ...Callback) methodDefCreator {
	return func(urlTemplate string) *MethodDefinition {
		return NewMethodDefinition(
			http.MethodDelete,
			append(callbacks, GetObjectExpects404),
			RequestCreator{creator: func(urlCreator urlCreator, body []byte) (*http.Request, error) {
				body = []byte(deleteOptions)
				return newDefaultRequest(http.MethodDelete, urlCreator(urlTemplate), body)
			}})
	}
}





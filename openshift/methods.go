package openshift

import (
	"net/http"
)

type MethodDefinition struct {
	action          string
	beforeCallbacks []BeforeCallback
	afterCallbacks  []AfterCallback
	requestCreator  RequestCreator
}

func NewMethodDefinition(action string, beforeCallbacks []BeforeCallback, afterCallbacks []AfterCallback, requestCreator RequestCreator) *MethodDefinition {
	return &MethodDefinition{
		action:          action,
		beforeCallbacks: beforeCallbacks,
		afterCallbacks:  afterCallbacks,
		requestCreator:  requestCreator,
	}
}

type methodDefCreator func(endpoint string) *MethodDefinition

func POST(afterCallbacks ...AfterCallback) methodDefCreator {
	return func(urlTemplate string) *MethodDefinition {
		return NewMethodDefinition(
			http.MethodPost,
			[]BeforeCallback{},
			append(afterCallbacks, GetObject),
			RequestCreator{creator: func(urlCreator urlCreator, body []byte) (*http.Request, error) {
				return newDefaultRequest(http.MethodPost, urlCreator(urlTemplate), body)
			}})
	}
}

func PUT(afterCallbacks ...AfterCallback) methodDefCreator {
	return func(urlTemplate string) *MethodDefinition {
		return NewMethodDefinition(
			http.MethodPut,
			[]BeforeCallback{},
			afterCallbacks,
			RequestCreator{creator: func(urlCreator urlCreator, body []byte) (*http.Request, error) {
				return newDefaultRequest(http.MethodPut, urlCreator(urlTemplate), body)
			}})
	}
}
func PATCH(afterCallbacks ...AfterCallback) methodDefCreator {
	return func(urlTemplate string) *MethodDefinition {
		return NewMethodDefinition(
			http.MethodPatch,
			[]BeforeCallback{GetObjectAndMerge},
			append(afterCallbacks, GetObject),
			RequestCreator{creator: func(urlCreator urlCreator, body []byte) (*http.Request, error) {
				req, err := newDefaultRequest(http.MethodPatch, urlCreator(urlTemplate), body)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", "application/strategic-merge-patch+json")
				return req, err
			}})
	}
}
func GET(afterCallbacks ...AfterCallback) methodDefCreator {
	return func(urlTemplate string) *MethodDefinition {
		return NewMethodDefinition(
			http.MethodGet,
			[]BeforeCallback{},
			afterCallbacks,
			RequestCreator{creator: func(urlCreator urlCreator, body []byte) (*http.Request, error) {
				return newDefaultRequest(http.MethodGet, urlCreator(urlTemplate), body)
			}})
	}
}

func DELETE(afterCallbacks ...AfterCallback) methodDefCreator {
	return func(urlTemplate string) *MethodDefinition {
		return NewMethodDefinition(
			http.MethodDelete,
			[]BeforeCallback{},
			append(afterCallbacks, GetObjectExpects404),
			RequestCreator{creator: func(urlCreator urlCreator, body []byte) (*http.Request, error) {
				body = []byte(deleteOptions)
				return newDefaultRequest(http.MethodDelete, urlCreator(urlTemplate), body)
			}})
	}
}

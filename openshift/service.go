package openshift

import (
	"net/http"
	"crypto/tls"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	"github.com/fabric8-services/fabric8-tenant/log"
	"github.com/fabric8-services/fabric8-tenant/cluster"
	"github.com/satori/go.uuid"
	"sync"
	"github.com/fabric8-services/fabric8-tenant/environment"
	"sort"
	"time"
	"fmt"
	"github.com/arquillian/ike-prow-plugins/pkg/retry"
	"github.com/fabric8-services/fabric8-tenant/auth"
)

type ServiceBuilder struct {
	service *Service
}

type Service struct {
	httpTransport *http.Transport
	nsTypes       []string
	log           log.Logger
	context       *ServiceContext
}

func NewService(log log.Logger, context *ServiceContext) *ServiceBuilder {
	httpTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: context.config.APIServerInsecureSkipTLSVerify(),
		},
	}
	return NewBuilderWithTransport(log, context, httpTransport)
}

func NewBuilderWithTransport(log log.Logger, context *ServiceContext, transport *http.Transport) *ServiceBuilder {
	return &ServiceBuilder{service: &Service{
		httpTransport: transport,
		log:           log,
		context:       context,
	}}
}

type ServiceContext struct {
	clusterMapping map[string]*cluster.Cluster
	user           *auth.User
	config         *configuration.Data
	space          *uuid.UUID
}

func NewServiceContext(config *configuration.Data, clusterMapping map[string]*cluster.Cluster, user *auth.User, space *uuid.UUID) *ServiceContext {
	return &ServiceContext{
		clusterMapping: clusterMapping,
		user:           user,
		config:         config,
		space:          space,
	}
}

func (b *ServiceBuilder) ApplyAll(nsTypes []string) *Service {
	b.service.nsTypes = nsTypes
	return b.service
}

func (s *Service) WithPostMethod() error {
	return s.processAndApplyAll(http.MethodPost)
}

func (s *Service) WithPatchMethod() error {
	return s.processAndApplyAll(http.MethodPatch)
}

func (s *Service) WithPutMethod() error {
	return s.processAndApplyAll(http.MethodPut)
}

func (s *Service) WithGetMethod() error {
	return s.processAndApplyAll(http.MethodGet)
}

func (s *Service) WithDeleteMethod() error {
	return s.processAndApplyAll(http.MethodDelete)
}

func (s *Service) processAndApplyAll(action string) error {
	var nsTypesWait sync.WaitGroup
	nsTypesWait.Add(len(s.nsTypes))

	for _, nsType := range s.nsTypes {
		go processAndApply(&nsTypesWait, s.log, *s.context, nsType, action, *s.httpTransport)
	}
	nsTypesWait.Wait()
	return nil
}

func processAndApply(nsTypeWait *sync.WaitGroup, log log.Logger, context ServiceContext, nsType string, action string, transport http.Transport) {
	defer nsTypeWait.Done()

	env, err := environment.NewService(context.config).GetEnvData(context.space, nsType)
	vars := environment.CollectVars(context.user.OpenshiftUserName, context.config)
	objects, err := env.Template.Process(vars)
	if err != nil {
		log.Error(err)
		return
	}

	if action == http.MethodDelete {
		sort.Reverse(environment.ByKind(objects))
	} else {
		sort.Sort(environment.ByKind(objects))
	}

	cluster := context.clusterMapping[nsType]
	client := newClient(log, &transport, cluster.APIURL, cluster.Token)

	var objectsWait sync.WaitGroup
	objectsWait.Add(len(objects))

	for _, object := range objects {
		go apply(&objectsWait, *client, action, object)
	}
	objectsWait.Wait()
}

func apply(objectsWait *sync.WaitGroup, client Client, action string, object environment.Object) {
	defer objectsWait.Done()
	errs := retry.Do(5, time.Millisecond*50, func() error {
		objectEndpoint, found := objectEndpoints[environment.GetKind(object)]
		if !found {
			return fmt.Errorf("there is no supported endpoint for the object %s", environment.GetKind(object))

		}
		_, err := objectEndpoint.Apply(&client, object, action)
		return err
	})
	if len(errs) != 0 {
		client.Log.Error(errs)
	}
}

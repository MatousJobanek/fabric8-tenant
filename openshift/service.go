package openshift

import (
	"net/http"
	"crypto/tls"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	"github.com/fabric8-services/fabric8-tenant/cluster"
	"github.com/satori/go.uuid"
	"sync"
	"github.com/fabric8-services/fabric8-tenant/environment"
	"sort"
	"time"
	"fmt"
	"github.com/fabric8-services/fabric8-tenant/retry"
	"github.com/fabric8-services/fabric8-tenant/dbsupport"
	log "github.com/sirupsen/logrus"
)

type ServiceBuilder struct {
	service *Service
}

type Service struct {
	httpTransport       http.RoundTripper
	nsTypes             []string
	context             *ServiceContext
	namespaceRepository dbsupport.NamespaceRepository
}

func NewService(context *ServiceContext, namespaceRepository dbsupport.NamespaceRepository) *ServiceBuilder {
	httpTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: context.config.APIServerInsecureSkipTLSVerify(),
		},
	}
	return NewBuilderWithTransport(context, namespaceRepository, httpTransport)
}

func NewBuilderWithTransport(context *ServiceContext, nsRepo dbsupport.NamespaceRepository, transport http.RoundTripper) *ServiceBuilder {
	return &ServiceBuilder{service: &Service{
		httpTransport:       transport,
		context:             context,
		namespaceRepository: nsRepo,
	}}
}

type ServiceContext struct {
	clusterMapping    ClusterMapping
	openShiftUsername string
	config            *configuration.Data
	space             *uuid.UUID
}

type ClusterMapping map[string]*cluster.Cluster

func NewServiceContext(config *configuration.Data, clusterMapping ClusterMapping, openShiftUsername string, space *uuid.UUID) *ServiceContext {
	return &ServiceContext{
		clusterMapping:    clusterMapping,
		openShiftUsername: openShiftUsername,
		config:            config,
		space:             space,
	}
}

func (b *ServiceBuilder) ApplyAll(nsTypes []string) *Service {
	b.service.nsTypes = nsTypes
	return b.service
}

func (b *ServiceBuilder) Apply(nsType string) *Service {
	b.service.nsTypes = []string{nsType}
	return b.service
}

func (s *Service) WithPostMethod() error {
	return s.processAndApplyAll(&create{})
}

func (s *Service) WithPatchMethod() error {
	return s.processAndApplyAll(&update{})
}

//func (s *Service) WithGetMethod() error {
//	return s.processAndApplyAll(&get{})
//}

func (s *Service) WithDeleteMethod() error {
	return s.processAndApplyAll(&delete{})
}

func (s *Service) processAndApplyAll(action NamespaceAction) error {
	var nsTypesWait sync.WaitGroup
	nsTypesWait.Add(len(s.nsTypes))

	for _, nsType := range s.nsTypes {
		go processAndApplyNs(&nsTypesWait, *s.context, nsType, action, s.httpTransport, s.namespaceRepository)
	}
	nsTypesWait.Wait()
	return nil
}

func processAndApplyNs(nsTypeWait *sync.WaitGroup, context ServiceContext, nsType string, action NamespaceAction, transport http.RoundTripper, nsRepo dbsupport.NamespaceRepository) {
	defer nsTypeWait.Done()

	namespace, err := action.getNamespaceEntity(context.space, nsType, nsRepo)
	if err != nil {
		log.Error(err)
		return
	}

	env, err := environment.NewService(context.config).GetEnvData(context.space, nsType)
	vars := environment.CollectVars(context.openShiftUsername, context.config)
	objects, err := env.Template.Process(vars)
	if err != nil {
		log.Error(err)
		return
	}

	if action.methodName() == http.MethodDelete {
		sort.Reverse(environment.ByKind(objects))
	} else {
		sort.Sort(environment.ByKind(objects))
	}

	cluster := context.clusterMapping[nsType]
	client := newClient(transport, cluster.APIURL, cluster.Token)

	var objectsWait sync.WaitGroup
	objectsWait.Add(len(objects))

	for _, object := range objects {
		go apply(&objectsWait, *client, action.methodName(), object)
	}

	objectsWait.Wait()
	action.updateTable(env, cluster, namespace, nsRepo)
}

func apply(objectsWait *sync.WaitGroup, client Client, action string, object environment.Object) {
	defer objectsWait.Done()

	objectEndpoint, found := objectEndpoints[environment.GetKind(object)]
	if !found {
		err := fmt.Errorf("there is no supported endpoint for the object %s", environment.GetKind(object))
		log.Error(err)
		return
	}

	errs := retry.Do(5, time.Millisecond*50, func() error {
		_, err := objectEndpoint.Apply(&client, object, action)
		return err
	})

	if len(errs) != 0 {
		log.Error(errs)
	}
}

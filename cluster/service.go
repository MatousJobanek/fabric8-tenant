package cluster

import (
	"context"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/fabric8-services/fabric8-tenant/auth"
	authclient "github.com/fabric8-services/fabric8-tenant/auth/client"
	"github.com/fabric8-services/fabric8-tenant/openshift"
	"github.com/fabric8-services/fabric8-wit/log"
	"github.com/pkg/errors"
	"fmt"
)

// Cluster a cluster
type Cluster struct {
	APIURL            string
	ConsoleURL        string
	MetricsURL        string
	LoggingURL        string
	AppDNS            string
	CapacityExhausted bool

	User  string
	Token string
}

func cleanURL(url string) string {
	if !strings.HasSuffix(url, "/") {
		return url + "/"
	}
	return url
}

// Service the interface for the cluster service
type Service interface {
	GetCluster(context.Context, string) (*Cluster, error)
	GetClusters(ctx context.Context) []Cluster
	Stats() Stats
	Stop()
}

// Stats some stats about the cached data, for verifying during the tests, at first.
type Stats struct {
	CacheHits      int
	CacheMissed    int
	CacheRefreshes int
}

// NewClusterService creates a Resolver that rely on the Auth service to retrieve tokens
func NewClusterService(refreshInt time.Duration, authClientService *auth.Service) (Service, error) {
	// setup a ticker to refresh the cluster cache at regular intervals
	cacheRefresher := time.NewTicker(refreshInt)
	s := &clusterService{
		authClientService: authClientService,
		cacheRefresher:    cacheRefresher,
		cacheRefreshLock:  &sync.RWMutex{},
	}
	// immediately load the list of clusters before returning
	err := s.refreshCache(context.Background())
	if err != nil {
		log.Error(nil, map[string]interface{}{"error_message": err}, "failed to load the list of clusters during service initialization")
		return nil, err
	}
	go func() {
		for range cacheRefresher.C { // while the `cacheRefresh` ticker is running
			s.refreshCache(context.Background())
		}
	}()
	return s, nil
}

type clusterService struct {
	authClientService *auth.Service
	cacheRefresher    *time.Ticker
	cacheRefreshLock  *sync.RWMutex
	cacheHits         int
	cacheMissed       int
	cacheRefreshes    int
	cachedClusters    []Cluster
}

func (s *clusterService) GetCluster(ctx context.Context, target string) (*Cluster, error) {
	for _, cluster := range s.GetClusters(ctx) {
		if cleanURL(target) == cleanURL(cluster.APIURL) {
			return &cluster, nil
		}
	}
	return nil, fmt.Errorf("unable to resolve cluster")

}

func (s *clusterService) GetClusters(ctx context.Context) []Cluster {
	s.cacheRefreshLock.RLock()
	log.Debug(ctx, nil, "read lock acquired")
	clusters := make([]Cluster, len(s.cachedClusters))
	copy(clusters, s.cachedClusters)
	s.cacheRefreshLock.RUnlock()
	log.Debug(ctx, nil, "read lock released")
	return clusters

}

func (s *clusterService) Stop() {
	s.cacheRefresher.Stop()
}

func (s *clusterService) refreshCache(ctx context.Context) error {
	log.Debug(ctx, nil, "refreshing cached list of clusters...")
	defer log.Debug(ctx, nil, "refreshed cached list of clusters.")
	s.cacheRefreshes = s.cacheRefreshes + 1
	client, err := s.authClientService.NewSaClient()

	res, err := client.ShowClusters(ctx, authclient.ShowClustersPath())
	if err != nil {
		return errors.Wrapf(err, "error while doing the request")
	}
	defer func() {
		ioutil.ReadAll(res.Body)
		res.Body.Close()
	}()

	validationerror := auth.ValidateResponse(ctx, client, res)
	if validationerror != nil {
		return errors.Wrapf(validationerror, "error from server %q", s.authClientService.GetAuthURL())
	}

	clusters, err := client.DecodeClusterList(res)
	if err != nil {
		return errors.Wrapf(err, "error from server %q", s.authClientService.GetAuthURL())
	}

	var cls []Cluster
	for _, cluster := range clusters.Data {
		// resolve/obtain the cluster token
		clusterUser, clusterToken, err :=
			s.authClientService.ResolveSaToken(ctx, cluster.APIURL)
		if err != nil {
			return errors.Wrapf(err, "Unable to resolve token for cluster %v", cluster.APIURL)
		}
		// verify the token
		_, err = openshift.WhoAmI(ctx, cluster.APIURL, clusterToken, s.authClientService.ClientOptions...)
		if err != nil {
			return errors.Wrapf(err, "token retrieved for cluster %v is invalid", cluster.APIURL)
		}

		cls = append(cls, Cluster{
			APIURL:            cluster.APIURL,
			AppDNS:            cluster.AppDNS,
			ConsoleURL:        cluster.ConsoleURL,
			MetricsURL:        cluster.MetricsURL,
			LoggingURL:        cluster.LoggingURL,
			CapacityExhausted: cluster.CapacityExhausted,

			User:  clusterUser,
			Token: clusterToken,
		})
	}
	// lock to avoid concurrent writes
	s.cacheRefreshLock.Lock()
	log.Debug(ctx, nil, "write lock acquired")
	s.cachedClusters = cls // only replace at the end of this function and within a Write lock scope, i.e., when all retrieved clusters have been processed
	s.cacheRefreshLock.Unlock()
	log.Debug(ctx, nil, "write lock released")
	return nil
}

func (s *clusterService) Stats() Stats {
	return Stats{
		CacheHits:      s.cacheHits,
		CacheMissed:    s.cacheMissed,
		CacheRefreshes: s.cacheRefreshes,
	}
}

package cluster_test

import (
	"context"
	"fmt"
	"gopkg.in/h2non/gock.v1"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/fabric8-services/fabric8-common/convert/ptr"
	"github.com/fabric8-services/fabric8-tenant/auth"
	authclient "github.com/fabric8-services/fabric8-tenant/auth/client"
	"github.com/fabric8-services/fabric8-tenant/cluster"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	"github.com/fabric8-services/fabric8-tenant/environment"
	testsupport "github.com/fabric8-services/fabric8-tenant/test"
	"github.com/fabric8-services/fabric8-tenant/test/doubles"
	"github.com/fabric8-services/fabric8-tenant/test/recorder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateCluster(t *testing.T) {
	// given
	defer gock.Off()
	testdoubles.MockCommunicationWithAuthSettingCapacityFlag(testsupport.ClusterURL, false, false)
	testdoubles.MockCommunicationWithAuthSettingCapacityFlag(testsupport.ClusterURL, false, false)

	// when
	clusterService, _, _, reset := testdoubles.PrepareConfigClusterAndAuthServiceWithRefreshInt(time.Second, t)
	defer reset()

	// then
	err := clusterService.Start()
	require.NoError(t, err)
	clusters := clusterService.GetClusters(context.Background())
	require.Len(t, clusters, 1)
	assert.False(t, clusters[0].CapacityExhausted)

	// and when
	testdoubles.MockCommunicationWithAuthSettingCapacityFlag(testsupport.ClusterURL, true, true)

	// then
	err = testsupport.WaitWithTimeout(3 * time.Second).Until(func() error {
		clusters := clusterService.GetClusters(context.Background())
		require.Len(t, clusters, 1)
		if !clusters[0].CapacityExhausted {
			return fmt.Errorf("CapacityExhausted flag is still false")
		}
		return nil
	})
	assert.NoError(t, err)
}

func TestResolveCluster(t *testing.T) {

	// given
	reset := testsupport.SetEnvironments(testsupport.Env("F8_AUTH_TOKEN_KEY", "foo"))
	defer reset()
	saToken, err := testsupport.NewToken(
		map[string]interface{}{
			"sub": "tenant_service",
		},
		"../test/private_key.pem",
	)
	require.NoError(t, err)
	authService, r, cleanup :=
		testdoubles.NewAuthServiceWithRecorder(t, "../test/data/cluster/resolve_cluster.fast", "http://fast.authservice", saToken.Raw, recorder.WithJWTMatcher)
	defer cleanup()

	clusterService := cluster.NewClusterService(time.Hour, authService, configuration.WithRoundTripper(r))
	err = clusterService.Start()

	require.NoError(t, err)
	defer clusterService.Stop()

	t.Run("cluster - end slash", func(t *testing.T) {
		// given
		target := "http://api.cluster1"
		// when
		found, err := clusterService.GetCluster(context.Background(), target)
		// then
		require.NoError(t, err)
		assert.Contains(t, found.APIURL, target)
	})

	t.Run("cluster - no end slash", func(t *testing.T) {
		// given
		target := "https://api.cluster2"
		// when
		found, err := clusterService.GetCluster(context.Background(), target+"/")
		// then
		require.NoError(t, err)
		assert.Contains(t, found.APIURL, target)
	})

	t.Run("both slash", func(t *testing.T) {
		// given
		target := "http://api.cluster1/"
		// when
		found, err := clusterService.GetCluster(context.Background(), target)
		// then
		require.NoError(t, err)
		assert.Contains(t, found.APIURL, target)
	})

	t.Run("no slash", func(t *testing.T) {
		// given
		target := "https://api.cluster2"
		// when
		found, err := clusterService.GetCluster(context.Background(), target)
		// then
		require.NoError(t, err)
		assert.Contains(t, found.APIURL, target)
	})
}

func TestGetClusters(t *testing.T) {
	// given
	reset := testsupport.SetEnvironments(testsupport.Env("F8_AUTH_TOKEN_KEY", "foo"))
	defer reset()
	saToken, err := testsupport.NewToken(
		map[string]interface{}{
			"sub": "tenant_service",
		},
		"../test/private_key.pem",
	)
	require.NoError(t, err)
	authService, r, cleanup :=
		testdoubles.NewAuthServiceWithRecorder(t, "../test/data/cluster/resolve_cluster.slow", "http://slow.authservice", saToken.Raw, recorder.WithJWTMatcher)
	defer cleanup()

	t.Run("ok", func(t *testing.T) {

		clusterService := cluster.NewClusterService(time.Hour, authService, configuration.WithRoundTripper(r))
		err := clusterService.Start()
		require.NoError(t, err)
		defer clusterService.Stop()
		// when
		clusters := clusterService.GetClusters(context.Background())
		// then
		require.Len(t, clusters, 2)
		assert.Equal(t, "http://api.cluster1/", clusters[0].APIURL)
		assert.Equal(t, "foo", clusters[0].AppDNS)
		assert.Equal(t, "http://console.cluster1/console/", clusters[0].ConsoleURL)
		assert.Equal(t, "http://metrics.cluster1/", clusters[0].MetricsURL)
		assert.Equal(t, "http://logging.cluster1/", clusters[0].LoggingURL)
		assert.Equal(t, saToken.Raw, clusters[0].Token) // see decode_test.go for decoded value of data in yaml file
		assert.Equal(t, "tenant_service", clusters[0].User)
		assert.Equal(t, false, clusters[0].CapacityExhausted)
	})

	t.Run("cache", func(t *testing.T) {

		t.Run("concurrent reads", func(t *testing.T) {
			clusterService := cluster.NewClusterService(time.Second, authService, configuration.WithRoundTripper(r))
			err := clusterService.Start()
			require.NoError(t, err)
			defer clusterService.Stop()
			// when 5 requests arrive at the same time (more or less)
			wg := sync.WaitGroup{}
			readersCount := 50
			results := make(chan int, readersCount)
			for i := 0; i < readersCount; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					// wait a random amount of time
					time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)
					clusters := clusterService.GetClusters(context.Background())
					// then
					results <- len(clusters)
					require.Len(t, clusters, 2)
				}()
			}
			wg.Wait()
			close(results)
			for i := 0; i < readersCount; i++ {
				result := <-results
				assert.Equal(t, 2, result)
			}
		})
	})

	t.Run("get cluster for user mapped by type", func(t *testing.T) {
		// given
		clusterService := cluster.NewClusterService(time.Hour, authService)
		err := clusterService.Start()
		require.NoError(t, err)
		defer clusterService.Stop()
		user := &auth.User{UserData: &authclient.UserDataAttributes{
			Cluster: ptr.String("http://api.cluster1/"),
		}}
		// when
		clusterForType, err := clusterService.GetUserClusterForType(context.Background(), user)
		// then
		assert.NoError(t, err)
		for _, envType := range environment.DefaultEnvTypes {
			assert.Equal(t, "http://api.cluster1/", clusterForType(envType).APIURL)
			assert.Equal(t, "foo", clusterForType(envType).AppDNS)
			assert.Equal(t, "http://console.cluster1/console/", clusterForType(envType).ConsoleURL)
			assert.Equal(t, "http://metrics.cluster1/", clusterForType(envType).MetricsURL)
			assert.Equal(t, "http://logging.cluster1/", clusterForType(envType).LoggingURL)
		}
	})
}

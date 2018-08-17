package cluster

import (
	"context"
	"testing"

	"github.com/fabric8-services/fabric8-tenant/openshift"
	testsupport "github.com/fabric8-services/fabric8-tenant/test"
	"github.com/fabric8-services/fabric8-tenant/test/recorder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commonConfig "github.com/fabric8-services/fabric8-common/configuration"
)

func TestWhoAmI(t *testing.T) {
	// given
	r, err := recorder.New("../test/data/openshift/whoami", recorder.WithJWTMatcher)
	require.NoError(t, err)
	defer r.Stop()
	tok, err := testsupport.NewToken(map[string]interface{}{
		"sub": "user_foo",
	}, "../test/private_key.pem")
	require.NoError(t, err)

	t.Run("ok", func(t *testing.T) {
		// when
		username, err := openshift.WhoAmI(context.Background(), "https://openshift.test", tok.Raw, commonConfig.WithRoundTripper(r))
		// then
		require.NoError(t, err)
		assert.Equal(t, "user_foo", username)
	})

	t.Run("forbidden", func(t *testing.T) {
		// when
		username, err := openshift.WhoAmI(context.Background(), "https://openshift.test", "", commonConfig.WithRoundTripper(r))
		// then
		require.Error(t, err)
		assert.Equal(t, "", username)
	})
}

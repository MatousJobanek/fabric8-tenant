package auth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testsupport "github.com/fabric8-services/fabric8-tenant/test"
	"github.com/fabric8-services/fabric8-tenant/auth"
	"github.com/fabric8-services/fabric8-tenant/test/doubles"
	"github.com/fabric8-services/fabric8-tenant/test/recorder"
)

func TestResolveUserToken(t *testing.T) {
	// given
	authService := testdoubles.NewAuthClientService(t, "../test/data/token/auth_resolve_target_token", "http://authservice", recorder.WithJWTMatcher)

	tok, err := testsupport.NewToken(
		map[string]interface{}{
			"sub": "user_foo",
		},
		"../test/private_key.pem",
	)
	require.NoError(t, err)

	t.Run("ok", func(t *testing.T) {
		// when
		username, accessToken, err := authService.ResolveUserToken(context.Background(), "some_valid_openshift_resource", tok.Raw)
		// then
		require.NoError(t, err)
		assert.Equal(t, "user_foo", username)
		assert.Equal(t, "an_openshift_token", accessToken)
	})

	t.Run("invalid resource", func(t *testing.T) {
		// when
		_, _, err := authService.ResolveUserToken(context.Background(), "some_invalid_resource", tok.Raw)
		// then
		require.Error(t, err)
	})

	t.Run("empty access token", func(t *testing.T) {
		// when
		_, _, err := authService.ResolveUserToken(context.Background(), "some_valid_openshift_resource", "")
		// then
		require.Error(t, err)
	})
}

func TestResolveServiceAccountToken(t *testing.T) {
	// given
	authService := testdoubles.NewAuthClientService(t, "../test/data/token/auth_resolve_target_token", "http://authservice", recorder.WithJWTMatcher)
	tok, err := testsupport.NewToken(
		map[string]interface{}{
			"sub": "tenant_service",
		},
		"../test/private_key.pem",
	)
	require.NoError(t, err)

	t.Run("ok", func(t *testing.T) {
		// when
		username, accessToken, err := authService.ResolveTargetToken(context.Background(), "some_valid_openshift_resource", tok.Raw, true, auth.PlainText)
		// then
		require.NoError(t, err)
		assert.Equal(t, "tenant_service", username)
		assert.Equal(t, "an_openshift_token", accessToken)
	})

	t.Run("expired token", func(t *testing.T) {
		// given
		tok, err := testsupport.NewToken(map[string]interface{}{
			"sub": "expired_tenant_service",
		}, "../test/private_key.pem")
		require.NoError(t, err)
		// when
		_, _, err = authService.ResolveTargetToken(context.Background(), "some_valid_openshift_resource", tok.Raw, true, auth.PlainText)
		// then
		require.Error(t, err)
	})

	t.Run("invalid resource", func(t *testing.T) {
		// when
		_, _, err := authService.ResolveTargetToken(context.Background(), "some_invalid_resource", tok.Raw, true, auth.PlainText)
		// then
		require.Error(t, err)
	})

	t.Run("empty access token", func(t *testing.T) {
		// when
		_, _, err := authService.ResolveTargetToken(context.Background(), "some_valid_openshift_resource", "", true, auth.PlainText)
		// then
		require.Error(t, err)
	})
}

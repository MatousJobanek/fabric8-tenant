package sentry

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/fabric8-services/fabric8-common/sentry"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	"github.com/fabric8-services/fabric8-tenant/test/doubles"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInitializeSentryLoggerAndSendRecord(t *testing.T) {
	// given
	reset := testdoubles.SetEnvironments(testdoubles.Env("F8_SENTRY_DSN", "https://abcdef123:abcde123@sentry.instance.server.io/1"))
	defer reset()
	config, err := configuration.NewData()
	require.NoError(t, err)

	// given
	claims := jwt.MapClaims{}
	claims["sub"] = uuid.NewV4().String()
	claims["preferred_username"] = "test-user"
	claims["email"] = "test@acme.com"

	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	ctx := goajwt.WithJWT(context.Background(), token)

	// when
	haltSentry, err := InitializeLogger(config, "123abc")
	sentry.Sentry().CaptureError(ctx, errors.New("test error"))
	defer haltSentry()

	// then
	require.NoError(t, err)
}

func TestExtractUserInfo(t *testing.T) {

	t.Run("valid token", func(t *testing.T) {
		// given
		id := uuid.NewV4().String()
		claims := jwt.MapClaims{}
		claims["sub"] = id
		claims["preferred_username"] = "test-user"
		claims["email"] = "test@acme.com"
		token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
		ctx := goajwt.WithJWT(context.Background(), token)

		// when
		user, err := extractUserInfo()(ctx)

		// then
		require.NoError(t, err)
		assert.Equal(t, id, user.ID)
		assert.Equal(t, "test-user", user.Username)
		assert.Equal(t, "test@acme.com", user.Email)
	})

	t.Run("token with missing user information", func(t *testing.T) {
		// given
		token := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{})
		ctx := goajwt.WithJWT(context.Background(), token)

		// when
		user, err := extractUserInfo()(ctx)

		// then
		require.NoError(t, err)
		assert.Equal(t, uuid.UUID{}.String(), user.ID)
		assert.Empty(t, "", user.Username)
		assert.Empty(t, "", user.Email)
	})

	t.Run("context without token", func(t *testing.T) {
		// when
		_, err := extractUserInfo()(context.Background())

		// then
		require.Error(t, err)
	})
}

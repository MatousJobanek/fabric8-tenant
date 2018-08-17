package testdoubles

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	"github.com/fabric8-services/fabric8-tenant/openshift"
	"net/http"
	"github.com/sirupsen/logrus"
	"github.com/fabric8-services/fabric8-tenant/log"
	"github.com/fabric8-services/fabric8-tenant/auth"
	"github.com/stretchr/testify/require"
	"github.com/fabric8-services/fabric8-tenant/test/recorder"
	commonConfig "github.com/fabric8-services/fabric8-common/configuration"
)

func LoadTestConfig(t *testing.T) *configuration.Data {
	data, err := configuration.GetData()
	assert.NoError(t, err, "config error")

	data.Set(configuration.VarTemplateRecommenderExternalName, "recommender.api.prod-preview.openshift.io")
	data.Set(configuration.VarTemplateRecommenderAPIToken, "HMs8laMmBSsJi8hpMDOtiglbXJ")
	data.Set(configuration.VarTemplateDomain, "d800.free-int.openshiftapps.com")
	data.Set(configuration.VarAPIServerInsecureSkipTLSVerify, "true")
	return data
}

func NewOpenshiftClient(clusterURL, token string, config *configuration.Data) *openshift.ServiceBuilder {
	return openshift.NewClientWithTransport(NewTestLogger(), clusterURL, token, config, http.DefaultTransport)
}

// NewTestLogger creates a logger instance not logging any output to Out Writer
// unless "LOG_TESTS" environment variable is set to "true"
func NewTestLogger() log.Logger {
	nullLogger := logrus.StandardLogger()
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.WarnLevel)
	//nullLogger.Out = ioutil.Discard // TODO rethink if we want to discard logging entirely for testing
	return logrus.NewEntry(nullLogger)
}

func NewAuthClientService(t *testing.T, cassetteName, authURL string, recorderOptions ...recorder.Option) *auth.Service {
	var options  []commonConfig.HTTPClientOption
	if cassetteName != "" {
		r, err := recorder.New(cassetteName, recorderOptions...)
		require.NoError(t, err)
		defer r.Stop()
		options = append(options, commonConfig.WithRoundTripper(r))
	}
	config := LoadTestConfig(t)
	config.Set(configuration.VarAuthURL, authURL)
	authService := &auth.Service{
		Config:        config,
		ClientOptions: options,
	}
	return authService
}
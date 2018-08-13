package auth

import (
	"context"
	"fmt"
	"io/ioutil"

	authclient "github.com/fabric8-services/fabric8-tenant/auth/client"
	"github.com/pkg/errors"
	commonConfig "github.com/fabric8-services/fabric8-common/configuration"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	goaclient "github.com/goadesign/goa/client"
	"github.com/satori/go.uuid"
)

type Service struct {
	Config        *configuration.Data
	ClientOptions []commonConfig.HTTPClientOption
	SaToken       string
}

// NewResolve creates a Resolver that rely on the Auth service to retrieve tokens
func NewAuthService(config *configuration.Data, options ...commonConfig.HTTPClientOption) (*Service, error) {
	c := &Service{
		Config:        config,
		ClientOptions: options,
	}
	saToken, err := c.GetOAuthToken(context.Background())
	if err != nil {
		return nil, err
	}
	c.SaToken = *saToken
	return c, nil
}

func (s *Service) GetAuthURL() string {
	return s.Config.GetAuthURL()
}

func (s *Service) NewSaClient() (*authclient.Client, error) {
	return s.NewClient(s.SaToken)
}

func (s *Service) NewClient(token string) (*authclient.Client, error) {
	client, err := NewClient(s.Config.GetAuthURL(), "", s.ClientOptions...)
	if err != nil {
		return nil, err
	}
	if token != "" {
		client.SetJWTSigner(
			&goaclient.JWTSigner{
				TokenSource: &goaclient.StaticTokenSource{
					StaticToken: &goaclient.StaticToken{
						Value: token,
						Type:  "Bearer"}}})
	}
	return client, nil
}

func (s *Service) GetOAuthToken(ctx context.Context) (*string, error) {
	c, err := s.NewClient("") // no need to specify a token in this request
	if err != nil {
		return nil, errors.Wrapf(err, "error while initializing the auth client")
	}

	path := authclient.ExchangeTokenPath()
	payload := &authclient.TokenExchange{
		ClientID: s.Config.GetAuthClientID(),
		ClientSecret: func() *string {
			sec := s.Config.GetClientSecret()
			return &sec
		}(),
		GrantType: s.Config.GetAuthGrantType(),
	}
	contentType := "application/x-www-form-urlencoded"

	res, err := c.ExchangeToken(ctx, path, payload, contentType)
	if err != nil {
		return nil, errors.Wrapf(err, "error while doing the request")
	}
	defer func() {
		ioutil.ReadAll(res.Body)
		res.Body.Close()
	}()

	validationerror := ValidateResponse(ctx, c, res)
	if validationerror != nil {
		return nil, errors.Wrapf(validationerror, "error from server %q", s.Config.GetAuthURL())
	}
	token, err := c.DecodeOauthToken(res)
	if err != nil {
		return nil, errors.Wrapf(err, "error from server %q", s.Config.GetAuthURL())
	}

	if token.AccessToken == nil || *token.AccessToken == "" {
		return nil, fmt.Errorf("received empty token from server %q", s.Config.GetAuthURL())
	}

	return token.AccessToken, nil
}

func (s *Service) ResolveUserToken(ctx context.Context, target, userToken string) (user, accessToken string, err error) {
	return s.ResolveTargetToken(ctx, target, userToken, false, PlainText)
}

func (s *Service) ResolveSaToken(ctx context.Context, target string) (username, accessToken string, err error) {
	// can't use "forcePull=true" to validate the `tenant service account` token since it's encrypted on auth
	return s.ResolveTargetToken(ctx, target, s.SaToken, false, NewGPGDecypter(s.Config.GetTokenKey()))
}

// ResolveTargetToken resolves the token for a human user or a service account user on the given target environment (can be GitHub, OpenShift Online, etc.)
func (s *Service) ResolveTargetToken(ctx context.Context, target, token string, forcePull bool, decode Decode) (username, accessToken string, err error) {
	// auth can return empty token so validate against that
	if token == "" {
		return "", "", fmt.Errorf("token must not be empty")
	}

	// check if the cluster is empty
	if target == "" {
		return "", "", fmt.Errorf("target must not be empty")
	}

	client, err := s.NewClient(token)
	if err != nil {
		return "", "", err
	}
	res, err := client.RetrieveToken(ctx, authclient.RetrieveTokenPath(), target, &forcePull)
	if err != nil {
		return "", "", errors.Wrapf(err, "error while resolving the token for %s", target)
	}
	defer func() {
		ioutil.ReadAll(res.Body)
		res.Body.Close()
	}()

	err = ValidateResponse(ctx, client, res)
	if err != nil {
		return "", "", errors.Wrapf(err, "error while resolving the token for %s", target)
	}

	externalToken, err := client.DecodeExternalToken(res)
	if err != nil {
		return "", "", errors.Wrapf(err, "error while decoding the token for %s", target)
	}
	if len(externalToken.Username) == 0 {
		return "", "", errors.Errorf("zero-length username from %s", s.Config.GetAuthURL())
	}

	t, err := decode(externalToken.AccessToken)
	return externalToken.Username, t, err
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*authclient.UserDataAttributes, error) {
	c, err := s.NewSaClient()
	if err != nil {
		return nil, err
	}

	res, err := c.ShowUsers(ctx, authclient.ShowUsersPath(id.String()), nil, nil)

	if err != nil {
		return nil, errors.Wrapf(err, "error while doing the request")
	}
	defer res.Body.Close()

	validationerror := ValidateResponse(ctx, c, res)
	if validationerror != nil {
		return nil, errors.Wrapf(validationerror, "error from server %q", s.GetAuthURL())
	}
	user, err := c.DecodeUser(res)
	if err != nil {
		return nil, errors.Wrapf(err, "error from server %q", s.GetAuthURL())
	}

	return user.Data.Attributes, nil
}

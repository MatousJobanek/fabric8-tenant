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
	"github.com/dgrijalva/jwt-go"
	"github.com/fabric8-services/fabric8-tenant/jsonapi"
	commonErrors "github.com/fabric8-services/fabric8-common/errors"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
	"github.com/fabric8-services/fabric8-tenant/log"
)

type Service struct {
	Config        *configuration.Data
	ClientOptions []commonConfig.HTTPClientOption
	SaToken       string
	log           log.Logger
}

// NewResolve creates a Resolver that rely on the Auth service to retrieve tokens
func NewAuthService(config *configuration.Data, log log.Logger, options ...commonConfig.HTTPClientOption) (*Service, error) {
	c := &Service{
		Config:        config,
		ClientOptions: options,
		SaToken:       "eyJhbGciOiJSUzI1NiIsImtpZCI6IjBsTDB2WHM5WVJWcVpNb3d5dzh1TkxSX3lyMGlGYW96ZFFrOXJ6cTJPVlUiLCJ0eXAiOiJKV1QifQ.eyJhY3IiOiIwIiwiYWxsb3dlZC1vcmlnaW5zIjpbImh0dHBzOi8vYXV0aC5vcGVuc2hpZnQuaW8iLCJodHRwczovL29wZW5zaGlmdC5pbyJdLCJhcHByb3ZlZCI6dHJ1ZSwiYXVkIjoiZmFicmljOC1vbmxpbmUtcGxhdGZvcm0iLCJhdXRoX3RpbWUiOjE1MzM4OTc2NjQsImF6cCI6ImZhYnJpYzgtb25saW5lLXBsYXRmb3JtIiwiZW1haWwiOiJtam9iYW5la0ByZWRoYXQuY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImV4cCI6MTUzNjQ4OTY2NCwiZmFtaWx5X25hbWUiOiJKb2JhbmVrIiwiZ2l2ZW5fbmFtZSI6Ik1hdG91cyIsImlhdCI6MTUzMzg5NzY2NCwiaXNzIjoiaHR0cHM6Ly9zc28ub3BlbnNoaWZ0LmlvL2F1dGgvcmVhbG1zL2ZhYnJpYzgiLCJqdGkiOiI1ZGYzMWNmMi00MzE4LTRlYzQtOTBiNy1iMTk2NDgxNWY0NTkiLCJuYW1lIjoiTWF0b3VzIEpvYmFuZWsiLCJuYmYiOjAsInByZWZlcnJlZF91c2VybmFtZSI6Im1qb2JhbmVrIiwicmVhbG1fYWNjZXNzIjp7InJvbGVzIjpbInVtYV9hdXRob3JpemF0aW9uIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsiYWNjb3VudCI6eyJyb2xlcyI6WyJtYW5hZ2UtYWNjb3VudCIsIm1hbmFnZS1hY2NvdW50LWxpbmtzIiwidmlldy1wcm9maWxlIl19LCJicm9rZXIiOnsicm9sZXMiOlsicmVhZC10b2tlbiJdfX0sInNlc3Npb25fc3RhdGUiOiIxMjQ2NjI0MC02ZTk2LTRhMTktOGE5Mi0wODQwMTZjYTQwOGQiLCJzdWIiOiI3NDI5M2RlNy1hNzk0LTQ1NTEtOGNlYy02N2U4OTkzMmU1YTYiLCJ0eXAiOiJCZWFyZXIifQ.AazwzRFwblWUfTIo2GJTHGtAPBod2AMMzepRfN5Lxw859V_zM65iHPhP7JEex3_G5J94XanRuNTtl-WpNt-yq--nfJjnpUeUAq2TTNtX1Kv_q-ZOBp1ZzKkWVob1tMQhQb6An5APW7mv1hlxb-SJsuOM3kARxOri4J6OJKyo0t3LIirTGsaU1pVT2FH3eIlU1QZVsyZHcYlrkwJYN2QbN_0YeE7LkscebRc8e2TcqLMh2M9bCjjs5urL-tpDqAgQ9fW4D7NWDv48_0rTPEDFrEmnyQdnrmsyPdrrnKh733l_rBZM5TMSymX_jOpBhRCaBLiTWElyxv6jE8mMYG3N3A",
		log:           log,
	}
	//saToken, err := c.GetOAuthToken(context.Background())
	//if err != nil {
	//	return nil, err
	//}
	//c.SaToken = *saToken
	return c, nil
}

type User struct {
	UserData           *authclient.UserDataAttributes
	OpenshiftUserName  string
	OpenshiftUserToken string
}

func (s *Service) NewUser(ctx context.Context) (*User, error) {
	userToken := goajwt.ContextJWT(ctx)
	if userToken == nil {
		return nil, jsonapi.JSONErrorResponse(ctx, commonErrors.NewUnauthorizedError("Missing JWT token"))
	}

	// fetch the cluster the user belongs to
	userData, err := s.GetAuthUserData(ctx, userToken)
	if err != nil {
		return nil, jsonapi.JSONErrorResponse(ctx, err)
	}

	if userData.Cluster == nil {
		s.log.Error(ctx, nil, "no cluster defined for tenant")
		return nil, jsonapi.JSONErrorResponse(ctx, commonErrors.NewInternalError(ctx, fmt.Errorf("unable to provision to undefined cluster")))
	}

	// fetch the users cluster token
	openshiftUserName, openshiftUserToken, err := s.ResolveUserToken(ctx, *userData.Cluster, userToken.Raw)
	if err != nil {
		s.log.Error(ctx, map[string]interface{}{
			"err":         err,
			"cluster_url": *userData.Cluster,
		}, "unable to fetch tenant token from auth")
		return nil, jsonapi.JSONErrorResponse(ctx, commonErrors.NewUnauthorizedError("Could not resolve user token"))
	}

	return &User{
		UserData:           userData,
		OpenshiftUserName:  openshiftUserName,
		OpenshiftUserToken: openshiftUserToken,
	}, nil
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

func (s *Service) GetAuthUserData(ctx context.Context, userToken *jwt.Token) (*authclient.UserDataAttributes, error) {
	c, err := s.NewSaClient()
	if err != nil {
		return nil, err
	}

	res, err := c.ShowUsers(ctx, authclient.ShowUsersPath(subject(userToken)), nil, nil)

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

func subject(token *jwt.Token) string {
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims["sub"].(string)
	}
	return ""
}

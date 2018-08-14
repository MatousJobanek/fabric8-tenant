package controller

import (
	"github.com/fabric8-services/fabric8-tenant/app"
	"github.com/goadesign/goa"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
	"github.com/fabric8-services/fabric8-tenant/log"
	"github.com/fabric8-services/fabric8-tenant/jsonapi"
	"github.com/fabric8-services/fabric8-common/errors"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	"github.com/fabric8-services/fabric8-tenant/openshift"
	"github.com/fabric8-services/fabric8-tenant/template"
	"github.com/fabric8-services/fabric8-tenant/cluster"
	"github.com/fabric8-services/fabric8-tenant/auth"
	"fmt"
	"context"
)

// TenantController implements the tenant resource.
type TenantController struct {
	*goa.Controller
	log               log.Logger
	config            *configuration.Data
	clusterService    cluster.Service
	authClientService *auth.Service
}

// NewTenantController creates a tenant controller.
func NewTenantController(service *goa.Service, clusterService cluster.Service, authClientService *auth.Service, log log.Logger, config *configuration.Data) *TenantController {
	return &TenantController{
		Controller:        service.NewController("TenantsController"),
		log:               log,
		config:            config,
		clusterService:    clusterService,
		authClientService: authClientService,
	}
}

// Clean runs the clean action.
func (c *TenantController) Clean(ctx *app.CleanTenantContext) error {
	cluster, openshiftUsername, err := c.parseTokens(ctx)
	if err != nil {
		return err
	}

	objects, err := template.RetrieveTemplatesObjects(ctx.Type, openshiftUsername, c.config)
	if err != nil {
		return err
	}

	err = openshift.NewClient(c.log, cluster.APIURL, cluster.Token, c.config).ApplyAll(objects).WithDeleteMethod()

	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, errors.NewInternalError(ctx, err))
	}

	return ctx.NoContent()
}

// Setup runs the setup action.
func (c *TenantController) Setup(ctx *app.SetupTenantContext) error {

	cluster, openshiftUsername, err := c.parseTokens(ctx)
	if err != nil {
		return err
	}

	objects, err := template.RetrieveTemplatesObjects(ctx.Type, openshiftUsername, c.config)
	if err != nil {
		return err
	}

	err = openshift.NewClient(c.log, cluster.APIURL, cluster.Token, c.config).ApplyAll(objects).WithPostMethod()
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, errors.NewInternalError(ctx, err))
	}

	return ctx.Accepted()
}

// Show runs the show action.
func (c *TenantController) Show(ctx *app.ShowTenantContext) error {
	// TenantController_Show: start_implement

	// Put your logic here

	// TenantController_Show: end_implement
	res := &app.TenantSingle{}
	return ctx.OK(res)
}

// Update runs the update action.
func (c *TenantController) Update(ctx *app.UpdateTenantContext) error {
	cluster, openshiftUsername, err := c.parseTokens(ctx)
	if err != nil {
		return err
	}

	objects, err := template.RetrieveTemplatesObjects(ctx.Type, openshiftUsername, c.config)
	if err != nil {
		return err
	}

	err = openshift.NewClient(c.log, cluster.APIURL, cluster.Token, c.config).ApplyAll(objects).WithPatchMethod()
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, errors.NewInternalError(ctx, err))
	}

	return nil
}

func (c *TenantController) parseTokens(ctx context.Context) (*cluster.Cluster, string, error) {
	userToken := goajwt.ContextJWT(ctx)
	if userToken == nil {
		return nil, "", jsonapi.JSONErrorResponse(ctx, errors.NewUnauthorizedError("Missing JWT token"))
	}
	ttoken := &auth.TenantToken{Token: userToken}

	// fetch the cluster the user belongs to
	user, err := c.authClientService.GetUser(ctx, ttoken.Subject())
	if err != nil {
		return nil, "", jsonapi.JSONErrorResponse(ctx, err)
	}

	if user.Cluster == nil {
		c.log.Error(ctx, nil, "no cluster defined for tenant")
		return nil, "", jsonapi.JSONErrorResponse(ctx, errors.NewInternalError(ctx, fmt.Errorf("unable to provision to undefined cluster")))
	}

	// fetch the users cluster token
	openshiftUsername, openshiftUserToken, err := c.authClientService.ResolveUserToken(ctx, *user.Cluster, userToken.Raw)
	if err != nil {
		c.log.Error(ctx, map[string]interface{}{
			"err":         err,
			"cluster_url": *user.Cluster,
		}, "unable to fetch tenant token from auth")
		return nil, "", jsonapi.JSONErrorResponse(ctx, errors.NewUnauthorizedError("Could not resolve user token"))
	}

	// fetch the cluster info
	cluster, err := c.clusterService.GetCluster(ctx, *user.Cluster)
	if err != nil {
		c.log.Error(ctx, map[string]interface{}{
			"err":         err,
			"cluster_url": *user.Cluster,
		}, "unable to fetch cluster")
		return nil, "", jsonapi.JSONErrorResponse(ctx, errors.NewInternalError(ctx, err))
	}
	// todo
	cluster.Token = openshiftUserToken
	cluster.Token = "SfxTgps5VblOEd4MPIWwh4ulcFDznvKkljJ0ViEUpSo"

	return cluster, openshiftUsername, nil
}

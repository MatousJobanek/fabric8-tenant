package controller

import (
	"github.com/fabric8-services/fabric8-tenant/app"
	"github.com/goadesign/goa"
	"github.com/fabric8-services/fabric8-tenant/log"
	"github.com/fabric8-services/fabric8-tenant/jsonapi"
	"github.com/fabric8-services/fabric8-common/errors"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	"github.com/fabric8-services/fabric8-tenant/openshift"
	"github.com/fabric8-services/fabric8-tenant/environment"
	"github.com/fabric8-services/fabric8-tenant/cluster"
	"github.com/fabric8-services/fabric8-tenant/auth"
	"github.com/fabric8-services/fabric8-tenant/utils"
	"context"
	"github.com/satori/go.uuid"
	"fmt"
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
	service, err := c.newOpenshiftService(ctx, ctx.Type, ctx.Space)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, errors.NewInternalError(ctx, err))
	}
	service.WithDeleteMethod()

	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, errors.NewInternalError(ctx, err))
	}

	return ctx.NoContent()
}

// Setup runs the setup action.
func (c *TenantController) Setup(ctx *app.SetupTenantContext) error {
	service, err := c.newOpenshiftService(ctx, ctx.Type, ctx.Space)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, errors.NewInternalError(ctx, err))
	}
	service.WithPostMethod()

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
	service, err := c.newOpenshiftService(ctx, ctx.Type, ctx.Space)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, errors.NewInternalError(ctx, err))
	}
	service.WithPatchMethod()
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, errors.NewInternalError(ctx, err))
	}

	return nil
}

func (c *TenantController) newOpenshiftService(ctx context.Context, nsType *string, space *uuid.UUID) (*openshift.Service, error) {
	// user
	user, err := c.authClientService.NewUser(ctx)
	if err != nil {
		return nil, err
	}

	// ns types
	nsTypes := environment.DefaultNamespaces
	fmt.Println(nsTypes)
	if !utils.IsEmpty(nsType) {
		fmt.Println("not nil", nsType)
		nsTypes = []string{*nsType}
	}

	// cluster
	user.UserData.Cluster = utils.String("https://192.168.42.241:8443")
	var clusterMapping map[string]*cluster.Cluster

	if space != nil && auth.IsCollaborator(space, user.UserData) {
		clusterMapping, err = c.clusterService.GetClusterNsMapping(space)
	} else {
		clusterMapping, err = c.clusterService.GetUserClusterNsMapping(ctx, user)
	}
	if err != nil {
		return nil, jsonapi.JSONErrorResponse(ctx, errors.NewInternalError(ctx, err))
	}

	// service
	serviceContext := openshift.NewServiceContext(c.config, clusterMapping, user, space)
	return openshift.NewService(c.log, serviceContext).ApplyAll(nsTypes), nil
}

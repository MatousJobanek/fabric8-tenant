package controller

import (
	"context"
	"fmt"
	"github.com/fabric8-services/fabric8-common/errors"
	"github.com/fabric8-services/fabric8-common/log"
	"github.com/fabric8-services/fabric8-tenant/app"
	"github.com/fabric8-services/fabric8-tenant/auth"
	"github.com/fabric8-services/fabric8-tenant/cluster"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	"github.com/fabric8-services/fabric8-tenant/environment"
	"github.com/fabric8-services/fabric8-tenant/jsonapi"
	"github.com/fabric8-services/fabric8-tenant/openshift"
	"github.com/fabric8-services/fabric8-tenant/tenant"
	"github.com/fabric8-services/fabric8-wit/rest"
	"github.com/goadesign/goa"
)

// TenantController implements the tenant resource.
type TenantController struct {
	*goa.Controller
	config            *configuration.Data
	clusterService    cluster.Service
	authClientService *auth.Service
	tenantRepository  tenant.Service
}

// NewTenantController creates a tenant controller.
func NewTenantController(
	service *goa.Service,
	clusterService cluster.Service,
	authClientService *auth.Service,
	config *configuration.Data,
	tenantRepository tenant.Service) *TenantController {

	return &TenantController{
		Controller:        service.NewController("TenantsController"),
		config:            config,
		clusterService:    clusterService,
		authClientService: authClientService,
		tenantRepository:  tenantRepository,
	}
}

// Clean runs the clean action.
func (c *TenantController) Clean(ctx *app.CleanTenantContext) error {
	// user
	user, err := c.authClientService.GetUser(ctx)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, err)
	}

	// todo verify when not exists
	namespaces, err := c.tenantRepository.GetNamespaces(user.ID)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, err)
	}

	// restrict deprovision from cluster to internal users only\
	removeFromCluster := false
	if user.UserData.FeatureLevel != nil && *user.UserData.FeatureLevel == "internal" {
		removeFromCluster = ctx.Remove
	}

	openShiftService, err := c.newOpenShiftService(ctx, user)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, err)
	}

	// todo solve remove from cluster
	err = openShiftService.WithDeleteMethod(namespaces, removeFromCluster).ApplyAll()
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, err)
	}

	return ctx.NoContent()
}

// Setup runs the setup action.
func (c *TenantController) Setup(ctx *app.SetupTenantContext) error {
	// user
	user, err := c.authClientService.GetUser(ctx)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, err)
	}

	var dbTenant *tenant.Tenant
	var namespaces []*tenant.Namespace
	if c.tenantRepository.Exists(user.ID) {
		namespaces, err = c.tenantRepository.GetNamespaces(user.ID)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}
	} else {
		dbTenant = &tenant.Tenant{
			ID:         user.ID,
			Email:      *user.UserData.Email,
			OSUsername: user.OpenShiftUsername,
		}
		err = c.tenantRepository.CreateTenant(dbTenant)
		if err != nil {
			log.Error(ctx, map[string]interface{}{
				"err": err,
			}, "unable to store tenant configuration")
			return jsonapi.JSONErrorResponse(ctx, err)
		}
	}

	missing, _ := filterMissingAndExisting(namespaces)
	if len(missing) == 0 {
		return ctx.Conflict()
	}

	service, err := c.newOpenShiftService(ctx, user)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, err)
	}

	err = service.WithPostMethod().ApplyAll(missing...)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, err)
	}

	ctx.ResponseData.Header().Set("Location", rest.AbsoluteURL(ctx.RequestData.Request, app.TenantHref()))
	return ctx.Accepted()
}

// Show runs the show action.
func (c *TenantController) Show(ctx *app.ShowTenantContext) error {
	// user
	user, err := c.authClientService.GetUser(ctx)
	if err != nil {
		return err
	}

	tenant, err := c.tenantRepository.GetTenant(user.ID)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, errors.NewNotFoundError("tenants", user.ID.String()))
	}

	namespaces, err := c.tenantRepository.GetNamespaces(user.ID)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, err)
	}

	return ctx.OK(&app.TenantSingle{Data: convertTenant(ctx, tenant, namespaces, c.clusterService.GetCluster)})
}

// Update runs the update action.
func (c *TenantController) Update(ctx *app.UpdateTenantContext) error {
	// user
	user, err := c.authClientService.GetUser(ctx)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, err)
	}

	// getting/creating tenant from DB
	tenant, err := c.tenantRepository.GetTenant(user.ID)
	if err != nil {
		return errors.NewNotFoundError("tenant", *user.UserData.IdentityID)
	}

	namespaces, err := c.tenantRepository.GetNamespaces(user.ID)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, err)
	}

	//update tenant config
	tenant.OSUsername = user.OpenShiftUsername

	if err = c.tenantRepository.SaveTenant(tenant); err != nil {
		log.Error(ctx, map[string]interface{}{
			"err": err,
		}, "unable to update tenant configuration")
		return jsonapi.JSONErrorResponse(ctx, errors.NewInternalError(ctx, fmt.Errorf("unable to update tenant configuration: %v", err)))
	}

	openShiftService, err := c.newOpenShiftService(ctx, user)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, err)
	}
	// todo check if callback was not used anywhere
	err = openShiftService.WithPatchMethod(namespaces).ApplyAll()
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, err)
	}

	ctx.ResponseData.Header().Set("Location", rest.AbsoluteURL(ctx.RequestData.Request, app.TenantHref()))
	return ctx.Accepted()
}

func (c *TenantController) newOpenShiftService(ctx context.Context, user *auth.User) (*openshift.ServiceBuilder, error) {
	// cluster
	clusterNsMapping, err := c.clusterService.GetUserClusterForType(ctx, user)
	if err != nil {
		log.Error(ctx, map[string]interface{}{
			"err":         err,
			"tenant":      user.ID,
			"cluster_url": *user.UserData.Cluster,
		}, "unable to fetch cluster for tenant")
		return nil, jsonapi.JSONErrorResponse(ctx, errors.NewInternalError(ctx, err))
	}

	nsRepo := c.tenantRepository.NewTenantRepository(user.ID)

	// service
	envService := environment.NewServiceForUserData(user.UserData)

	serviceContext := openshift.NewServiceContext(ctx, c.config, clusterNsMapping, user.OpenShiftUsername, user.OpenShiftUserToken)
	return openshift.NewService(serviceContext, nsRepo, envService), nil
}

func filterMissingAndExisting(namespaces []*tenant.Namespace) (missing []environment.Type, existing []environment.Type) {
	exitingTypes := GetNamespaceByType(namespaces)

	missingNamespaces := make([]environment.Type, 0)
	existingNamespaces := make([]environment.Type, 0)
	for _, nsType := range environment.DefaultEnvTypes {
		_, exits := exitingTypes[nsType]
		if !exits {
			missingNamespaces = append(missingNamespaces, nsType)
		} else {
			existingNamespaces = append(existingNamespaces, nsType)
		}
	}
	return missingNamespaces, existingNamespaces
}

func GetNamespaceByType(namespaces []*tenant.Namespace) map[environment.Type]*tenant.Namespace {
	var nsTypes map[environment.Type]*tenant.Namespace
	for _, namespace := range namespaces {
		nsTypes[namespace.Type] = namespace
	}
	return nsTypes
}

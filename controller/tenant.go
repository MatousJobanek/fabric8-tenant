package controller

import (
	"github.com/fabric8-services/fabric8-tenant/app"
	"github.com/goadesign/goa"
	"github.com/fabric8-services/fabric8-tenant/log"
	"github.com/fabric8-services/fabric8-common/errors"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	"github.com/fabric8-services/fabric8-tenant/cluster"
	"github.com/fabric8-services/fabric8-tenant/auth"
	"github.com/fabric8-services/fabric8-tenant/dbsupport"
	"github.com/fabric8-services/fabric8-tenant/utils"
	"github.com/fabric8-services/fabric8-tenant/environment"
	"fmt"
	"github.com/satori/go.uuid"
	"github.com/fabric8-services/fabric8-tenant/openshift"
	"context"
)

// TenantController implements the tenant resource.
type TenantController struct {
	*goa.Controller
	log               log.Logger
	config            *configuration.Data
	clusterService    cluster.Service
	authClientService *auth.Service
	tenantRepository  dbsupport.TenantRepository
}

// NewTenantController creates a tenant controller.
func NewTenantController(service *goa.Service, clusterService cluster.Service, authClientService *auth.Service,
	log log.Logger, config *configuration.Data, tenantRepository dbsupport.TenantRepository) *TenantController {

	return &TenantController{
		Controller:        service.NewController("TenantsController"),
		log:               log,
		config:            config,
		clusterService:    clusterService,
		authClientService: authClientService,
		tenantRepository:  tenantRepository,
	}
}

// Clean runs the clean action.
func (c *TenantController) Clean(ctx *app.CleanTenantContext) error {
	// user
	user, err := c.authClientService.NewUser(ctx)
	if err != nil {
		return err
	}

	// getting/creating tenant from DB
	tenant, err := c.tenantRepository.Load(*user.UserData.IdentityID)
	if err != nil {
		return errors.NewNotFoundError("tenant", *user.UserData.IdentityID)
	}

	_, toDelete := filterMissingAndExisting(tenant, ctx.Type, ctx.Space)
	if len(toDelete) == 0 {
		return errors.NewNotFoundError("tenant", *user.UserData.IdentityID)
	}

	openShiftService, err := c.newOpenShiftService(ctx, user, ctx.Space)
	if err != nil {
		return err
	}

	err = openShiftService.ApplyAll(toDelete).WithDeleteMethod()
	if err != nil {
		return err
	}

	return ctx.NoContent()
}

// Setup runs the setup action.
func (c *TenantController) Setup(ctx *app.SetupTenantContext) error {
	// user
	user, err := c.authClientService.NewUser(ctx)
	if err != nil {
		return err
	}

	tenantService := NewTenantService(ctx, user, ctx.Space)

	// getting/creating tenant from DB
	// todo should be write=true by default or rather customizable
	tenant, err := tenantService.GetTenant(c.tenantRepository, true)
	if err != nil {
		return err
	}

	missing, _ := filterMissingAndExisting(tenant, ctx.Type, ctx.Space)
	if len(missing) == 0 {
		return ctx.Conflict()
	}

	service, err := c.newOpenShiftService(ctx, user, ctx.Space)
	if err != nil {
		return err
	}
	return service.ApplyAll(missing).WithPostMethod()

	if err != nil {
		return err
	}

	return ctx.Accepted()
}

// Show runs the show action.
func (c *TenantController) Show(ctx *app.ShowTenantContext) error {
	// user
	user, err := c.authClientService.NewUser(ctx)
	if err != nil {
		return err
	}

	tenantService := NewTenantService(ctx, user, ctx.Space)

	// getting/creating tenant from DB
	tenant, err := tenantService.GetTenant(c.tenantRepository, ctx.W)
	if err != nil {
		return err
	}

	if !utils.IsEmpty(ctx.Type) {
		_, found := tenant.GetNamespace(*ctx.Type, utils.UuidValue(ctx.Space))
		if !found {
			if !ctx.W {
				return errors.NewNotFoundError(fmt.Sprintf("namespace %s for tenant", *ctx.Type), *user.UserData.IdentityID)
			}
			service, err := c.newOpenShiftService(ctx, user, ctx.Space)
			if err != nil {
				return err
			}
			err = service.Apply(*ctx.Type).WithPostMethod()
			if err != nil {
				return err
			}
			tenant, err = c.tenantRepository.Load(*user.UserData.IdentityID)
			if err != nil {
				return err
			}
		}
	}
	return ctx.OK(convertTenant(tenant, c.clusterService, ctx, filterByNsAndSpace(ctx.Type, ctx.Space)))
}

// Update runs the update action.
func (c *TenantController) Update(ctx *app.UpdateTenantContext) error {
	// user
	user, err := c.authClientService.NewUser(ctx)
	if err != nil {
		return err
	}

	// getting/creating tenant from DB
	tenant, err := c.tenantRepository.Load(*user.UserData.IdentityID)
	if err != nil {
		return errors.NewNotFoundError("tenant", *user.UserData.IdentityID)
	}

	// todo how to manage the missing ones
	_, toUpdate := filterMissingAndExisting(tenant, ctx.Type, ctx.Space)
	if len(toUpdate) == 0 {
		return errors.NewNotFoundError("tenant", *user.UserData.IdentityID)
	}

	openShiftService, err := c.newOpenShiftService(ctx, user, ctx.Space)
	if err != nil {
		return err
	}

	err = openShiftService.ApplyAll(toUpdate).WithPatchMethod()
	if err != nil {
		return err
	}

	return ctx.Accepted()
}

func (c *TenantController) newOpenShiftService(ctx context.Context, user *auth.User, space *uuid.UUID) (*openshift.ServiceBuilder, error) {
	// cluster
	clusterNsMapping, err := c.clusterService.GetClusterNsMapping(ctx, user, space)
	if err != nil {
		return nil, err
	}

	id, err := utils.UuidFromString(user.UserData.IdentityID)
	if err != nil {
		return nil, err
	}
	nsRepo := c.tenantRepository.NewNamespaceRepository(id)

	// service
	serviceContext := openshift.NewServiceContext(c.config, clusterNsMapping, user.OpenshiftUserName, space)
	return openshift.NewService(serviceContext, nsRepo), nil
}

func filterMissingAndExisting(tenant *dbsupport.Tenant, nsType *string, space *uuid.UUID) (missing []string, existing []string) {
	nsTypes := requiredNsTypes(nsType)
	exitingTypes := tenant.GetNamespaceByType(utils.UuidValue(space))

	missingNamespaces := make([]string, 0)
	existingNamespaces := make([]string, 0)
	for _, nsType := range nsTypes {
		_, exits := exitingTypes[nsType]
		if !exits {
			missingNamespaces = append(missingNamespaces, nsType)
		} else {
			existingNamespaces = append(existingNamespaces, nsType)
		}
	}
	return missingNamespaces, existingNamespaces
}

func requiredNsTypes(nsType *string) []string {
	nsTypes := environment.DefaultNamespaces
	fmt.Println(nsTypes)
	if !utils.IsEmpty(nsType) {
		nsTypes = []string{*nsType}
	}
	return nsTypes
}

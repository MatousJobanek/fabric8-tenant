package controller

import (
	"github.com/fabric8-services/fabric8-tenant/app"
	"github.com/goadesign/goa"
	"github.com/fabric8-services/fabric8-auth/token"
	"github.com/fabric8-services/fabric8-tenant/jsonapi"
	"github.com/fabric8-services/fabric8-common/errors"
	"strings"
	"github.com/fabric8-services/fabric8-tenant/dbsupport"
	"github.com/fabric8-services/fabric8-tenant/openshift"
	"github.com/fabric8-services/fabric8-tenant/cluster"
	"github.com/fabric8-services/fabric8-tenant/auth"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	"context"
	"github.com/sirupsen/logrus"
)

var (
	SERVICE_ACCOUNTS = []string{"fabric8-jenkins-idler", "rh-che"}
	wrongTokenError  = errors.NewUnauthorizedError(
		"Wrong token. Only these service accounts are allowed to perform such a request: " + strings.Join(SERVICE_ACCOUNTS, ", "))
)

// TenantsController implements the tenants resource.
type TenantsController struct {
	*goa.Controller
	config            *configuration.Data
	clusterService    cluster.Service
	authClientService *auth.Service
	tenantRepository  dbsupport.TenantRepository
}

// NewTenantsController creates a tenants controller.
func NewTenantsController(service *goa.Service) *TenantsController {
	return &TenantsController{Controller: service.NewController("TenantsController")}
}

// Delete runs the delete action.
func (c *TenantsController) Delete(ctx *app.DeleteTenantsContext) error {
	if !token.IsSpecificServiceAccount(ctx, SERVICE_ACCOUNTS...) {
		return jsonapi.JSONErrorResponse(ctx, wrongTokenError)
	}

	tenant, err := c.tenantRepository.Load(ctx.TenantID.String())
	if err != nil {
		return errors.NewNotFoundError("tenant", ctx.TenantID.String())
	}

	namespacesBySpace := tenant.GetNamespacesBySpace()

	nsRepo := c.tenantRepository.NewNamespaceRepository(ctx.TenantID)
	var clusterCache map[string]*cluster.Cluster

	for space, namespaces := range namespacesBySpace {
		clusterNsMapping := c.getClusterMapping(ctx, namespaces, clusterCache)
		serviceContext := openshift.NewServiceContext(c.config, clusterNsMapping, tenant.OSUsername, &space)

		openshift.NewService(serviceContext, nsRepo).ApplyAll(getNsTypes(namespaces)).WithDeleteMethod()
	}
	return ctx.NoContent()
}

// Search runs the search action.
func (c *TenantsController) Search(ctx *app.SearchTenantsContext) error {
	if !token.IsSpecificServiceAccount(ctx, SERVICE_ACCOUNTS...) {
		return jsonapi.JSONErrorResponse(ctx, wrongTokenError)
	}

	res := &app.TenantList{}
	return ctx.OK(res)
}

// Show runs the show action.
func (c *TenantsController) Show(ctx *app.ShowTenantsContext) error {
	if !token.IsSpecificServiceAccount(ctx, SERVICE_ACCOUNTS...) {
		return jsonapi.JSONErrorResponse(ctx, wrongTokenError)
	}


	res := &app.TenantSingle{}
	return ctx.OK(res)
}

func getNsTypes(namespaces []dbsupport.Namespace) []string {
	nsTypes := make([]string, 0, len(namespaces))
	for _, ns := range namespaces {
		nsTypes = append(nsTypes, ns.Type)
	}
	return nsTypes
}

func (c *TenantsController) getClusterMapping(ctx context.Context, namespaces []dbsupport.Namespace, clusterCache map[string]*cluster.Cluster) openshift.ClusterMapping {
	var mapping openshift.ClusterMapping
	for _, ns := range namespaces {
		var cluster *cluster.Cluster

		if cached, found := clusterCache[ns.ClusterURL]; found {
			cluster = cached

		} else {
			var err error
			cluster, err = c.clusterService.GetCluster(ctx, ns.ClusterURL)
			if err != nil {
				logrus.Error(err)
				continue
			}
			clusterCache[ns.ClusterURL] = cluster
		}
		mapping[ns.Type] = cluster
	}
	return mapping
}

package controller

import (
	"github.com/fabric8-services/fabric8-tenant/app"
	"github.com/goadesign/goa"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
	"github.com/dgrijalva/jwt-go"
	"github.com/satori/go.uuid"
	"github.com/fabric8-services/fabric8-tenant/log"
	"github.com/fabric8-services/fabric8-tenant/jsonapi"
	"github.com/fabric8-services/fabric8-common/errors"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	"github.com/fabric8-services/fabric8-tenant/openshift"
	"github.com/fabric8-services/fabric8-tenant/template"
)

// TenantController implements the tenant resource.
type TenantController struct {
	*goa.Controller
	log     log.Logger
	config  *configuration.Data
	cluster Cluster
}

// NewTenantController creates a tenant controller.
func NewTenantController(service *goa.Service, cluster Cluster, log log.Logger, config *configuration.Data) *TenantController {
	return &TenantController{
		Controller: service.NewController("TenantsController"),
		log:        log,
		config:     config,
		cluster:    cluster,
	}
}

// Clean runs the clean action.
func (c *TenantController) Clean(ctx *app.CleanTenantContext) error {
	userToken := goajwt.ContextJWT(ctx)
	if userToken == nil {
		return jsonapi.JSONErrorResponse(ctx, errors.NewUnauthorizedError("Missing JWT token"))
	}

	objects, err := template.RetrieveTemplatesObjects(ctx.Type, "mjobanek", c.config)
	if err != nil {
		return err
	}

	err = openshift.NewClient(c.log, c.cluster.APIURL, c.cluster.Token, c.config).ApplyAll(objects).WithDeleteMethod()

	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, errors.NewInternalError(ctx, err))
	}

	return ctx.NoContent()
}

// Setup runs the setup action.
func (c *TenantController) Setup(ctx *app.SetupTenantContext) error {
	userToken := goajwt.ContextJWT(ctx)
	if userToken == nil {
		return jsonapi.JSONErrorResponse(ctx, errors.NewUnauthorizedError("Missing JWT token"))
	}

	objects, err := template.RetrieveTemplatesObjects(ctx.Type, "mjobanek", c.config)
	if err != nil {
		return err
	}

	err = openshift.NewClient(c.log, c.cluster.APIURL, c.cluster.Token, c.config).ApplyAll(objects).WithPostMethod()
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
	userToken := goajwt.ContextJWT(ctx)
	if userToken == nil {
		return jsonapi.JSONErrorResponse(ctx, errors.NewUnauthorizedError("Missing JWT token"))
	}

	objects, err := template.RetrieveTemplatesObjects(ctx.Type, "mjobanek", c.config)
	if err != nil {
		return err
	}

	err = openshift.NewClient(c.log, c.cluster.APIURL, c.cluster.Token, c.config).ApplyAll(objects).WithPatchMethod()
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, errors.NewInternalError(ctx, err))
	}

	return nil
}

// Subject returns the value of the `sub` claim in the token
func Subject(token *jwt.Token) uuid.UUID {
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		id, err := uuid.FromString(claims["sub"].(string))
		if err != nil {
			return uuid.UUID{}
		}
		return id
	}
	return uuid.UUID{}
}

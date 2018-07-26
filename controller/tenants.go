package controller

import (
	"github.com/fabric8-services/fabric8-tenant/app"
	"github.com/goadesign/goa"
)

// TenantsController implements the tenants resource.
type TenantsController struct {
	*goa.Controller
}

// NewTenantsController creates a tenants controller.
func NewTenantsController(service *goa.Service) *TenantsController {
	return &TenantsController{Controller: service.NewController("TenantsController")}
}

// Delete runs the delete action.
func (c *TenantsController) Delete(ctx *app.DeleteTenantsContext) error {
	// TenantsController_Delete: start_implement

	// Put your logic here

	// TenantsController_Delete: end_implement
	return nil
}

// Search runs the search action.
func (c *TenantsController) Search(ctx *app.SearchTenantsContext) error {
	// TenantsController_Search: start_implement

	// Put your logic here

	// TenantsController_Search: end_implement
	res := &app.TenantList{}
	return ctx.OK(res)
}

// Show runs the show action.
func (c *TenantsController) Show(ctx *app.ShowTenantsContext) error {
	// TenantsController_Show: start_implement

	// Put your logic here

	// TenantsController_Show: end_implement
	res := &app.TenantSingle{}
	return ctx.OK(res)
}

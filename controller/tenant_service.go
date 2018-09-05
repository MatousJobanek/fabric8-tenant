package controller

import (
	"github.com/fabric8-services/fabric8-tenant/jsonapi"
	"github.com/fabric8-services/fabric8-common/errors"
	"github.com/fabric8-services/fabric8-tenant/auth"
	"github.com/fabric8-services/fabric8-tenant/utils"
	"context"
	"github.com/satori/go.uuid"
	"github.com/fabric8-services/fabric8-tenant/dbsupport"
)

type TenantService struct {
	ctx   context.Context
	user  *auth.User
	space *uuid.UUID
}

func NewTenantService(ctx context.Context, user *auth.User, space *uuid.UUID) *TenantService {
	return &TenantService{
		ctx:   ctx,
		user:  user,
		space: space,
	}
}

func (s *TenantService) GetTenant(repository dbsupport.TenantRepository, write bool) (*dbsupport.Tenant, error) {
	tenant, err := repository.Load(*s.user.UserData.IdentityID)
	if err != nil && !write {
		return nil, jsonapi.JSONErrorResponse(s.ctx, errors.NewNotFoundError("tenant", *s.user.UserData.IdentityID))
	} else {
		tenant, err = s.createTenant(repository)
		if err != nil {
			return nil, err
		}
	}
	return tenant, nil
}

func (s *TenantService) createTenant(repository dbsupport.TenantRepository) (*dbsupport.Tenant, error) {
	id, err := utils.UuidFromString(s.user.UserData.IdentityID)
	if err != nil {
		return nil, err
	}
	tenant, err := &dbsupport.Tenant{
		ID:         id,
		Email:      *s.user.UserData.Email,
		OSUsername: s.user.OpenshiftUserName,
	}, nil

	if err != nil {
		return nil, jsonapi.JSONErrorResponse(s.ctx, errors.NewInternalErrorFromString("unable to create a new tenant with an error: "+err.Error()))
	}
	err = repository.Create(tenant)
	if err != nil {
		return nil, jsonapi.JSONErrorResponse(s.ctx, errors.NewInternalError(s.ctx, err))
	}
	return tenant, err
}

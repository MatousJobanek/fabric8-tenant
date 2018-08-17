package auth

import (
	authclient "github.com/fabric8-services/fabric8-tenant/auth/client"
	uuid "github.com/satori/go.uuid"
)

func IsCollaborator(space *uuid.UUID, user *authclient.UserDataAttributes) bool {
	return false
}

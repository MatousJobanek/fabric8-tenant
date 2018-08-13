package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/satori/go.uuid"
)

// TenantToken the token on the tenant
type TenantToken struct {
	Token *jwt.Token
}

// Subject returns the value of the `sub` claim in the token
func (t TenantToken) Subject() uuid.UUID {
	if claims, ok := t.Token.Claims.(jwt.MapClaims); ok {
		id, err := uuid.FromString(claims["sub"].(string))
		if err != nil {
			return uuid.UUID{}
		}
		return id
	}
	return uuid.UUID{}
}

// Username returns the value of the `preferred_username` claim in the token
func (t TenantToken) Username() string {
	if claims, ok := t.Token.Claims.(jwt.MapClaims); ok {
		answer := claims["preferred_username"].(string)
		if len(answer) == 0 {
			answer = claims["username"].(string)
		}
		return answer
	}
	return ""
}

// Email returns the value of the `email` claim in the token
func (t TenantToken) Email() string {
	if claims, ok := t.Token.Claims.(jwt.MapClaims); ok {
		return claims["email"].(string)
	}
	return ""
}

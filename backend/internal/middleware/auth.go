package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yven/backend/internal/config"
	"github.com/yven/backend/internal/models"
	"gorm.io/gorm"
)

const (
	CtxUserKey = "authUser"
)

// AuthUser is the resolved identity attached to the request context
// after JWT validation + a User table lookup.
type AuthUser struct {
	UserID uuid.UUID
	Role   models.Role
	Email  string
}

// RequireAuth validates the Auth0-issued JWT on the Authorization header
// and loads the corresponding local User row.
//
// NOTE — this is a scaffold: real signature verification against Auth0's
// JWKS endpoint (https://{AUTH0_DOMAIN}/.well-known/jwks.json) needs to
// be wired in with a library such as github.com/auth0/go-jwt-middleware
// before this ships. The TODOs below mark exactly where.
func RequireAuth(cfg config.Config, database *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		_ = strings.TrimPrefix(header, "Bearer ")

		// TODO: verify JWT signature + claims against Auth0 JWKS
		// (cfg.Auth0Domain, cfg.Auth0Audience). Extract the `sub` claim
		// into auth0Sub below once verification is wired in.
		auth0Sub := c.GetHeader("X-Debug-Auth0-Sub") // dev-only stand-in until JWKS verification lands

		if auth0Sub == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		var user models.User
		if err := database.Where("auth0_sub = ?", auth0Sub).First(&user).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unrecognized identity"})
			return
		}

		c.Set(CtxUserKey, AuthUser{UserID: user.ID, Role: user.Role, Email: user.Email})
		c.Next()
	}
}

// CurrentUser pulls the AuthUser set by RequireAuth. Panics if called on
// a route not behind RequireAuth — that's intentional, it's a
// programmer error to reach a handler without auth context.
func CurrentUser(c *gin.Context) AuthUser {
	return c.MustGet(CtxUserKey).(AuthUser)
}

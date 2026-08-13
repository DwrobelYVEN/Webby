package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yven/backend/internal/models"
)

// RequireRole restricts a route to one or more platform roles. This is
// the coarse (platform-level) gate; org-scoped checks — e.g. "is this
// user an active supervisor for THIS organization" — happen inside the
// handler via OrgMembership lookups, since that can't be decided from
// the JWT alone.
func RequireRole(roles ...models.Role) gin.HandlerFunc {
	allowed := make(map[models.Role]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		user := CurrentUser(c)
		if !allowed[user.Role] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role for this action"})
			return
		}
		c.Next()
	}
}

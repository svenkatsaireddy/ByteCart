package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"oolio-backend-challenge/internal/errs"
)

const headerAPIKey = "api_key"

// APIAuth holds the configured key -> scopes map.
type APIAuth struct {
	keys map[string][]string
}

// NewAPIAuth builds middleware helpers from config.
func NewAPIAuth(keys map[string][]string) *APIAuth {
	return &APIAuth{keys: keys}
}

func hasScope(scopes []string, want string) bool {
	for _, s := range scopes {
		if s == want {
			return true
		}
	}
	return false
}

// RequireScope returns a Gin handler that enforces a valid api_key and scope.
func (a *APIAuth) RequireScope(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := strings.TrimSpace(c.GetHeader(headerAPIKey))
		if key == "" {
			errs.JSON(c, 401, errs.CodeUnauthorized, "missing api_key header")
			c.Abort()
			return
		}
		scopes, ok := a.keys[key]
		if !ok {
			errs.JSON(c, 401, errs.CodeUnauthorized, "invalid api_key")
			c.Abort()
			return
		}
		if !hasScope(scopes, scope) {
			errs.JSON(c, 403, errs.CodeForbidden, "insufficient scope for this operation")
			c.Abort()
			return
		}
		c.Next()
	}
}

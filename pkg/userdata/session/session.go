package session

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type Role string

const (
	StaffRole    Role = "staff"
	ResidentRole Role = "resident"
)

func RequireRoles(allowed ...Role) gin.HandlerFunc {
	acl := make(map[Role]struct{}, len(allowed))
	for _, role := range allowed {
		acl[role] = struct{}{}
	}

	return func(c *gin.Context) {
		session := sessions.Default(c)
		roleValue := session.Get("role")
		role, ok := roleValue.(Role)
		if !ok {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}
		if _, exists := acl[role]; !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

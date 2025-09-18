package session

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GinSessionManager struct {
	Logger *zap.SugaredLogger
}

func (sm *GinSessionManager) RequireRoles(allowed ...Role) gin.HandlerFunc {
	acl := make(map[Role]struct{}, len(allowed))
	for _, role := range allowed {
		acl[role] = struct{}{}
	}

	return func(c *gin.Context) {
		roleValue, exists := c.Get(sessKeyRole)
		role, ok := roleValue.(string)

		if !exists || !ok {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			sm.Logger.Infof("no such role detected: %s", role)

			return
		}

		if _, exists := acl[Role(role)]; !exists {
			sm.Logger.Infof("Role %s not found in acl", role)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})

			return
		}
		c.Next()
	}
}

func (sm *GinSessionManager) UserFromSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		sess := sessions.Default(c)
		if sess == nil {
			c.Next()
			return
		}

		if sessPhoneVal := sess.Get("phoneNumber"); sessPhoneVal != nil {
			if phone, ok := sessPhoneVal.(string); ok && phone != "" {
				c.Set("phoneNumber", phone)
			}
		}

		if sessRoleVal := sess.Get("role"); sessRoleVal != nil {
			switch val := sessRoleVal.(type) {
			case string:
				if val != "" {
					c.Set("currentUserRole", val)
				}
			default:
				if s := fmt.Sprintf("%v", val); s != "" {
					c.Set("currentUserRole", s)
				}
			}
		}

		c.Next()
	}
}

func (sm *GinSessionManager) setSessionValue(c *gin.Context, key, value string) {
	userSession := sessions.Default(c)
	userSession.Set(key, value)
}

func (sm *GinSessionManager) SetUserSessionRole(c *gin.Context, role Role) {
	sm.setSessionValue(c, sessKeyRole, string(role))
}

func (sm *GinSessionManager) SetUserSessionPhone(c *gin.Context, phone string) {
	sm.setSessionValue(c, sessKeyPhone, phone)
}

func (sm *GinSessionManager) SaveSession(c *gin.Context) error {
	userSession := sessions.Default(c)
	if err := userSession.Save(); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	return nil
}

func (sm *GinSessionManager) ClearSession(c *gin.Context) {
	userSession := sessions.Default(c)
	userSession.Clear()
}

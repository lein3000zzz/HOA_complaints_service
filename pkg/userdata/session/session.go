package session

import (
	"github.com/gin-gonic/gin"
)

type Role string

const (
	StaffRole    Role = "staff"
	ResidentRole Role = "resident"
)

const (
	sessKeyRole  string = "role"
	sessKeyPhone string = "phoneNumber"
)

type GinSessionManagerRepo interface {
	RequireRoles(allowed ...Role) gin.HandlerFunc
	UserFromSession() gin.HandlerFunc
	SetUserSessionRole(c *gin.Context, role Role)
	SetUserSessionPhone(c *gin.Context, phone string)
	SaveSession(c *gin.Context) error
}

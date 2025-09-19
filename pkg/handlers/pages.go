package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PageHandler struct {
	Logger *zap.SugaredLogger
}

func (h *PageHandler) MainPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		phoneVal, _ := c.Get("phoneNumber")
		c.HTML(http.StatusOK, "main.tmpl", gin.H{
			"title":       "main",
			"content":     "",
			"phoneNumber": phoneVal,
		})
	}
}

func (h *PageHandler) LoginPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		phoneVal, exists := c.Get("phoneNumber")
		if exists {
			c.Redirect(http.StatusFound, "/")
		}
		c.HTML(http.StatusOK, "login.tmpl", gin.H{
			"title":       "Login",
			"content":     "login.content",
			"phoneNumber": phoneVal,
		})
	}
}

func (h *PageHandler) RegisterPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		phoneVal, _ := c.Get("phoneNumber")
		c.HTML(http.StatusOK, "register.tmpl", gin.H{
			"title":       "Register",
			"content":     "register.content",
			"phoneNumber": phoneVal,
		})
	}
}

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
		c.HTML(http.StatusOK, "userFormMain", gin.H{
			"title":   "main",
			"content": "",
		})
	}
}

func (h *PageHandler) LoginPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.tmpl", gin.H{
			"title":   "Login",
			"content": "login.content",
		})
	}
}

func (h *PageHandler) RegisterPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.tmpl", gin.H{
			"title":   "Register",
			"content": "register.content",
		})
	}
}

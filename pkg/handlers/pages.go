package handlers

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PageHandler struct {
	Logger    *zap.SugaredLogger
	Templates map[string]*template.Template
}

func (h *PageHandler) InitHTML() {
	baseTemplatesPath := "web/templates/"
	baseTemplate := "base.tmpl"

	pagesExtendingBase := []string{
		"main.tmpl",
		"register.tmpl",
		"login.tmpl",
		"create_request.tmpl",
		"my_requests.tmpl",
		"admin.tmpl",
		"admin_requests.tmpl",
		"users.tmpl",
	}

	h.Templates = make(map[string]*template.Template)

	for _, page := range pagesExtendingBase {
		tmpl := template.Must(template.ParseFiles(
			baseTemplatesPath+baseTemplate,
			baseTemplatesPath+page,
		))
		h.Templates[page] = tmpl
	}
}

func (h *PageHandler) respondWithHTML(c *gin.Context, templateName string, data gin.H) {
	c.Status(http.StatusOK)

	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := h.Templates[templateName].ExecuteTemplate(c.Writer, templateName, data)

	if err != nil {
		h.Logger.Errorf("error rendering page %q, err %s", templateName, err)

		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
}

func (h *PageHandler) MainPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		phoneVal, exists := c.Get("phoneNumber")

		if !exists {
			c.Redirect(http.StatusSeeOther, "/login")
		}

		roleVal, _ := c.Get("role")

		data := gin.H{
			"title":       "main",
			"role":        roleVal,
			"phoneNumber": phoneVal,
		}

		h.respondWithHTML(c, "main.tmpl", data)
	}
}

func (h *PageHandler) AdminRequests() gin.HandlerFunc {
	return func(c *gin.Context) {
		phoneVal, exists := c.Get("phoneNumber")

		if !exists {
			c.Redirect(http.StatusSeeOther, "/login")
		}

		roleVal, _ := c.Get("role")

		data := gin.H{
			"title":       "admin requests",
			"role":        roleVal,
			"phoneNumber": phoneVal,
		}

		h.respondWithHTML(c, "admin_requests.tmpl", data)
	}
}

func (h *PageHandler) UsersManager() gin.HandlerFunc {
	return func(c *gin.Context) {
		phoneVal, exists := c.Get("phoneNumber")

		if !exists {
			c.Redirect(http.StatusSeeOther, "/login")
		}

		roleVal, _ := c.Get("role")

		data := gin.H{
			"title":       "user management",
			"role":        roleVal,
			"phoneNumber": phoneVal,
		}

		h.respondWithHTML(c, "users.tmpl", data)
	}
}

func (h *PageHandler) LoginPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		phoneVal, exists := c.Get("phoneNumber")

		if exists {
			c.Redirect(http.StatusFound, "/")
		}

		roleVal, _ := c.Get("role")

		data := gin.H{
			"title":       "main",
			"role":        roleVal,
			"phoneNumber": phoneVal,
		}

		h.respondWithHTML(c, "login.tmpl", data)
	}
}

func (h *PageHandler) RegisterPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		phoneVal, _ := c.Get("phoneNumber")
		roleVal, _ := c.Get("role")

		data := gin.H{
			"title":       "Register",
			"role":        roleVal,
			"phoneNumber": phoneVal,
		}

		h.respondWithHTML(c, "register.tmpl", data)
	}
}

func (h *PageHandler) CreateRequestPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		phoneVal, _ := c.Get("phoneNumber")
		roleVal, _ := c.Get("role")

		data := gin.H{
			"title":       "main",
			"role":        roleVal,
			"phoneNumber": phoneVal,
		}

		h.respondWithHTML(c, "create_request.tmpl", data)
	}
}

func (h *PageHandler) UserRequestsPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		phoneVal, _ := c.Get("phoneNumber")
		roleVal, _ := c.Get("role")

		data := gin.H{
			"title":       "my requests",
			"role":        roleVal,
			"phoneNumber": phoneVal,
		}

		h.respondWithHTML(c, "my_requests.tmpl", data)
	}
}

func (h *PageHandler) AdminPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		phoneVal, _ := c.Get("phoneNumber")
		roleVal, _ := c.Get("role")

		data := gin.H{
			"title":       "admin panel",
			"role":        roleVal,
			"phoneNumber": phoneVal,
		}

		h.respondWithHTML(c, "admin.tmpl", data)
	}
}

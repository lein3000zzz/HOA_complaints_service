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
		phoneVal, _ := c.Get("phoneNumber")
		data := gin.H{
			"title":       "main",
			"phoneNumber": phoneVal,
		}

		h.respondWithHTML(c, "main.tmpl", data)
	}
}

func (h *PageHandler) LoginPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		phoneVal, exists := c.Get("phoneNumber")

		if exists {
			c.Redirect(http.StatusFound, "/")
		}

		data := gin.H{
			"title":       "main",
			"phoneNumber": phoneVal,
		}

		h.respondWithHTML(c, "login.tmpl", data)
	}
}

func (h *PageHandler) RegisterPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		phoneVal, _ := c.Get("phoneNumber")
		//c.HTML(http.StatusOK, "register.tmpl", )
		data := gin.H{
			"title":       "Register",
			"phoneNumber": phoneVal,
		}

		h.respondWithHTML(c, "register.tmpl", data)
	}
}

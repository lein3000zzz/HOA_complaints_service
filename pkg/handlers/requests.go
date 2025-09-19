package handlers

import (
	"DBPrototyping/pkg/requests"
	"DBPrototyping/pkg/residence"
	"DBPrototyping/pkg/staffdata"
	"DBPrototyping/pkg/userdata"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ComplaintsHandler struct {
	RequestsRepo  requests.RequestRepo
	ResidentsRepo residence.ResidentsController
	StaffRepo     staffdata.StaffRepo
	UserRepo      userdata.UserRepo
	Logger        *zap.SugaredLogger
}

func (h *ComplaintsHandler) CreateRequest() func(c *gin.Context) {
	return func(c *gin.Context) {
		phone, exists := c.Get("PhoneNumber")

		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized)
		}

		requestData := requests.InitialRequestData{
			ResidentID: h.ResidentsRepo.GetResidentByPhoneNumber(phone),
		}
	}
}

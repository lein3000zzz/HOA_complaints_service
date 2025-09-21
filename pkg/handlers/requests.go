package handlers

import (
	"DBPrototyping/pkg/requests"
	"DBPrototyping/pkg/residence"
	"DBPrototyping/pkg/staffdata"
	"DBPrototyping/pkg/userdata"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RequestsHandler struct {
	RequestsRepo  requests.RequestRepo
	ResidentsRepo residence.ResidentsController
	StaffRepo     staffdata.StaffRepo
	UserRepo      userdata.UserRepo
	Logger        *zap.SugaredLogger
}

func (h *RequestsHandler) CreateRequest() func(c *gin.Context) {
	return func(c *gin.Context) {
		phoneVal, exists := c.Get("phoneNumber")
		phoneString, ok := phoneVal.(string)

		responseJSON := gin.H{}

		if !exists || !ok {
			responseJSON["error"] = "session error, try to re-login"
			h.Logger.Errorf("create request phone conversion fail for phone value %v", phoneVal)

			c.AbortWithStatusJSON(http.StatusUnauthorized, responseJSON)
			return
		}

		houseIDString := c.PostForm("houseID")
		houseID, errHouseIDConversion := strconv.Atoi(houseIDString)

		requestType := requests.RequestType(c.PostForm("requestType"))
		complaint := c.PostForm("complaint")

		if houseIDString == "" || errHouseIDConversion != nil || !requestType.IsValid() || complaint == "" {
			responseJSON["error"] = "proper request type and complaint are required"
			h.Logger.Infof("create request type and complaint are required, but wrong info provided, %s %s %s %s", houseIDString, requestType, complaint, responseJSON)

			c.AbortWithStatusJSON(http.StatusBadRequest, responseJSON)
			return
		}

		resident, errGetResident := h.ResidentsRepo.GetResidentByPhoneNumber(phoneString)

		if errGetResident != nil {
			h.Logger.Errorf("failed to find resident by phone: %s", phoneString)
			responseJSON["error"] = "no permission"

			c.AbortWithStatusJSON(http.StatusUnauthorized, responseJSON)
			return
		}

		isValid, errValidating := h.ResidentsRepo.ValidateResidentHouse(resident.ID, houseID)

		if !isValid || errValidating != nil {
			h.Logger.Errorf("failed to validate resident house: %s", resident.ID)
			responseJSON["error"] = "no permission to send request for house " + houseIDString

			c.AbortWithStatusJSON(http.StatusUnauthorized, responseJSON)
			return
		}

		requestData := requests.InitialRequestData{
			ResidentID:  resident.ID,
			HouseID:     houseID,
			RequestType: requestType,
			Complaint:   complaint,
		}

		request, errCreatingRequest := h.RequestsRepo.CreateRequest(requestData)

		if errCreatingRequest != nil {
			h.Logger.Errorf("failed to create request: %v", errCreatingRequest)
			responseJSON["error"] = "failed to create request"

			c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			return
		}

		h.Logger.Infof("created request: %v", request)

		c.JSON(http.StatusOK, request)
		return
	}
}

func (h *RequestsHandler) GetRequestsForUser() func(c *gin.Context) {
	return func(c *gin.Context) {
		phoneVal, exists := c.Get("phoneNumber")
		phoneString, ok := phoneVal.(string)

		responseJSON := gin.H{}

		if !exists || !ok {
			responseJSON["error"] = "session error, try to re-login"
			h.Logger.Errorf("create request phone conversion fail for phone value %v", phoneVal)

			c.AbortWithStatusJSON(http.StatusUnauthorized, responseJSON)
			return
		}

		page := 1
		limit := 10

		if pageStr := c.Query("page"); pageStr != "" {
			if pageInt, err := strconv.Atoi(pageStr); err == nil && pageInt > 0 {
				page = pageInt
			}
		}
		if limitStr := c.Query("limit"); limitStr != "" {
			if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 && limitInt <= 100 {
				limit = limitInt
			}
		}
		sort := c.Query("sort")
		offset := (page - 1) * limit

		userRequests, total, errGetRequests := h.RequestsRepo.GetResidentRequestsByPhone(phoneString, limit, offset, sort)

		if errGetRequests != nil {
			h.Logger.Errorf("failed to get userRequests: %v", errGetRequests)
			responseJSON["error"] = "failed to get userRequests"

			c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			return
		}

		pages := max(1, total/limit)
		if total%limit != 0 {
			pages++
		}

		meta := gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": pages,
		}
		
		responseJSON["requests"] = userRequests
		responseJSON["meta"] = meta

		c.JSON(http.StatusOK, responseJSON)
		return
	}
}

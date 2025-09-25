package handlers

import (
	"DBPrototyping/pkg/requests"
	"DBPrototyping/pkg/residence"
	"DBPrototyping/pkg/staffdata"
	"DBPrototyping/pkg/userdata"
	"errors"
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

func (h *RequestsHandler) GetRequestsForAdmin() func(c *gin.Context) {
	return func(c *gin.Context) {
		page := 1
		limit := 20
		if pageStr := c.Query("page"); pageStr != "" {
			if pageInt, err := strconv.Atoi(pageStr); err == nil && pageInt > 0 {
				page = pageInt
			}
		}
		if limitStr := c.Query("limit"); limitStr != "" {
			if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 && limitInt <= 500 {
				limit = limitInt
			}
		}
		sort := c.Query("sort")

		var filter requests.RequestFilter
		filter.Limit = limit
		filter.Offset = (page - 1) * limit
		filter.Sort = sort

		if id := c.Query("id"); id != "" {
			filter.ID = &id
		}
		if residentID := c.Query("residentID"); residentID != "" {
			filter.ResidentID = &residentID
		}
		if organizationID := c.Query("organizationID"); organizationID != "" {
			filter.OrganizationID = &organizationID
		}
		if complaint := c.Query("complaint"); complaint != "" {
			filter.Complaint = &complaint
		}

		if reqTypeString := c.Query("type"); reqTypeString != "" {
			reqType := requests.RequestType(reqTypeString)
			if reqType.IsValid() {
				filter.RequestType = &reqType
			} else {
				h.Logger.Debugf("ignore invalid request type filter: %s", reqTypeString)
			}
		}

		if statusStr := c.Query("status"); statusStr != "" {
			reqStatus := requests.RequestStatus(statusStr)
			if reqStatus.IsValid() {
				filter.Status = &reqStatus
			} else {
				h.Logger.Debugf("ignore invalid request status filter: %s", reqStatus)
			}
		}

		if houseStr := c.Query("houseID"); houseStr != "" {
			if houseInt, err := strconv.Atoi(houseStr); err == nil {
				filter.HouseID = &houseInt
			} else {
				h.Logger.Debugf("ignore invalid houseID filter: %s", houseStr)
			}
		}
		if respStr := c.Query("responsibleID"); respStr != "" {
			if responsibleInt, err := strconv.Atoi(respStr); err == nil {
				filter.ResponsibleID = &responsibleInt
			} else {
				h.Logger.Debugf("ignore invalid responsibleID filter: %s", respStr)
			}
		}

		responseJSON := gin.H{}

		requestsList, total, err := h.RequestsRepo.GetByFilter(filter)
		if err != nil {
			h.Logger.Errorf("failed to get filtered requests: %v", err)
			responseJSON["error"] = "failed to get filtered requests"
			c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			return
		}

		pages := 1
		if total > 0 {
			pages = total / limit
			if total%limit != 0 {
				pages++
			}
		}

		meta := gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": pages,
		}

		responseJSON["requests"] = requestsList
		responseJSON["meta"] = meta

		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *RequestsHandler) GetLeastBusyByJobID() func(c *gin.Context) {
	return func(c *gin.Context) {
		jobIDStr := c.Query("jobID")

		responseJSON := gin.H{}

		staffMember, errFindMember := h.StaffRepo.FindLeastBusyByJobID(jobIDStr)

		if errFindMember != nil {
			h.Logger.Errorf("failed to find least busy by jobID: %v", errFindMember)
			responseJSON["error"] = "failed to find least busy by jobID" + errFindMember.Error()

			if errors.Is(errFindMember, staffdata.ErrStaffMemberNotFound) {
				c.AbortWithStatusJSON(http.StatusBadRequest, responseJSON)
				return
			}

			c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			return
		}

		responseJSON["leastBusy"] = staffMember.ID
		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *RequestsHandler) UpdateRequest() func(c *gin.Context) {
	return func(c *gin.Context) {
		responseJSON := gin.H{}

		id := c.PostForm("id")
		residentID := c.PostForm("residentID")
		houseIDStr := c.PostForm("houseID")
		houseID, errConvertHouse := strconv.Atoi(houseIDStr)
		reqType := requests.RequestType(c.PostForm("type"))
		complaint := c.PostForm("complaint")
		costStr := c.PostForm("cost")
		reqStatus := requests.RequestStatus(c.PostForm("status"))
		respIDStr := c.PostForm("respID")
		organizationIDStr := c.PostForm("organizationID")

		if id == "" || residentID == "" || errConvertHouse != nil || !reqType.IsValid() || !reqStatus.IsValid() {
			responseJSON["error"] = "invalid request"
			h.Logger.Debugf("ignore invalid request")
			c.AbortWithStatusJSON(http.StatusBadRequest, responseJSON)
			return
		}

		requestUpdates := requests.Request{
			ID:          id,
			ResidentID:  residentID,
			Complaint:   complaint,
			HouseID:     houseID,
			RequestType: reqType,
			Status:      reqStatus,
		}

		if costStr != "" {
			if costVal, err := strconv.ParseFloat(costStr, 64); err == nil {
				requestUpdates.Cost = &costVal
			} else {
				responseJSON["error"] = "invalid cost"
				h.Logger.Debugf("invalid cost provided: %s", costStr)
				c.AbortWithStatusJSON(http.StatusBadRequest, responseJSON)
				return
			}
		}

		if respIDStr != "" {
			if respID, err := strconv.Atoi(respIDStr); err == nil {
				requestUpdates.ResponsibleID = &respID
			} else {
				responseJSON["error"] = "invalid responsible ID"
				h.Logger.Debugf("invalid respID provided: %s", respIDStr)

				c.AbortWithStatusJSON(http.StatusBadRequest, responseJSON)
				return
			}
		}

		if organizationIDStr != "" {
			requestUpdates.OrganizationID = &organizationIDStr
		}

		h.Logger.Infof("update request payload: %v", requestUpdates)

		errUpdating := h.RequestsRepo.UpdateRequest(&requestUpdates)

		if errUpdating != nil {
			h.Logger.Errorf("failed to update request: %v", errUpdating)
			responseJSON["error"] = "failed to update request"
			c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			return
		}

		responseJSON["message"] = requestUpdates.ResponsibleID
		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *RequestsHandler) DeleteRequest() func(c *gin.Context) {
	return func(c *gin.Context) {
		responseJSON := gin.H{}

		id := c.Param("id")
		if id == "" {
			responseJSON["error"] = "invalid id"
			h.Logger.Debugf("delete request invalid id")
			c.AbortWithStatusJSON(http.StatusBadRequest, responseJSON)
			return
		}

		if err := h.RequestsRepo.DeleteByID(id); err != nil {
			h.Logger.Errorf("failed to delete request: %v", err)
			responseJSON["error"] = "failed to delete request"
			c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			return
		}

		responseJSON["message"] = "deleted " + id
		c.JSON(http.StatusOK, responseJSON)
	}
}

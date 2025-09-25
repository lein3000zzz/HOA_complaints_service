package handlers

import (
	"DBPrototyping/pkg/staffdata"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type StaffHandler struct {
	StaffRepo staffdata.StaffRepo
	Logger    *zap.SugaredLogger
}

func (h *StaffHandler) GetLeastBusyByJobID() func(c *gin.Context) {
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

func (h *StaffHandler) GetSpecializationsForStaffMember() func(c *gin.Context) {
	return func(c *gin.Context) {
		staffMemberIDStr := c.Query("staffMemberID")
		staffMemberID, errConv := strconv.Atoi(staffMemberIDStr)

		responseJSON := gin.H{}

		if errConv != nil {
			h.Logger.Errorf("failed to convert staffMemberID to int: %v", errConv)
			responseJSON["error"] = "failed to convert staffMemberID to int" + errConv.Error()

			c.AbortWithStatusJSON(http.StatusBadRequest, responseJSON)
		}

		staffMemberSpecs, errGetSpecs := h.StaffRepo.FindCurrentSpecializations(staffMemberID)
		if errGetSpecs != nil {
			h.Logger.Errorf("failed to find specializations: %v", errGetSpecs)
			responseJSON["error"] = "failed to find specializations: " + errGetSpecs.Error()

			if errors.Is(errGetSpecs, staffdata.ErrStaffMemberNotFound) {
				c.AbortWithStatusJSON(http.StatusBadRequest, responseJSON)
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			}

			return
		}

		responseJSON["specializations"] = staffMemberSpecs
		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *StaffHandler) DeactivateSpecialization() func(c *gin.Context) {
	return func(c *gin.Context) {
		jobID := c.Query("jobID")
		staffMemberIDStr := c.Query("staffMemberID")
		staffMemberID, errConv := strconv.Atoi(staffMemberIDStr)

		responseJSON := gin.H{}

		if errConv != nil || jobID == "" {
			h.Logger.Errorf("failed to convert staffMemberID to int: %v", errConv)
			responseJSON["error"] = "failed to convert staffMemberID to int" + errConv.Error()

			c.AbortWithStatusJSON(http.StatusBadRequest, responseJSON)
			return
		}

		err := h.StaffRepo.DeactivateStaffMemberSpecialization(staffMemberID, jobID)

		if err != nil {
			h.Logger.Errorf("failed to deactivate staff member specialization: %v", errConv)
			responseJSON["error"] = err.Error()

			c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			return
		}

		responseJSON["message"] = "success"
		c.JSON(http.StatusOK, responseJSON)
	}
}

package handlers

import (
	"DBPrototyping/pkg/company"
	"DBPrototyping/pkg/utils"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type StaffHandler struct {
	StaffRepo company.StaffRepo
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

			if errors.Is(errFindMember, company.ErrStaffMemberNotFound) {
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

			if errors.Is(errGetSpecs, company.ErrStaffMemberNotFound) {
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

func (h *StaffHandler) GetAllSpecs() func(c *gin.Context) {
	return func(c *gin.Context) {
		responseJSON := gin.H{}

		page, limit := utils.GetPageAndLimitFromContext(c)

		offset := (page - 1) * limit

		pattern := c.Query("pattern")

		specs, total, err := h.StaffRepo.GetSpecializations(pattern, limit, offset)
		if err != nil {
			h.Logger.Errorf("failed to get all specializations: %v", err)
			responseJSON["error"] = "failed to get all specializations: " + err.Error()
			c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			return
		}

		pages := utils.CountPages(total, limit)

		meta := gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": pages,
		}

		responseJSON["specializations"] = specs
		responseJSON["meta"] = meta

		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *StaffHandler) CreateSpecialization() func(c *gin.Context) {
	return func(c *gin.Context) {
		responseJSON := gin.H{}

		jobName := c.PostForm("jobName")

		if jobName == "" {
			responseJSON["error"] = "jobID and name are required"
			h.Logger.Infof("create specialization missing fields: jobID=%s", jobName)

			c.AbortWithStatusJSON(http.StatusBadRequest, responseJSON)
			return
		}

		spec, err := h.StaffRepo.RegisterNewSpecialization(jobName)
		if err != nil {
			h.Logger.Errorf("failed to create specialization: %v", err)
			responseJSON["error"] = "failed to create specialization"

			c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			return
		}

		responseJSON["message"] = "success"
		h.Logger.Infof("created specialization: %v", spec)
		c.JSON(http.StatusOK, spec)
	}
}

func (h *StaffHandler) AddStaffSpecialization() func(c *gin.Context) {
	return func(c *gin.Context) {
		responseJSON := gin.H{}

		staffMemberIDStr := c.Query("staffMemberID")
		specializationID := c.PostForm("specializationID")

		if staffMemberIDStr == "" || specializationID == "" {
			responseJSON["error"] = "staffMemberID and specializationID are required"
			h.Logger.Infof("add staff specialization missing fields: staffMemberID=%s spec=%s", staffMemberIDStr, specializationID)

			c.AbortWithStatusJSON(http.StatusBadRequest, responseJSON)
			return
		}

		staffMemberID, err := strconv.Atoi(staffMemberIDStr)
		if err != nil {
			h.Logger.Errorf("failed to convert staffMemberID to int: %v", err)
			responseJSON["error"] = "failed to convert staffMemberID to int: " + err.Error()

			c.AbortWithStatusJSON(http.StatusBadRequest, responseJSON)
			return
		}

		if err := h.StaffRepo.AddStaffMemberSpecializationAssoc(staffMemberID, specializationID); err != nil {
			h.Logger.Errorf("failed to add staff specialization: %v", err)
			responseJSON["error"] = "failed to add staff specialization: " + err.Error()

			c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			return
		}

		responseJSON["message"] = "success"
		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *StaffHandler) CreateOrganization() func(c *gin.Context) {
	return func(c *gin.Context) {
		responseJSON := gin.H{}

		name := c.PostForm("name")
		if name == "" {
			responseJSON["error"] = "name is required"
			h.Logger.Infof("create organization missing name")

			c.AbortWithStatusJSON(http.StatusBadRequest, responseJSON)
			return
		}

		org, err := h.StaffRepo.CreateOrganization(name)
		if err != nil {
			h.Logger.Errorf("failed to create organization: %v", err)
			responseJSON["error"] = "failed to create organization"

			c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			return
		}

		responseJSON["message"] = "success"
		h.Logger.Infof("created organization: %v", org)
		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *StaffHandler) GetOrganizations() func(c *gin.Context) {
	return func(c *gin.Context) {
		responseJSON := gin.H{}

		page, limit := utils.GetPageAndLimitFromContext(c)

		offset := (page - 1) * limit
		pattern := c.Query("pattern")

		orgs, total, err := h.StaffRepo.GetOrganizationsByPattern(pattern, limit, offset)
		if err != nil {
			h.Logger.Errorf("failed to get organizations: %v", err)
			responseJSON["error"] = "failed to get organizations: " + err.Error()
			c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			return
		}

		pages := utils.CountPages(total, limit)

		meta := gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": pages,
		}

		responseJSON["organizations"] = orgs
		responseJSON["meta"] = meta

		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *StaffHandler) UpdateOrganizationName() func(c *gin.Context) {
	return func(c *gin.Context) {
		responseJSON := gin.H{}

		name := c.PostForm("name")
		orgID := c.PostForm("organizationID")

		if name == "" || orgID == "" {
			responseJSON["error"] = "organizationID and name are required"
			h.Logger.Infof("update organization missing fields: organizationID=%s name=%s", orgID, name)

			c.AbortWithStatusJSON(http.StatusBadRequest, responseJSON)
			return
		}

		if err := h.StaffRepo.UpdateOrganizationByID(orgID, name); err != nil {
			h.Logger.Errorf("failed to update organization name: %v", err)
			responseJSON["error"] = "failed to update organization"

			if errors.Is(err, company.ErrCreatingOrganization) {
				c.AbortWithStatusJSON(http.StatusBadRequest, responseJSON)
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			}
			return
		}

		responseJSON["message"] = "success"
		c.JSON(http.StatusOK, responseJSON)
	}
}

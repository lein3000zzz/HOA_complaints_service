package handlers

import (
	"DBPrototyping/pkg/residence"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ResidentsHandler struct {
	ResidentsRepo residence.ResidentsController
	Logger        *zap.SugaredLogger
}

func (h *ResidentsHandler) GetHousesForResident() func(c *gin.Context) {
	return func(c *gin.Context) {
		residentID := c.Query("residentID")

		responseJSON := gin.H{}

		resHouses, errGetHouses := h.ResidentsRepo.FindResidentHouses(residentID)
		if errGetHouses != nil {
			h.Logger.Errorf("failed to find houses: %v", errGetHouses)
			responseJSON["error"] = "failed to find houses: " + errGetHouses.Error()

			if errors.Is(errGetHouses, residence.ErrResidentNotFound) {
				c.AbortWithStatusJSON(http.StatusBadRequest, responseJSON)
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			}

			return
		}

		responseJSON["houses"] = resHouses
		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *ResidentsHandler) DeleteHouseForResident() func(c *gin.Context) {
	return func(c *gin.Context) {
		residentID := c.Query("residentID")
		houseIDStr := c.Query("houseID")
		houseID, errConv := strconv.Atoi(houseIDStr)

		responseJSON := gin.H{}

		if houseIDStr == "" || residentID == "" || errConv != nil {
			responseJSON["error"] = "invalid data provided"
			h.Logger.Debugf("invalid data provided")

			c.AbortWithStatusJSON(http.StatusBadRequest, responseJSON)
			return
		}

		err := h.ResidentsRepo.DeleteResidentHouse(residentID, houseID)

		if err != nil {
			h.Logger.Errorf("failed to delete house: %v", errConv)
			responseJSON["error"] = "failed to delete house: " + errConv.Error()

			c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			return
		}

		responseJSON["message"] = "success"
		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *ResidentsHandler) GetHouses() func(c *gin.Context) {
	return func(c *gin.Context) {
		responseJSON := gin.H{}

		page := 1
		limit := 10

		if p := c.Query("page"); p != "" {
			if v, err := strconv.Atoi(p); err == nil && v > 0 {
				page = v
			}
		}
		if l := c.Query("limit"); l != "" {
			if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 1000 {
				limit = v
			}
		}

		offset := (page - 1) * limit

		pattern := c.Query("pattern")

		houses, total, err := h.ResidentsRepo.GetHouses(pattern, limit, offset)
		if err != nil {
			h.Logger.Errorf("failed to get houses: %v", err)
			responseJSON["error"] = "failed to get houses"

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

		responseJSON["houses"] = houses
		responseJSON["meta"] = meta

		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *ResidentsHandler) CreateHouse() func(c *gin.Context) {
	return func(c *gin.Context) {
		responseJSON := gin.H{}

		address := c.PostForm("address")

		if address == "" {
			responseJSON["error"] = "address is required"
			h.Logger.Infof("create house missing fields: address=%s", address)

			c.AbortWithStatusJSON(http.StatusBadRequest, responseJSON)
			return
		}

		house, err := h.ResidentsRepo.RegisterNewHouse(address)
		if err != nil {
			h.Logger.Errorf("failed to create house: %v", err)
			responseJSON["error"] = "failed to create house"

			c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			return
		}

		h.Logger.Infof("created house: %v", house)
		responseJSON["message"] = "success"
		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *ResidentsHandler) AddResidentHouse() func(c *gin.Context) {
	return func(c *gin.Context) {
		responseJSON := gin.H{}

		residentID := c.Query("residentID")
		houseIDStr := c.PostForm("houseID")

		houseID, errConv := strconv.Atoi(houseIDStr)
		if residentID == "" || houseIDStr == "" || errConv != nil {
			responseJSON["error"] = "invalid data provided"
			h.Logger.Debugf("invalid data provided residentID=%s houseID=%s err=%v", residentID, houseIDStr, errConv)

			c.AbortWithStatusJSON(http.StatusBadRequest, responseJSON)
			return
		}

		if err := h.ResidentsRepo.AddResidentAddressAssoc(residentID, houseID); err != nil {
			h.Logger.Errorf("failed to assign house: %v", err)
			responseJSON["error"] = "failed to assign house: " + err.Error()

			if errors.Is(err, residence.ErrResidentNotFound) {
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

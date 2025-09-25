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

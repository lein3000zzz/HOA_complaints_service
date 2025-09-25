package handlers

import (
	"DBPrototyping/pkg/residence"

	"go.uber.org/zap"
)

type ResidentsHandler struct {
	ResidentsRepo residence.ResidentsController
	Logger        *zap.SugaredLogger
}

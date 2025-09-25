package handlers

import (
	"DBPrototyping/pkg/userdata"

	"go.uber.org/zap"
)

type StaffHandler struct {
	UserRepo userdata.UserRepo
	Logger   *zap.SugaredLogger
}

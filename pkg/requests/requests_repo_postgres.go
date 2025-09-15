package requests

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RequestPg Request

func (RequestPg) TableName() string {
	return "requests"
}

type RequestPgRepo struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewRequestPgRepo(logger *zap.SugaredLogger, db *gorm.DB) *RequestPgRepo {
	return &RequestPgRepo{
		logger: logger,
		db:     db,
	}
}

package residents

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ResidentPg Resident

func (ResidentPg) TableName() string {
	return "residents"
}

type HousePg House

func (HousePg) TableName() string {
	return "houses"
}

type ResidentPgRepo struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewResidentPgRepo(logger *zap.SugaredLogger, db *gorm.DB) *ResidentPgRepo {
	return &ResidentPgRepo{
		logger: logger,
		db:     db,
	}
}

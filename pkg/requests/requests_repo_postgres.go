package requests

import (
	"DBPrototyping/pkg/residence"
	"DBPrototyping/pkg/staffdata"
	"DBPrototyping/pkg/utils"
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RequestPg Request

func (RequestPg) TableName() string {
	return "requests"
}

var (
	ErrCreatingRequestPg = errors.New("error creating request pg object")
)

type RequestPgRepo struct {
	logger              *zap.SugaredLogger
	db                  *gorm.DB
	staffController     staffdata.StaffRepo
	residentsController residence.ResidentsController
}

func NewRequestPgRepo(logger *zap.SugaredLogger, db *gorm.DB, staffRepo staffdata.StaffRepo, residentsRepo residence.ResidentsController) *RequestPgRepo {
	return &RequestPgRepo{
		logger:              logger,
		db:                  db,
		staffController:     staffRepo,
		residentsController: residentsRepo,
	}
}

func (repo *RequestPgRepo) CreateRequest(requestData InitialRequestData) (*Request, error) {
	retryFactor := os.Getenv("RETRY_FACTOR")
	retries, errConversion := strconv.Atoi(retryFactor)
	if errConversion != nil || retries <= 0 {
		retries = 1
	}

	requestPg := RequestPg{
		ResidentID:     requestData.ResidentID,
		HouseID:        requestData.HouseID,
		RequestType:    requestData.RequestType,
		Complaint:      requestData.Complaint,
		Cost:           nil,
		Status:         StatusCreated,
		ResponsibleID:  nil,
		OrganizationID: nil,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	createdFlag := false
	for i := 0; i < retries && !createdFlag; i++ {
		requestID, err := utils.GenerateID()
		if err != nil {
			repo.logger.Warnf("failed to generate newResident ID, %v", err)
			continue
		}

		requestPg.ID = requestID

		upsertRes := repo.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&requestPg)
		if upsertRes.Error != nil || upsertRes.RowsAffected != 1 {
			repo.logger.Warnf("failed to insert new Resident ID, err %v", upsertRes.Error)
			continue
		}
		createdFlag = true
	}

	if !createdFlag {
		return nil, ErrCreatingRequestPg
	}

	request := Request(requestPg)

	return &request, nil
}

//func (repo *RequestPgRepo) ChooseResponsibleByJobTitle(jobTitle string) {
//
//}

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
	ErrCreatingRequestPg      = errors.New("error creating request pg object")
	ErrGettingResidentByPhone = errors.New("error matching resident by phone number")
	ErrNoRequestsFound        = errors.New("no requests found")
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
		CreatedAt:      time.Now(),
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
		if upsertRes.Error != nil {
			repo.logger.Warnf("failed to insert new Resident ID, err %v", upsertRes.Error)
			continue
		}

		if upsertRes.RowsAffected != 1 {
			repo.logger.Warnf("failed to insert new Resident, already exists, %v", upsertRes.Error)
			return nil, ErrCreatingRequestPg
		}

		createdFlag = true
	}

	if !createdFlag {
		return nil, ErrCreatingRequestPg
	}

	request := Request(requestPg)

	return &request, nil
}

func (repo *RequestPgRepo) GetResidentRequestsByPhone(phoneNumber string, limit, offset int, sort string) ([]*Request, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var requests []*Request

	reqTable := RequestPg{}.TableName()            // req
	resTable := residence.ResidentPg{}.TableName() // res

	queryRes := repo.db.WithContext(ctx).
		Table(reqTable+" AS req").
		Select("*").
		Joins("JOIN "+resTable+" AS res ON res.id = req.id_resident").
		Where("res.phone_number = ?", phoneNumber)

	var total int64
	if err := queryRes.Count(&total).Error; err != nil {
		repo.logger.Warnf("failed to count requests by phone %s: %v", phoneNumber, err)
		return nil, 0, err
	}

	if total == 0 {
		return []*Request{}, 0, nil
	}

	var order string
	switch sort {
	case "status_asc":
		order = "req.status ASC"
	case "type_asc":
		order = "req.type ASC"
	default:
		order = "req.created_at DESC"
	}

	if limit > 0 {
		queryRes = queryRes.Limit(limit)
	}
	if offset > 0 {
		queryRes = queryRes.Offset(offset)
	}
	queryRes = queryRes.Order(order)

	if err := queryRes.Find(&requests).Error; err != nil {
		repo.logger.Warnf("failed to query requests by phone %s: %v", phoneNumber, err)
		return nil, 0, err
	}

	return requests, int(total), nil
}

//func (repo *RequestPgRepo) GetAll() ([]*Request, error) {
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	var requestsPg []*RequestPg
//	queryRes := repo.db.WithContext(ctx).Find(&requestsPg)
//
//	if queryRes.Error != nil {
//		repo.logger.Errorf("error getting all requests, %v", queryRes.Error)
//		return nil, queryRes.Error
//	}
//
//	requests := make([]*Request, len(requestsPg))
//	for i, requestPg := range requestsPg {
//		requests[i] = (*Request)(requestPg)
//	}
//
//	return requests, nil
//}

func (repo *RequestPgRepo) GetAll(limit, offset int, sort string) ([]*Request, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var requests []*Request

	reqTable := RequestPg{}

	queryRes := repo.db.WithContext(ctx).Find(&reqTable)

	var total int64
	if err := queryRes.Count(&total).Error; err != nil {
		repo.logger.Warnf("failed to count requests: %v", err)
		return nil, 0, err
	}

	if total == 0 {
		return []*Request{}, 0, nil
	}

	var order string
	switch sort {
	case "status_asc":
		order = "req.status ASC"
	case "type_asc":
		order = "req.type ASC"
	default:
		order = "req.created_at DESC"
	}

	if limit > 0 {
		queryRes = queryRes.Limit(limit)
	}
	if offset > 0 {
		queryRes = queryRes.Offset(offset)
	}
	queryRes = queryRes.Order(order)

	if err := queryRes.Find(&requests).Error; err != nil {
		repo.logger.Warnf("failed to query requests: %v", err)
		return nil, 0, err
	}

	return requests, int(total), nil
}

//func (repo *RequestPgRepo) ChooseResponsibleByJobTitle(jobTitle string) {
//
//}

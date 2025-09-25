package requests

import (
	"DBPrototyping/pkg/residence"
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
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewRequestPgRepo(logger *zap.SugaredLogger, db *gorm.DB) *RequestPgRepo {
	return &RequestPgRepo{
		logger: logger,
		db:     db,
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

func (repo *RequestPgRepo) GetByFilter(filter RequestFilter) ([]*Request, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var requestsPg []*RequestPg

	reqTable := RequestPg{}.TableName()
	query := repo.db.WithContext(ctx).Table(reqTable + " AS req").Model(&RequestPg{})

	if filter.ID != nil {
		query = query.Where("req.ID LIKE ?", "%"+*filter.ID+"%")
	}
	if filter.Status != nil {
		query = query.Where("req.status = ?", string(*filter.Status))
	}
	if filter.RequestType != nil {
		query = query.Where("req.type = ?", string(*filter.RequestType))
	}
	if filter.ResidentID != nil {
		query = query.Where("req.id_resident LIKE ?", "%"+*filter.ResidentID+"%")
	}
	if filter.HouseID != nil {
		query = query.Where("req.id_house = ?", *filter.HouseID)
	}
	if filter.ResponsibleID != nil {
		query = query.Where("req.id_responsible = ?", *filter.ResponsibleID)
	}
	if filter.OrganizationID != nil {
		query = query.Where("req.id_organization LIKE ?", "%"+*filter.OrganizationID+"%")
	}
	if filter.Complaint != nil {
		query = query.Where("req.complaint LIKE ?", "%"+*filter.Complaint+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		repo.logger.Warnf("failed to count requests with filter: %v", err)
		return nil, 0, err
	}
	if total == 0 {
		return []*Request{}, 0, nil
	}

	var order string
	switch filter.Sort {
	case "status_asc":
		order = "req.status ASC"
	case "status_desc":
		order = "req.status DESC"
	case "type_asc":
		order = "req.type ASC"
	case "type_desc":
		order = "req.type DESC"
	case "created_asc":
		order = "req.created_at ASC"
	default:
		order = "req.created_at DESC"
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}
	query = query.Order(order)

	if err := query.Find(&requestsPg).Error; err != nil {
		repo.logger.Warnf("failed to query filtered requests: %v", err)
		return nil, 0, err
	}

	requests := make([]*Request, len(requestsPg))
	for i, r := range requestsPg {
		requests[i] = (*Request)(r)
	}

	return requests, int(total), nil
}

func (repo *RequestPgRepo) DeleteByID(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res := repo.db.WithContext(ctx).Where("id = ?", id).Delete(&RequestPg{})
	if res.Error != nil {
		repo.logger.Warnf("failed to delete request id %s: %v", id, res.Error)
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNoRequestsFound
	}

	return nil
}

func (repo *RequestPgRepo) UpdateRequest(updatedRequest *Request) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	requestPg := RequestPg(*updatedRequest)
	res := repo.db.WithContext(ctx).Model(&RequestPg{}).Where("id = ?", updatedRequest.ID).Updates(requestPg)
	if res.Error != nil {
		repo.logger.Warnf("failed to update request id %s: %v", updatedRequest.ID, res.Error)
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNoRequestsFound
	}

	return nil
}

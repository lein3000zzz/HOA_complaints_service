package residence

import (
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

type ResidentPg Resident

func (ResidentPg) TableName() string {
	return "residents"
}

type HousePg House

func (HousePg) TableName() string {
	return "houses"
}

type ResidentHousePg ResidentHouse

func (ResidentHousePg) TableName() string {
	return "residents_houses"
}

var (
	ErrResidentExists         = errors.New("resident already exists")
	ErrRegisterResidentFailed = errors.New("register resident fail")
	ErrResidentNotFound       = errors.New("resident not found")
	ErrHouseExists            = errors.New("house already exists")
	ErrAddingResidentAddress  = errors.New("error adding residential address")
)

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

// TODO в хэндлерах можно при регистрации пользователя также отправлять дополнительные поля, чтобы регистрировать и резидента
func (repo *ResidentPgRepo) RegisterNewResident(phone, fullName string) (*Resident, error) {
	retryFactor := os.Getenv("RETRY_FACTOR")
	retries, errConversion := strconv.Atoi(retryFactor)
	if errConversion != nil || retries <= 0 {
		retries = 1
	}

	newResidentPg := ResidentPg{
		Phone:    phone,
		FullName: fullName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	createdFlag := false
	for i := 0; i < retries && !createdFlag; i++ {
		residentID, err := utils.GenerateID()
		if err != nil {
			repo.logger.Warnf("failed to generate newResident ID, %v", err)
			continue
		}

		newResidentPg.ID = residentID

		upsertRes := repo.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&newResidentPg)
		if upsertRes.Error != nil {
			repo.logger.Warnf("failed to register newResident, %v", upsertRes.Error)
			continue
		}

		if upsertRes.RowsAffected != 1 {
			repo.logger.Warnf("failed to register new resident, already exists, %v", phone)
			return nil, ErrResidentExists
		}

		createdFlag = true
	}

	if !createdFlag {
		return nil, ErrRegisterResidentFailed
	}

	newResident := Resident(newResidentPg)

	return &newResident, nil
}

func (repo *ResidentPgRepo) RegisterNewHouse(address string) (*House, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	newHousePg := HousePg{
		Address: address,
	}

	createRes := repo.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&newHousePg)
	if createRes.Error != nil {
		return nil, createRes.Error
	}
	if createRes.RowsAffected != 1 {
		return nil, ErrHouseExists
	}

	newHouse := House(newHousePg)
	return &newHouse, nil
}

func (repo *ResidentPgRepo) AddResidentAddressAssoc(residentID string, houseID int) error {
	mapping := ResidentHousePg{
		ResidentID: residentID,
		HouseID:    houseID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := repo.db.WithContext(ctx).
		Table(ResidentHousePg{}.TableName()).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&mapping).Error; err != nil {
		repo.logger.Errorf("failed to create resident-house mapping resident=%s house=%d: %v", residentID, houseID, err)
		return err
	}

	repo.logger.Infof("assigned house %d to resident %s", houseID, residentID)
	return nil
}

func (repo *ResidentPgRepo) GetResidentByPhoneNumber(phoneNumber string) (*Resident, error) {
	var residentPg ResidentPg

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := repo.db.WithContext(ctx).Where("phone_number = ?", phoneNumber).First(&residentPg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			repo.logger.Debugf("resident not found: %s", phoneNumber)
			return nil, ErrResidentNotFound
		}

		repo.logger.Errorf("failed to find resident: %s", err.Error())
		return nil, err
	}

	repo.logger.Debugf("Successfully found resident: %s", phoneNumber)

	resident := Resident(residentPg)

	return &resident, nil
}

func (repo *ResidentPgRepo) ValidateResidentHouse(residentID string, houseID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var residentHousePg ResidentHousePg

	if err := repo.db.WithContext(ctx).Where("id_resident = ?", residentID).Where("id_house = ?", houseID).First(&residentHousePg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			repo.logger.Debugf("resident and house do not match: %s, %d", residentID, houseID)
			return false, ErrResidentNotFound
		}

		repo.logger.Errorf("failed to find resident and/or house id: %s", err.Error())
		return false, err
	}

	return true, nil
}

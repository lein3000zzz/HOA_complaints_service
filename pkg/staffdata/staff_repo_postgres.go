package staffdata

import (
	"DBPrototyping/pkg/utils"
	"errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"os"
	"strconv"
)

var (
	ErrCreatingSpecialization = errors.New("error creating specialization")
	ErrCreatingMember         = errors.New("error creating a new member")
)

type StaffMemberPg StaffMember

func (StaffMemberPg) TableName() string {
	return "staff_members"
}

type StaffRepoPostgres struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewStaffRepoPostgres(logger *zap.SugaredLogger, db *gorm.DB) *StaffRepoPostgres {
	return &StaffRepoPostgres{
		logger: logger,
		db:     db,
	}
}

func (repo *StaffRepoPostgres) RegisterNewSpecialization(jobTitle string) (string, error) {
	retryFactor := os.Getenv("RETRY_FACTOR")
	retries, errConversion := strconv.Atoi(retryFactor)
	if errConversion != nil || retries <= 0 {
		retries = 1
	}

	spec := Specialization{
		Title: jobTitle,
	}

	for i := 0; i < retries; i++ {
		jobID, err := utils.GenerateID()
		if err != nil {
			repo.logger.Warnf("failed to generate specialization ID, %v", err)
			continue
		}

		spec.ID = jobID

		upsertRes := repo.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&spec)
		if upsertRes.Error != nil {
			return "", upsertRes.Error
		}
		if upsertRes.RowsAffected == 1 {
			return spec.ID, nil
		}
	}

	return "", ErrCreatingSpecialization
}

// TODO на хэндлере проверять количество символов в номере
func (repo *StaffRepoPostgres) RegisterNewMember(phone, fullName string) error {
	newMember := StaffMemberPg{
		Phone:    phone,
		FullName: fullName,
		Status:   StatusActive,
	}

	upsertRes := repo.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&newMember)

	if upsertRes.Error != nil {
		return upsertRes.Error
	}
	if upsertRes.RowsAffected != 1 {
		repo.logger.Warnf("failed to insert new staff member %v", newMember)
		return ErrCreatingMember
	}

	return nil
}

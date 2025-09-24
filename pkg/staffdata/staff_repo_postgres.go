package staffdata

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

var (
	ErrCreatingSpecialization = errors.New("error creating specialization")
	ErrStaffMemberNotFound    = errors.New("staff member not found")
	ErrCreatingMember         = errors.New("error creating a new member")
)

type StaffMemberPg StaffMember

func (StaffMemberPg) TableName() string {
	return "staff_members"
}

type SpecializationPg Specialization

func (SpecializationPg) TableName() string {
	return "specializations"
}

type StaffMemberSpecializationPg StaffMemberSpecialization

func (StaffMemberSpecializationPg) TableName() string {
	return "staff_member_specialization"
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

func (repo *StaffRepoPostgres) RegisterNewSpecialization(jobTitle string) (*Specialization, error) {
	retryFactor := os.Getenv("RETRY_FACTOR")
	retries, errConversion := strconv.Atoi(retryFactor)
	if errConversion != nil || retries <= 0 {
		retries = 1
	}

	specPg := SpecializationPg{
		Title: jobTitle,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	createdFlag := false
	for i := 0; i < retries && !createdFlag; i++ {
		jobID, err := utils.GenerateID()
		if err != nil {
			repo.logger.Warnf("failed to generate specialization ID, %v", err)
			continue
		}

		specPg.ID = jobID

		upsertRes := repo.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&specPg)
		if upsertRes.Error != nil || upsertRes.RowsAffected != 1 {
			continue
		}
		createdFlag = true
	}

	if !createdFlag {
		return nil, ErrCreatingSpecialization
	}

	spec := Specialization(specPg)

	return &spec, nil
}

// TODO на хэндлере проверять количество символов в номере
func (repo *StaffRepoPostgres) RegisterNewMember(phone, fullName string) (*StaffMember, error) {
	newMemberPg := StaffMemberPg{
		Phone:    phone,
		FullName: fullName,
		Status:   StatusActive,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	upsertRes := repo.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&newMemberPg)

	if upsertRes.Error != nil {
		return nil, upsertRes.Error
	}
	if upsertRes.RowsAffected != 1 {
		repo.logger.Warnf("failed to insert new staff member %v", newMemberPg)
		return nil, ErrCreatingMember
	}

	newMember := StaffMember(newMemberPg)

	return &newMember, nil
}

func (repo *StaffRepoPostgres) AddStaffMemberSpecializationAssoc(staffMemberID int, specializationID string) error {
	mapping := StaffMemberSpecializationPg{
		MemberID:         staffMemberID,
		SpecializationID: specializationID,
		IsActive:         true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := repo.db.WithContext(ctx).
		Table(StaffMemberSpecializationPg{}.TableName()).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&mapping).Error; err != nil {
		repo.logger.Errorf("failed to create staff-spec mapping staffMemberID=%d spec=%s: %v", staffMemberID, specializationID, err)
		return err
	}

	repo.logger.Infof("assigned member %d to spec %s", staffMemberID, specializationID)
	return nil
}

func (repo *StaffRepoPostgres) GetStaffMemberByPhoneNumber(phoneNumber string) (*StaffMember, error) {
	var staffMemberPg StaffMemberPg

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := repo.db.WithContext(ctx).Where("phone_number = ?", phoneNumber).First(&staffMemberPg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			repo.logger.Debugf("staff member not found: %s", phoneNumber)
			return nil, ErrStaffMemberNotFound
		}

		repo.logger.Errorf("failed to find staff member: %s", err.Error())
		return nil, err
	}

	repo.logger.Debugf("Successfully found staff member: %s", staffMemberPg.Phone)
	staffMember := StaffMember(staffMemberPg)
	return &staffMember, nil
}

func (repo *StaffRepoPostgres) DeleteByPhone(phoneNumber string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	deleteRes := repo.db.WithContext(ctx).Where("phone_number = ?", phoneNumber).Delete(&StaffMember{})

	if deleteRes.Error != nil {
		repo.logger.Errorf("failed to delete staff member: %s", deleteRes.Error)
		return deleteRes.Error
	}

	if deleteRes.RowsAffected != 1 {
		repo.logger.Warnf("failed to delete staff member, no such phone %s: %v", phoneNumber, deleteRes.RowsAffected)
		return ErrStaffMemberNotFound
	}

	return nil
}

package staffdata

import (
	"DBPrototyping/pkg/requests"
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
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "id_member"},
				{Name: "id_specialization"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{"is_active": true}),
		}).
		Create(&mapping).Error; err != nil {
		repo.logger.Errorf("failed to create/update staff-spec mapping staffMemberID=%d spec=%s: %v", staffMemberID, specializationID, err)
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
	
	var member StaffMemberPg
	if err := repo.db.WithContext(ctx).Where("phone_number = ?", phoneNumber).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			repo.logger.Debugf("staff member not found: %s", phoneNumber)
			return ErrStaffMemberNotFound
		}
		repo.logger.Errorf("failed to query staff member: %v", err)
		return err
	}

	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		mappingTable := StaffMemberSpecializationPg{}.TableName()
		if del := tx.Table(mappingTable).Where("id_member = ?", member.ID).Delete(&StaffMemberSpecializationPg{}); del.Error != nil {
			repo.logger.Errorf("failed to delete staff-member specializations for member %d: %v", member.ID, del.Error)
			return del.Error
		}

		deleteRes := tx.Where("id = ?", member.ID).Delete(&StaffMember{})
		if deleteRes.Error != nil {
			repo.logger.Errorf("failed to delete staff member: %v", deleteRes.Error)
			return deleteRes.Error
		}
		if deleteRes.RowsAffected != 1 {
			repo.logger.Warnf("failed to delete staff member, no such id %d: %v", member.ID, deleteRes.RowsAffected)
			return ErrStaffMemberNotFound
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (repo *StaffRepoPostgres) FindLeastBusyByJobID(jobID string) (*StaffMember, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var staffPg StaffMemberPg

	staffTable := StaffMemberPg{}.TableName()
	specAssocTable := StaffMemberSpecializationPg{}.TableName()
	reqTable := requests.RequestPg{}.TableName()

	query := repo.db.WithContext(ctx).
		Table(staffTable+" AS staff").
		Select("staff.*, COALESCE(COUNT(req.id), 0) AS active_count").
		Joins("JOIN "+specAssocTable+" AS staffspec ON staffspec.id_member = staff.id").
		Joins("LEFT JOIN "+reqTable+" AS req ON req.id_responsible = staff.id AND req.status = ?", string(requests.StatusAssigned)).
		Where("staffspec.id_specialization = ?", jobID).
		Group("staff.id").
		Order("active_count ASC").
		Limit(1)

	if err := query.Scan(&staffPg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStaffMemberNotFound
		}
		return nil, err
	}

	if staffPg.ID == 0 {
		return nil, ErrStaffMemberNotFound
	}

	staff := StaffMember(staffPg)
	return &staff, nil
}

func (repo *StaffRepoPostgres) FindCurrentSpecializations(staffMemberID int) ([]*Specialization, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var specsPg []SpecializationPg

	assocTable := StaffMemberSpecializationPg{}.TableName()
	specTable := SpecializationPg{}.TableName()

	query := repo.db.WithContext(ctx).
		Table(specTable+" AS spec").
		Select("spec.*").
		Joins("JOIN "+assocTable+" AS assoc ON assoc.id_specialization = spec.id").
		Where("assoc.id_member = ? AND assoc.is_active = ?", staffMemberID, true)

	if err := query.Find(&specsPg).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		repo.logger.Errorf("failed to query specializations for member %d: %v", staffMemberID, err)
		return nil, err
	}

	result := make([]*Specialization, 0, len(specsPg))
	for _, s := range specsPg {
		result = append(result, (*Specialization)(&s))
	}

	return result, nil
}

func (repo *StaffRepoPostgres) DeactivateStaffMemberSpecialization(staffMemberID int, jobID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	table := StaffMemberSpecializationPg{}.TableName()
	updateRes := repo.db.WithContext(ctx).
		Table(table).
		Where("id_member = ? AND id_specialization = ? AND is_active = ?", staffMemberID, jobID, true).
		Update("is_active", false)

	if updateRes.Error != nil {
		repo.logger.Errorf("failed to deactivate staff-spec mapping staffMemberID=%d spec=%s: %v", staffMemberID, jobID, updateRes.Error)
		return updateRes.Error
	}

	if updateRes.RowsAffected != 1 {
		repo.logger.Warnf("no active staff-spec mapping found to deactivate staffMemberID=%d spec=%s", staffMemberID, jobID)
		return ErrStaffMemberNotFound
	}

	repo.logger.Infof("deactivated specialization %s for member %d", jobID, staffMemberID)
	return nil
}

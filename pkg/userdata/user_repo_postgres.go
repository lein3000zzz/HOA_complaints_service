package userdata

import (
	"DBPrototyping/pkg/utils"
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
	ErrCreatingUser = errors.New("error creating a new user")
)

type UserRepoPg struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

type UserPg User

func (UserPg) TableName() string {
	return "login_credentials"
}

func NewUserRepoPg(db *gorm.DB, logger *zap.SugaredLogger) *UserRepoPg {
	return &UserRepoPg{
		db:     db,
		logger: logger,
	}
}

func (repo *UserRepoPg) Authorize(login, password string) (*User, error) {
	var userPg UserPg

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := repo.db.WithContext(ctx).Where("phone_number = ?", login).First(&userPg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			repo.logger.Debugf("User %s not found", login)
			return nil, ErrUserNotFound
		}

		repo.logger.Debugf("User %s not found, err, %v", login, err)
		return nil, err
	}

	if ok := utils.CheckPassword(userPg.PasswordHash, password); !ok {
		repo.logger.Debugf("User %s password not match", login)
		return nil, ErrUserNotFound
	}

	user := User(userPg)
	repo.logger.Debugf("User %s authorized to %s", login, user.Phone)
	return &user, nil
}

// TODO в хэндлере замутить проверку количества символов
func (repo *UserRepoPg) Register(phone, password string) (*User, error) {
	passwordHash, errHashing := utils.HashPassword(password)

	if errHashing != nil {
		repo.logger.Debugf("Error hashing password: %v", errHashing)
		return nil, errHashing
	}

	userPg := UserPg{
		Phone:        phone,
		PasswordHash: passwordHash,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	upsertRes := repo.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&userPg)

	if upsertRes.Error != nil {
		return nil, upsertRes.Error
	}
	if upsertRes.RowsAffected != 1 {
		repo.logger.Warnf("failed to insert new user %v", userPg)
		return nil, ErrUserExists
	}

	user := User(userPg)

	return &user, nil
}

//func (repo *UserRepoPg) checkUserExists(phone string) (bool, error) {
//	var user UserPg
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	if err := repo.db.WithContext(ctx).Where("phone = ?", phone).First(user).Error; err != nil {
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			return false, nil
//		}
//		return false, err
//	}
//
//	return true, nil
//}

//func (repo *UserRepoPg) isStaffMember(user *User) (bool, error) {
//	if user == nil || user.Phone == "" {
//		return false, nil
//	}
//
//	var staffMember staffdata.StaffMemberPg
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	if err := repo.db.WithContext(ctx).Where("phone = ?", user.Phone).First(&staffMember).Error; err != nil {
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			repo.logger.Debugf("staff member not found: %s", user.Phone)
//			return false, nil
//		}
//
//		repo.logger.Errorf("failed to find staff member: %s", err.Error())
//		return false, err
//	}
//
//	repo.logger.Debugf("Successfully found staff member: %s", staffMember.Phone)
//	return true, nil
//}

//func (repo *UserRepoPg) isResident(user *User) (bool, error) {
//	if user == nil || user.Phone == "" {
//		return false, nil
//	}
//
//	var resident residence.ResidentPg
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	if err := repo.db.WithContext(ctx).Where("phone = ?", user.Phone).First(&resident).Error; err != nil {
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			repo.logger.Debugf("resident not found: %s", user.Phone)
//			return false, nil
//		}
//
//		repo.logger.Errorf("failed to find resident: %s", err.Error())
//		return false, err
//	}
//
//	repo.logger.Debugf("Successfully found resident: %s", resident.Phone)
//	return true, nil
//}

//func (repo *UserRepoPg) GenerateUserToken(user *User) (*jwt.Token, error) {
//	isStaffMember, errStaffMember := repo.isStaffMember(user)
//	isResident, errResident := repo.isResident(user)
//
//	if errStaffMember != nil || errResident != nil {
//		repo.logger.Errorf("Staff member not found for user: %s", user.Phone)
//		return nil, errStaffMember
//	}
//
//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
//		"phone_number":  user.Phone,
//		"isStaffMember": strconv.FormatBool(isStaffMember),
//		"isResident":    strconv.FormatBool(isResident),
//	})
//	// "exp": time.Now().Add(1 * time.Hour).Unix(),
//	repo.logger.Debugf("Successfully generate user token for user: %s", user.Phone)
//	return token, nil
//}

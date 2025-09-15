package userdata

import (
	"DBPrototyping/pkg/residents"
	"DBPrototyping/pkg/staffdata"
	"DBPrototyping/pkg/utils"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strconv"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

type UserRepoPg struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

type UserPg User

func NewUserRepoPg(db *gorm.DB, logger *zap.SugaredLogger) *UserRepoPg {
	return &UserRepoPg{
		db:     db,
		logger: logger,
	}
}

func (repo *UserRepoPg) Authorize(login, password string) (*User, error) {
	var userPg UserPg
	if err := repo.db.Where("phone = ?", login).First(&userPg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			repo.logger.Debugf("User %s not found", login)
			return nil, ErrUserNotFound
		}
		repo.logger.Debugf("User %s not found, err, %v", login, err)
		return nil, err
	}

	if ok := utils.CheckPassword(password, userPg.PasswordHash); !ok {
		repo.logger.Debugf("User %s password not match", login)
		return nil, ErrUserNotFound
	}

	user := User(userPg)
	repo.logger.Debugf("User %s authorized to %s", login, user.Phone)
	return &user, nil
}

// TODO в хэндлере замутить проверку количества символов
func (repo *UserRepoPg) Register(phone, password string) (*User, error) {
	exists, err := repo.checkUserExists(phone)
	if err != nil {
		repo.logger.Errorf("checkUserExists error: %s", err.Error())
		return nil, err
	}

	if exists {
		repo.logger.Debugf("user %s already exists", phone)
		return nil, ErrUserExists
	}

	passwordHash, errHashing := utils.HashPassword(password)

	if errHashing != nil {
		repo.logger.Errorf("hashing error: %s", errHashing.Error())
		return nil, errHashing
	}

	userPg := UserPg{
		Phone:        phone,
		PasswordHash: passwordHash,
	}

	repo.db.Create(&userPg)

	repo.logger.Debugf("Successfully created user: %s", userPg.Phone)
	user := User(userPg)

	return &user, nil
}

func (repo *UserRepoPg) checkUserExists(phone string) (bool, error) {
	var user UserPg

	if err := repo.db.Where("phone = ?", phone).First(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (repo *UserRepoPg) isStaffMember(user *User) (bool, error) {
	if user == nil || user.Phone == "" {
		return false, nil
	}

	var staffMember staffdata.StaffMemberPg
	if err := repo.db.Where("phone = ?", user.Phone).First(&staffMember).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			repo.logger.Debugf("staff member not found: %s", user.Phone)
			return false, nil
		}

		repo.logger.Errorf("failed to find staff member: %s", err.Error())
		return false, err
	}

	repo.logger.Debugf("Successfully found staff member: %s", staffMember.Phone)
	return true, nil
}

func (repo *UserRepoPg) isResident(user *User) (bool, error) {
	if user == nil || user.Phone == "" {
		return false, nil
	}

	var resident residents.ResidentPg
	if err := repo.db.Where("phone = ?", user.Phone).First(&resident).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			repo.logger.Debugf("resident not found: %s", user.Phone)
			return false, nil
		}

		repo.logger.Errorf("failed to find resident: %s", err.Error())
		return false, err
	}

	repo.logger.Debugf("Successfully found resident: %s", resident.Phone)
	return true, nil
}

func (repo *UserRepoPg) GenerateUserToken(u *User) (*jwt.Token, error) {
	isStaffMember, errStaffMember := repo.isStaffMember(u)
	isResident, errResident := repo.isResident(u)

	if errStaffMember != nil || errResident != nil {
		repo.logger.Errorf("Staff member not found for user: %s", u.Phone)
		return nil, errStaffMember
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"phone_number":  u.Phone,
		"isStaffMember": strconv.FormatBool(isStaffMember),
		"isResident":    strconv.FormatBool(isResident),
	})
	// "exp": time.Now().Add(1 * time.Hour).Unix(),
	repo.logger.Debugf("Successfully generate user token for user: %s", u.Phone)
	return token, nil
}

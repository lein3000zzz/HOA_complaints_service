package handlers

import (
	"DBPrototyping/pkg/residence"
	"DBPrototyping/pkg/staffdata"
	"DBPrototyping/pkg/userdata"
	"DBPrototyping/pkg/userdata/session"
	"DBPrototyping/pkg/utils"
	"errors"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	ErrWrongFormat = errors.New("wrong format")
	ErrRegistering = errors.New("error registering")
)

type UserHandler struct {
	ResidentsRepo residence.ResidentsController
	StaffRepo     staffdata.StaffRepo
	UserRepo      userdata.UserRepo
	Logger        *zap.SugaredLogger
}

// TODO залогировать тут всё!
func (h *UserHandler) Register(c *gin.Context) func(ctx *gin.Context) {

	return func(ctx *gin.Context) {
		phoneNumber := c.PostForm("phoneNumber")
		password := c.PostForm("password")
		// TODO: value может быть каким-то еще, ПОТОМУ ЧТО Я ХОЧУ ЧЕКБОКСЫЫЫЫЫЫЫЫЫЫ, не забыть про это
		isResident := c.PostForm("isResident") == "on"
		isStaffMember := c.PostForm("isStaffMember") == "on"

		if phoneNumber == "" || password == "" || len(password) < 5 || len(phoneNumber) < 5 ||
			len(phoneNumber) > 30 || len(password) > 30 || !utils.IsNumbers(phoneNumber) || !utils.IsNumbersAndLetters(password) {

			h.Logger.Error("phone number or password are in the wrong format")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": ErrWrongFormat.Error()})
		}

		responseJSON := gin.H{}

		user, errUserReg := h.UserRepo.Register(phoneNumber, password)
		if errUserReg != nil {
			responseJSON["userErr"] = errUserReg
			h.Logger.Errorf("register phone number error: %s", errUserReg.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			return
		}

		if isResident || isStaffMember {
			fullName := c.PostForm("fullName")
			if isResident {
				_, errResidentReg := h.ResidentsRepo.RegisterNewResident(phoneNumber, fullName)
				if errResidentReg != nil {
					h.Logger.Errorf("register resident error: %s", errResidentReg.Error())
					responseJSON["residentErr"] = errResidentReg
				}
			}

			if isStaffMember {
				_, errStaffReg := h.StaffRepo.RegisterNewMember(phoneNumber, fullName)
				if errStaffReg != nil {
					h.Logger.Errorf("register resident error: %s", errStaffReg.Error())
					responseJSON["residentErr"] = errStaffReg
				}
			}
		}

		h.Logger.Debugf("user registered %s", user.Phone)
		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *UserHandler) Login() func(c *gin.Context) {

	return func(c *gin.Context) {
		phoneNumber := c.PostForm("phoneNumber")
		password := c.PostForm("password")

		if phoneNumber == "" || password == "" || len(password) < 5 || len(phoneNumber) < 5 ||
			len(phoneNumber) > 30 || len(password) > 30 || !utils.IsNumbers(phoneNumber) || !utils.IsNumbersAndLetters(password) {

			h.Logger.Error("phone number or password are in the wrong format")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": ErrWrongFormat.Error()})

			return
		}

		responseJSON := gin.H{}

		userToLogin, errUserLogin := h.UserRepo.Authorize(phoneNumber, password)
		if errUserLogin != nil {
			responseJSON["userErr"] = errUserLogin.Error()

			h.Logger.Errorf("authorize phone number error: %s", errUserLogin.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, responseJSON)

			return
		}

		userSession := sessions.Default(c)

		userSession.Set("phoneNumber", userToLogin.Phone)

		setRoleAndSave := func(role session.Role) bool {
			userSession.Set("role", role)

			if err := userSession.Save(); err != nil {
				h.Logger.Infof("user session save failure: %s", err.Error())
				c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)

				return false
			}
			return true
		}

		isStaff, errStaff := h.StaffRepo.IsStaffMember(userToLogin.Phone)
		if errStaff != nil {
			responseJSON["staffErr"] = errStaff.Error()
			h.Logger.Errorf("isStaff error: %s", errStaff.Error())
		}
		if isStaff {
			if !setRoleAndSave(session.StaffRole) {
				return
			}

			c.JSON(http.StatusOK, responseJSON)
			return
		}

		isResident, errResident := h.ResidentsRepo.IsResident(userToLogin.Phone)
		if errResident != nil {
			responseJSON["residentErr"] = errResident.Error()
			h.Logger.Errorf("isResident error: %s", errResident.Error())
		}
		if isResident {
			if !setRoleAndSave(session.ResidentRole) {
				return
			}
		}

		c.JSON(http.StatusOK, responseJSON)
		return
	}
}

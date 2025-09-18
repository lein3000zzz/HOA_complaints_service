package handlers

import (
	"DBPrototyping/pkg/residence"
	"DBPrototyping/pkg/staffdata"
	"DBPrototyping/pkg/userdata"
	"DBPrototyping/pkg/userdata/session"
	"DBPrototyping/pkg/utils"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	ErrWrongFormat     = errors.New("wrong format: the data provided is <5 or >30 symbols or some symbols are not ascii")
	ErrRegisteringRole = errors.New("error registering a user: no role specified")
)

type UserHandler struct {
	SessionManager session.GinSessionManagerRepo
	ResidentsRepo  residence.ResidentsController
	StaffRepo      staffdata.StaffRepo
	UserRepo       userdata.UserRepo
	Logger         *zap.SugaredLogger
}

// TODO залогировать тут всё!
func (h *UserHandler) Register() func(c *gin.Context) {
	return func(c *gin.Context) {
		phoneNumber := c.PostForm("phoneNumber")
		password := c.PostForm("password")
		// TODO: value может быть каким-то еще, ПОТОМУ ЧТО Я ХОЧУ ЧЕКБОКСЫЫЫЫЫЫЫЫЫЫ, не забыть про это
		isResident := c.PostForm("isResident") == "on"
		isStaffMember := c.PostForm("isStaffMember") == "on"
		fullName := c.PostForm("fullName")

		if phoneNumber == "" || password == "" || fullName == "" || len(password) < 5 || len(phoneNumber) < 5 ||
			len(phoneNumber) > 40 || len(password) > 30 || !utils.IsNumbers(phoneNumber) || !utils.IsNumbersAndLetters(password) {

			h.Logger.Error("phone number, password or full name are in the wrong format")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": ErrWrongFormat.Error()})

			return
		}

		if !isResident && !isStaffMember {
			h.Logger.Error("At least one role should be specified in order to be registered")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": ErrRegisteringRole.Error()})

			return
		}

		responseJSON := gin.H{}

		user, errUserReg := h.UserRepo.Register(phoneNumber, password)
		if errUserReg != nil {
			responseJSON["userErr"] = errUserReg.Error()
			h.Logger.Errorf("register phone number error: %s", errUserReg.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)

			return
		}

		if isResident || isStaffMember {
			if isResident {
				_, errResidentReg := h.ResidentsRepo.RegisterNewResident(phoneNumber, fullName)
				if errResidentReg != nil {
					h.Logger.Errorf("register resident error: %s", errResidentReg.Error())
					responseJSON["residentErr"] = fmt.Sprintf("resident registration error (it failed): %s", errResidentReg.Error())
				}
			}

			if isStaffMember {
				_, errStaffReg := h.StaffRepo.RegisterNewMember(phoneNumber, fullName)
				if errStaffReg != nil {
					h.Logger.Errorf("register staff error: %s", errStaffReg.Error())
					responseJSON["residentErr"] = fmt.Sprintf("staff registration error (it failed): %s", errStaffReg.Error())
				}
			}
		}

		responseJSON["registered"] = user.Phone
		h.Logger.Debugf("user registered %s", user.Phone)
		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *UserHandler) Login() func(c *gin.Context) {
	return func(c *gin.Context) {
		phoneNumber := c.PostForm("phoneNumber")
		password := c.PostForm("password")

		if phoneNumber == "" || password == "" || len(password) < 5 || len(phoneNumber) < 5 ||
			len(phoneNumber) > 40 || len(password) > 30 || !utils.IsNumbers(phoneNumber) || !utils.IsNumbersAndLetters(password) {

			h.Logger.Error("phone number or password are in the wrong format")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": ErrWrongFormat.Error()})

			return
		}

		errResponseJSON := gin.H{}

		userToLogin, errUserLogin := h.UserRepo.Authorize(phoneNumber, password)
		if errUserLogin != nil {
			errResponseJSON["userErr"] = errUserLogin.Error()

			h.Logger.Errorf("authorize phone number error: %s", errUserLogin.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, errResponseJSON)

			return
		}

		h.SessionManager.SetUserSessionPhone(c, userToLogin.Phone)

		saveRole := func(role session.Role) error {
			h.SessionManager.SetUserSessionRole(c, role)

			if err := h.SessionManager.SaveSession(c); err != nil {
				errResponseJSON["sessionErr"] = err.Error()
				h.Logger.Errorf("save session error: %s", err.Error())

				return err
			}

			return nil
		}

		isStaff, errStaff := h.StaffRepo.IsStaffMember(userToLogin.Phone)
		if errStaff != nil {
			errResponseJSON["staffErr"] = errStaff.Error()
			h.Logger.Errorf("isStaff error: %s", errStaff.Error())

			c.AbortWithStatusJSON(http.StatusInternalServerError, errResponseJSON)
			return
		}

		if isStaff {
			if err := saveRole(session.StaffRole); err != nil {
				errResponseJSON["staffSaveErr"] = err.Error()
				h.Logger.Errorf("save staff role error: %s", err.Error())

				c.AbortWithStatusJSON(http.StatusInternalServerError, errResponseJSON)
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"phone": userToLogin.Phone,
				"type":  "login",
			})
			return
		}

		isResident, errResident := h.ResidentsRepo.IsResident(userToLogin.Phone)
		if errResident != nil {
			errResponseJSON["residentErr"] = errResident.Error()
			h.Logger.Errorf("isResident error: %s", errResident.Error())

			c.AbortWithStatusJSON(http.StatusInternalServerError, errResponseJSON)
			return
		}

		if isResident {
			if err := saveRole(session.ResidentRole); err != nil {
				errResponseJSON["residentSaveErr"] = err.Error()
				h.Logger.Errorf("save resident error: %s", err.Error())

				c.AbortWithStatusJSON(http.StatusInternalServerError, errResponseJSON)
				return
			}
		}

		//c.JSON(http.StatusOK, gin.H{"HEHHE": "HEHE"})
		c.JSON(http.StatusOK, gin.H{
			"phone": userToLogin.Phone,
			"type":  "login",
		})
		return
	}
}

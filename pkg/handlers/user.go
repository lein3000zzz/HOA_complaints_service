package handlers

import (
	"DBPrototyping/pkg/residence"
	"DBPrototyping/pkg/staffdata"
	"DBPrototyping/pkg/userdata"
	"DBPrototyping/pkg/userdata/session"
	"DBPrototyping/pkg/utils"
	"errors"
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
		responseJSON["type"] = "login"

		var finalErr error

		user, errUserReg := h.UserRepo.Register(phoneNumber, password)
		if errUserReg != nil {
			finalErr = errors.Join(finalErr, errUserReg)
			responseJSON["error"] = finalErr.Error()

			h.Logger.Errorf("register phone number error: %s", errUserReg.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)

			return
		}

		if isResident || isStaffMember {
			if isResident {
				_, errResidentReg := h.ResidentsRepo.RegisterNewResident(phoneNumber, fullName)
				if errResidentReg != nil {
					h.Logger.Errorf("register resident error: %s", errResidentReg.Error())

					finalErr = errors.Join(finalErr, errResidentReg)
				}
			}

			if isStaffMember {
				_, errStaffReg := h.StaffRepo.RegisterNewMember(phoneNumber, fullName)
				if errStaffReg != nil {
					h.Logger.Errorf("register staff error: %s", errStaffReg.Error())
					finalErr = errors.Join(finalErr, errStaffReg)
				}
			}
		}

		responseJSON["error"] = finalErr.Error()
		responseJSON["message"] = user.Phone
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

		responseJSON := gin.H{}
		responseJSON["type"] = "login"

		var finalErr error

		userToLogin, errUserLogin := h.UserRepo.Authorize(phoneNumber, password)
		if errUserLogin != nil {
			finalErr = errors.Join(finalErr, errUserLogin)
			responseJSON["error"] = finalErr.Error()

			h.Logger.Errorf("authorize phone number error: %s", errUserLogin.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, responseJSON)

			return
		}

		h.SessionManager.SetUserSessionPhone(c, userToLogin.Phone)

		saveRole := func(role session.Role) error {
			h.SessionManager.SetUserSessionRole(c, role)

			if err := h.SessionManager.SaveSession(c); err != nil {
				finalErr = errors.Join(finalErr, err)
				h.Logger.Errorf("save session error: %s", err.Error())

				return err
			}

			return nil
		}

		staffMember, errStaff := h.StaffRepo.GetStaffMemberByPhoneNumber(userToLogin.Phone)
		if errStaff != nil && errors.Is(errStaff, staffdata.ErrStaffMemberNotFound) {
			finalErr = errors.Join(finalErr, errStaff)
			responseJSON["error"] = finalErr.Error()

			h.Logger.Errorf("staffMember error: %s", errStaff.Error())

			c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			return
		}

		if staffMember != nil {
			if err := saveRole(session.StaffRole); err != nil {
				finalErr = errors.Join(finalErr, err)
				responseJSON["error"] = finalErr.Error()

				h.Logger.Errorf("save staff role error: %s", err.Error())

				c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
				return
			}

			responseJSON["message"] = userToLogin.Phone

			c.JSON(http.StatusOK, responseJSON)
			return
		}

		resident, errResident := h.ResidentsRepo.GetResidentByPhoneNumber(userToLogin.Phone)
		if errResident != nil && errors.Is(errResident, residence.ErrResidentNotFound) {
			finalErr = errors.Join(finalErr, errResident)
			responseJSON["error"] = finalErr.Error()

			h.Logger.Errorf("resident error: %s", errResident.Error())

			c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
			return
		}

		if resident != nil {
			if err := saveRole(session.ResidentRole); err != nil {
				finalErr = errors.Join(finalErr, err)
				responseJSON["error"] = finalErr.Error()

				h.Logger.Errorf("save resident error: %s", err.Error())

				c.AbortWithStatusJSON(http.StatusInternalServerError, responseJSON)
				return
			}
		}

		//c.JSON(http.StatusOK, gin.H{"HEHHE": "HEHE"})
		//c.JSON(http.StatusOK, gin.H{
		//	"phone": userToLogin.Phone,
		//	"type":  "login",
		//})
		responseJSON["error"] = finalErr.Error()
		responseJSON["message"] = userToLogin.Phone
		return
	}
}

func (h *UserHandler) DeleteUser() func(c *gin.Context) {
	return func(c *gin.Context) {
		phoneNumber := c.Param("phoneNumber")

		var finalErr error
		responseJSON := gin.H{}

		userDeleteErr := h.UserRepo.DeleteByPhone(phoneNumber)
		if userDeleteErr != nil {
			finalErr = errors.Join(finalErr, userDeleteErr)
		}

		residentDeleteErr := h.ResidentsRepo.DeleteResidentByPhone(phoneNumber)
		if residentDeleteErr != nil {
			finalErr = errors.Join(finalErr, residentDeleteErr)
		}

		staffDeleteErr := h.StaffRepo.DeleteByPhone(phoneNumber)
		if staffDeleteErr != nil {
			finalErr = errors.Join(finalErr, staffDeleteErr)
		}

		if finalErr != nil {
			h.Logger.Errorf("user delete error: %s", finalErr.Error())
			responseJSON["error"] = finalErr.Error()

			c.JSON(http.StatusInternalServerError, responseJSON)
			return
		}

		responseJSON["message"] = "success"
		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *UserHandler) GetAllUsers() func(c *gin.Context) {
	return func(c *gin.Context) {
		users, errGetUsers := h.UserRepo.GetAll()

		responseJSON := gin.H{}

		if errGetUsers != nil {
			h.Logger.Errorf("get users error: %s", errGetUsers.Error())
			responseJSON["error"] = errGetUsers.Error()
			c.JSON(http.StatusInternalServerError, responseJSON)
		}

		responseJSON["users"] = users
		c.JSON(http.StatusOK, responseJSON)
	}
}

func (h *UserHandler) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.SessionManager.ClearSession(c)
		err := h.SessionManager.SaveSession(c)
		if err != nil {
			h.Logger.Errorf("save session error: %s", err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		}
		c.Redirect(http.StatusSeeOther, "/login")
	}
}

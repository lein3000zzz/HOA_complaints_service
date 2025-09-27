package main

import (
	"DBPrototyping/pkg/company"
	"DBPrototyping/pkg/handlers"
	"DBPrototyping/pkg/requests"
	"DBPrototyping/pkg/residence"
	"DBPrototyping/pkg/userdata"
	"DBPrototyping/pkg/userdata/session"
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ТСЖ - HomeOwner Association (HOA)
func main() {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		fmt.Println("Error initializing zap logger:", err)
		return
	}

	defer func(zapLogger *zap.Logger) {
		err := zapLogger.Sync()
		if err != nil {
			fmt.Println("Error syncing zap logger:", err)
		}
	}(zapLogger)

	logger := zapLogger.Sugar()

	err = godotenv.Load("tools/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := "host=localhost user=postgres password=lein dbname=HouseholdTickets port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	if errAuto := db.AutoMigrate(
		&residence.ResidentPg{},
		&residence.HousePg{},
		&residence.ResidentHousePg{},
		&company.StaffMemberPg{},
		&company.SpecializationPg{},
		&company.StaffMemberSpecializationPg{},
		&requests.RequestPg{},
		&userdata.UserPg{},
	); errAuto != nil {
		logger.Errorf("AutoMigrate failed: %v", errAuto)
		return
	}

	pageHandler := &handlers.PageHandler{Logger: logger}

	r := gin.Default()
	store, errRedisStore := redis.NewStore(10, "tcp", "localhost:6379", "", os.Getenv("REDIS_PASSWORD"), []byte(os.Getenv("UNIFIED_PASSWORD")))
	if errRedisStore != nil {
		fmt.Println("Error initializing redis store:", errRedisStore)
		log.Fatal(errRedisStore)
	}
	r.Use(sessions.Sessions("hoa_project", store))

	sm := &session.GinSessionManager{
		Logger: logger,
	}

	userRepo := userdata.NewUserRepoPg(db, logger)
	residentsRepo := residence.NewResidentPgRepo(logger, db)
	staffRepo := company.NewStaffRepoPostgres(logger, db)
	reqRepo := requests.NewRequestPgRepo(logger, db)

	userHandler := handlers.UserHandler{
		SessionManager: sm,
		StaffRepo:      staffRepo,
		ResidentsRepo:  residentsRepo,
		UserRepo:       userRepo,
		Logger:         logger,
	}

	reqHandler := handlers.RequestsHandler{
		RequestsRepo:  reqRepo,
		Logger:        logger,
		StaffRepo:     staffRepo,
		UserRepo:      userRepo,
		ResidentsRepo: residentsRepo,
	}

	staffHandler := handlers.StaffHandler{
		StaffRepo: staffRepo,
		Logger:    logger,
	}

	resHandler := handlers.ResidentsHandler{
		ResidentsRepo: residentsRepo,
		Logger:        logger,
	}

	r.Use(sm.UserFromSession())

	api := r.Group("/api")
	staffGroup := r.Group("/staff")
	residentGroup := r.Group("/resident")

	staffApiGroup := api.Group("/staff")
	residentApiGroup := api.Group("/resident")

	staffGroup.Use(sm.UserFromSession(), sm.RequireRoles(session.StaffRole))
	staffApiGroup.Use(sm.UserFromSession(), sm.RequireRoles(session.StaffRole))
	residentGroup.Use(sm.UserFromSession(), sm.RequireRoles(session.ResidentRole, session.StaffRole))
	residentApiGroup.Use(sm.UserFromSession(), sm.RequireRoles(session.ResidentRole, session.StaffRole))

	r.Static("/static", "./web/static")
	pageHandler.InitHTML()

	residentGroup.GET("/my-requests", pageHandler.UserRequestsPage())
	residentApiGroup.GET("/requests", reqHandler.GetRequestsForUser())

	api.POST("/login", userHandler.Login())
	staffApiGroup.POST("/register", userHandler.Register())
	r.GET("/login", pageHandler.LoginPage())
	residentGroup.GET("/create-request", pageHandler.CreateRequestPage())
	residentApiGroup.POST("/create-request", reqHandler.CreateRequest())
	r.GET("/logout", userHandler.Logout())
	r.GET("/", pageHandler.MainPage())
	staffGroup.GET("/register", pageHandler.RegisterPage())
	staffGroup.GET("/admin-panel", pageHandler.AdminPage())

	staffApiGroup.GET("/users/list", userHandler.GetAllUsersFiltered())
	staffApiGroup.DELETE("/users/delete/:phoneNumber", userHandler.DeleteUser())

	staffApiGroup.GET("/users/info/:phoneNumber", userHandler.GetUserDetails())

	staffApiGroup.GET("/users/resident/info", resHandler.GetHousesForResident())
	staffApiGroup.DELETE("/users/resident/remove-house", resHandler.DeleteHouseForResident())
	staffApiGroup.POST("/users/resident/add-house", resHandler.AddResidentHouse())
	staffApiGroup.POST("/users/resident/update-house", resHandler.UpdateHouseAddress())
	staffApiGroup.GET("/users/resident/get-number", resHandler.GetResidentPhoneNumberByID())

	staffApiGroup.GET("/users/staff/info", staffHandler.GetSpecializationsForStaffMember())
	staffApiGroup.DELETE("/users/staff/delete-spec", staffHandler.DeactivateSpecialization())
	staffApiGroup.POST("/users/staff/add-specialization", staffHandler.AddStaffSpecialization())

	staffApiGroup.GET("/organizations/list", staffHandler.GetOrganizations())
	staffApiGroup.POST("/organizations/create", staffHandler.CreateOrganization())
	staffApiGroup.POST("/organizations/update", staffHandler.UpdateOrganizationName())
	staffGroup.GET("/organizations/panel", pageHandler.OrganizationsPage())

	staffApiGroup.GET("/specializations/list", staffHandler.GetAllSpecs())
	staffApiGroup.POST("/specializations/create", staffHandler.CreateSpecialization())
	staffGroup.GET("/specializations/info", pageHandler.SpecializationsPage())

	staffApiGroup.GET("/houses/list", resHandler.GetHouses())
	staffApiGroup.POST("/houses/create", resHandler.CreateHouse())
	staffGroup.GET("/houses/info", pageHandler.HousesPage())

	staffApiGroup.GET("/requests/panel", reqHandler.GetRequestsForAdmin())
	staffApiGroup.POST("/requests/panel/update", reqHandler.UpdateRequest())
	staffApiGroup.GET("/requests/panel/update/random-assign", staffHandler.GetLeastBusyByJobID())
	staffApiGroup.DELETE("/requests/panel/delete/:id", reqHandler.DeleteRequest())

	staffGroup.GET("/requests/panel", pageHandler.AdminRequestsPage())
	staffGroup.GET("/users/panel", pageHandler.UsersManagerPage())

	log.Fatal(r.Run(":8000"))
}

package main

import (
	"DBPrototyping/pkg/handlers"
	"DBPrototyping/pkg/requests"
	"DBPrototyping/pkg/residence"
	"DBPrototyping/pkg/staffdata"
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

// ТСЖ - HomeOwner Association
// TODO сделать контекстный таймаут в каждой операции с бд (тяжело...)
func main() {
	//err := godotenv.Load()
	//if err != nil {
	//	log.Fatal("Error loading .env file")
	//}
	//
	//dsn := "host=localhost user=postgres password=lein dbname=HouseholdTickets port=5432 sslmode=disable"
	//db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	//
	//router := gin.Default()
	//
	//router.GET("/ping", func(c *gin.Context) {
	//	c.JSON(http.StatusOK, gin.H{
	//		"message": "pong",
	//	})
	//})
	//
	//err2 := router.Run()
	//if err2 != nil {
	//	log.Fatal("Error starting the server:", err2)
	//} // listen and serve on 0.0.0.0:8080
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
		&staffdata.StaffMemberPg{},
		&staffdata.SpecializationPg{},
		&staffdata.StaffMemberSpecializationPg{},
		&requests.RequestPg{},
		&userdata.UserPg{},
	); errAuto != nil {
		logger.Errorf("AutoMigrate failed: %v", errAuto)
		return
	}

	pageH := &handlers.PageHandler{Logger: logger}

	r := gin.Default()
	store, errRedisStore := redis.NewStore(10, "tcp", "localhost:6379", "", os.Getenv("REDIS_PASSWORD"), []byte(os.Getenv("UNIFIED_PASSWORD")))
	if errRedisStore != nil {
		fmt.Println("Error initializing redis store:", errRedisStore)
		log.Fatal(errRedisStore)
	}
	r.Use(sessions.Sessions("hoa_project", store))

	api := r.Group("/api")
	//
	//apiStaffGroup := api.Group("/staff", session.RequireRoles("staff"))
	//apiResidentsGroup := api.Group("/residents", session.RequireRoles("resident"))

	userRepo := userdata.NewUserRepoPg(db, logger)

	residentsRepo := residence.NewResidentPgRepo(logger, db)
	staffRepo := staffdata.NewStaffRepoPostgres(logger, db)

	sm := &session.GinSessionManager{
		Logger: logger,
	}

	userHandler := handlers.UserHandler{
		SessionManager: sm,
		StaffRepo:      staffRepo,
		ResidentsRepo:  residentsRepo,
		UserRepo:       userRepo,
		Logger:         logger,
	}

	//api.GET("/sd", userHandler.Login())

	//api.GET("/incr", func(c *gin.Context) {
	//	session := sessions.Default(c)
	//	var count int
	//	v := session.Get("count")
	//	if v == nil {
	//		count = 0
	//	} else {
	//		count = v.(int)
	//		count++
	//	}
	//	session.Set("count", count)
	//	session.Save()
	//	c.JSON(200, gin.H{"count": count})
	//})

	//_, err1 := userRepo.Register("7777777", "abobus")
	//_, err2 := staffRepo.RegisterNewMember("7777777", "lein3000")
	//if err1 != nil || err2 != nil {
	//	fmt.Println(err1)
	//	fmt.Println(err2)
	//	return
	//}
	//log.Fatal("Successfully registered")

	staffApiGroup := api.Group("/staff")
	staffApiGroup.Use(sm.RequireRoles(session.StaffRole))

	staffGroup := r.Group("/staff")
	staffGroup.Use(sm.RequireRoles(session.StaffRole))

	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*.tmpl")
	//r.LoadHTMLGlob("web/static/html/*.html")
	api.POST("/login", userHandler.Login())
	staffApiGroup.POST("/register", userHandler.Register())
	r.GET("/login", pageH.LoginPage())
	r.GET("/", pageH.MainPage())
	staffGroup.GET("/register", pageH.RegisterPage())

	r.Run(":8000")

	//err2 := GenerateSeed(db, logger)
	//fmt.Println(err2)
}

//func GenerateSeed(db *gorm.DB, logger *zap.SugaredLogger) error {
//	rand.Seed(time.Now().UnixNano())
//
//	// repos using existing constructors
//	resRepo := residence.NewResidentPgRepo(logger, db)
//	staffRepo := staffdata.NewStaffRepoPostgres(logger, db)
//	reqRepo := requests.NewRequestPgRepo(logger, db)
//	userRepo := userdata.NewUserRepoPg(db, logger)
//
//	numHouses := 200
//	numResidents := 500
//	numSpecializations := 12
//	numStaff := 150
//	numRequests := 1000
//	password := "Password123!"
//
//	houseIDs := make([]int, 0, numHouses)
//	residentIDs := make([]string, 0, numResidents)
//	residentToHouses := make(map[string][]int)
//
//	for i := 1; i <= numHouses; i++ {
//		addr := fmt.Sprintf("Seed St. %d, Building %d", (i/10)+1, i%10+1)
//		h, err := resRepo.RegisterNewHouse(addr)
//		if err != nil {
//			logger.Warnf("RegisterNewHouse failed for %s: %v", addr, err)
//			continue
//		}
//		houseIDs = append(houseIDs, h.ID)
//	}
//
//	if len(houseIDs) == 0 {
//		return fmt.Errorf("no houses created, aborting")
//	}
//
//	pickHouse := func() int {
//		return houseIDs[rand.Intn(len(houseIDs))]
//	}
//
//	for i := 1; i <= numResidents; i++ {
//		phone := fmt.Sprintf("700%07d", i) // unique-ish phone
//		fullName := fmt.Sprintf("Resident Seed %d", i)
//
//		if _, err := userRepo.Register(phone, password); err != nil {
//			logger.Debugf("user register skipped for %s: %v", phone, err)
//		}
//
//		r, err := resRepo.RegisterNewResident(phone, fullName)
//		if err != nil {
//			logger.Warnf("RegisterNewResident failed for %s: %v", phone, err)
//			continue
//		}
//		residentIDs = append(residentIDs, r.ID)
//
//		numAssoc := 1 + rand.Intn(3)
//		for a := 0; a < numAssoc; a++ {
//			hid := pickHouse()
//			if err := resRepo.AddResidentAddressAssoc(r.ID, hid); err != nil {
//				logger.Warnf("AddResidentAddressAssoc failed resident=%s house=%d: %v", r.ID, hid, err)
//				continue
//			}
//			residentToHouses[r.ID] = append(residentToHouses[r.ID], hid)
//		}
//	}
//
//	specIDs := make([]string, 0, numSpecializations)
//	jobTitles := []string{
//		"plumber", "electrician", "locksmith", "carpenter",
//		"painter", "heating_specialist", "roofer", "glazier",
//		"cleaning", "gardener", "inspector", "mason",
//	}
//	for i := 0; i < numSpecializations && i < len(jobTitles); i++ {
//		title := jobTitles[i]
//		spec, err := staffRepo.RegisterNewSpecialization(title)
//		if err != nil {
//			logger.Warnf("RegisterNewSpecialization failed %s: %v", title, err)
//			continue
//		}
//		specIDs = append(specIDs, spec.ID)
//	}
//
//	for i := 1; i <= numStaff; i++ {
//		phone := fmt.Sprintf("800%07d", i)
//		fullName := fmt.Sprintf("Staff Seed %d", i)
//
//		// create user account for staff first (satisfy FK)
//		if _, err := userRepo.Register(phone, password); err != nil {
//			logger.Debugf("user register skipped for staff %s: %v", phone, err)
//		}
//
//		member, err := staffRepo.RegisterNewMember(phone, fullName)
//		if err != nil {
//			logger.Warnf("RegisterNewMember failed %s: %v", phone, err)
//			continue
//		}
//
//		if len(specIDs) > 0 {
//			numAssign := 1 + rand.Intn(2)
//			for a := 0; a < numAssign; a++ {
//				specID := specIDs[rand.Intn(len(specIDs))]
//
//				// use the repo method instead of direct DB insert
//				if err := staffRepo.AddStaffMemberSpecializationAssoc(member.ID, specID); err != nil {
//					logger.Debugf("failed to create staff-specialization mapping (member=%d spec=%s): %v", member.ID, specID, err)
//				}
//			}
//		}
//	}
//
//	complaints := []string{
//		"Leaking pipe in kitchen", "No hot water", "Broken window",
//		"Electrical short in corridor", "Clogged drain", "Broken lock",
//		"Peeling paint in stairwell", "Elevator not working",
//	}
//
//	createdRequests := 0
//	for i := 0; i < numRequests; i++ {
//		if len(residentIDs) == 0 {
//			break
//		}
//		resID := residentIDs[rand.Intn(len(residentIDs))]
//		// pick a house associated with this resident if possible
//		hlist := residentToHouses[resID]
//		var hid int
//		if len(hlist) > 0 {
//			hid = hlist[rand.Intn(len(hlist))]
//		} else {
//			hid = pickHouse()
//		}
//
//		reqType := requests.TypeApartmentInternal
//		if rand.Intn(2) == 0 {
//			reqType = requests.TypeHouseCommon
//		}
//		complaint := complaints[rand.Intn(len(complaints))]
//
//		ir := requests.InitialRequestData{
//			ResidentID:  resID,
//			HouseID:     hid,
//			RequestType: reqType,
//			Complaint:   complaint,
//		}
//		if _, err := reqRepo.CreateRequest(ir); err != nil {
//			logger.Debugf("CreateRequest failed (resident=%s house=%d): %v", resID, hid, err)
//			continue
//		}
//		createdRequests++
//	}
//
//	logger.Infof("Seed finished: houses=%d residents=%d staff=%d specs=%d requests=%d",
//		len(houseIDs), len(residentIDs), numStaff, len(specIDs), createdRequests)
//
//	return nil
//}

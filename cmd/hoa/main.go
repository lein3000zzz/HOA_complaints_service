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
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
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

	r := gin.Default()
	store, _ := redis.NewStore(10, "tcp", "localhost:6379", "", os.Getenv("REDIS_PASSWORD"), []byte(os.Getenv("UNIFIED_PASSWORD")))
	r.Use(sessions.Sessions("hoa_project", store))

	api := r.Group("/api")

	apiStaffGroup := api.Group("/staff", session.RequireRoles("staff"))
	apiResidentsGroup := api.Group("/residents", session.RequireRoles("resident"))

	h := handlers.UserHandler{}
	api.GET("/sd", h.Login())

	api.GET("/incr", func(c *gin.Context) {
		session := sessions.Default(c)
		var count int
		v := session.Get("count")
		if v == nil {
			count = 0
		} else {
			count = v.(int)
			count++
		}
		session.Set("count", count)
		session.Save()
		c.JSON(200, gin.H{"count": count})
	})

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
//	// configuration
//	numHouses := 200
//	numResidents := 500
//	numSpecializations := 12
//	numStaff := 150
//	numRequests := 1000
//	password := "Password123!" // sample password for all created users
//
//	// Keep track of created IDs and associations
//	houseIDs := make([]int, 0, numHouses)
//	residentIDs := make([]string, 0, numResidents)
//	residentToHouses := make(map[string][]int)
//
//	// 1) Create houses
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
//	// helper to pick a random house
//	pickHouse := func() int {
//		return houseIDs[rand.Intn(len(houseIDs))]
//	}
//
//	// 2) Create residents and users, and link each resident to 1-3 houses
//	for i := 1; i <= numResidents; i++ {
//		phone := fmt.Sprintf("700%07d", i) // unique-ish phone
//		fullName := fmt.Sprintf("Resident Seed %d", i)
//
//		// create user account first (satisfy foreign key)
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
//		// attach 1-3 houses to resident
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
//	// 3) Create specializations
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
//	// 4) Create staff members, users, and link some specializations via direct DB inserts
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
//		// randomly assign 1-2 specializations (use repo helper to create join record)
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
//	// sample complaints
//	complaints := []string{
//		"Leaking pipe in kitchen", "No hot water", "Broken window",
//		"Electrical short in corridor", "Clogged drain", "Broken lock",
//		"Peeling paint in stairwell", "Elevator not working",
//	}
//
//	// 5) Create requests - try to favor houses associated with the resident
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

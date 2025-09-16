package main

import (
	"DBPrototyping/pkg/requests"
	"DBPrototyping/pkg/residence"
	"DBPrototyping/pkg/staffdata"
	"DBPrototyping/pkg/userdata"
	"fmt"
	"log"
	"strings"

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
		// Ignore missing-constraint error produced by Postgres when GORM attempts to DROP a non-existing constraint.
		// Treat other errors as fatal.
		if strings.Contains(errAuto.Error(), "constraint \"uni_residents_phone_number\"") ||
			strings.Contains(errAuto.Error(), "SQLSTATE 42704") {
			logger.Warnf("AutoMigrate returned missing-constraint error; ignoring: %v", errAuto)
		} else {
			logger.Errorf("AutoMigrate failed: %v", errAuto)
			return
		}
	}

	//err2 := GenerateSeed(db, logger)
	//fmt.Println(err2)
}

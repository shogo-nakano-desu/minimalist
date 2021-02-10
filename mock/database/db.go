package database

import (
	"log"
	"os"

	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
)

// Connect to DB
func GormConnect() *gorm.DB {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	DBMS := os.Getenv("belongings_DBMS")
	USER := os.Getenv("belongings_USER")
	PASS := os.Getenv("belongings_PASS")
	DBNAME := os.Getenv("belongings_DBNAME")
	CONNECT := USER + ":" + PASS + "@/" + DBNAME + "?parseTime=true"
	db, err := gorm.Open(DBMS, CONNECT)
	if err != nil {
		panic(err.Error())
	}
	return db
}

// DB migration
func DbInit() {
	db := GormConnect()
	defer db.Close()
	db.AutoMigrate(&BModel{})
	db.AutoMigrate(&User{})
}

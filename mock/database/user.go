package database

import (
	"go-web-app/mock/crypto"
	"log"
)

// User Model
type User struct {
	ID       int    `json:"id" gorm:"praimaly_key"`
	Username string `form:"username" binding:"required" gorm:"unique;not null"`
	Password string `form:"password" binding:"required"`
}

//User sign up procecc func
func CreateUser(username string, password string) []error {
	passwordEncrypt, err := crypto.PasswordEncrypt(password)
	if err != nil {
		log.Fatal(err)
	}
	db := GormConnect()
	defer db.Close()
	//Insert process
	if cerr := db.Create(&User{Username: username, Password: passwordEncrypt}).GetErrors(); len(cerr) > 0 {
		// FAIL
		log.Println("Insert error")
		log.Println(cerr)
		return cerr
	}
	// SUCCESS
	log.Println("nil?")
	return nil
}

// Get one user info
func GetUser(username string) User {
	db := GormConnect()
	var user User
	db.First(&user, "username = ?", username)
	db.Close()
	return user
}

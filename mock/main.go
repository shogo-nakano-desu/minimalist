package main

import (
	"go-web-app/mock/crypto"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
)

// BModel = BelongigsModel
type BModel struct {
	gorm.Model
	// Id       int `gorm:"primary_key`
	Name     string
	Price    int
	Quantity int
	SellBuy  string
	Date     string // ex) 2020/01/01 It will CAST in SQL later
}

// Totalを計算するためのstruct
type Presult struct {
	Totalprice int
}

type Qresult struct {
	Totalquantity int
}

// User モデルの宣言
type User struct {
	gorm.Model
	Username string `form:"username" binding:"required" gorm:"unique;not null"`
	Password string `form:"password" binding:"required"`
}

func gormConnect() *gorm.DB {
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
func dbInit() {
	db := gormConnect()
	defer db.Close()
	db.AutoMigrate(&BModel{})
	db.AutoMigrate(&User{})
}

// DB Create
func dbCreate(name string, price int, quantity int, sellbuy string, date string) {
	db := gormConnect()
	defer db.Close()
	db.Create(&BModel{
		Name:     name,
		Price:    price,
		Quantity: quantity,
		SellBuy:  sellbuy,
		Date:     date,
	})
}

// DB Update
func dbUpdate(id int, name string, price int, quantity int, sellbuy string, date string) {
	db := gormConnect()
	defer db.Close()
	var belongings BModel
	db.First(&belongings, id)
	belongings.Name = name
	belongings.Price = price
	belongings.Quantity = quantity
	belongings.SellBuy = sellbuy
	belongings.Date = date
	db.Save(&belongings)
}

// DB Delete
func dbDelete(id int) {
	db := gormConnect()
	defer db.Close()
	var belongings BModel
	db.First(&belongings, id)
	db.Unscoped().Delete(&belongings)
}

// DB Get All
func dbGetAll() []BModel {
	db := gormConnect()
	defer db.Close()
	var b_models []BModel
	db.Order("date desc").Find(&b_models)
	return b_models
}

// DB Get One
func dbGetOne(id int) BModel {
	db := gormConnect()
	defer db.Close()
	var belongings BModel
	db.First(&belongings, id)
	return belongings
}

// DB Get number of belongings list
func dbGetNum() int {
	db := gormConnect()
	defer db.Close()
	var num int
	db.Table("b_models").Count(&num)
	return num
}

// DB Get Sum of quantity
func dbGetSumQuantity() Qresult {
	db := gormConnect()
	defer db.Close()
	var qresult Qresult
	db.Table("b_models").Select("sum(quantity) as totalquantity").Scan(&qresult)
	return qresult

}

// DB Get Sum of price
func dbGetSumPrice() Presult {
	db := gormConnect()
	defer db.Close()
	var presult Presult
	db.Table("b_models").Select("sum(price) as totalprice").Scan(&presult)
	return presult
}

// ユーザー登録処理
func createUser(username string, password string) []error {
	passwordEncrypt, _ := crypto.PasswordEncrypt(password)
	db := gormConnect()
	defer db.Close()
	// Insert処理
	if err := db.Create(&User{Username: username, Password: passwordEncrypt}).GetErrors(); err != nil {
		return err
	}
	return nil
}

// ユーザーを一件取得
func getUser(username string) User {
	db := gormConnect()
	var user User
	db.First(&user, "username = ?", username)
	db.Close()
	return user
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("front/*.html")

	dbInit()

	// ユーザー登録画面
	router.GET("/signup", func(c *gin.Context) {

		c.HTML(200, "signup.html", gin.H{})
	})

	// ユーザー登録
	router.POST("/signup", func(c *gin.Context) {
		var form User
		// バリデーション処理
		if err := c.Bind(&form); err != nil {
			c.HTML(http.StatusBadRequest, "signup.html", gin.H{"err": err})
			c.Abort()
		} else {
			username := c.PostForm("username")
			password := c.PostForm("password")
			// 登録ユーザーが重複していた場合にはじく処理
			if err := createUser(username, password); err != nil {
				c.HTML(http.StatusBadRequest, "signup.html", gin.H{"err": err})
			}
			c.Redirect(302, "/")
		}
	})

	// ユーザーログイン画面
	router.GET("/login", func(c *gin.Context) {

		c.HTML(200, "login.html", gin.H{})
	})

	// ユーザーログイン
	router.POST("/login", func(c *gin.Context) {

		// DBから取得したユーザーパスワード(Hash)
		dbPassword := getUser(c.PostForm("username")).Password
		log.Println(dbPassword)
		// フォームから取得したユーザーパスワード
		formPassword := c.PostForm("password")

		// ユーザーパスワードの比較
		if err := crypto.CompareHashAndPassword(dbPassword, formPassword); err != nil {
			log.Println("ログインできませんでした")
			c.HTML(http.StatusBadRequest, "login.html", gin.H{"err": err})
			c.Abort()
		} else {
			log.Println("ログインできました")
			c.Redirect(302, "/")
		}
	})

	// index
	router.GET("/", func(c *gin.Context) {
		b_models := dbGetAll()
		num := dbGetNum()
		sumQuantity := dbGetSumQuantity()
		sumPrice := dbGetSumPrice()
		c.HTML(200, "belongings.html", gin.H{"b_models": b_models, "num": num, "sumQuantity": sumQuantity.Totalquantity, "sumPrice": sumPrice.Totalprice})
	})

	// Create
	router.POST("/new", func(c *gin.Context) {
		name := c.PostForm("name")
		price, perr := strconv.Atoi(c.PostForm("price"))
		if perr != nil {
			panic(perr)
		}
		quantity, qerr := strconv.Atoi(c.PostForm("quantity"))
		if qerr != nil {
			panic(qerr)
		}
		sellbuy := c.PostForm("sellbuy")
		date := c.PostForm("date")
		dbCreate(name, price, quantity, sellbuy, date)
		c.Redirect(302, "/")
	})

	// Edit
	router.GET("/edit/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			panic(err)
		}
		belongings := dbGetOne(id)
		c.HTML(200, "edit.html", gin.H{"belongings": belongings})
	})

	// Update
	router.POST("/updated/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			panic(err)
		}
		name := c.PostForm("name")
		price, perr := strconv.Atoi(c.PostForm("price"))
		if perr != nil {
			panic(perr)
		}
		quantity, qerr := strconv.Atoi(c.PostForm("quantity"))
		if qerr != nil {
			panic(qerr)
		}
		sellbuy := c.PostForm("sellbuy")
		date := c.PostForm("date")
		dbUpdate(id, name, price, quantity, sellbuy, date)
		c.Redirect(302, "/")
	})

	// delete
	router.POST("/delete/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			panic(err)
		}
		dbDelete(id)
		c.Redirect(302, "/")
	})

	// delete_confirm
	router.GET("/delete_confirm/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			panic(err)
		}
		belongings := dbGetOne(id)
		c.HTML(200, "delete.html", gin.H{"belongings": belongings})
	})

	router.Run()

}

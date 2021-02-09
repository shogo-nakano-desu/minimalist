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

// Total Price Struct
type Presult struct {
	Totalprice int
}

// Total Quantity struct
type Qresult struct {
	Totalquantity int
}

// User Model
type User struct {
	gorm.Model
	Username string `form:"username" binding:"required" gorm:"unique;not null"`
	Password string `form:"password" binding:"required"`
	BModel   BModel // one user has one Belongings Model
}

// Connect to DB
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

// DB Get Sum of quantity, Sell Item
func dbGetSumQuantitySell() Qresult {
	db := gormConnect()
	defer db.Close()
	var qresultSell Qresult
	db.Table("b_models").Select("sum(-quantity) as totalquantity").Where("sell_buy = ?", "sell").Scan(&qresultSell) // sum of sell items
	return qresultSell
}

// DB Get Sum of quantity, Buy Item
func dbGetSumQuantityBuy() Qresult {
	db := gormConnect()
	defer db.Close()
	var qresultBuy Qresult
	db.Table("b_models").Select("sum(quantity) as totalquantity").Where("sell_buy = ?", "buy").Scan(&qresultBuy) // sum of buy items
	log.Println(qresultBuy.Totalquantity)
	return qresultBuy
}

// calculation quantity
func calcQuantity() int {
	return dbGetSumQuantitySell().Totalquantity + dbGetSumQuantityBuy().Totalquantity
}

// DB Get Sum of price, Sell Item
func dbGetSumPriceSell() Presult {
	db := gormConnect()
	defer db.Close()
	var presultsell Presult
	db.Table("b_models").Select("sum(price) as totalprice").Scan(&presultsell)
	return presultsell
}

// DB Get Sum of price, Buy Item
func dbGetSumPriceBuy() Presult {
	db := gormConnect()
	defer db.Close()
	var presultbuy Presult
	db.Table("b_models").Select("sum(price) as totalprice").Scan(&presultbuy)
	return presultbuy
}

// calculation price
func calcPrice() int {
	return dbGetSumPriceSell().Totalprice + dbGetSumPriceBuy().Totalprice
}

//User sign up procecc func
func createUser(username string, password string) []error {
	passwordEncrypt, err := crypto.PasswordEncrypt(password)
	if err != nil {
		log.Fatal(err)
	}
	db := gormConnect()
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

	// index
	router.GET("/", func(c *gin.Context) {
		b_models := dbGetAll()
		num := dbGetNum()
		sumQuantity := calcQuantity()
		sumPrice := calcPrice()
		c.HTML(200, "belongings.html", gin.H{"b_models": b_models, "num": num, "sumQuantity": sumQuantity, "sumPrice": sumPrice})
	})

	// User sign up page
	router.GET("/signup", func(c *gin.Context) {
		c.HTML(200, "signup.html", gin.H{})
	})

	// User sign up process
	router.POST("/signup", func(c *gin.Context) {
		var form User
		// Validation
		if err := c.Bind(&form); err != nil {
			c.HTML(http.StatusBadRequest, "signup.html", gin.H{"err": err})
			log.Println("fail to login because your info is invalid")
			c.Abort()
		} else {
			username := c.PostForm("username")
			password := c.PostForm("password")

			// Process to reject duplicate registered users
			if err := createUser(username, password); err != nil {
				log.Printf("%T\n", err)
				c.HTML(http.StatusBadRequest, "signup.html", gin.H{"err": err})
				c.Abort()
			} else {
				log.Println("success to signup!")
				c.Redirect(302, "/login")
			}
		}
	})

	// User login page
	router.GET("/login", func(c *gin.Context) {

		c.HTML(200, "login.html", gin.H{})
	})

	// User login
	router.POST("/login", func(c *gin.Context) {

		// UserPassword from DB(Hash)
		dbPassword := getUser(c.PostForm("username")).Password
		log.Println(dbPassword)
		// UserPassword from Form(non-Hash)
		formPassword := c.PostForm("password")

		// Compare User password(from DB & Form)
		if err := crypto.CompareHashAndPassword(dbPassword, formPassword); err != nil {
			log.Println("Failed to login")
			c.HTML(http.StatusBadRequest, "login.html", gin.H{"err": err})
			c.Abort()
		} else {
			log.Println("Success to login")
			c.Redirect(302, "/")
		}
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

package database

import "github.com/jinzhu/gorm"

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

package database

import (
	"log"
	"sip_Smart/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Db *gorm.DB
var err error

func Initialize(c *gin.Context) {

	dsn := "host=localhost user=postgres password=5102 dbname=sipsmart port=5432 sslmode=disable TimeZone=Asia/Shanghai"

	Db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	Db.AutoMigrate(&models.Admin_Credentials{})
	Db.AutoMigrate(&models.Temp_User{})
	Db.AutoMigrate(&models.User{})
	Db.AutoMigrate(&models.Category{})
	Db.AutoMigrate(&models.Review{})
	Db.AutoMigrate(&models.Product{})
	Db.AutoMigrate(&models.Address{})
	Db.AutoMigrate(&models.Cart{})
	Db.AutoMigrate(&models.Cart_Item{})
	Db.AutoMigrate(&models.Order{})
	Db.AutoMigrate(&models.Order_Item{})
	Db.AutoMigrate(&models.Coupon{})
	Db.AutoMigrate(&models.Product_Offer{})
	Db.AutoMigrate(&models.Wallet{})
	Db.AutoMigrate(&models.Transaction{})
	Db.AutoMigrate(&models.Wishlist_item{})
}

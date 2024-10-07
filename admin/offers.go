package admin

import (
	"fmt"
	"net/http"
	"sip_Smart/database"
	"sip_Smart/helper"
	"sip_Smart/models"
	"time"

	"github.com/gin-gonic/gin"
)

func Create_Coupon(c *gin.Context) {

	var coupon models.Coupon
	if err := c.ShouldBindJSON(&coupon); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if time.Now().After(coupon.ExpiryDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Expiry date must be in the future"})
		return
	}

	err := helper.Validate(coupon)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var count int64
	if err := database.Db.Model(&models.Coupon{}).Where("code = ?", coupon.Code).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database operation failed"})
		return
	}

	if count != 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Coupon code already exists"})
		return
	}

	if err := database.Db.Create(&coupon).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create coupon"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Coupon created successfully", "coupon": coupon})
}

func List_Coupons(c *gin.Context) {

	var coupons []models.Coupon

	if err := database.Db.Order("id").Find(&coupons).Error; err != nil {
		c.JSON(500, gin.H{"error": "database operation failed"})
		return
	}
	c.JSON(200, gin.H{
		"coupons": coupons,
	})
}

func Delete_Coupon(c *gin.Context) {
	coupon_Id := c.Param("id")

	var coupon models.Coupon
	if err := database.Db.Where("id = ?", coupon_Id).First(&coupon).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coupon not found"})
		return
	}

	if err := database.Db.Delete(&coupon).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete coupon"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Coupon deleted successfully"})
}

func Create_Productoffer(c *gin.Context) {
	var offer models.Product_Offer
	if err := c.ShouldBindJSON(&offer); err != nil {
		fmt.Println(offer)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var product models.Product
	if err := database.Db.Where("id = ?", offer.Product_Id).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	if time.Now().After(offer.ExpiryDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Expiry date must be in the future"})
		return
	}

	if err := database.Db.Create(&offer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create offer"})
		return
	}

	discountedPrice := float32(product.Price) * (1 - (offer.Discount / 100))
	product.Offer_Price = float64(discountedPrice)

	if err := database.Db.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product offer price"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product offer created successfully", "offer": offer, "updated_price": discountedPrice})
}

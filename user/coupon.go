package user

import (
	"fmt"
	"net/http"
	"sip_Smart/database"
	"sip_Smart/helper"
	"sip_Smart/models"
	"time"

	"github.com/gin-gonic/gin"
)

func Apply_Coupon(c *gin.Context) {

	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var request struct {
		Coupon_code string `json:"coupon_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var cart models.Cart
	if err := database.Db.Where("user_id = ?", user_id).Preload("Items").First(&cart).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}

	if cart.Coupon_code != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A coupon has already been applied to this cart"})
		return
	}

	var coupon models.Coupon
	if err := database.Db.Where("code = ?", request.Coupon_code).First(&coupon).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coupon not found"})
		return
	}

	if !coupon.Status || time.Now().After(coupon.ExpiryDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Coupon is invalid or expired"})
		return
	}

	if cart.Total_Price < float32(coupon.Min_Purchase) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("This coupon requires a minimum purchase amount of %d", coupon.Min_Purchase),
		})
		cart.Final_Price = float32(cart.Total_Price)
		return
	}

	discountAmount := (coupon.Discount / 100) * (cart.Total_Price)
	if discountAmount > (float32(coupon.Max_Discount)) {
		discountAmount = (float32(coupon.Max_Discount))
	}

	cart.Final_Price = float32(cart.Total_Price) - float32(discountAmount)
	cart.Coupon_code = coupon.Code

	for _, cart_item := range cart.Items {
		cart_item.Offer_Price = cart_item.Price - ((cart_item.Price / cart.Total_Price) * discountAmount)
		if err := database.Db.Save(&cart_item).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart"})
			return
		}
	}

	if err := database.Db.Save(&cart).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply coupon to cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Coupon applied successfully",
		"discount_amount": discountAmount,
		"final_price":     cart.Final_Price,
	})
}

func Remove_Coupon(c *gin.Context) {
	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var cart models.Cart
	if err := database.Db.Where("user_id = ?", user_id).Preload("Items").First(&cart).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}

	if cart.Coupon_code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No coupon applied to this cart"})
		return
	}

	var coupon models.Coupon
	if err := database.Db.Where("code = ?", cart.Coupon_code).First(&coupon).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coupon not found"})
		return
	}

	cart.Coupon_code = ""
	cart.Final_Price = cart.Total_Price

	for _, cart_item := range cart.Items {
		cart_item.Offer_Price = cart_item.Price
		if err := database.Db.Save(&cart_item).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart"})
			return
		}
	}

	if err := database.Db.Save(&cart).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Coupon removed successfully",
		"updated_price": cart.Final_Price,
	})
}

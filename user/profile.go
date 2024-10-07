package user

import (
	"net/http"
	"sip_Smart/database"
	"sip_Smart/helper"
	"sip_Smart/models"
	"sip_Smart/responsemodels"

	"github.com/gin-gonic/gin"
)

func User_Profile(c *gin.Context) {

	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var user models.User
	if err := database.Db.Where("id = ?", user_id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":            user.ID,
			"name":          user.Name,
			"phone":         user.Phone,
			"email":         user.Email,
			"status":        user.Status,
			"signup_method": user.Signup_Method,
		},
	})
}

func Edit_Profile(c *gin.Context) {

	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var user_info models.Edit_profile

	if err := c.ShouldBindJSON(&user_info); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	err := helper.Validate(user_info)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.Db.Where("id = ?", user_id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var count int64

	if err := database.Db.Model(&models.User{}).Where("phone = ?", user.Phone).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database operation failed"})
	}

	if count != 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Record exists with the given phone number"})
		return
	}

	var count1 int64

	if err := database.Db.Model(&models.User{}).Where("email = ?", user.Email).Count(&count1).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database operation failed"})
	}

	if count1 != 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Record exists with the given email"})
		return
	}

	user.Name = user_info.Name
	user.Phone = user_info.Phone
	user.Email = user_info.Email
	user.Password = user_info.Password

	if err := database.Db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user": gin.H{
			"id":            user.ID,
			"name":          user.Name,
			"phone":         user.Phone,
			"email":         user.Email,
			"status":        user.Status,
			"signup_method": user.Signup_Method,
		},
	})
}

func Get_Wallet(c *gin.Context) {
	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var wallet models.Wallet
	if err := database.Db.Preload("Transactions").Where("user_id = ?", user_id).First(&wallet).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		return
	}

	var wallet_response responsemodels.Wallet
	wallet_response.Balance = wallet.Balance
	wallet_response.User_Id = wallet.User_Id

	for _, trans := range wallet.Transactions {
		transaction := responsemodels.Transaction{
			Amount: trans.Amount,
			Type:   trans.Type,
			Time:   trans.UpdatedAt,
		}

		wallet_response.Transactions = append(wallet_response.Transactions, transaction)
	}

	c.JSON(http.StatusOK, gin.H{
		"wallet": wallet_response,
	})
}

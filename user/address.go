package user

import (
	"fmt"
	"net/http"

	"sip_Smart/database"
	"sip_Smart/helper"
	"sip_Smart/models"
	"sip_Smart/responsemodels"

	"github.com/gin-gonic/gin"
)

func List_address(c *gin.Context) {

	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found from claims"})
		return
	}

	var addresses []models.Address

	if err := database.Db.Where("user_id = ?", user_id).Find(&addresses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database operation failed"})
		return
	}

	if len(addresses) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No addresses found for the user"})
		return
	}

	var response []responsemodels.Address
	for _, address := range addresses {
		response = append(response, responsemodels.Address{
			Id:          address.ID,
			Country:     address.Country,
			State:       address.State,
			District:    address.District,
			StreetName:  address.Street,
			Pincode:     address.Pincode,
			PhoneNumber: address.Phone,
			Default:     address.Default,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"addresses": response,
	})
}

func Add_address(c *gin.Context) {

	var address_info models.Address_info

	if err := c.ShouldBindJSON(&address_info); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := helper.Validate(address_info)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		fmt.Println("user id not found from claims")
	}

	address := models.Address{
		User_Id:  user_id,
		Country:  address_info.Country,
		State:    address_info.State,
		District: address_info.District,
		Pincode:  address_info.Pincode,
		Phone:    address_info.PhoneNumber,
		Street:   address_info.StreetName,
	}

	var count1 int64
	database.Db.Model(&models.Address{}).Where("user_id = ?", address.User_Id).Count(&count1)

	if count1 == 0 {
		address.Default = true
	} else {
		address.Default = false
	}

	if err := database.Db.Create(&address).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add address"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Address created successfully",
		"address": gin.H{
			"id":           address.ID,
			"user_id":      address.User_Id,
			"country":      address.Country,
			"state":        address.State,
			"district":     address.District,
			"street_name":  address.Street,
			"pincode":      address.Pincode,
			"phone_number": address.Phone,
			"default":      address.Default,
		},
	})
}

func Edit_address(c *gin.Context) {

	var address models.Address

	id := c.Param("id")

	if err := database.Db.First(&address, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "address not found"})
		return
	}
	if err := c.ShouldBindJSON(&address); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	err := helper.Validate(address)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		fmt.Println("user id not found from claims")
	}

	var count2 int64
	database.Db.Model(&models.Address{}).Where("user_id = ?", user_id).Count(&count2)

	if count2 == 0 {
		c.JSON(400, gin.H{"error": "No address found for this user"})
		return
	}

	if err := database.Db.Save(&address).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database operation failed",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "address updated successfully",
	})
}

func Delete_address(c *gin.Context) {

	id := c.Param("id")
	var address models.Address

	if err := database.Db.First(&address, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "address not found"})
		return
	}

	if err := database.Db.Delete(&address).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete address",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "address deleted successfully",
	})
}

package admin

import (
	"net/http"
	"sip_Smart/database"
	"sip_Smart/models"

	"github.com/gin-gonic/gin"
)

func List_Users(c *gin.Context) {

	var users []models.User

	result := database.Db.Order("id").Find(&users)
	if result.Error != nil {
		c.JSON(500, gin.H{
			"error": "Failed to retrieve users",
		})
		return
	}

	c.JSON(200, gin.H{
		"users": users,
	})
}

func Block_User(c *gin.Context) {

	ID := c.Param("id")
	var user models.User

	if err := database.Db.First(&user, ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Status = "Blocked"
	database.Db.Save(&user)

	c.JSON(http.StatusOK, gin.H{
		"message": "User blocked successfully",
		"user":    user,
	})
}

func Unblock_User(c *gin.Context) {

	userID := c.Param("id")
	var user models.User

	if err := database.Db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Status = "Active"
	database.Db.Save(&user)

	c.JSON(http.StatusOK, gin.H{
		"message": "User unblocked successfully",
		"user":    user,
	})
}

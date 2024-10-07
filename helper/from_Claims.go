package helper

import (
	"net/http"
	"sip_Smart/database"
	"sip_Smart/models"

	"github.com/gin-gonic/gin"
)

func Get_Userid(c *gin.Context) int {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Claims not found"})
		return 0
	}

	custom_Claims, ok := claims.(*Jwt_Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
		return 0
	}

	var user models.User

	if err := database.Db.Model(&models.User{}).Where("email = ?", custom_Claims.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database operation failed"})
		return 0
	}
	return int(user.ID)
}

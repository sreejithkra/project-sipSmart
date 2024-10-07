package admin

import (
	"net/http"
	"sip_Smart/database"
	"sip_Smart/helper"
	"sip_Smart/models"

	"github.com/gin-gonic/gin"
)

func Admin_Login(c *gin.Context) {
	var admin models.Admin_Credentials

	if err := c.ShouldBindJSON(&admin); err != nil {
		c.JSON(400, gin.H{"error": "Invalid json format"})
		return
	}

	err := helper.Validate(admin)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var data models.Admin_Credentials

	if err := database.Db.Where("email = ?", admin.Email).First(&data).Error; err != nil {
		c.JSON(404, gin.H{"error": "Invalid username"})
		return
	}
	if admin.Password != data.Password {
		c.JSON(404, gin.H{"error": "Incorrect password"})
		return
	}

	role := "admin"

	token, err := helper.Generate_Token(admin.Email, role)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(200, gin.H{
		"message": "Admin login success!",
		"token":   token,
	})

}

func Adminauth_Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		token_String := c.GetHeader("Authorization")
		if token_String == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		claims, err := helper.ValidateToken(token_String)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		if claims.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to access this resource"})
			c.Abort()
			return
		}

		c.Next()
	}
}

package user

import (
	"net/http"
	"sip_Smart/database"
	"sip_Smart/helper"
	"sip_Smart/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type Custom_Claims struct {
	User_Email string `json:"user_email"`
	Role       string `json:"role"`
	jwt.StandardClaims
}

func Signup_User(c *gin.Context) {
	var user models.Temp_User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid json format"})
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

	if err := database.Db.Model(&models.User{}).Where("email = ?", user.Email).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database operation failed"})
		return
	}

	if count1 != 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Record exists with the given email"})
		return
	}

	err := helper.Validate(user)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	otp, err := helper.Generate_Otp()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP "})
	}

	helper.Send_Otp(user.Email, otp)

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP sent successfully",
		"otp":     otp,
	})

	helper.StoreOtp(user.Email, otp)

	if err := database.Db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database operation failed"})
		return
	}
}

func User_Login(c *gin.Context) {
	var loginuser models.User_Login
	var user models.User

	if err := c.ShouldBindJSON(&loginuser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := database.Db.Where("email = ?", loginuser.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if user.Status == "Blocked" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user blocked by admin"})
		return
	}

	if loginuser.Password != user.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect password"})
		return
	}

	role := "user"
	tokenString, err := helper.Generate_Token(loginuser.Email, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user login success",
		"token":   tokenString,
	})
}

func Userauth_Middleware() gin.HandlerFunc {
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

		if claims.Role != "user" {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to access this resource"})
			c.Abort()
			return
		}

		c.Set("claims", claims)

		c.Next()
	}
}

package helper

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/smtp"
	"sip_Smart/database"
	"sip_Smart/models"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var OTPStore = struct {
	sync.RWMutex
	store map[string]models.Otp
}{
	store: make(map[string]models.Otp),
}

func Generate_Otp() (string, error) {
	const otpLength = 6
	const digits = "0123456789"

	otp := make([]byte, otpLength)
	for i := range otp {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		otp[i] = digits[num.Int64()]
	}

	return string(otp), nil
}

func Send_Otp(email string, otp string) {

	from := "sreejith9961782021@gmail.com"
	to := []string{email}
	subject := "Subject: Your OTP Code\n"
	body := "Your OTP is: " + otp
	msg := subject + "\n" + body

	auth := smtp.PlainAuth(
		"",
		from,
		"yetr ufvr colf xqky",
		"smtp.gmail.com",
	)

	err := smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		from,
		to,
		[]byte(msg),
	)
	if err != nil {
		log.Fatalf("Failed to send OTP email: %v", err)
	}
	fmt.Println("OTP sent successfully to", to[0])

}

func StoreOtp(email string, otp string) {
	OTPStore.Lock()
	defer OTPStore.Unlock()

	OTPStore.store[email] = models.Otp{
		Otp:       otp,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	go func() {
		time.Sleep(5 * time.Minute)
		OTPStore.Lock()
		delete(OTPStore.store, email)
		OTPStore.Unlock()
	}()
}

var tempuser_Email string

func Validate_otp(c *gin.Context) {

	var input models.User_Otp

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
		return
	}

	OTPStore.RLock()
	storedOtp, exists := OTPStore.store[input.Email]
	OTPStore.RUnlock()

	var temp models.Temp_User

	if err := database.Db.Where("email = ?", input.Email).First(&temp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database operation failed"})
		}
		return
	}

	tempuser_Email = temp.Email

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No OTP found for this email"})
		return
	}

	if time.Now().After(storedOtp.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP has expired"})
		return
	}

	if input.Otp != storedOtp.Otp {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP verified successfully"})

	var newuser models.User

	newuser.Name = temp.Name
	newuser.Email = temp.Email
	newuser.Phone = temp.Phone
	newuser.Password = temp.Password

	if err := database.Db.Create(&newuser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database operation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Signup successful",
	})
}

func DeleteTempUserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Next()

		if err := database.Db.Where("email = ?", tempuser_Email).Delete(&models.Temp_User{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete temporary user"})
			return
		}
	}
}

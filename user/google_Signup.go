package user

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sip_Smart/database"
	"sip_Smart/helper"
	"sip_Smart/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "https://sreejith.shop/auth/google/callback",
		ClientID:     "330301856979-36gtuensp83bpr92e9nc1vp61c4mp02v.apps.googleusercontent.com",
		ClientSecret: "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",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
	oauthStateString = "randomstate"
)

func GoogleLogin(c *gin.Context) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func GoogleCallback(c *gin.Context) {

	state := c.Query("state")
	code := c.Query("code")

	if state != oauthStateString {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid state"})
		return
	}

	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	client := googleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	userinfo, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"msg": err.Error()})
	}

	var new models.Google_Response
	err = json.Unmarshal(userinfo, &new)
	if err != nil {
		fmt.Println("", err)
	}

	var count int64
	role := "user"
	database.Db.Model(&models.User{}).Where("email = ?", new.Email).Count(&count)
	if count != 0 {
		tokenString, err := helper.Generate_Token(new.Email, role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "user login with google success",
			"token":   tokenString,
		})
		return
	}

	var user models.User

	user.Email = new.Email
	user.Name = new.Name
	user.Signup_Method = "google"

	if err := database.Db.Create(&user).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database operation failed"})
		return
	}

	tokenString, err := helper.Generate_Token(user.Email, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user signup with google success",
		"token":   tokenString,
	})
}

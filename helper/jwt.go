package helper

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

var secretKey []byte = []byte("your_secret_key")

type Jwt_Claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.StandardClaims
}

func Generate_Token(email string, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Jwt_Claims{
		Email: email,
		Role:  role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(tokenString string) (*Jwt_Claims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &Jwt_Claims{}, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Jwt_Claims); ok && token.Valid {

		return claims, nil
	}
	return nil, errors.New("invalid token")
}

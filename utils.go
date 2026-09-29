package main

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("rahasia_lms_super_aman_123")

// Fungsi buat acak pw
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// Fungsi buat ngecek kecocokan pw asli atau pw acak
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Fungsi buat "tiket masuk" JWT
func GenerateJWT(userID string, email string, roleID uint) (string, error) {
	claims := jwt.MapClaims{
		"id":      userID,
		"email":   email,
		"role_id": roleID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
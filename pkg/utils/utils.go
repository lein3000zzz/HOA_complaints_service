package utils

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func GenerateID() (string, error) {
	bytes := make([]byte, 20)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func IsNumbers(s string) bool {
	for i := range len(s) {
		if !(s[i] >= '0' && s[i] <= '9') {
			return false
		}
	}
	return true
}

func IsNumbersAndLetters(s string) bool {
	for i := range len(s) {
		if !((s[i] >= 'a' && s[i] <= 'z') ||
			(s[i] >= 'A' && s[i] <= 'Z') ||
			(s[i] >= '0' && s[i] <= '9')) {
			return false
		}
	}
	return true
}

func GetPageAndLimitFromContext(c *gin.Context) (int, int) {
	page := 1
	limit := 10

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 1000 {
			limit = v
		}
	}

	return page, limit
}

func CountPages(total, limit int) int {
	pages := 1
	if total > 0 {
		pages = total / limit
		if total%limit != 0 {
			pages++
		}
	}
	return pages
}

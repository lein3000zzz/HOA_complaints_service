package utils

import (
	"crypto/rand"
	"encoding/hex"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("whoever_reads_it_is_gay")

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

//func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
//	w.Header().Set("Content-Type", "application/json")
//	w.WriteHeader(status)
//	out, err := json.Marshal(data)
//	if err != nil {
//		return
//	}
//	_, err = w.Write(out)
//	if err != nil {
//		return
//	}
//}

//func SendJwtToken(w http.ResponseWriter, token *jwt.Token) error {
//	tokenString, err := token.SignedString(jwtSecret)
//	if err != nil {
//		WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
//		return err
//	}
//
//	fmt.Println("SendJwtToken token:", tokenString)
//
//	WriteJSON(w, http.StatusOK, map[string]interface{}{
//		"token": tokenString,
//	})
//	return nil
//}

//func IsNumbersAndLetters(s string) bool {
//	for i := range len(s) {
//		if !((s[i] >= 'a' && s[i] <= 'z') ||
//			(s[i] >= 'A' && s[i] <= 'Z') ||
//			(s[i] >= '0' && s[i] <= '9')) {
//			return false
//		}
//	}
//	return true
//}

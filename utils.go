package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

const (
  ContextKeyUser ContextKey = "user"
)

type ContextKey string

type UserJSON struct {
	Id             int    `json:"id"`
	Username       string `json:"username"`
	PasswordHashed string `json:"passwordHashed"`
	ChatIds        []int  `json:"chatIds"`
}

func (u *UserJSON) ValidatePassword(pass string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHashed), []byte(pass)) == nil
}

type ChatJSON struct {
	Id        int           `json:"id"`
	Usernames []string      `json:"usernames"`
	Messages  []MessageJSON `json:"messages"`
}

type MessageJSON struct {
	Id         int    `json:"id"`
	ChatId     int    `json:"chatId"`
	Text       string `json:"text"`
	AuthorName string `json:"authorName"`
	Timestamp  int    `json:"timestamp"`
}

func CreateCookie(name string, value string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     Path,
		Domain:   Domain,
		Expires:  time.Now().Add(CookieLifeTime),
		MaxAge:   int(CookieLifeTime),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
	}
}

func CreateJWT(userId int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": strconv.Itoa(userId),
	})

	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString(JwtSecret)
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(JwtSecret), nil
	})
}

func WriteJSON(w http.ResponseWriter, v any, status int) error {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

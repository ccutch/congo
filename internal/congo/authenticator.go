package congo

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Authenticator struct {
}

func (auth *Authenticator) Secure(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.Authenticate(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		fn(w, r)
	}
}

func (auth *Authenticator) Authenticate(r *http.Request) bool {
	cookie, err := r.Cookie("academy-user")
	if err != nil {
		return false
	}
	if _, err = auth.ParseToken(cookie.Value); err != nil {
		return false
	}
	return true
}

func (auth *Authenticator) Token(subject string) string {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": subject,
		"iat": now.Unix(),
		"exp": now.Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		log.Printf("Failed to sign token: %v", err)
		return ""
	}
	return signedToken
}

func (auth *Authenticator) ParseToken(tokenString string) (string, error) {
	if tokenString == "" {
		return "", errors.New("empty token string")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET")), nil
	})
	if err != nil {
		return "", err
	} else if !token.Valid {
		return "", errors.New("invlaid token")
	}
	claims, valid := token.Claims.(jwt.MapClaims)
	if !valid {
		return "", errors.New("invalid token")
	}
	return claims["sub"].(string), nil
}

func (auth *Authenticator) StartSession(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "academy-user",
		Path:     "/",
		Value:    value,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})
}

func (auth *Authenticator) StopSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "academy-user",
		Path:     "/",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour), // Set the expiration to the past to delete the cookie
		HttpOnly: true,
	})
}

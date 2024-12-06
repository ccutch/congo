package congo_auth

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

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

func (auth *Authenticator) ParseToken(tokenString string) (*User, error) {
	if tokenString == "" {
		return nil, errors.New("empty token string")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET")), nil
	})
	if err != nil {
		return nil, err
	} else if !token.Valid {
		return nil, errors.New("invlaid token")
	}
	claims, valid := token.Claims.(jwt.MapClaims)
	if !valid {
		return nil, errors.New("invalid token")
	}
	userID, valid := claims["sub"].(string)
	if !valid {
		return nil, errors.New("invalid token")
	}
	return auth.GetUser(userID)
}

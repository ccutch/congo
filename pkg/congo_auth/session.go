package congo_auth

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Session struct {
	congo.Model
	IdentID string
}

func (i *Identity) NewSession() (*Session, error) {
	s := &Session{i.DB.NewModel(uuid.NewString()), i.ID}
	return s, s.DB.Query(`
		INSERT INTO sessions (id, identity_id)
		VALUES (?, ?)
		RETURNING created_at, updated_at
	`, s.ID, s.IdentID).Scan(&s.CreatedAt, &s.UpdatedAt)
}

func (i *Identity) Sessions() ([]*Session, error) {
	sessions := []*Session{}
	return sessions, i.DB.Query(`
		SELECT id, identity_id, created_at, updated_at
		FROM sessions
		WHERE identity_id = ?
		ORDER BY created_at DESC
	`, i.ID).All(func(scan congo.Scanner) error {
		s := Session{Model: congo.Model{DB: i.DB}}
		sessions = append(sessions, &s)
		return scan(&s.ID, &s.IdentID, &s.CreatedAt, &s.UpdatedAt)
	})
}

func (auth *CongoAuth) GetSession(id string) (*Session, error) {
	s := &Session{Model: congo.Model{DB: auth.DB}}
	return s, s.DB.Query(`
		SELECT id, identity_id, created_at, updated_at
		FROM sessions
		WHERE id = ?
	`, id).Scan(&s.ID, &s.IdentID, &s.CreatedAt, &s.UpdatedAt)
}

func (s *Session) Token() string {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": s.ID,
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

func (s *Session) Delete() error {
	return s.DB.Query(`
		DELETE FROM sessions
		WHERE id = ?
	`, s.ID).Exec()
}

func (auth *CongoAuth) Authenticate(r *http.Request, roles ...string) (*Identity, *Session) {
	for _, role := range roles {
		cookie, err := r.Cookie(auth.CookieName + "-" + role)
		if err != nil {
			continue
		}

		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("SECRET")), nil
		})

		if err != nil || !token.Valid {
			continue
		}

		claims, valid := token.Claims.(jwt.MapClaims)
		if !valid {
			continue
		}

		sessionID, valid := claims["sub"].(string)
		if !valid {
			continue
		}

		session, err := auth.GetSession(sessionID)
		if err != nil {
			log.Printf("Failed to lookup session %s: %s", sessionID, err)
			continue
		}

		identity, err := auth.Lookup(session.IdentID)
		if err != nil {
			log.Printf("Failed to lookup identity %s: %s", session.IdentID, err)
			continue
		}

		return identity, session
	}
	return nil, nil
}

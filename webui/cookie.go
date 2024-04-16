package webui

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
)

var ErrCookieNotFound = errors.New("cookie not found")
var ErrCookieExpired = errors.New("cookie has expired")

const ExpirationHours = 30 * 24 // Cookie 有效期改为一个月

func GenerateCookie(db *sql.DB) (string, error) {
	cookie := uuid.New().String()
	expiration := time.Now().Add(ExpirationHours * time.Hour).Unix()

	_, err := db.Exec("INSERT INTO cookies (cookie_id, expiration) VALUES (?, ?)", cookie, expiration)
	if err != nil {
		log.Fatalf("Failed to insert new cookie: %v", err)
		return "", err
	}
	return cookie, nil
}

func ValidateCookie(db *sql.DB, cookie string) (bool, error) {
	var expiration int64
	err := db.QueryRow("SELECT expiration FROM cookies WHERE cookie_id = ?", cookie).Scan(&expiration)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, ErrCookieNotFound
		}
		log.Fatalf("Failed to query cookie: %v", err)
		return false, err
	}
	if time.Now().Unix() > expiration {
		return false, ErrCookieExpired
	}
	return true, nil
}

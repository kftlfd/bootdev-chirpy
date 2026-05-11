package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func CheckPasswordHash(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

func extractToken(authHeader string) (string, error) {
	fields := strings.Split(authHeader, " ")

	if len(fields) != 2 || fields[0] != "Bearer" {
		return "", errors.New("invalid format")
	}

	return fields[1], nil
}

func GetBearerToken(header http.Header) (string, error) {
	t := header.Get("Authorization")
	return extractToken(t)
}

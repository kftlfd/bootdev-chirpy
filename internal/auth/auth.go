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

func getAuthToken(header http.Header, format string) (string, error) {
	t := header.Get("Authorization")

	fields := strings.Fields(t)

	if len(fields) != 2 || fields[0] != format {
		return "", errors.New("invalid format")
	}

	return fields[1], nil
}

func GetBearerToken(header http.Header) (string, error) {
	return getAuthToken(header, "Bearer")
}

func GetAPIKey(header http.Header) (string, error) {
	return getAuthToken(header, "ApiKey")
}

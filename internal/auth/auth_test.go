package auth_test

import (
	"chirpy/internal/auth"
	"testing"
)

func TestMakePassword(t *testing.T) {
	password := "password"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Error(err)
		return
	}

	if len(hash) < 1 {
		t.Error("empty hash")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "password"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Error(err)
		return
	}

	ok, err := auth.CheckPasswordHash(password, hash)
	if err != nil {
		t.Error(err)
		return
	}

	if !ok {
		t.Error("check")
	}
}

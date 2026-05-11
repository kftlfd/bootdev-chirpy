package auth_test

import (
	"chirpy/internal/auth"
	"testing"
	"time"

	"github.com/google/uuid"
)

var tokenSecret = "token-secret"

func TestMakeJwt(t *testing.T) {
	id := uuid.New()

	token, err := auth.MakeJWT(id, tokenSecret, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	if len(token) < 1 {
		t.Fatal("empty token")
	}
}

func TestValidateJWT(t *testing.T) {
	id := uuid.New()

	token, err := auth.MakeJWT(id, tokenSecret, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	userId, err := auth.ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatal(err)
	}

	if id != userId {
		t.Fatal("ids don't match")
	}
}

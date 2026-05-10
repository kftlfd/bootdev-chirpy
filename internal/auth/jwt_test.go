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
		t.Error(err)
		return
	}

	if len(token) < 1 {
		t.Error("empty token")
	}
}

func TestValidateJWT(t *testing.T) {
	id := uuid.New()

	token, err := auth.MakeJWT(id, tokenSecret, time.Minute)
	if err != nil {
		t.Error(err)
		return
	}

	userId, err := auth.ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Error(err)
		return
	}

	if id != userId {
		t.Error("ids don't match")
	}
}

package auth_test

import (
	"chirpy/internal/auth"
	"net/http"
	"testing"
)

func TestMakePassword(t *testing.T) {
	password := "password"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	if len(hash) < 1 {
		t.Fatal("empty hash")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "password"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := auth.CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("check")
	}
}

func TestExtractToken(t *testing.T) {

	testCases := []struct {
		name  string
		in    string
		out   string
		isErr bool
	}{
		{"ok", "Bearer hello", "hello", false},
		{"bad", "Bearer hey hey", "", true},
		{"empty", "", "", true},
		{"invalid", "Token hello", "", true},
		{"invalid2", "BEARER hello", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := http.Header{}
			h.Add("Authorization", tc.in)

			bearer, err := auth.GetBearerToken(h)

			if err != nil {
				if !tc.isErr {
					t.Fatal(err)
				}
				return
			}
			if tc.isErr {
				t.Fatal("expected error")
			}

			if bearer != tc.out {
				t.Fatalf("expected: %s, got: %s", tc.out, bearer)
			}
		})
	}
}

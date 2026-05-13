package e2e_test

import (
	"chirpy/tests/testutils"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHealth(t *testing.T) {
	app := testutils.NewTestApp(t)

	srv := httptest.NewServer(app)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/healthz")

	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
}

func TestCreateUser(t *testing.T) {
	app := testutils.NewTestApp(t)

	testCases := []struct {
		name   string
		body   string
		status int
	}{
		{"no email", `{"email2": ""}`, 400},
		{"no pass", `{"passw": ""}`, 400},
		{"ok", `{"email": "hey1", "password": "ho"}`, 201},
		{"duplicate email", `{"email": "hey1", "password": "ho"}`, 400},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/users",
				strings.NewReader(tc.body),
			)

			app.ServeHTTP(rr, req)

			require.Equal(t, tc.status, rr.Code)
		})
	}
}

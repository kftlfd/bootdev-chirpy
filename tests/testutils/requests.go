package testutils

import (
	"net/http"
)

func setAuthToken(h http.Header, format, token string) {
	h.Set("Authorization", format+" "+token)
}

func SetAuthBearer(h http.Header, token string) {
	setAuthToken(h, "Bearer", token)
}

func SetAuthApiKey(h http.Header, token string) {
	setAuthToken(h, "ApiKey", token)
}

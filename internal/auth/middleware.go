package auth

import (
	"chirpy/internal/config"
	u "chirpy/internal/utils"
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const authUserKey contextKey = "auth_user"

func WithAuth(cfg *config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := GetBearerToken(r.Header)
		if err != nil {
			u.SendJSON(w, http.StatusUnauthorized, u.D{
				"error": fmt.Sprintf("No auth token: %s", err),
			})
			return
		}

		userId, err := ValidateJWT(token, cfg.Env.ServerSecret)
		if err != nil {
			u.SendJSON(w, http.StatusUnauthorized, u.D{
				"error": fmt.Sprintf("Invalid token: %s", err),
			})
			return
		}

		ctx := context.WithValue(
			r.Context(),
			authUserKey,
			userId,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireAuthUser(ctx context.Context) uuid.UUID {
	userId, ok := ctx.Value(authUserKey).(uuid.UUID)

	if !ok {
		panic("auth missing from context")
	}

	return userId
}

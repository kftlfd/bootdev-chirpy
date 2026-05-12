package auth

import (
	"chirpy/internal/config"
	"chirpy/internal/httpx"
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const authUserKey contextKey = "auth_user"

func WithAuth(cfg *config.Config, logger *slog.Logger, next http.Handler) http.Handler {
	log := logger.With("module", "middleware-with-auth")
	res := httpx.NewResponder(log)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := GetBearerToken(r.Header)
		if err != nil {
			res.JSON(w, http.StatusUnauthorized, httpx.D{
				"error": fmt.Sprintf("No auth token: %s", err),
			})
			return
		}

		userId, err := ValidateJWT(token, cfg.Env.ServerSecret)
		if err != nil {
			res.JSON(w, http.StatusUnauthorized, httpx.D{
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

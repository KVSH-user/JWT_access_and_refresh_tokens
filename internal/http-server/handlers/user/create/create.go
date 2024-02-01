package create

import (
	"auth/internal/entity"
	resp "auth/internal/lib/api/response"
	"auth/internal/token/access"
	"auth/internal/token/refresh"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Request struct {
	GUID string `json:"guid" validate:"required"`
}

type Response struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenSaver interface {
	SaveToken(user *entity.User) error
	ValidateToken(refreshToken string) (string, error)
}

func New(log *slog.Logger, tokenSaver TokenSaver, secret string, ttl time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.user.create.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			render.JSON(w, r, resp.Error("empty request"))

			return
		}
		if err != nil {
			log.Error("failed to decode request body: ", err)

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		accessToken, err := access.Generate(secret, req.GUID, ttl)

		refreshToken, err := refresh.Generate()

		user := &entity.User{
			Guid:         req.GUID,
			RefreshToken: refreshToken,
		}

		err = tokenSaver.SaveToken(user)
		if err != nil {
			log.Error("failed to save user: ", err)

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("user successfully added")

		render.JSON(w, r, Response{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		})
	}
}

func Refresh(log *slog.Logger, tokenSaver TokenSaver, secret string, ttl time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.user.create.Refresh"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Error("Authorization header is missing")
			render.JSON(w, r, resp.Error("Authorization header is missing"))
			return
		}

		splitToken := strings.Split(authHeader, "Bearer ")
		if len(splitToken) != 2 {
			log.Error("Invalid Authorization token format")
			render.JSON(w, r, resp.Error("Invalid Authorization token format"))
			return
		}
		refreshToken := splitToken[1]

		guid, err := tokenSaver.ValidateToken(refreshToken)
		if err != nil {
			log.Error("invalid refresh token", err)
			render.JSON(w, r, resp.Error("invalid refresh token"))
			return
		}

		newAccessToken, err := access.Generate(secret, guid, ttl)
		if err != nil {
			log.Error("failed to generate access token", err)
			render.JSON(w, r, resp.Error("failed to generate access token"))
			return
		}

		newRefreshToken, err := refresh.Generate()
		if err != nil {
			log.Error("failed to generate refresh token", err)
			render.JSON(w, r, resp.Error("failed to generate refresh token"))
			return
		}

		user := &entity.User{
			Guid:         guid,
			RefreshToken: newRefreshToken,
		}
		if err := tokenSaver.SaveToken(user); err != nil {
			log.Error("failed to save new refresh token", err)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		render.JSON(w, r, Response{
			AccessToken:  newAccessToken,
			RefreshToken: newRefreshToken,
		})
	}
}

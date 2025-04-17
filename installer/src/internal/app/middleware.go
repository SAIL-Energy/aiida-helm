package app

import (
	"context"
	"github.com/eddie-energy/aiida-helm/installer/internal/dto"
	"log"
	"net/http"
	"strings"
)

func (app *App) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		log.Printf("Authenticating request %s %s", request.Method, request.URL.Path)

		authHeader := request.Header.Get("Authorization")
		if authHeader == "" {
			app.respondWithJSON(responseWriter, http.StatusUnauthorized, dto.NewErrorResponseDto("missing authorization header"))
			return
		}

		if !strings.HasPrefix(authHeader, bearerPrefix) {
			app.respondWithJSON(responseWriter, http.StatusUnauthorized, dto.NewErrorResponseDto("invalid authorization header"))
			return
		}

		jwtTokenString := strings.TrimPrefix(authHeader, bearerPrefix)
		if jwtTokenString == "" {
			app.respondWithJSON(responseWriter, http.StatusUnauthorized, dto.NewErrorResponseDto("missing jwt token"))
			return
		}

		userId, err := app.verifyJwtToken(jwtTokenString)
		if err != nil {
			app.respondWithJSON(responseWriter, http.StatusUnauthorized, dto.NewErrorResponseDto(err.Error()))
			return
		}

		ctx := context.WithValue(request.Context(), userIdKey, *userId)
		next.ServeHTTP(responseWriter, request.WithContext(ctx))
	})
}

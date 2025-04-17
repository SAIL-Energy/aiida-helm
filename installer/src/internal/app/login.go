package app

import (
	"encoding/json"
	"github.com/eddie-energy/aiida-helm/installer/internal/dto"
	"io"
	"log"
	"net/http"
)

func (app *App) login(responseWriter http.ResponseWriter, request *http.Request) {
	log.Println("Login")

	body, err := io.ReadAll(request.Body)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(request.Body)

	var loginDto dto.LoginDto
	if err := json.Unmarshal(body, &loginDto); err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	err = app.verifyInstallerToken(loginDto.InstallerToken)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	jwtToken, err := app.createJwtToken(loginDto.UserId)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessDataResponseDto(jwtToken))
}

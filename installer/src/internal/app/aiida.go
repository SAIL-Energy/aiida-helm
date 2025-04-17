package app

import (
	"github.com/eddie-energy/aiida-helm/installer/internal/dto"
	"log"
	"net/http"
)

func (app *App) getAiidaVersion(responseWriter http.ResponseWriter, _ *http.Request) {
	log.Println("Get version info of AIIDA")

	versionInfo, err := app.AiidaInstaller.GetVersionInfo()
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessDataResponseDto(versionInfo))
}

func (app *App) installOrUpgradeAiida(responseWriter http.ResponseWriter, _ *http.Request) {
	log.Println("Install or upgrade AIIDA")

	versionInfo, err := app.AiidaInstaller.InstallOrUpgrade()
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessDataResponseDto(versionInfo))
}

func (app *App) uninstallAiida(responseWriter http.ResponseWriter, _ *http.Request) {
	log.Println("Uninstall AIIDA")

	if err := app.AiidaInstaller.Uninstall(); err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessResponseDto())
}

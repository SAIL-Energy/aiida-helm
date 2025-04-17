package app

import (
	"encoding/json"
	"github.com/eddie-energy/aiida-helm/installer/internal/dto"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
)

func (app *App) getServicesVersionsAll(responseWriter http.ResponseWriter, _ *http.Request) {
	log.Println("Get version info of all services")

	versionInfos, err := app.ServicesInstaller.GetVersionInfosRepo(nil, nil)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessDataResponseDto(versionInfos))
}

func (app *App) getServicesVersionsUser(responseWriter http.ResponseWriter, request *http.Request) {
	log.Println("Get version info of all services for user")

	userId, err := app.getUserIDFromContext(request)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	versionInfos, err := app.ServicesInstaller.GetVersionInfosRepo(&userId, nil)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessDataResponseDto(versionInfos))
}

func (app *App) getServicesVersionsChart(responseWriter http.ResponseWriter, request *http.Request) {
	log.Println("Get version info of services for a chart name")

	chartName := mux.Vars(request)[varChartName]
	if chartName == "" {
		app.respondWithJSON(responseWriter, http.StatusBadRequest, dto.NewErrorResponseDto("missing chart name"))
		return
	}

	userId, err := app.getUserIDFromContext(request)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	versionInfos, err := app.ServicesInstaller.GetVersionInfosRepo(&userId, &chartName)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessDataResponseDto(versionInfos))
}

func (app *App) installNewService(responseWriter http.ResponseWriter, request *http.Request) {
	log.Println("Install new service")

	chartName := mux.Vars(request)[varChartName]
	if chartName == "" {
		app.respondWithJSON(responseWriter, http.StatusBadRequest, dto.NewErrorResponseDto("missing chart name"))
		return
	}

	userId, err := app.getUserIDFromContext(request)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(request.Body)

	var setupDto dto.SetupDto
	_ = json.Unmarshal(body, &setupDto)

	versionInfo, err := app.ServicesInstaller.InstallNew(userId, chartName, &setupDto)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessDataResponseDto(versionInfo))
}

func (app *App) getVersionInfoService(responseWriter http.ResponseWriter, request *http.Request) {
	log.Println("Get version info service")

	chartName, releaseName, err := app.getChartAndReleaseName(request)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusBadRequest, dto.NewErrorResponseDto(err.Error()))
		return
	}

	userId, err := app.getUserIDFromContext(request)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	if err := app.verifyServiceReleaseName(releaseName, userId); err != nil {
		app.respondWithJSON(responseWriter, http.StatusUnauthorized, dto.NewErrorResponseDto(err.Error()))
		return
	}

	versionInfo, err := app.ServicesInstaller.GetVersionInfo(chartName, releaseName)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessDataResponseDto(versionInfo))
}

func (app *App) installOrUpgradeService(responseWriter http.ResponseWriter, request *http.Request) {
	log.Println("Install or upgrade service")

	chartName, releaseName, err := app.getChartAndReleaseName(request)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusBadRequest, dto.NewErrorResponseDto(err.Error()))
		return
	}

	userId, err := app.getUserIDFromContext(request)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	if err := app.verifyServiceReleaseName(releaseName, userId); err != nil {
		app.respondWithJSON(responseWriter, http.StatusUnauthorized, dto.NewErrorResponseDto(err.Error()))
		return
	}

	versionInfo, err := app.ServicesInstaller.InstallOrUpgrade(chartName, releaseName)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessDataResponseDto(versionInfo))
}

func (app *App) uninstallService(responseWriter http.ResponseWriter, request *http.Request) {
	log.Println("Uninstall service")

	chartName, releaseName, err := app.getChartAndReleaseName(request)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusBadRequest, dto.NewErrorResponseDto(err.Error()))
		return
	}

	userId, err := app.getUserIDFromContext(request)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	if err := app.verifyServiceReleaseName(releaseName, userId); err != nil {
		app.respondWithJSON(responseWriter, http.StatusUnauthorized, dto.NewErrorResponseDto(err.Error()))
		return
	}

	if err := app.ServicesInstaller.Uninstall(chartName, releaseName); err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessResponseDto())
}

package main

import (
	"encoding/json"
	"fmt"
	"github.com/eddie-energy/aiida-helm/installer/internal/helm/aiida"
	"github.com/eddie-energy/aiida-helm/installer/internal/helm/services"
	"github.com/gorilla/mux"
	"helm.sh/helm/v3/pkg/repo"
	"log"
	"net/http"
	"os"
)

const authScheme = "Basic"

const repositoryUrl = "https://eddie-energy.github.io/aiida-helm/"
const repositoryName = "aiida"
const repositoryDefaultTrack = "stable"

const pathAiida = "/aiida"
const pathService = "/service"
const varChartName = "chartName"

type App struct {
	Router            *mux.Router
	Repository        repo.Entry
	AiidaInstaller    *aiida.Installer
	ServicesInstaller *services.Installer
}

func (app *App) Run(addr string) {
	repositoryTrack := os.Getenv("REPOSITORY_TRACK")
	if repositoryTrack == "" {
		repositoryTrack = repositoryDefaultTrack
	}

	router := mux.NewRouter()
	router.Use(app.authMiddleware)

	app.Router = router
	app.Repository = repo.Entry{
		Name:               repositoryName,
		URL:                repositoryUrl + repositoryTrack,
		PassCredentialsAll: true,
	}
	app.AiidaInstaller = aiida.NewInstaller(app.Repository)
	app.ServicesInstaller = services.NewInstaller(app.Repository)

	app.initializeRoutes()
	log.Fatal(http.ListenAndServe(addr, app.Router))
}

func (app *App) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		authHeader := request.Header.Get("Authorization")

		if app.AiidaInstaller.Token != nil && authHeader != authScheme+" "+*app.AiidaInstaller.Token {
			app.respondWithJSON(responseWriter, http.StatusUnauthorized, NewErrorResponse("unauthorized"))
			return
		}

		next.ServeHTTP(responseWriter, request)
	})
}

func (app *App) initializeRoutes() {
	fullServicePath := fmt.Sprintf("%s/{%s:.*}", pathService, varChartName)

	app.Router.HandleFunc("/health", app.getHealthCheck).Methods("GET")

	app.Router.HandleFunc(pathAiida, app.getVersionInfoAiida).Methods("GET")
	app.Router.HandleFunc(pathAiida, app.installOrUpgradeAiida).Methods("POST")
	app.Router.HandleFunc(pathAiida, app.uninstallAiida).Methods("DELETE")

	app.Router.HandleFunc(pathService, app.getVersionInfosRepoServices).Methods("GET")
	app.Router.HandleFunc(fullServicePath, app.getVersionInfoService).Methods("GET")
	app.Router.HandleFunc(fullServicePath, app.installOrUpgradeService).Methods("POST")
	app.Router.HandleFunc(fullServicePath, app.uninstallService).Methods("DELETE")
}

func (app *App) getHealthCheck(responseWriter http.ResponseWriter, _ *http.Request) {
	app.respondWithJSON(responseWriter, http.StatusOK, NewSuccessResponse())
}

func (app *App) getVersionInfoAiida(responseWriter http.ResponseWriter, _ *http.Request) {
	versionInfo, err := app.AiidaInstaller.GetVersionInfo()
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, NewErrorResponse(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, NewSuccessDataResponse(versionInfo))
}

func (app *App) installOrUpgradeAiida(responseWriter http.ResponseWriter, _ *http.Request) {
	versionInfo, err := app.AiidaInstaller.InstallOrUpgrade()
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, NewErrorResponse(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, NewSuccessDataResponse(versionInfo))
}

func (app *App) uninstallAiida(responseWriter http.ResponseWriter, _ *http.Request) {
	if err := app.AiidaInstaller.Uninstall(); err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, NewErrorResponse(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, NewSuccessResponse())
}

func (app *App) getVersionInfosRepoServices(responseWriter http.ResponseWriter, _ *http.Request) {
	versionInfos, err := app.ServicesInstaller.GetVersionInfosRepo()
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, NewErrorResponse(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, NewSuccessDataResponse(versionInfos))
}

func (app *App) getVersionInfoService(responseWriter http.ResponseWriter, request *http.Request) {
	chartName := mux.Vars(request)[varChartName]

	versionInfo, err := app.ServicesInstaller.GetVersionInfo(chartName)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, NewErrorResponse(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, NewSuccessDataResponse(versionInfo))
}

func (app *App) installOrUpgradeService(responseWriter http.ResponseWriter, request *http.Request) {
	chartName := mux.Vars(request)[varChartName]

	versionInfo, err := app.ServicesInstaller.InstallOrUpgrade(chartName)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, NewErrorResponse(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, NewSuccessDataResponse(versionInfo))
}

func (app *App) uninstallService(responseWriter http.ResponseWriter, request *http.Request) {
	chartName := mux.Vars(request)[varChartName]

	if err := app.ServicesInstaller.Uninstall(chartName); err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, NewErrorResponse(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, NewSuccessResponse())
}

func (app *App) respondWithJSON(responseWriter http.ResponseWriter, code int, response Response) {
	responseJson, _ := json.Marshal(response)

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(code)

	_, _ = responseWriter.Write(responseJson)
}

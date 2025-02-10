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
)

const repositoryUrl = "https://eddie-energy.github.io/aiida-helm/stable"
const repositoryName = "aiida"

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
	app.Router = mux.NewRouter()
	app.Repository = repo.Entry{
		Name:               repositoryName,
		URL:                repositoryUrl,
		PassCredentialsAll: true,
	}
	app.AiidaInstaller = aiida.NewInstaller(app.Repository)
	app.ServicesInstaller = services.NewInstaller(app.Repository)

	app.initializeRoutes()
	log.Fatal(http.ListenAndServe(addr, app.Router))
}

func (app *App) initializeRoutes() {
	fullServicePath := fmt.Sprintf("%s/{%s:.*}", pathService, varChartName)

	app.Router.HandleFunc(pathAiida, app.getVersionInfoAiida).Methods("GET")
	app.Router.HandleFunc(pathAiida, app.installAiida).Methods("POST")
	app.Router.HandleFunc(pathAiida, app.upgradeAiida).Methods("PUT")
	app.Router.HandleFunc(pathAiida, app.uninstallAiida).Methods("DELETE")

	app.Router.HandleFunc(pathService, app.getsVersionInfosRepoServices).Methods("GET")
	app.Router.HandleFunc(fullServicePath, app.getVersionInfoService).Methods("GET")
	app.Router.HandleFunc(fullServicePath, app.installService).Methods("POST")
	app.Router.HandleFunc(fullServicePath, app.upgradeService).Methods("PUT")
	app.Router.HandleFunc(fullServicePath, app.uninstallService).Methods("DELETE")
}

func (app *App) getVersionInfoAiida(responseWriter http.ResponseWriter, _ *http.Request) {
	versionInfo, err := app.AiidaInstaller.GetVersionInfo()
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, NewSuccessDataResponse(versionInfo))
}

func (app *App) installAiida(responseWriter http.ResponseWriter, _ *http.Request) {
	versionInfo, err := app.AiidaInstaller.Install()
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, NewSuccessDataResponse(versionInfo))
}

func (app *App) upgradeAiida(responseWriter http.ResponseWriter, _ *http.Request) {
	versionInfo, err := app.AiidaInstaller.Upgrade()
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, NewSuccessDataResponse(versionInfo))
}

func (app *App) uninstallAiida(responseWriter http.ResponseWriter, _ *http.Request) {
	if err := app.AiidaInstaller.Uninstall(); err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, NewSuccessResponse())
}

func (app *App) getsVersionInfosRepoServices(responseWriter http.ResponseWriter, _ *http.Request) {
	versionInfos, err := app.ServicesInstaller.GetVersionInfosRepo()
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, NewSuccessDataResponse(versionInfos))
}

func (app *App) getVersionInfoService(responseWriter http.ResponseWriter, request *http.Request) {
	chartName := mux.Vars(request)[varChartName]

	versionInfo, err := app.ServicesInstaller.GetVersionInfo(chartName)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, NewSuccessDataResponse(versionInfo))
}

func (app *App) installService(responseWriter http.ResponseWriter, request *http.Request) {
	chartName := mux.Vars(request)[varChartName]

	versionInfo, err := app.ServicesInstaller.Install(chartName)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, NewSuccessDataResponse(versionInfo))
}

func (app *App) upgradeService(responseWriter http.ResponseWriter, request *http.Request) {
	chartName := mux.Vars(request)[varChartName]

	versionInfo, err := app.ServicesInstaller.Upgrade(chartName)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, NewSuccessDataResponse(versionInfo))
}

func (app *App) uninstallService(responseWriter http.ResponseWriter, request *http.Request) {
	chartName := mux.Vars(request)[varChartName]

	if err := app.ServicesInstaller.Uninstall(chartName); err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, NewErrorResponse(err))
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

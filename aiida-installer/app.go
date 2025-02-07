package main

import (
	"encoding/json"
	"github.com/eddie-energy/aiida-helm/aiida-installer/internal/installer"
	"github.com/gorilla/mux"
	"helm.sh/helm/v3/pkg/repo"
	"log"
	"net/http"
)

const RepositoryUrl = "https://eddie-energy.github.io/aiida-helm/stable"
const RepositoryName = "aiida"

type App struct {
	Router           *mux.Router
	Repository       repo.Entry
	AiidaInstaller   *installer.AiidaInstaller
	ServiceInstaller *installer.ServiceInstaller
}

func (app *App) Run(addr string) {
	app.Router = mux.NewRouter()
	app.Repository = repo.Entry{
		Name:               RepositoryName,
		URL:                RepositoryUrl,
		PassCredentialsAll: true,
	}
	app.AiidaInstaller = installer.NewAiidaInstaller(app.Repository)
	app.ServiceInstaller = installer.NewServiceInstaller(app.Repository)

	app.initializeRoutes()
	log.Fatal(http.ListenAndServe(addr, app.Router))
}

func (app *App) initializeRoutes() {
	app.Router.HandleFunc("/install/aiida", app.installAiida).Methods("GET")
	app.Router.HandleFunc("/install/service/{chartName:.+}", app.installService).Methods("GET")
}

func respondWithError(responseWriter http.ResponseWriter, code int, message string) {
	respondWithJSON(responseWriter, code, map[string]string{"error": message})
}

func respondWithJSON(responseWriter http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(code)

	_, err := responseWriter.Write(response)
	if err != nil {
		respondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}
}

func (app *App) installAiida(responseWriter http.ResponseWriter, _ *http.Request) {
	err := app.AiidaInstaller.Install()
	if err != nil {
		respondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(responseWriter, http.StatusOK, map[string]string{"message": "Installation successful"})
}

func (app *App) installService(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	chartName := vars["chartName"]

	err := app.ServiceInstaller.Install(chartName)
	if err != nil {
		respondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(responseWriter, http.StatusOK, map[string]string{"message": "Installation successful"})
}

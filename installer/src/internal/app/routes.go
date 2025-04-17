package app

import (
	"fmt"
	"github.com/eddie-energy/aiida-helm/installer/internal/dto"
	"github.com/gorilla/mux"
	"net/http"
)

func (app *App) initializeRoutes() {
	app.Router = mux.NewRouter()

	secureRouter := app.Router.PathPrefix("/auth").Subrouter()
	secureRouter.Use(app.authMiddleware)
	app.SecureRouter = secureRouter

	pathServicesUserChart := fmt.Sprintf("%s/{%s:.*}", pathServicesUser, varChartName)
	pathServicesUserRelease := fmt.Sprintf("%s/{%s:.*}", pathServicesUserChart, varReleaseName)

	app.Router.HandleFunc("/health", app.getHealthCheck).Methods("GET")
	app.Router.HandleFunc("/login", app.login).Methods("POST")

	app.SecureRouter.HandleFunc(pathAiida, app.getAiidaVersion).Methods("GET")
	app.SecureRouter.HandleFunc(pathAiida, app.installOrUpgradeAiida).Methods("POST")
	app.SecureRouter.HandleFunc(pathAiida, app.uninstallAiida).Methods("DELETE")

	app.SecureRouter.HandleFunc(pathServices, app.getServicesVersionsAll).Methods("GET")
	app.SecureRouter.HandleFunc(pathServicesUser, app.getServicesVersionsUser).Methods("GET")
	app.SecureRouter.HandleFunc(pathServicesUserChart, app.getServicesVersionsChart).Methods("GET")
	app.SecureRouter.HandleFunc(pathServicesUserChart, app.installNewService).Methods("POST")
	app.SecureRouter.HandleFunc(pathServicesUserRelease, app.getVersionInfoService).Methods("GET")
	app.SecureRouter.HandleFunc(pathServicesUserRelease, app.installOrUpgradeService).Methods("POST")
	app.SecureRouter.HandleFunc(pathServicesUserRelease, app.uninstallService).Methods("DELETE")
}

func (app *App) getHealthCheck(responseWriter http.ResponseWriter, _ *http.Request) {
	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessResponseDto())
}

package app

import (
	"github.com/awnumar/memguard"
	"github.com/eddie-energy/aiida-helm/installer/internal/aiida"
	"github.com/eddie-energy/aiida-helm/installer/internal/services"
	"github.com/gorilla/mux"
	"helm.sh/helm/v3/pkg/repo"
	"log"
	"net/http"
	"os"
	"time"
)

const repositoryUrl = "https://eddie-energy.github.io/aiida-helm/"
const repositoryName = "aiida"
const repositoryDefaultTrack = "stable"

const pathAiida = "/aiida"
const pathServices = "/services"
const pathServicesUser = pathServices + "/user"
const varChartName = "chartName"
const varReleaseName = "releaseName"

const userIdKey = "userId"

const bearerPrefix = "Bearer "
const jwtIssuer = "aiida-installer.local"
const jwtTokenLifetime = 30 * time.Minute

type App struct {
	Router            *mux.Router
	SecureRouter      *mux.Router
	Repository        repo.Entry
	AiidaInstaller    *aiida.Installer
	ServicesInstaller *services.Installer
	JwtKeyEnclave     *memguard.Enclave
}

func (app *App) Run(addr string) {
	app.initializeRepository()

	keyBuffer := memguard.NewBufferRandom(32)
	defer keyBuffer.Destroy()
	app.JwtKeyEnclave = keyBuffer.Seal()

	app.AiidaInstaller = aiida.NewInstaller(app.Repository)
	app.ServicesInstaller = services.NewInstaller(app.Repository)

	app.initializeRoutes()
	log.Fatal(http.ListenAndServe(addr, app.Router))
}

func (app *App) initializeRepository() {
	repositoryTrack := os.Getenv("REPOSITORY_TRACK")
	if repositoryTrack == "" {
		repositoryTrack = repositoryDefaultTrack
	}

	app.Repository = repo.Entry{
		Name:               repositoryName,
		URL:                repositoryUrl + repositoryTrack,
		PassCredentialsAll: true,
	}
}

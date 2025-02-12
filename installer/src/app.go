package main

import (
	"encoding/json"
	"fmt"
	"github.com/awnumar/memguard"
	"github.com/eddie-energy/aiida-helm/installer/internal/helm/aiida"
	"github.com/eddie-energy/aiida-helm/installer/internal/helm/services"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"helm.sh/helm/v3/pkg/repo"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const repositoryUrl = "https://eddie-energy.github.io/aiida-helm/"
const repositoryName = "aiida"
const repositoryDefaultTrack = "stable"

const pathAiida = "/aiida"
const pathService = "/service"
const varChartName = "chartName"

const bearerPrefix = "Bearer "
const jwtIssuer = "aiida-installer.local"

type App struct {
	Router            *mux.Router
	SecureRouter      *mux.Router
	Repository        repo.Entry
	AiidaInstaller    *aiida.Installer
	ServicesInstaller *services.Installer
	JwtKeyEnclave     *memguard.Enclave
}

func (app *App) Run(addr string) {
	repositoryTrack := os.Getenv("REPOSITORY_TRACK")
	if repositoryTrack == "" {
		repositoryTrack = repositoryDefaultTrack
	}

	app.Router = mux.NewRouter()

	secureRouter := app.Router.PathPrefix("/auth").Subrouter()
	secureRouter.Use(app.authMiddleware)
	app.SecureRouter = secureRouter

	keyBuffer := memguard.NewBufferRandom(32)
	defer keyBuffer.Destroy()
	app.JwtKeyEnclave = keyBuffer.Seal()

	app.Repository = repo.Entry{
		Name:               repositoryName,
		URL:                repositoryUrl + repositoryTrack,
		PassCredentialsAll: true,
	}
	app.AiidaInstaller = aiida.NewInstaller(app.Repository)
	app.ServicesInstaller = services.NewInstaller(app.Repository)

	versionInfo, err := app.AiidaInstaller.InstallOrUpgrade()
	if err != nil {
		log.Fatalf("Failed to install or upgrade aiida: %v", err.Error())
	}
	log.Printf("Installed or upgraded aiida: %v", versionInfo.InstalledChart.Version)

	app.initializeRoutes()
	log.Fatal(http.ListenAndServe(addr, app.Router))
}

func (app *App) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		authHeader := request.Header.Get("Authorization")
		if authHeader == "" {
			app.respondWithJSON(responseWriter, http.StatusUnauthorized, NewErrorResponse("missing authorization header"))
			return
		}

		if !strings.HasPrefix(authHeader, bearerPrefix) {
			app.respondWithJSON(responseWriter, http.StatusUnauthorized, NewErrorResponse("invalid authorization header"))
			return
		}

		jwtTokenString := strings.TrimPrefix(authHeader, bearerPrefix)
		if jwtTokenString == "" {
			app.respondWithJSON(responseWriter, http.StatusUnauthorized, NewErrorResponse("missing jwt token"))
			return
		}

		err := app.verifyJwtToken(jwtTokenString)
		if err != nil {
			app.respondWithJSON(responseWriter, http.StatusUnauthorized, NewErrorResponse(err.Error()))
			return
		}

		next.ServeHTTP(responseWriter, request)
	})
}

func (app *App) initializeRoutes() {
	fullServicePath := fmt.Sprintf("%s/{%s:.*}", pathService, varChartName)

	app.Router.HandleFunc("/health", app.getHealthCheck).Methods("GET")
	app.Router.HandleFunc("/login", app.login).Methods("POST")

	app.SecureRouter.HandleFunc(pathAiida, app.getVersionInfoAiida).Methods("GET")
	app.SecureRouter.HandleFunc(pathAiida, app.installOrUpgradeAiida).Methods("POST")
	app.SecureRouter.HandleFunc(pathAiida, app.uninstallAiida).Methods("DELETE")

	app.SecureRouter.HandleFunc(pathService, app.getVersionInfosRepoServices).Methods("GET")
	app.SecureRouter.HandleFunc(fullServicePath, app.getVersionInfoService).Methods("GET")
	app.SecureRouter.HandleFunc(fullServicePath, app.installOrUpgradeService).Methods("POST")
	app.SecureRouter.HandleFunc(fullServicePath, app.uninstallService).Methods("DELETE")
}

func (app *App) getHealthCheck(responseWriter http.ResponseWriter, _ *http.Request) {
	app.respondWithJSON(responseWriter, http.StatusOK, NewSuccessResponse())
}

func (app *App) login(responseWriter http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, NewErrorResponse(err.Error()))
	}

	err = app.verifyInstallerToken(string(body))
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, NewErrorResponse(err.Error()))
	}

	jwtToken, err := app.createJwtToken()
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, NewErrorResponse(err.Error()))
	}

	app.respondWithJSON(responseWriter, http.StatusOK, jwtToken)
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

func (app *App) respondWithJSON(responseWriter http.ResponseWriter, code int, response interface{}) {
	responseJson, _ := json.Marshal(response)

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(code)

	_, _ = responseWriter.Write(responseJson)
}

func (app *App) createJwtToken() (string, error) {
	jwtKeyBuffer, err := app.JwtKeyEnclave.Open()
	if err != nil {
		return "", err
	}
	defer jwtKeyBuffer.Destroy()

	now := time.Now()
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iss": jwtIssuer,
			"exp": now.Add(time.Minute * 30).Unix(),
			"iat": now.Unix(),
		},
	)
	return jwtToken.SignedString(jwtKeyBuffer.Bytes())
}

func (app *App) verifyJwtToken(tokenString string) error {
	jwtKeyBuffer, err := app.JwtKeyEnclave.Open()
	if err != nil {
		return err
	}
	defer jwtKeyBuffer.Destroy()

	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {
			return jwtKeyBuffer.Bytes(), nil
		},
		jwt.WithIssuer(jwtIssuer),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid jwt token")
	}

	return nil
}

func (app *App) verifyInstallerToken(token string) error {
	tokenBuffer, err := app.AiidaInstaller.Encryptor.InstallerTokenEnclave.Open()
	if err != nil {
		return err
	}
	defer tokenBuffer.Destroy()

	if tokenBuffer.String() != token {
		return fmt.Errorf("invalid installer token")
	}

	return nil
}

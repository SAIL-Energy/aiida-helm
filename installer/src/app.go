package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/awnumar/memguard"
	"github.com/eddie-energy/aiida-helm/installer/internal/dto"
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
const pathServiceUser = pathService + "/user"
const varChartName = "chartName"

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

func (app *App) initializeRoutes() {
	fullServicePath := fmt.Sprintf("%s/{%s:.*}", pathServiceUser, varChartName)

	app.Router.HandleFunc("/health", app.getHealthCheck).Methods("GET")
	app.Router.HandleFunc("/login", app.login).Methods("POST")

	app.SecureRouter.HandleFunc(pathAiida, app.getVersionInfoAiida).Methods("GET")
	app.SecureRouter.HandleFunc(pathAiida, app.installOrUpgradeAiida).Methods("POST")
	app.SecureRouter.HandleFunc(pathAiida, app.uninstallAiida).Methods("DELETE")

	app.SecureRouter.HandleFunc(pathService, app.getVersionInfosRepoServicesAll).Methods("GET")
	app.SecureRouter.HandleFunc(pathServiceUser, app.getVersionInfosRepoServices).Methods("GET")
	app.SecureRouter.HandleFunc(fullServicePath, app.getVersionInfoService).Methods("GET")
	app.SecureRouter.HandleFunc(fullServicePath, app.installOrUpgradeService).Methods("POST")
	app.SecureRouter.HandleFunc(fullServicePath, app.uninstallService).Methods("DELETE")
}

func (app *App) getHealthCheck(responseWriter http.ResponseWriter, _ *http.Request) {
	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessResponseDto())
}

func (app *App) login(responseWriter http.ResponseWriter, request *http.Request) {
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

func (app *App) getVersionInfoAiida(responseWriter http.ResponseWriter, _ *http.Request) {
	versionInfo, err := app.AiidaInstaller.GetVersionInfo()
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessDataResponseDto(versionInfo))
}

func (app *App) installOrUpgradeAiida(responseWriter http.ResponseWriter, _ *http.Request) {
	versionInfo, err := app.AiidaInstaller.InstallOrUpgrade()
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessDataResponseDto(versionInfo))
}

func (app *App) uninstallAiida(responseWriter http.ResponseWriter, _ *http.Request) {
	if err := app.AiidaInstaller.Uninstall(); err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessResponseDto())
}

func (app *App) getVersionInfosRepoServicesAll(responseWriter http.ResponseWriter, request *http.Request) {
	versionInfos, err := app.ServicesInstaller.GetVersionInfosRepo(nil)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessDataResponseDto(versionInfos))
}

func (app *App) getVersionInfosRepoServices(responseWriter http.ResponseWriter, request *http.Request) {
	userId, err := app.getUserIDFromContext(request)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	versionInfos, err := app.ServicesInstaller.GetVersionInfosRepo(&userId)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessDataResponseDto(versionInfos))
}

func (app *App) getVersionInfoService(responseWriter http.ResponseWriter, request *http.Request) {
	chartName := mux.Vars(request)[varChartName]
	if chartName == "" {
		app.respondWithJSON(responseWriter, http.StatusBadRequest, dto.NewErrorResponseDto("missing chart name"))
	}

	userId, err := app.getUserIDFromContext(request)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	versionInfo, err := app.ServicesInstaller.GetVersionInfo(userId, chartName)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessDataResponseDto(versionInfo))
}

func (app *App) installOrUpgradeService(responseWriter http.ResponseWriter, request *http.Request) {
	chartName := mux.Vars(request)[varChartName]
	if chartName == "" {
		app.respondWithJSON(responseWriter, http.StatusBadRequest, dto.NewErrorResponseDto("missing chart name"))
	}

	userId, err := app.getUserIDFromContext(request)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	versionInfo, err := app.ServicesInstaller.InstallOrUpgrade(userId, chartName)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessDataResponseDto(versionInfo))
}

func (app *App) uninstallService(responseWriter http.ResponseWriter, request *http.Request) {
	chartName := mux.Vars(request)[varChartName]
	if chartName == "" {
		app.respondWithJSON(responseWriter, http.StatusBadRequest, dto.NewErrorResponseDto("missing chart name"))
	}

	userId, err := app.getUserIDFromContext(request)
	if err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	if err := app.ServicesInstaller.Uninstall(userId, chartName); err != nil {
		app.respondWithJSON(responseWriter, http.StatusInternalServerError, dto.NewErrorResponseDto(err.Error()))
		return
	}

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessResponseDto())
}

func (app *App) respondWithJSON(responseWriter http.ResponseWriter, code int, response interface{}) {
	responseJson, _ := json.Marshal(response)

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(code)

	_, _ = responseWriter.Write(responseJson)
}

func (app *App) createJwtToken(userId string) (string, error) {
	jwtKeyBuffer, err := app.JwtKeyEnclave.Open()
	if err != nil {
		return "", err
	}
	defer jwtKeyBuffer.Destroy()

	now := time.Now()
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iss": jwtIssuer,
			"exp": now.Add(jwtTokenLifetime).Unix(),
			"iat": now.Unix(),
			"sub": userId,
		},
	)
	return jwtToken.SignedString(jwtKeyBuffer.Bytes())
}

func (app *App) verifyJwtToken(tokenString string) (*string, error) {
	jwtKeyBuffer, err := app.JwtKeyEnclave.Open()
	if err != nil {
		return nil, err
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
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid jwt token")
	}

	userId, err := token.Claims.GetSubject()
	if err != nil {
		return nil, err
	}

	return &userId, nil
}

func (app *App) verifyInstallerToken(token string) error {
	tokenBuffer, err := app.AiidaInstaller.Encryptor.InstallerTokenEnclave.Open()
	if err != nil {
		return err
	}
	defer tokenBuffer.Destroy()

	if tokenBuffer.String() != token {
		return errors.New("invalid installer token")
	}

	return nil
}

func (app *App) getUserIDFromContext(request *http.Request) (string, error) {
	userId, ok := request.Context().Value(userIdKey).(string)
	if !ok {
		return "", errors.New("user ID not found in context")
	}
	return userId, nil
}

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
	repositoryTrack := os.Getenv("REPOSITORY_TRACK")
	if repositoryTrack == "" {
		repositoryTrack = repositoryDefaultTrack
	}

	servicesNamespace := os.Getenv("SERVICES_NAMESPACE")

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
	app.ServicesInstaller = services.NewInstaller(app.Repository, servicesNamespace)

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
		log.Printf("Authenticating request %s %s", request.Method, request.URL.Path)

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
	log.Println("Health check")

	app.respondWithJSON(responseWriter, http.StatusOK, dto.NewSuccessResponseDto())
}

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

func (app *App) respondWithJSON(responseWriter http.ResponseWriter, code int, response dto.ResponseDto) {
	log.Printf("Responding with code %d with success %t and error %s", code, response.Success, response.Error)

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

func (app *App) getChartAndReleaseName(request *http.Request) (string, string, error) {
	chartName := mux.Vars(request)[varChartName]
	releaseName := mux.Vars(request)[varReleaseName]
	if chartName == "" || releaseName == "" {
		return "", "", errors.New("missing chart or release name")
	}
	return chartName, releaseName, nil
}

func (app *App) verifyServiceReleaseName(releaseName string, userId string) error {
	if !strings.HasSuffix(releaseName, userId) {
		return errors.New("no permission to access service release")
	}
	return nil
}

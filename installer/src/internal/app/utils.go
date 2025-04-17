package app

import (
	"encoding/json"
	"errors"
	"github.com/eddie-energy/aiida-helm/installer/internal/dto"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
)

func (app *App) respondWithJSON(responseWriter http.ResponseWriter, code int, response dto.ResponseDto) {
	if !response.Success {
		log.Printf("Responding with error: %s", response.Error)
	}

	responseJson, _ := json.Marshal(response)

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(code)

	_, _ = responseWriter.Write(responseJson)
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

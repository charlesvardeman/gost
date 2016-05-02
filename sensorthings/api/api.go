package api

import (
	"github.com/geodan/gost/configuration"

	"github.com/geodan/gost/sensorthings/models"
	"github.com/geodan/gost/sensorthings/rest"
)

// APIv1 is the default implementation of SensorThingsApi, API needs a database
// provider, config, endpoint information to setup te needed services
type APIv1 struct {
	db        models.Database
	config    configuration.Config
	endPoints []models.Endpoint
	//mqtt      mqtt.MQTTServer
}

// NewAPI Initialise a new SensorThings API
func NewAPI(database models.Database, config configuration.Config) models.API {
	return &APIv1{
		db: database,
		//mqtt:   mqtt,
		config: config,
	}
}

// GetConfig return the current configuration.Config set for the api
func (a *APIv1) GetConfig() *configuration.Config {
	return &a.config
}

// GetVersionInfo retrieves the version info of the current supported SensorThings API Version and running server version
func (a *APIv1) GetVersionInfo() *models.VersionInfo {
	versionInfo := models.VersionInfo{
		GostServerVersion: models.GostServerVersion{Version: configuration.ServerVersion},
		APIVersion:        models.APIVersion{Version: configuration.SensorThingsAPIVersion},
	}

	return &versionInfo
}

// GetBasePathInfo when navigating to the base resource path will return a JSON array of the available SensorThings resource endpoints.
func (a *APIv1) GetBasePathInfo() *models.ArrayResponse {
	var ep interface{} = a.GetEndpoints()
	basePathInfo := models.ArrayResponse{
		Data: &ep,
	}

	return &basePathInfo
}

// GetEndpoints returns all configured endpoints for the HTTP server
func (a *APIv1) GetEndpoints() *[]models.Endpoint {
	if a.endPoints == nil {
		a.endPoints = rest.CreateEndPoints(a.config.GetExternalServerURI())
	}

	return &a.endPoints
}

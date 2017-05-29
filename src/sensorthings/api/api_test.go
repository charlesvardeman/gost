package api

import (
	"github.com/geodan/gost/src/configuration"
	"github.com/geodan/gost/src/database/postgis"
	"github.com/geodan/gost/src/mqtt"
	"github.com/geodan/gost/src/sensorthings/entities"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateApi(t *testing.T) {
	// arrange
	cfg := configuration.Config{}
	mqttServer := mqtt.CreateMQTTClient(configuration.MQTTConfig{})
	database := postgis.NewDatabase("", 123, "", "", "", "", false, 50, 100, 200)
	stAPI := NewAPI(database, cfg, mqttServer)

	// act
	ep := stAPI.GetEndpoints()
	config := stAPI.GetConfig()
	paths := stAPI.GetAcceptedPaths()
	versionInfo := stAPI.GetVersionInfo()
	basePathInfo := stAPI.GetBasePathInfo()

	endpoints := *ep

	// assert
	assert.NotNil(t, basePathInfo)
	assert.NotNil(t, versionInfo)
	assert.NotNil(t, paths)
	assert.NotNil(t, config)
	assert.NotNil(t, ep)
	assert.NotEqual(t, len(endpoints), 0, "Endpoints empty")
}

func TestGetTopics(t *testing.T) {
	// arrange
	cfg := configuration.Config{}
	mqttServer := mqtt.CreateMQTTClient(configuration.MQTTConfig{})
	database := postgis.NewDatabase("", 123, "", "", "", "", false, 50, 100, 200)
	stAPI := NewAPI(database, cfg, mqttServer)

	// act
	topics := stAPI.GetTopics()

	// assert
	assert.NotNil(t, topics)
	firsttopic := (*topics)[0]
	assert.True(t, firsttopic.Path == "GOST/#")
}

func TestAppendQueryPart(t *testing.T) {
	// act
	result := appendQueryPart("base", "q")
	result1 := appendQueryPart("base?", "q")

	// assert
	assert.True(t, result == "base?q")
	assert.True(t, result1 == "base?&q")

}

func TestSetLinks(t *testing.T) {
	// arrange
	cfg := configuration.Config{}
	mqttServer := mqtt.CreateMQTTClient(configuration.MQTTConfig{})
	database := postgis.NewDatabase("", 123, "", "", "", "", false, 50, 100, 200)
	stAPI := NewAPI(database, cfg, mqttServer)
	ds := entities.Datastream{}

	// act
	stAPI.SetLinks(&ds, nil)

	// assert
	assert.True(t, ds.GetSelfLink() == "/v1.0/Datastreams")
}

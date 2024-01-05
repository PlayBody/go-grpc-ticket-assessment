package server

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerConfig_initConfig(t *testing.T) {
	dumpfile, err := os.CreateTemp("", "config-test.yaml")
	assert.NoError(t, err)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {

		}
	}(dumpfile.Name())

	yamlContent := `
train:
  sections:
    - A
    - B
  seat_count: 5
  routes:
    - from: London
      to: Paris
      price: 20
    - from: Osaka
      to: London
      price: 200
auth:
  secret_key: abc
  expire: 3600
roles:
  - email: a@a.com
    caps:
      - admin
      - read
      - write
  - email: b@b.com
    caps:
      - read
  - email: c@c.com
    caps:
      - write
`
	err = os.WriteFile(dumpfile.Name(), []byte(yamlContent), 0644)
	assert.NoError(t, err)

	// Create a SConfig instance and initialize it with the temporary file
	serverConfig := &Config{}
	err = serverConfig.InitConfig(dumpfile.Name())

	// Assert that initialization was successful
	assert.NoError(t, err)
	assert.Equal(t, []string{"A", "B"}, serverConfig.Train.Sections)
	assert.Equal(t, 5, serverConfig.Train.SeatCount)
	assert.Len(t, serverConfig.Train.Routes, 2)

	route1 := serverConfig.Train.Routes[0]
	assert.Equal(t, "London", route1.From)
	assert.Equal(t, "Paris", route1.To)
	assert.Equal(t, int32(20), route1.Price)

	route2 := serverConfig.Train.Routes[1]
	assert.Equal(t, "Osaka", route2.From)
	assert.Equal(t, "London", route2.To)
	assert.Equal(t, int32(200), route2.Price)

	assert.Equal(t, "abc", serverConfig.Auth.SecretKey)
	assert.Equal(t, int64(3600), serverConfig.Auth.Expire)

	admin := serverConfig.RoleUsers[0]
	assert.Equal(t, "a@a.com", admin.Email)
	assert.Equal(t, []string{"admin", "read", "write"}, admin.Capabilities)

	read := serverConfig.RoleUsers[1]
	assert.Equal(t, "b@b.com", read.Email)
	assert.Equal(t, []string{"read"}, read.Capabilities)

	write := serverConfig.RoleUsers[2]
	assert.Equal(t, "c@c.com", write.Email)
	assert.Equal(t, []string{"write"}, write.Capabilities)
}

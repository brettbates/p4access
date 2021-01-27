package config

import (
	"os"
	"testing"

	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	os.Setenv("P4ACCESS_P4PORT", "port:1666")
	os.Setenv("P4ACCESS_P4USER", "usr")
	os.Setenv("P4ACCESS_P4CLIENT", "client_ws")
	os.Setenv("P4ACCESS_LEAK", "true")

	var c Config
	err := envconfig.Process("p4access", &c)
	assert := assert.New(t)
	assert.Nil(err)
	assert.Equal("port:1666", c.P4Port)
	assert.Equal("usr", c.P4User)
	assert.Equal("client_ws", c.P4Client)
	assert.Equal(true, c.Leak)
}

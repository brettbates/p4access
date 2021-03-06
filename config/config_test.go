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
	os.Setenv("P4ACCESS_RESULTS", "/path/to/template.go.tpl")
	os.Setenv("P4ACCESS_HELP", "/path/to/help.txt")
	os.Setenv("P4ACCESS_LOG", "/path/to/p4access.log")

	var c Config
	err := envconfig.Process("p4access", &c)
	assert := assert.New(t)
	assert.Nil(err)
	assert.Equal("port:1666", c.P4Port)
	assert.Equal("usr", c.P4User)
	assert.Equal("client_ws", c.P4Client)
	assert.Equal("/path/to/template.go.tpl", c.Results)
	assert.Equal("/path/to/help.txt", c.Help)
	assert.Equal("/path/to/p4access.log", c.Log)
}

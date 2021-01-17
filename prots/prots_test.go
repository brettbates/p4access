package prots

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type FakeP4Runner struct {
	mock.Mock
}

func (mock *FakeP4Runner) Run(args []string) ([]map[interface{}]interface{}, error) {
	ags := mock.Called(args)
	return ags.Get(0).([]map[interface{}]interface{}), ags.Error(1)
}

var basicTest = []map[interface{}]interface{}{{
	"perm":      "super",
	"host":      "host",
	"user":      "user",
	"line":      "1",
	"depotFile": "//...",
	// No unmap, so should be false
}}

func TestProtections(t *testing.T) {
	fp4 := &FakeP4Runner{}
	fp4.On("Run", []string{"protects", "-a", "//depot/path/afile.txt"}).Return(basicTest, nil)
	res, err := Protections(fp4, "//depot/path/afile.txt")
	assert := assert.New(t)
	assert.Nil(err)
	assert.NotNil(res)
	fp4.AssertExpectations(t)
	assert.Equal(res[0].Perm, "super")
	assert.Equal(res[0].Host, "host")
	assert.Equal(res[0].User, "user")
	assert.Equal(res[0].Line, 1)
	assert.Equal(res[0].DepotFile, "//...")
	assert.Equal(res[0].Unmap, false)
}

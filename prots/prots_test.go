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
	res, _ := Protections(fp4, "//depot/path/afile.txt")
	// assert.NotNil(t, err)
	assert.NotNil(t, res)
	fp4.AssertExpectations(t)
}

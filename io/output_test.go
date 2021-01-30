package io

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/brettbates/p4access/config"
	"github.com/brettbates/p4access/prots"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type FakeP4Runner struct {
	mock.Mock
}

// Mocks p4.Run, so we can run fake perforce commands
func (mock *FakeP4Runner) Run(args []string) ([]map[interface{}]interface{}, error) {
	ags := mock.Called(args)
	return ags.Get(0).([]map[interface{}]interface{}), ags.Error(1)
}

type testGroup struct {
	group  string
	owners []prots.Owner
}

type resultsTestInput struct {
	adv    *prots.Advice
	args   Args
	groups testGroup
}

type resultsTest struct {
	input resultsTestInput
	want  string
}

var resultsTests = []resultsTest{
	{ // Single result
		resultsTestInput{
			&prots.Advice{
				Ps: prots.Prots{
					{
						Perm:      "read",
						Host:      "host",
						User:      "P_group_for_somewhere",
						IsGroup:   true,
						Line:      1,
						DepotFile: "//path/to/somewhere/...",
						Unmap:     false,
						Segments:  4,
					},
				},
				Context: "",
			},
			Args{
				"a.user",
				"read",
				"//path/to/somewhere/...",
			},
			testGroup{
				"P_group_for_somewhere",
				[]prots.Owner{
					{
						User:     "owner.first",
						FullName: "Owner First",
						Email:    "owner.first@email.com"},
				},
			},
		},
		"./want/single_result.txt",
	},
	{ // Context result
		resultsTestInput{
			&prots.Advice{
				Ps: prots.Prots{
					{
						Perm:      "read",
						Host:      "host",
						User:      "P_group_for_somewhere",
						IsGroup:   true,
						Line:      1,
						DepotFile: "//path/to/somewhere/...",
						Unmap:     false,
						Segments:  4,
					},
				},
				Context: "User a.user already has read access or higher to //path/to/somewhere/...",
			},
			Args{
				"a.user",
				"read",
				"//path/to/somewhere/...",
			},
			testGroup{
				"P_group_for_somewhere",
				[]prots.Owner{
					{
						User:     "owner.first",
						FullName: "Owner First",
						Email:    "owner.first@email.com"},
				},
			},
		},
		"./want/single_result_context.txt",
	},
}

// TODO share this with prots_test.go
func FakeOutput(fp4 *FakeP4Runner, groups testGroup) {
	gret := []map[interface{}]interface{}{{}}
	for i, o := range groups.owners {
		gret[0][fmt.Sprintf("Owners%d", i)] = o.User
		fp4.On("Run", []string{"user", "-o", o.User}).Return(
			[]map[interface{}]interface{}{{"Email": o.Email, "FullName": o.FullName}}, nil)
	}
	fp4.On("Run", []string{"group", "-o", groups.group}).Return(gret, nil)
}

func TestResults(t *testing.T) {
	var c config.Config
	err := envconfig.Process("p4access", &c)
	if err != nil {
		t.Errorf("Failed to set up config %v", err)
	}
	assert := assert.New(t)
	for _, tst := range resultsTests {
		fp4 := &FakeP4Runner{}
		FakeOutput(fp4, tst.input.groups)
		wantF, err := ioutil.ReadFile(tst.want)
		if err != nil {
			t.Errorf("Failed to read in file %s, %v", wantF, err)
		}
		wants := string(wantF)
		actual := Results(fp4, tst.input.adv, tst.input.args, c)
		// assert.Equal(wants, actual)
		// This makes it easier to see line differences
		assert.Equal(strings.Split(wants, "\n"), strings.Split(actual, "\n"))
	}
}

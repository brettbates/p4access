package prots

import (
	"errors"
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

type protTest struct {
	input []map[interface{}]interface{}
	want  []Prot
}

var protTests = []protTest{
	{input: []map[interface{}]interface{}{{
		"perm":      "super",
		"host":      "host",
		"user":      "user",
		"line":      "1",
		"depotFile": "//...",
		// No unmap, so should be false
	}},
		want: []Prot{{
			Perm:      "super",
			Host:      "host",
			User:      "user",
			IsGroup:   false,
			Line:      1,
			DepotFile: "//...",
			Unmap:     false,
		}},
	},
	{input: []map[interface{}]interface{}{{
		"perm":      "super",
		"host":      "host",
		"user":      "grp",
		"isgroup":   "",
		"line":      "1",
		"depotFile": "//...",
		// No unmap, so should be false
	}},
		want: []Prot{{
			Perm:      "super",
			Host:      "host",
			User:      "grp",
			IsGroup:   true,
			Line:      1,
			DepotFile: "//...",
			Unmap:     false,
		}},
	},
	{
		input: []map[interface{}]interface{}{
			{
				"perm":      "super",
				"host":      "host",
				"user":      "user",
				"line":      "1",
				"depotFile": "//...",
				// No unmap, so should be false
			},
			{
				"perm":      "list",
				"host":      "*",
				"unmap":     "", // negative
				"user":      "user",
				"line":      "2",
				"depotFile": "//depot/...",
			}},
		want: []Prot{
			{
				Perm:      "super",
				Host:      "host",
				User:      "user",
				IsGroup:   false,
				Line:      1,
				DepotFile: "//...",
				Unmap:     false,
			}, {
				Perm:      "list",
				Host:      "*",
				User:      "user",
				IsGroup:   false,
				Line:      2,
				DepotFile: "//depot/...",
				Unmap:     true,
			}},
	},
}

func TestProtections(t *testing.T) {
	for _, tst := range protTests {
		fp4 := &FakeP4Runner{}
		fp4.On("Run", []string{"protects", "-a", "//depot/path/afile.txt"}).Return(tst.input, nil)
		res, err := Protections(fp4, "//depot/path/afile.txt")
		assert := assert.New(t)
		assert.Nil(err)
		assert.Equal(res, tst.want)
	}
}

type AccessInput struct {
	user      string
	path      string
	reqAccess string                        // What access level we are are asking for
	retAccess []map[interface{}]interface{} // Access level returned by p4.Run()
}

type AccessResult struct {
	group  string
	access string
}

type accessTest struct {
	input AccessInput
	want  bool
	err   error
}

var accessTests = []accessTest{
	{
		input: AccessInput{
			"usr",
			"//depot/path/afile",
			"super",
			[]map[interface{}]interface{}{{"permMax": "super"}},
		},
		want: true,
		err:  nil,
	},
	{
		input: AccessInput{
			"usr",
			"//depot/path/afile",
			"read",
			[]map[interface{}]interface{}{{"permMax": "list"}},
		},
		want: false,
		err:  nil,
	},
	{
		input: AccessInput{
			"usr",
			"//notreal/path/afile",
			"read",
			[]map[interface{}]interface{}{{
				"code": "error",
				"data": "//notreal/... - must refer to client 'NP-B-BATES'."}},
		},
		want: false,
		err:  errors.New("exit status 1"),
	},
}

func TestHasAccess(t *testing.T) {
	// If a user already has access, don't look, just report that
	for _, tst := range accessTests {
		fp4 := &FakeP4Runner{}
		fp4.On("Run", []string{"protects", "-M", "-u", tst.input.user, tst.input.path}).Return(tst.input.retAccess, tst.err)
		res, err := hasAccess(fp4, tst.input.user, tst.input.path, tst.input.reqAccess)
		assert := assert.New(t)
		if tst.err == nil {
			assert.Nil(err)
		} else {
			assert.EqualError(err, tst.err.Error())
		}
		assert.Equal(res, tst.want)
	}
}

/**
So I remember for later, I was mixing up MaxAccess and the 'advise' function
type adviseInput struct {
	user    string
	group   string
	currMax string
	path    string
	prots   []Prot
}

type adviseResult struct {
	group  string
	access string
}

type adviseTest struct {
	input adviseInput
	want  adviseResult
	err   string
}

var adviseTests = []adviseTest{
	{
		input: adviseInput{
			"usr",
			"super",
			"//depot/path/afile",
			[]Prot{{
				Perm:      "super",
				Host:      "host",
				User:      "usr",
				Line:      1,
				DepotFile: "//...",
				Unmap:     false,
			}}},
		want: adviseResult{"", ""},
		err:  "User usr already has super advise to //depot/path/afile",
	},
}

func TestAdvise(t *testing.T) {
	for _, tst := range accessTests {
		fp4 := &FakeP4Runner{}
		fp4.On("Run", []string{"protects", "-M", "-u", tst.input.user, tst.input.path}).Return(tst.input.currMax, nil)
		res, err := MaxAccess(fp4, tst.input.user, tst.input.group, test.input.path, test.input.prots)
		assert := assert.New(t)
		if tst.err == "" {
			assert.Nil(err)
		} else {
			assert.EqualError(err, tst.err)
		}
		assert.Equal(res, tst.want)
		fmt.Printf("%v == %v", res, tst.want)
	}
}
}

**/

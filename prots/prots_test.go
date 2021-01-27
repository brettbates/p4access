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

// Mocks p4.Run, so we can run fake perforce commands
func (mock *FakeP4Runner) Run(args []string) ([]map[interface{}]interface{}, error) {
	ags := mock.Called(args)
	return ags.Get(0).([]map[interface{}]interface{}), ags.Error(1)
}

type protTest struct {
	input []map[interface{}]interface{}
	want  Prots
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
		want: Prots{{
			Perm:      "super",
			Host:      "host",
			User:      "user",
			IsGroup:   false,
			Line:      1,
			DepotFile: "//...",
			Unmap:     false,
			Segments:  1,
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
		want: Prots{{
			Perm:      "super",
			Host:      "host",
			User:      "grp",
			IsGroup:   true,
			Line:      1,
			DepotFile: "//...",
			Unmap:     false,
			Segments:  1,
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
		want: Prots{
			{
				Perm:      "super",
				Host:      "host",
				User:      "user",
				IsGroup:   false,
				Line:      1,
				DepotFile: "//...",
				Unmap:     false,
				Segments:  1,
			}, {
				Perm:      "list",
				Host:      "*",
				User:      "user",
				IsGroup:   false,
				Line:      2,
				DepotFile: "//depot/...",
				Unmap:     true,
				Segments:  2,
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
		assert.Equal(tst.want, res)
	}
}

type filterInput struct {
	path      string
	reqAccess string
	prots     Prots
}

type filterTest struct {
	input filterInput
	want  Prots
	err   error
}

var filterTests = []filterTest{
	{ // We should return the input prots
		input: filterInput{
			"//depot/mapped",
			"write",
			Prots{{
				Perm:      "write",
				Host:      "host",
				User:      "grp",
				IsGroup:   true,
				Line:      1,
				DepotFile: "//...",
				Unmap:     false,
				Segments:  1,
			}},
		},
		want: Prots{
			{
				Perm:      "write",
				Host:      "host",
				User:      "grp",
				IsGroup:   true,
				Line:      1,
				DepotFile: "//...",
				Unmap:     false,
				Segments:  1,
			},
		},
		err: nil,
	},
	{ // We should ignore prots that have more segments that the request
		// This may be too much of a heuristic, but if i ask for //depot/...
		// I shouldn't receive //depot/path/to/file protections
		input: filterInput{
			"//depot/mapped",
			"write",
			Prots{
				{
					Perm:      "write",
					Host:      "host",
					User:      "grp",
					IsGroup:   true,
					Line:      1,
					DepotFile: "//depot/mapped/longer",
					Unmap:     false,
					Segments:  3,
				},
				{
					Perm:      "write",
					Host:      "host",
					User:      "grp",
					IsGroup:   true,
					Line:      2,
					DepotFile: "//depot/mapped",
					Unmap:     false,
					Segments:  2,
				},
			},
		},
		want: Prots{
			{
				Perm:      "write",
				Host:      "host",
				User:      "grp",
				IsGroup:   true,
				Line:      2,
				DepotFile: "//depot/mapped",
				Unmap:     false,
				Segments:  2,
			},
		},
		err: nil,
	},
	{
		input: filterInput{
			"//depot/unmapped",
			"write",
			Prots{
				{
					Perm:      "write",
					Host:      "host",
					User:      "grp",
					IsGroup:   true,
					Line:      1,
					DepotFile: "//...",
					Unmap:     false,
					Segments:  1,
				},
				{
					Perm:      "write",
					Host:      "host",
					User:      "grp",
					IsGroup:   true,
					Line:      2,
					DepotFile: "//depot/...",
					Unmap:     true,
					Segments:  2,
				},
				{
					Perm:      "write",
					Host:      "host",
					User:      "grp2",
					IsGroup:   true,
					Line:      3,
					DepotFile: "//depot/...",
					Unmap:     false,
					Segments:  2,
				},
			},
		},
		want: Prots{
			{
				Perm:      "write",
				Host:      "host",
				User:      "grp2",
				IsGroup:   true,
				Line:      3,
				DepotFile: "//depot/...",
				Unmap:     false,
				Segments:  2,
			},
		},
		err: nil,
	},
}

func TestFilter(t *testing.T) {
	// Advise the user on which groups, to use
	for _, tst := range filterTests {
		fp4 := &FakeP4Runner{}
		pnone := []map[interface{}]interface{}{{"permMax": "none"}}
		pwrite := []map[interface{}]interface{}{{"permMax": "write"}}
		fp4.On("Run", []string{"protects", "-M", "-g", "grp", "//depot/mapped"}).Return(pwrite, nil).
			On("Run", []string{"protects", "-M", "-g", "grp", "//depot/unmapped"}).Return(pnone, nil).
			On("Run", []string{"protects", "-M", "-g", "grp2", "//depot/unmapped"}).Return(pwrite, nil)
		// TODO check that the group gives the correct access? or are we sure already
		// fp4.On("Run", []string{"protects", "-M", "-g", tst.input.user, tst.input.path}).Return("super", nil)
		res, err := tst.input.prots.filter(fp4, tst.input.path, tst.input.reqAccess)
		assert := assert.New(t)
		if tst.err == nil {
			assert.Nil(err)
		} else {
			assert.EqualError(err, tst.err.Error())
		}
		assert.Equal(tst.want, res)
	}
}

type adviseInput struct {
	user      string
	path      string
	reqAccess string
	prots     Prots
}

type adviseTest struct {
	input adviseInput
	want  Prots
	err   error
}

var adviseTests = []adviseTest{
	{ // This should fail as the user already has acccess
		input: adviseInput{
			"usr",
			"//depot/hasAccess",
			"write",
			Prots{{
				Perm:      "write",
				Host:      "host",
				User:      "grp",
				IsGroup:   true,
				Line:      1,
				DepotFile: "//...",
				Unmap:     false,
				Segments:  1,
			}}},
		want: nil,
		err:  errors.New("User usr already has write access or higher to //depot/hasAccess"),
	},
	{ // This should fail as we aren't requesting read or write
		// TODO This may be more appropriate in main.input()
		input: adviseInput{
			"usr",
			"//depot/hasAccess",
			"super",
			Prots{{
				Perm:      "write",
				Host:      "host",
				User:      "grp",
				IsGroup:   true,
				Line:      1,
				DepotFile: "//...",
				Unmap:     false,
				Segments:  1,
			}}},
		want: nil,
		err:  errors.New("Must request either read or write access"),
	},
	{ // Very simple test with a correct write group
		input: adviseInput{
			"usr",
			"//depot/path/afile",
			"write",
			Prots{{
				Perm:      "write",
				Host:      "host",
				User:      "grp",
				IsGroup:   true,
				Line:      1,
				DepotFile: "//...",
				Unmap:     false,
				Segments:  1,
			}}},
		want: Prots{{
			Perm:      "write",
			Host:      "host",
			User:      "grp",
			IsGroup:   true,
			Line:      1,
			DepotFile: "//...",
			Unmap:     false,
			Segments:  1,
		}},
		err: nil,
	},
	{ // Don't advise groups with paths further down the tree
		input: adviseInput{
			"usr",
			"//depot/...",
			"write",
			Prots{
				{
					Perm:      "write",
					Host:      "host",
					User:      "grp",
					IsGroup:   true,
					Line:      1,
					DepotFile: "//depot/path/to/afile",
					Unmap:     false,
					Segments:  4,
				},
				{
					Perm:      "write",
					Host:      "host",
					User:      "grp2",
					IsGroup:   true,
					Line:      2,
					DepotFile: "//depot/...",
					Unmap:     false,
					Segments:  2,
				},
			}},
		want: Prots{{
			Perm:      "write",
			Host:      "host",
			User:      "grp2",
			IsGroup:   true,
			Line:      2,
			DepotFile: "//depot/...",
			Unmap:     false,
			Segments:  2,
		}},
		err: nil,
	},
	{ // Correct group following higher access group
		input: adviseInput{
			"usr",
			"//depot/path/afile",
			"read",
			Prots{
				{
					Perm:      "super",
					Host:      "host",
					User:      "grp",
					IsGroup:   true,
					Line:      1,
					DepotFile: "//...",
					Unmap:     false,
					Segments:  1,
				},
				{
					Perm:      "read",
					Host:      "host",
					User:      "grp2",
					IsGroup:   true,
					Line:      2,
					DepotFile: "//depot/...",
					Unmap:     false,
					Segments:  2,
				}}},
		want: Prots{{
			Perm:      "read",
			Host:      "host",
			User:      "grp2",
			IsGroup:   true,
			Line:      2,
			DepotFile: "//depot/...",
			Unmap:     false,
			Segments:  2,
		}},
		err: nil,
	},
	{ // Request read with read and open available
		input: adviseInput{
			"usr",
			"//depot/path/afile",
			"read",
			Prots{
				{
					Perm:      "open",
					Host:      "host",
					User:      "grp",
					IsGroup:   true,
					Line:      1,
					DepotFile: "//...",
					Unmap:     false,
					Segments:  1,
				},
				{
					Perm:      "read",
					Host:      "host",
					User:      "grp2",
					IsGroup:   true,
					Line:      2,
					DepotFile: "//depot/...",
					Unmap:     false,
					Segments:  2,
				}}},
		// I only want to know of the closer 2nd line
		want: Prots{{
			Perm:      "read",
			Host:      "host",
			User:      "grp2",
			IsGroup:   true,
			Line:      2,
			DepotFile: "//depot/...",
			Unmap:     false,
			Segments:  2,
		}},
		err: nil,
	},
	{ // Request read with read and open available reverse order
		input: adviseInput{
			"usr",
			"//depot/path/afile",
			"read",
			Prots{
				{
					Perm:      "read",
					Host:      "host",
					User:      "grp2",
					IsGroup:   true,
					Line:      1,
					DepotFile: "//depot/...",
					Unmap:     false,
					Segments:  2,
				},
				{
					Perm:      "open",
					Host:      "host",
					User:      "grp",
					IsGroup:   true,
					Line:      2,
					DepotFile: "//...",
					Unmap:     false,
					Segments:  1,
				},
			}},
		// I only want to know of the closer 1st line
		want: Prots{{
			Perm:      "read",
			Host:      "host",
			User:      "grp2",
			IsGroup:   true,
			Line:      1,
			DepotFile: "//depot/...",
			Unmap:     false,
			Segments:  2,
		}},
		err: nil,
	},
	{ // Request read with read, open and write available, differing reads
		input: adviseInput{
			"usr",
			"//depot/path/afile",
			"read",
			Prots{
				{
					Perm:      "read",
					Host:      "host",
					User:      "grp2",
					IsGroup:   true,
					Line:      1,
					DepotFile: "//depot/...",
					Unmap:     false,
					Segments:  2,
				},
				{
					Perm:      "open",
					Host:      "host",
					User:      "grp",
					IsGroup:   true,
					Line:      2,
					DepotFile: "//...",
					Unmap:     false,
					Segments:  1,
				},
				{
					Perm:      "write",
					Host:      "host",
					User:      "grp",
					IsGroup:   true,
					Line:      3,
					DepotFile: "//depot/...",
					Unmap:     false,
					Segments:  2,
				},
			}},
		// I only want to know of the closer 1st line
		want: Prots{{
			Perm:      "read",
			Host:      "host",
			User:      "grp2",
			IsGroup:   true,
			Line:      1,
			DepotFile: "//depot/...",
			Unmap:     false,
			Segments:  2,
		}},
		err: nil,
	},
	{ // Request read with read, open and write available, same read paths
		input: adviseInput{
			"usr",
			"//depot/path/afile",
			"read",
			Prots{
				{
					Perm:      "read",
					Host:      "host",
					User:      "grp2",
					IsGroup:   true,
					Line:      1,
					DepotFile: "//depot/...",
					Unmap:     false,
					Segments:  2,
				},
				{
					Perm:      "open",
					Host:      "host",
					User:      "grp",
					IsGroup:   true,
					Line:      2,
					DepotFile: "//depot/...",
					Unmap:     false,
					Segments:  2,
				},
				{
					Perm:      "write",
					Host:      "host",
					User:      "grp",
					IsGroup:   true,
					Line:      3,
					DepotFile: "//depot/...",
					Unmap:     false,
					Segments:  2,
				},
			}},
		// We should get both groups read groups back as they give the same
		// It will be up to the user which they pick
		want: Prots{
			{
				Perm:      "open",
				Host:      "host",
				User:      "grp",
				IsGroup:   true,
				Line:      2,
				DepotFile: "//depot/...",
				Unmap:     false,
				Segments:  2,
			},
			{
				Perm:      "read",
				Host:      "host",
				User:      "grp2",
				IsGroup:   true,
				Line:      1,
				DepotFile: "//depot/...",
				Unmap:     false,
				Segments:  2,
			},
		},
		err: nil,
	},
	{
		// Request read with read, open and write available, same read paths
		// Unmap between reads
		input: adviseInput{
			"usr",
			"//unmap/path/afile",
			"read",
			Prots{
				{
					Perm:      "read",
					Host:      "host",
					User:      "grpunmap",
					IsGroup:   true,
					Line:      1,
					DepotFile: "//unmap/...",
					Unmap:     false,
					Segments:  2,
				},
				{
					Perm:      "read",
					Host:      "host",
					User:      "grpunmap",
					IsGroup:   true,
					Line:      2,
					DepotFile: "//unmap/...",
					Unmap:     true,
					Segments:  2,
				},
				{
					Perm:      "open",
					Host:      "host",
					User:      "grp",
					IsGroup:   true,
					Line:      3,
					DepotFile: "//unmap/...",
					Unmap:     false,
					Segments:  2,
				},
				{
					Perm:      "write",
					Host:      "host",
					User:      "grp",
					IsGroup:   true,
					Line:      4,
					DepotFile: "//unmap/...",
					Unmap:     false,
					Segments:  2,
				},
			}},
		// We should only get the open
		want: Prots{
			{
				Perm:      "open",
				Host:      "host",
				User:      "grp",
				IsGroup:   true,
				Line:      3,
				DepotFile: "//unmap/...",
				Unmap:     false,
				Segments:  2,
			},
		},
		err: nil,
	},
}

func TestAdvise(t *testing.T) {
	// Advise the user on which groups, to use
	for _, tst := range adviseTests {
		fp4 := &FakeP4Runner{}
		pnone := []map[interface{}]interface{}{{"permMax": "none"}}
		pwrite := []map[interface{}]interface{}{{"permMax": "write"}}
		psuper := []map[interface{}]interface{}{{"permMax": "super"}}
		fp4.On("Run", []string{"protects", "-M", "-u", tst.input.user, "//depot/hasAccess"}).Return(psuper, nil).
			On("Run", []string{"protects", "-M", "-u", tst.input.user, tst.input.path}).Return(pnone, nil).
			On("Run", []string{"protects", "-M", "-g", "grpunmap", "//unmap/path/afile"}).Return(pnone, nil).
			On("Run", []string{"protects", "-M", "-g", "grp", "//unmap/path/afile"}).Return(pwrite, nil).
			On("Run", []string{"protects", "-M", "-g", "grp", tst.input.path}).Return(pwrite, nil).
			On("Run", []string{"protects", "-M", "-g", "grp2", tst.input.path}).Return(pwrite, nil)
		// TODO check that the group gives the correct access? or are we sure already
		// fp4.On("Run", []string{"protects", "-M", "-g", tst.input.user, tst.input.path}).Return("super", nil)
		res, err := tst.input.prots.Advise(fp4, tst.input.user, tst.input.path, tst.input.reqAccess)
		assert := assert.New(t)
		if tst.err == nil {
			assert.Nil(err)
		} else {
			assert.EqualError(err, tst.err.Error())
		}
		assert.Equal(tst.want, res)
	}
}

func TestOwners(t *testing.T) {
	tst := []map[interface{}]interface{}{{
		"code":            "stat",
		"Group":           "P_group_name",
		"Timeout":         "43200",
		"Subgroups0":      "A_subgroup",
		"Users0":          "some.guy",
		"Users1":          "some.person",
		"Users3":          "a.user",
		"Users2":          "not.real",
		"PasswordTimeout": "unset",
		"MaxOpenFiles":    "unset",
		"MaxResults":      "unset",
		"Owners0":         "owner.first",
		"Owners1":         "owner.second",
		"MaxScanRows":     "unset",
		"MaxLockTime":     "unset",
	}}
	owner1 := []map[interface{}]interface{}{{
		"code":           "stat",
		"AuthMethod":     "ldap",
		"Update":         "2016/02/09 11:41:06",
		"passwordChange": "2018/11/25 13:43:16",
		"Access":         "2020/03/09 10:18:01",
		"extraTagType0":  "date",
		"User":           "owner.first",
		"FullName":       "Owner First",
		"Type":           "standard",
		"Email":          "owner.first@p4access.com",
		"extraTag0":      "passwordChange",
	}}
	owner2 := []map[interface{}]interface{}{{
		"code":           "stat",
		"AuthMethod":     "ldap",
		"Update":         "2016/02/09 11:41:06",
		"passwordChange": "2018/11/25 13:43:16",
		"Access":         "2020/03/09 10:18:01",
		"extraTagType0":  "date",
		"User":           "owner.second",
		"FullName":       "Owner Second",
		"Type":           "standard",
		"Email":          "owner.second@p4access.com",
		"extraTag0":      "passwordChange",
	}}
	tstProt := Prot{IsGroup: true, User: "P_group_name"}
	fp4 := &FakeP4Runner{}
	fp4.On("Run", []string{"group", "-o", "P_group_name"}).Return(tst, nil)
	fp4.On("Run", []string{"user", "-o", "owner.first"}).Return(owner1, nil)
	fp4.On("Run", []string{"user", "-o", "owner.second"}).Return(owner2, nil)

	res, err := tstProt.owners(fp4)

	assert := assert.New(t)
	assert.Nil(err)
	assert.Equal([]Owner{
		{"owner.first", "owner.first@p4access.com"},
		{"owner.second", "owner.second@p4access.com"}}, res)
}

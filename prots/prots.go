package prots

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	// This should be rcowham/go-libp4, but he needs to accept the pull request
	p4 "github.com/brettbates/go-libp4"
	"github.com/brettbates/p4access/config"
)

// P4Runner is an interface for testing without calling p4
type P4Runner interface {
	Run([]string) ([]map[interface{}]interface{}, error)
}

// P4C Is a wrapper around the p4 connection
type P4C struct {
	p4.P4
}

// NewP4C connects to p4 and returns a P4C wrapper
func NewP4C() *P4C {
	return &P4C{P4: *p4.NewP4()}
}

// NewP4CParams TODO This needs to read from .p4config files
func NewP4CParams(c config.Config) *P4C {
	return &P4C{P4: *p4.NewP4Params(c.P4Port, c.P4User, c.P4Client)}
}

// permMap maps permission levels to their hex value
var permMap map[string]uint8

func init() {
	permMap = map[string]uint8{
		"none":   0x000000, // Grants no access
		"list":   0x000001, // Grants list access
		"read":   0x000002, // Grants read access
		"branch": 0x000004, // Grants ability to branch/integ from - used with unmaps
		"open":   0x000008, // Grants open access
		"write":  0x000010, // Grants write access
		"review": 0x000020, // Grants review access
		"admin":  0x000080, // Grants admin access
		"super":  0x000040, // Grants super-user access
	}
}

// Prot is a single line of a protections table
type Prot struct {
	Perm      string
	Unmap     bool
	Host      string
	User      string
	IsGroup   bool
	Line      int
	DepotFile string
	Segments  int
}

// Owner represents the username and password of a group owner
type Owner struct {
	User     string
	FullName string
	Email    string
}

// owners returns the owners for a given prots group
func (p Prot) owners(p4r P4Runner) ([]Owner, error) {
	res, err := p4r.Run([]string{"group", "-o", p.User})
	if err != nil {
		return nil, err
	}

	i := 0
	out := []Owner{}
	// There is an indeterminate amount of OwnersX: owner.name
	// so we have to just try them all until we run out
	for {
		key := fmt.Sprintf("Owners%d", i)
		if v, ok := res[0][key]; ok {
			user := v.(string)
			// We have the username, find their email address
			ures, err := p4r.Run([]string{"user", "-o", user})
			if err != nil {
				return nil, err
			}
			var fullname, email string
			if uv, ok := ures[0]["Email"]; ok {
				email = uv.(string)
			}
			if uv, ok := ures[0]["FullName"]; ok {
				fullname = uv.(string)
			}
			out = append(out, Owner{user, fullname, email})
			i++
		} else {
			break
		}
	}
	return out, nil
}

// Prots is a set of protections
type Prots []Prot

func segments(path string) int {
	return len(strings.FieldsFunc(path, func(c rune) bool {
		return c == '/'
	}))
}

// Protections takes a path in p4 depot syntax
func Protections(p4r P4Runner, path string) (Prots, error) {
	res, err := p4r.Run([]string{"protects", "-a", path})
	if err != nil {
		log.Printf("Failed to get protects for %s\nRes: %v\nErr: %v\n", path, res, err)
	}

	prots := Prots{}
	for _, r := range res {
		p := Prot{}
		if v, ok := r["perm"]; ok {
			p.Perm = v.(string)
		}
		if v, ok := r["host"]; ok {
			p.Host = v.(string)
		}
		if v, ok := r["user"]; ok {
			p.User = v.(string)
		}
		if v, ok := r["line"]; ok {
			p.Line, err = strconv.Atoi((v.(string)))
		}
		if v, ok := r["depotFile"]; ok {
			p.DepotFile = v.(string)
		}
		if _, ok := r["unmap"]; ok {
			p.Unmap = ok
		}
		if _, ok := r["isgroup"]; ok {
			p.IsGroup = ok
		}
		p.Segments = segments(p.DepotFile)
		prots = append(prots, p)
	}
	return prots, err
}

// Info is the path and owners of a group
type Info struct {
	Path   string
	Access string
	Group  string
	Owners []Owner
}

// OutputInfo prepares the output for use in a template
func (ps *Prots) OutputInfo(p4r P4Runner, path, reqAccess string) ([]Info, error) {
	out := []Info{}
	for _, p := range *ps {
		owners, err := p.owners(p4r)
		if err != nil {
			return nil, err
		}
		// Don't report on ownerless groups
		if len(owners) > 0 {
			out = append(out, Info{
				path,
				reqAccess,
				p.User,
				owners,
			})
		}
	}
	if len(out) == 0 {
		return nil, errors.New("No matching groups found, try again with a more specific path")
	}
	return out, nil
}

// filterProts filters the output Prots from 'p4 protects' for those that pertain to the the request
func (ps *Prots) filter(p4r P4Runner, path, reqAccess string) (Prots, error) {
	out := Prots{}

	// read can be read or open, write is just write
	// TODO may need to make this configurable
	var minA, maxA uint8
	if reqAccess == "read" {
		minA = permMap["read"]
		maxA = permMap["open"]
	} else if reqAccess == "write" {
		minA = permMap["write"]
		maxA = permMap["write"]
	}

	// Reverse prots and filter out non-matching prots
	for i := len(*ps) - 1; i >= 0; i-- {
		c := (*ps)[i]

		/* We should ignore prots that have more segments that the request
		This may be too much of a heuristic, but if i ask for //depot/...
		I shouldn't receive //depot/path/to/file protections */
		pseg := segments(path)
		if c.Segments > pseg {
			continue
		}

		// Check that the group actually gives the correct access
		if permMap[c.Perm] >= minA && permMap[c.Perm] <= maxA {
			res, err := p4r.Run([]string{"protects", "-M", "-g", c.User, path})
			if err != nil {
				return nil, err
			}

			var permMax uint8
			if v, ok := res[0]["permMax"]; ok {
				permMax = permMap[v.(string)]
			} else {
				permMax = permMap["none"]
			}

			if permMax >= permMap[reqAccess] {
				out = append(out, c)
			}
		}
	}

	return out, nil
}

// sort reorders the given protections so that the long the path is (in segments), the earlier it is
// This might be too simplistic, but it seems to give decent results
func (ps *Prots) sort(path string) Prots {
	out := *ps
	// Stable means protections with the same number of segments are returned in reverse order
	// of the protections table
	sort.SliceStable(out, func(i, j int) bool {
		return (*ps)[i].Segments > (*ps)[j].Segments
	})
	return out
}

// Advise running user on probable group to join
// Returns one or more possible protections in order of how likely they are correct
func (ps *Prots) Advise(p4r P4Runner, user, path, reqAccess string) (Prots, error) {
	// TODO Move this to the command line parsing func
	if reqAccess != "read" && reqAccess != "write" {
		return nil, errors.New("Must request either read or write access")
	}

	a, err := hasAccess(p4r, user, path, reqAccess)
	if err != nil {
		return nil, err
	} else if a {
		return nil, fmt.Errorf("User %s already has %s access or higher to %s", user, reqAccess, path)
	}

	// Filter the prots for those that matter
	psf, err := ps.filter(p4r, path, reqAccess)
	if err != nil {
		return nil, fmt.Errorf("Failed to filter %v", err)
	}
	psf = psf.sort(path)
	l := psf[0].Segments
	out := Prots{psf[0]}

	// All matching prots with the same Segments length should be returned
	for i, p := range psf {
		if i > 0 && p.Segments == l {
			out = append(out, p)
		}
	}

	return out, nil
}

// hasAccess checks whether the given user already has access
func hasAccess(p4r P4Runner, user, path, reqAccess string) (bool, error) {
	res, err := p4r.Run([]string{"protects", "-M", "-u", user, path})
	if err != nil {
		log.Printf("\nFailed to run protects for user %s to path %s\n%v\n", user, path, err)
		return false, err
	}

	var permMax uint8
	if v, ok := res[0]["permMax"]; ok {
		permMax = permMap[v.(string)]
	} else {
		permMax = permMap["none"]
	}
	if permMax >= permMap[reqAccess] {
		return true, nil
	}
	return false, nil
}

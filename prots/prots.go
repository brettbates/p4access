package prots

import (
	"errors"
	"log"
	"sort"
	"strconv"
	"strings"

	p4 "github.com/rcowham/go-libp4"
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

// permMap maps permission levels to their hex value
var permMap map[string]uint8

func init() {
	permMap = map[string]uint8{
		"none":   0x000000, // none
		"list":   0x000001, // Grants list access
		"read":   0x000002, // Grants read access
		"branch": 0x000004, // Grants ability to branch/integ from
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

// Prots is a set of protections
type Prots []Prot

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
		p.Segments = len(strings.FieldsFunc(p.DepotFile, func(c rune) bool {
			return c == '/'
		}))
		prots = append(prots, p)
	}
	return prots, err
}

// filterProts filters the given prots for
func (ps *Prots) filter(reqAccess string) Prots {
	out := Prots{}

	var minA, maxA uint8
	if reqAccess == "read" {
		minA = permMap["read"]
		maxA = permMap["open"]
	} else if reqAccess == "write" {
		minA = permMap["write"]
		maxA = permMap["write"]
	}
	for i := len(*ps) - 1; i >= 0; i-- {
		c := (*ps)[i]
		// TODO this won't work if there are only groups available above the reqAccess
		if permMap[c.Perm] >= minA && permMap[c.Perm] <= maxA {
			out = append(out, c)
		}
	}

	return out
}

// sort reorders the given protections so that the closer the path is, the earlier it is
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
	// TODO Check !hasAccess
	// TODO Check that reqAccess is read or write only
	a, err := hasAccess(p4r, user, path, reqAccess)
	if err != nil {
		return nil, err
	} else if a {
		return nil, errors.New("User usr already has super access to //depot/hasAccess")
	}
	// Filter the prots for those that matter
	psf := ps.filter(reqAccess)
	psf = psf.sort(path)
	l := psf[0].Segments
	out := Prots{psf[0]}

	for i, p := range out {
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

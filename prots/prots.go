package prots

import (
	"log"
	"strconv"

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
}

// Protections takes a path in p4 depot syntax
func Protections(p4r P4Runner, path string) ([]Prot, error) {
	res, err := p4r.Run([]string{"protects", "-a", path})
	if err != nil {
		log.Printf("Failed to get protects for %s\nRes: %v\nErr: %v\n", path, res, err)
	}

	prots := []Prot{}
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
		prots = append(prots, p)
	}
	return prots, err
}

// Advise running user on probable group to join
func Advise(p4r P4Runner, user, path, reqAccess string, prots []Prot) ([]Prot, error) {
	out := []Prot{}
	for i := len(prots) - 1; i >= 0; i-- {
		c := prots[i]
		if permMap[c.Perm] >= permMap[reqAccess] {
			out = append(out, c)
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

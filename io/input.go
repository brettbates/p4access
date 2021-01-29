package io

import (
	"log"
	"os"

	p4b "github.com/brettbates/p4broker-reader/reader"
)

// Args are the arguments from 'p4 access reqAccess path'
type Args struct {
	User      string
	ReqAccess string
	Path      string
}

// Input gathers all the information p4broker has passed on
// Arg0 is read/write
// Arg1 is the path
func Input() Args {
	res, err := p4b.Read(os.Stdin)
	if err != nil {
		log.Fatalf("Failed to read in stdin, %v", err)
	}
	a := Args{
		res["user"],
		res["Arg0"],
		res["Arg1"],
	}
	return a
}

package main

import (
	"log"
	"os"

	"github.com/brettbates/p4access/config"
	"github.com/brettbates/p4access/io"
	"github.com/brettbates/p4access/prots"
	"github.com/kelseyhightower/envconfig"
)

func main() {
	var c config.Config
	io.Reject(envconfig.Process("p4access", &c))
	f, err := os.OpenFile(c.Log, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	args := io.Input()
	if args.ReqAccess == "-h" {
		io.Help(c)
		return
	}
	p4c := prots.NewP4CParams(c)
	res, err := prots.Protections(p4c, args.Path)
	io.Reject(err)
	advice, err := res.Advise(p4c, args.User, args.Path, args.ReqAccess)
	io.Reject(err)
	io.Results(p4c, advice, args, c)
}

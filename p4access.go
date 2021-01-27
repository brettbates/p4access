package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"

	"github.com/brettbates/p4access/config"
	"github.com/brettbates/p4access/prots"
	p4b "github.com/brettbates/p4broker-reader/reader"
	"github.com/kelseyhightower/envconfig"
)

// args are the arguments from 'p4 access reqAccess path'
type args struct {
	user      string
	reqAccess string
	path      string
}

// output creates p4broker friendly text to send back to the user
func output(p4r prots.P4Runner, ps prots.Prots, args args) {
	tmp, err := ioutil.ReadFile("response.go.tpl")
	if err != nil {
		log.Fatalln("Failed to find response.txt template")
	}
	t := template.Must(template.New("response").Parse(string(tmp)))
	info, err := ps.OutputInfo(p4r, args.path, args.reqAccess)
	if err != nil {
		log.Fatalf("Failed to retrieve the output data %v", err)
	}
	out := struct{ Groups []prots.Info }{Groups: info}
	err = t.Execute(os.Stdout, out)
	if err != nil {
		log.Fatalf("Failed to execute template\n%v", err)
	}
}

// input gathers all the information p4broker has passed on
// Arg0 is read/write
// Arg1 is the path
func input() args {
	res, err := p4b.Read(os.Stdin)
	if err != nil {
		log.Fatalf("Failed to read in stdin, %v", err)
	}
	a := args{
		res["user"],
		res["Arg0"], // TODO This should error if not read/write ?
		res["Arg1"],
	}
	return a
}

// reject will send a failure message to the user and record the error in a log file
func reject(err error) {
	if err != nil {
		fmt.Println("action: REJECT")
		fmt.Println("message: Failed to get protections, please contact support")
		log.Println("Failing, err recvd:")
		log.Println(err)
		os.Exit(0)
	}
}

func main() {
	f, err := os.OpenFile("output.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	args := input()
	var c config.Config
	reject(envconfig.Process("p4access", &c))
	p4c := prots.NewP4CParams(c)
	res, err := prots.Protections(p4c, args.path)
	reject(err)
	advice, err := res.Advise(p4c, args.user, args.path, args.reqAccess)
	reject(err)
	output(p4c, advice, args)
}

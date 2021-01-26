package main

import (
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/brettbates/p4access/prots"
	p4b "github.com/brettbates/p4broker-reader/reader"
)

type args struct {
	user      string
	path      string
	reqAccess string
}

// Info is the path and owners of a group
type Info struct {
	Access string
	Path   string
	Owners []string
}

// Groups maps Group names to their path/owners
type Groups map[string]Info

func output(ps prots.Prots, args args) {
	tmp, err := ioutil.ReadFile("response.txt")
	if err != nil {
		log.Fatalln("Failed to find answer.txt template")
	}
	t := template.Must(template.New("response").Parse(string(tmp)))
	out := struct{ Groups Groups }{Groups: Groups{
		"Test group": Info{
			"Read",
			"Test Path",
			[]string{"owner 1", "owner 2"},
		},
	}}
	err = t.Execute(os.Stdout, out)
	if err != nil {
		log.Fatalf("Failed to execute template\n%v", err)
	}
}

func input() args {
	res, err := p4b.Read(os.Stdin)
	if err != nil {
		log.Fatalf("Failed to read in stdin, %v", err)
	}
	a := args{
		res["user"],
		res["Arg0"],
		res["Arg1"],
	}
	return a
}

func logSetup() {
	// If running from main, make sure to print to a log file and stdout
	f, err := os.OpenFile("output.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	wrt := io.MultiWriter(os.Stdout, f)
	log.SetOutput(wrt)
}

func reject(err error) {
	if err != nil {
		log.Printf("action: REJECT")
		log.Printf("message: Failed to get protections, please contact support")
		log.Println(err)
	}
}

func main() {
	logSetup()
	args := input()
	p4c := prots.NewP4C()
	res, err := prots.Protections(p4c, args.path)
	reject(err)
	advice, err := res.Advise(p4c, args.user, args.path, args.reqAccess)
	reject(err)
	output(advice, args)
}

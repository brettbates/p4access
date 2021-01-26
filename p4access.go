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
	reqAccess string
	path      string
}

func output(ps prots.Prots, args args) {
	tmp, err := ioutil.ReadFile("response.txt")
	if err != nil {
		log.Fatalln("Failed to find answer.txt template")
	}
	t := template.Must(template.New("response").Parse(string(tmp)))
	out := struct{ Groups []prots.Info }{Groups: ps.OutputInfo(args.path, args.reqAccess)}
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
		panic(err)
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

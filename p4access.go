package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"os"

	p4b "github.com/brettbates/p4broker-reader"
)

// Info is the path and owners of a group
type Info struct {
	Access string
	Path   string
	Owners []string
}

// Groups maps Group names to their path/owners
type Groups map[string]Info

func output() {
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

func input() {
	p4b.Read()
}

func main() {
	output()
}

package io

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"text/template"

	"github.com/brettbates/p4access/config"
	"github.com/brettbates/p4access/prots"
)

// templateInfo is the struct fed to the result template
// any information you need in the template must be contained here
type templateInfo struct {
	Groups  []prots.Info
	Context string
}

// Results places successful Advise output into a p4broker friendly format
func Results(p4r prots.P4Runner, adv *prots.Advice, args Args, c config.Config) string {
	tmp, err := ioutil.ReadFile(c.Results)
	if err != nil {
		log.Fatalf("Failed to find response template %s", c.Results)
	}
	t := template.Must(template.New("response").Parse(string(tmp)))
	info, err := adv.OutputInfo(p4r, args.Path, args.ReqAccess)
	if err != nil {
		Reject(err)
	}
	out := templateInfo{info, adv.Context}
	var ob bytes.Buffer
	err = t.Execute(&ob, out)
	obs := ob.Bytes() // So we can write to Stdout and return the value
	os.Stdout.Write(obs)
	if err != nil {
		log.Fatalf("Failed to execute template\n%v", err)
	}
	return string(obs)
}

// Reject will send a failure message to the user and record the error in a log file
func Reject(err error) {
	if err != nil {
		out := "action: REJECT\n" +
			"message: \"Failing, error received:\n" +
			err.Error() +
			"\""
		fmt.Printf(out)
		// Write to log too
		log.Println("Failing, err recvd:")
		log.Println(out)
		os.Exit(0)
	}
}

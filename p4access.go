package main

import (
	"fmt"

	"github.com/brettbates/p4access/prots"
)

func main() {
	p4c := prots.NewP4C()
	res, err := prots.Protections(p4c, "//depot/...")
	fmt.Println(err)
	fmt.Println(res)
}

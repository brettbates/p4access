package main

import (
	"fmt"

	"github.com/brettbates/p4access/prots"
)

func main() {
	p4c := prots.NewP4C()
	fmt.Println(p4c.Protections("//depot/..."))
}

package main

import "fmt"
import "github.com/brettbates/p4access/prots"

func main() {
	fmt.Println(prots.Protections("//perforce/..."))
}

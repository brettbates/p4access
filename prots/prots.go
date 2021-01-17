package prots

import p4 "github.com/rcowham/go-libp4"
import "log"

// Just using env variables for now
var p4c *p4.P4 = p4.NewP4()

func getKeys(mp []map[interface{}]interface{}) []string{

	keys := make([]string, len(mp))

	i := 0
	for k := range mp {
		keys[i] = k
		i++
	}
	return keys
}

func Protections (path string) string {
	res, err := p4c.Run([]string{"protects", "-a", path})
	if err != nil {
		log.Fatalf("Failed to get protects for %s\n", path)
	}
	log.Println(res)
	return res
}

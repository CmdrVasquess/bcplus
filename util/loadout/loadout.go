package main

import (
	"encoding/json"
	"log"
	"os"

	"git.fractalqb.de/fractalqb/ggja"
)

func main() {
	dec := json.NewDecoder(os.Stdin)
	jobj := new(ggja.GenObj)
	if err := dec.Decode(&jobj); err != nil {
		log.Fatal(err)
	}
}

package ship

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"git.fractalqb.de/fractalqb/ggja"
)

func ExampleRefine() {
	jldo, _ := ioutil.ReadFile("prometheus.json")
	var ldo ggja.GenObj
	json.Unmarshal(jldo, &ldo)
	var shty ShipType
	shty.Refine(ggja.Obj{Bare: ldo})
	fmt.Println(shty)
	shipTypes[shty.Name] = &shty
	var ship Ship
	ship.Update(ggja.Obj{Bare: ldo})
	jenc := json.NewEncoder(os.Stdout)
	jenc.SetIndent("", "  ")
	err := jenc.Encode(&ship)
	fmt.Println(err)
	// Output:
	// _
}

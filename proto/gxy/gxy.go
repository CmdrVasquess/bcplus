package main

type BodyType int

const (
	UnknkowBody BodyType = iota
	Star
	Planet
	Belt
)

type Body struct {
	Type BodyType
	Name string
}

type System struct {
	Addr   uint64
	Name   string
	Bodies []Body
}

func main() {

}

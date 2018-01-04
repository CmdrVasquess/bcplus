package main

type Engie struct {
	Name        string
	Discover    string
	Requirement string
	Unlock      string
	GainRep     []string
}

type Group struct {
	Key     string
	Modules []*Module
}

type Module struct {
	Key   string
	Group *Group
}

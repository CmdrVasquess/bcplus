include VERSION

#PACK:=BCplus-$(major).$(minor).$(bugfix)$(quality).zip

all: godepgraph.svg

%.svg: %.dot
	dot -Tsvg $< > $@

.PHONY: godepgraph.dot
godepgraph.dot: Makefile
	godepgraph -l 2 -horizontal github.com/CmdrVasquess/BCplus > $@

pack: pack/pack BCplus
	pack/pack -pack zip

BCplus: version.go
	go build

version.go: VERSION
	go generate

pack/pack: $(wildcard pack/*.go)
	cd pack; go build

pack/vbcp.go: VERSION
	cd pack; go generate

all: godepgraph.svg

%.svg: %.dot
	dot -Tsvg $< > $@

.PHONY: godepgraph.dot
godepgraph.dot: Makefile
	godepgraph -l 2 -horizontal github.com/CmdrVasquess/BCplus > $@

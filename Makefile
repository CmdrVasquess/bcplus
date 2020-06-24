.PHONY: build deps utils clean

build:
	go generate ./...
	go build
	cd util/screenshot; go build

clean:
	rm -f BCplus BCplus.debug util/screenshot/screenshot

pack: build utils
	cd pack; go build
	pack/pack --pack zip --ddap BCplus

utils:
	cd util/screenshot; go build

deps: depgraph.svg

depgraph.svg:
	graphdot -p example.gprops | dot -Tsvg -o $@
#	graphdot -p 'node [shape=box]' | dot -Tsvg -o $@

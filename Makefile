.PHONY: build deps utils clean

build:
	go generate ./...
	go build -tags release
	cd util/screenshot; go build

clean:
	rm -f BCplus BCplus.debug util/screenshot/screenshot
	$(MAKE) -C test clean

test: build
	go test
	go test ./util/...
	go test ./pack
	$(MAKE) -C test

pack: build utils
	cd pack; go build
	pack/pack --pack zip --ddap BCplus

utils:
	cd util/screenshot; go build

deps: depgraph.svg

depgraph.svg:
	graphdot -p example.gprops | dot -Tsvg -o $@
#	graphdot -p 'node [shape=box]' | dot -Tsvg -o $@

JSDIR:=../res/s/js
GOSRC:=$(wildcard *.go)
GOEXE:=$(patsubst %.go,%,$(GOSRC))

all: $(JSDIR)/en/synth.js

goexes: $(GOEXE)

%: %.go
	go build $<

$(JSDIR)/%/synth.js: synth.xml synth
	test -d $(dir $@) || mkdir $(dir $@)
	./synth > $@

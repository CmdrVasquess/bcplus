RNCS:=$(wildcard *.rnc)
XMLS:=$(RNCS:.rnc=.xml)
JSON:=$(XMLS:.xml=.json)

all: $(JSON)

%.json: %.xml %2json.xsl
	xsltproc $*2json.xsl $< > $@

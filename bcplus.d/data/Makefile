JSON:=$(patsubst %2json.xsl,%.json,$(wildcard *2json.xsl))

all: $(JSON) ../shipyard.html

%.json: %.xml %2json.xsl
	xsltproc $*2json.xsl $< > $@

../shipyard.html: ships.xml ships2html.xsl
	xsltproc ships2html.xsl $< > $@

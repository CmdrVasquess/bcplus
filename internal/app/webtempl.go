package app

import (
	"encoding/json"
	"io"
	"path/filepath"
	"strings"
	"unicode"

	"git.fractalqb.de/fractalqb/goxic"
	"git.fractalqb.de/fractalqb/goxic/html"
	"git.fractalqb.de/fractalqb/goxic/js"
	"git.fractalqb.de/fractalqb/qbsllm"
)

type TmplLoader struct {
	parser *goxic.Parser
	dir    string
}

func NewTmplLoader(dir string) *TmplLoader {
	res := &TmplLoader{
		parser: html.NewParser(),
		dir:    dir,
	}
	if !App.debugMode {
		res.parser.PrepLine = func(line string) string {
			if line == "" {
				return ""
			}
			space := unicode.IsSpace(rune(line[0]))
			line = strings.TrimSpace(line)
			if space {
				line = " " + line
			}
			return line
		}
	}
	res.parser.CxfMap = map[string]goxic.CntXformer{
		"xml":   html.EscWrap,
		"jsStr": js.EscWrap,
	}
	return res
}

func (tld *TmplLoader) load(name, lang string) map[string]*goxic.Template {
	var fname string
	if len(lang) > 0 {
		fname = filepath.Join(tld.dir, lang, name)
	} else {
		fname = filepath.Join(tld.dir, name)
	}
	tname := name
	if ext := filepath.Ext(name); len(ext) > 0 {
		tname = name[:len(name)-len(ext)]
	}
	log.Debuga("load `template` from `file`", tname, fname)
	res := make(map[string]*goxic.Template)
	err := tld.parser.ParseFile(fname, tname, res)
	if err != nil {
		log.Fatale(err)
	}
	if log.Logs(qbsllm.Ltrace) {
		for nm, t := range res {
			if len(t.Phs()) == 0 {
				log.Tracea("`template` (`aka`) is static", nm, t.Name)
			} else {
				log.Tracea("`template` (`aka`) has `placeholders`",
					nm,
					t.Name,
					strings.Join(t.Phs(), ";"))
			}
		}
	}
	return res
}

func jsonContent(j interface{}) func(io.Writer) int {
	return func(wr io.Writer) int {
		enc := json.NewEncoder(wr)
		if App.debugMode {
			enc.SetIndent("", "\t")
		}
		err := enc.Encode(j)
		if err != nil {
			panic(err)
		}
		return 1
	}
}

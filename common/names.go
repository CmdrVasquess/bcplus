package common

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"git.fractalqb.de/fractalqb/ggja"
	"git.fractalqb.de/fractalqb/namemap"
	"git.fractalqb.de/fractalqb/qbsllm"
)

var (
	log       = qbsllm.New(qbsllm.Lnormal, "bc+cmn", nil, nil)
	LogConfig = qbsllm.Config(log)
)

const (
	DefaultLang    = "en"
	DomGame        = "ED"
	DomLocal       = "local"
	DomMaterialSym = "sym"
	namesDir       = "nms"
)

type NMap struct {
	namemap.FromTo
	save bool
	file string
}

func (nm *NMap) SetL10n(ed, loc string) {
	ed = strings.ToLower(ed)
	mapNm, mapDom := nm.Map(ed)
	if mapDom < 0 || mapNm != loc {
		nm.Base().Set(nm.FromIdx(), ed, nm.FromTo.ToIdxs()[0], loc)
		nm.save = true
	}
}

func (nm *NMap) PickL10n(evt ggja.Obj, edAtt, locAtt string) {
	edNm := evt.Str(edAtt, "")
	if len(edNm) == 0 {
		return
	}
	locNm := evt.Str(locAtt, "")
	if len(locNm) == 0 {
		return
	}
	nm.SetL10n(edNm, locNm)
}

func (nm *NMap) load(nmps *NameMaps, dataDir, lang, mapName string, xDoms ...string) {
	nmap := nmps.loadL10n(lang, mapName, xDoms)
	nm.FromTo = nmap.From(DomGame, true).To(false, DomLocal)
	nm.save = false
	nm.file = filepath.Join(dataDir, namesDir, lang, mapName)
}

func (nm *NMap) jicSave() {
	if !nm.save {
		return
	}
	log.Infoa("writing namemap `file`", nm.file)
	dir := filepath.Dir(nm.file)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Infoa("create `directoty`", dir)
		err = os.MkdirAll(dir, 0777)
		if err != nil {
			log.Fatala("`err`", err)
		}
	}
	tmpNm := nm.file + "~"
	wr, err := os.Create(tmpNm)
	if err != nil {
		log.Errora("`err`", err)
	}
	defer func() {
		if wr != nil {
			wr.Close()
		}
	}()
	err = nm.Base().Save(wr, "_")
	if err != nil {
		log.Errora("`err`", err)
	}
	wr.Close()
	wr = nil
	os.Rename(tmpNm, nm.file)
}

type NameMaps struct {
	resDir, dataDir string
	Lang            *namemap.NameMap
	LangEd, LangLoc int
	ShipType        NMap
	Material        NMap
	MaterialSym     int
}

func (nm *NameMaps) Load(resDir, dataDir, edLang string) {
	nm.resDir, nm.dataDir = resDir, dataDir
	nm.loadLangs()
	locale, _ := nm.Lang.Map(nm.LangEd, edLang, nm.LangLoc)
	nm.ShipType.load(nm, dataDir, locale, "ship-types.xsx")
	nm.Material.load(nm, dataDir, locale, "materials.xsx", DomMaterialSym)
	nm.MaterialSym = nm.Material.Base().DomainIdx(DomMaterialSym)
}

func (nm *NameMaps) Save() {
	nm.ShipType.jicSave()
}

func (nm *NameMaps) loadLangs() {
	if nm.Lang == nil {
		nm.Lang = &namemap.NameMap{}
	}
	err := nm.Lang.LoadFile(filepath.Join(nm.resDir, namesDir, "lang.xsx"))
	if err != nil {
		log.Fatala("`err`", err)
	}
	nm.LangEd = nm.Lang.DomainIdx(DomGame)   // TODO error
	nm.LangLoc = nm.Lang.DomainIdx(DomLocal) // TODO error
	sep := ""
	langs := bytes.NewBuffer(nil)
	nm.Lang.ForEach(nm.LangEd,
		func(v string) { fmt.Fprintf(langs, "%s%s", sep, v) })
	log.Infoa("available `languages`", langs.String())
}

// TODO If there is something like stat.CanRead() use that
func tryInDirs(base, lang string, do func(dir string) error) (hitDir string) {
	hitDir = filepath.Join(base, lang)
	err := do(hitDir)
	if err == nil {
		return hitDir
	}
	if len(lang) > 2 {
		hitDir = filepath.Join(base, lang[:2])
		err = do(hitDir)
		if err == nil {
			return hitDir
		}
	}
	hitDir = filepath.Join(base, DefaultLang)
	err = do(hitDir)
	if err == nil {
		return hitDir
	}
	return ""
}

func (nm *NameMaps) tryLocale(lang string, do func(dir string) error) (hitDir string) {
	base := filepath.Join(nm.resDir, namesDir)
	hitDir = tryInDirs(base, lang, do)
	if len(hitDir) > 0 {
		return hitDir
	}
	base = filepath.Join(nm.dataDir, namesDir)
	hitDir = tryInDirs(base, lang, do)
	return hitDir
}

func (nm *NameMaps) loadL10n(lang, mapName string, xDom []string) (nmap *namemap.NameMap) {
	log.Infoa("loading name `map` for `locale`", mapName, lang)
	loadMap := func(dir string) error {
		log.Tracea("try name `map` in `dir`", mapName, dir)
		rd, err := os.Open(filepath.Join(dir, mapName))
		if err != nil {
			return err
		}
		defer rd.Close()
		nmap = &namemap.NameMap{}
		err = nmap.Load(rd)
		return err
	}
	hitDir := nm.tryLocale(lang, loadMap)
	if len(hitDir) == 0 {
		log.Errora("could not load name `map`", mapName)
	} else {
		log.Debuga("took name `map` for `locale` from `dir`", mapName, lang, hitDir)
		if nmap == nil {
			log.Errora("found name map `dir`/`file` but result is nil", hitDir, mapName)
		}
	}
	if nmap == nil {
		nmap = namemap.NewNameMap(DomGame, DomLocal)
		nmap.SetStdDomain(DomGame)
	}
	return nmap
}

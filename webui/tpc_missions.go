package webui

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/CmdrVasquess/BCplus/cmdr"
	"github.com/CmdrVasquess/BCplus/galaxy"
	"github.com/ungerik/go3d/float64/vec3"
)

const (
	tkeyMissions = "missions"
)

var (
	gxtMissions gxtTopic
)

type tpcMissionData struct {
	Title   string
	DstId   int64
	DstSys  string
	Faction string
	Rep     float32
}

type tpcMsnSolver struct {
	msns []*cmdr.Mission
	stat []int
}

func newMsnSolver(c *cmdr.State) *tpcMsnSolver {
	res := new(tpcMsnSolver)
	for _, m := range c.Missions {
		if len(m.Dests) > 0 {
			res.msns = append(res.msns, m)
		}
	}
	res.stat = make([]int, len(res.msns))
	return res
}

func (slv *tpcMsnSolver) best(
	start *galaxy.Vec3D,
	chose int,
) (
	dist float64,
	end *galaxy.System,
	path []int,
) {
	chMsn := slv.msns[chose]
	chSys, err := theGalaxy.GetSystem(chMsn.Dests[slv.stat[chose]])
	if err != nil {
		sysId := slv.msns[chose].Dests[slv.stat[chose]]
		log.Errorf("cannot find mission destination system %d", sysId)
	}
	d := vec3.Distance(start, &chSys.Coos)
	slv.stat[chose]++
	var (
		optChs  = -1
		optDist float64
		optEnd  *galaxy.System
		optPath []int
	)
	for i := 0; i < len(slv.msns); i++ {
		if slv.stat[i] >= len(slv.msns[i].Dests) {
			continue
		}
		subDist, subEnd, subPath := slv.best(&chSys.Coos, i)
		if subDist < optDist || optChs < 0 {
			optChs = i
			optDist = subDist
			optEnd = subEnd
			optPath = subPath
		}
	}
	slv.stat[chose]--
	if optChs < 0 {
		return d, chSys, []int{chose}
	} else {
		return d + optDist, optEnd, append(optPath, chose)
	}
}

func (slv *tpcMsnSolver) solve(start *galaxy.Vec3D) (path []int, len float64) {
	if len(slv.msns) == 0 {
		return
	}
	optDist, optEnd, optPath := slv.best(start, 0)
	optDist += vec3.Distance(&optEnd.Coos, start)
	for i := 1; i < len(slv.msns); i++ {
		sd, se, sp := slv.best(start, i)
		sd += vec3.Distance(&se.Coos, start)
		if sd < optDist {
			optDist = sd
			optEnd = se
			optPath = sp
		}
	}
	return optPath
}

func (slv *tpcMsnSolver) visit(path []int, do func(m *cmdr.Mission, dest int)) {
	for i := range slv.stat {
		slv.stat[i] = 0
	}
	for _, d := range path {
		m := slv.msns[d]
		do(m, slv.stat[d])
		slv.stat[d]++
	}
}

func newMissions() (res []*tpcMissionData) {
	c := theCmdr()
	mslv := newMsnSolver(c)
	sys, _ := theGalaxy.GetSystem(c.Loc.SysId) // TODO error
	path := mslv.solve(&sys.Coos)
	log.Info(path)
	mslv.visit(path, func(m *cmdr.Mission, dst int) {
		dsys, err := theGalaxy.GetSystem(m.Dests[dst])
		if err != nil {
			log.Errorf("cannot find mission destination system %d", m.Dests[dst])
		}
		res = append(res, &tpcMissionData{
			Title:   m.Title,
			DstId:   dsys.Id,
			DstSys:  dsys.Name,
			Faction: m.Faction,
			Rep:     m.Reputation,
		})
	})
	return res
}

func tpcMissions(w http.ResponseWriter, r *http.Request) {
	bt := gxtMissions.NewBounT(nil)
	bindTpcHeader(bt, &gxtMissions)
	data := newMissions()
	bt.BindGen(gxtMissions.TopicData, func(wr io.Writer) int {
		enc := json.NewEncoder(wr)
		enc.SetIndent("", "\t")
		err := enc.Encode(data)
		if err != nil {
			panic(err)
		}
		return 1 // TODO howto determine the correct length
	})
	bt.Emit(w)
}

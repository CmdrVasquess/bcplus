package webui

import (
	"net/http"
	"time"

	"github.com/CmdrVasquess/BCplus/cmdr"
	"github.com/CmdrVasquess/BCplus/common"
	"github.com/CmdrVasquess/BCplus/galaxy"
	"github.com/ungerik/go3d/float64/vec3"
)

const (
	tkeyMissions = "missions"
)

var (
	gxtMissions gxtTopic
)

type tpcMsnDest struct {
	Title   string
	DstId   int64
	DstSys  string
	Faction string
	Rep     float32
}

type tpcMissionData struct {
	Dist  float64
	Dests []*tpcMsnDest
}

type tpcMsnSolver struct {
	msns    []*cmdr.Mission
	stat    []int
	resolve map[string]int
}

func newMsnSolver(c *cmdr.State) *tpcMsnSolver {
	res := new(tpcMsnSolver)
	for _, m := range c.Missions {
		if len(m.Dests) > 0 {
			res.msns = append(res.msns, m)
		}
	}
	res.stat = make([]int, len(res.msns))
	res.resolve = make(map[string]int)
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
		log.Errorf("failed to load mission destination system %d: %s", sysId, err)
		return 0.0, nil, nil
	}
	if chSys == nil {
		sysId := slv.msns[chose].Dests[slv.stat[chose]]
		log.Errorf("cannot find mission destination system %d", sysId)
		return 0.0, nil, nil
	}
	if !galaxy.V3dValid(chSys.Coos) {
		log.Tracef("system w/o coos: %d '%s'", chSys.Id, chSys.Name)
		slv.resolve[chSys.Name] = 1
		return 0.0, nil, nil
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
		if subEnd == nil {
			return 0.0, nil, nil
		}
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

func (slv *tpcMsnSolver) solve(start *galaxy.Vec3D) (path []int, dist float64) {
	if len(slv.msns) == 0 {
		return
	}
	optDist, optEnd, optPath := slv.best(start, 0)
	if optEnd == nil {
		return nil, 0.0
	}
	optDist += vec3.Distance(&optEnd.Coos, start)
	for i := 1; i < len(slv.msns); i++ {
		sd, se, sp := slv.best(start, i)
		if se == nil {
			return nil, 0.0
		}
		sd += vec3.Distance(&se.Coos, start)
		if sd < optDist {
			optDist = sd
			optEnd = se
			optPath = sp
		}
	}
	return optPath, optDist
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

func newMissions() (res *tpcMissionData) {
	c := theCmdr()
	mslv := newMsnSolver(c)
	sys, err := theGalaxy.GetSystem(c.Loc.SysId)
	if err != nil {
		log.Panicf("cannot resolve current system %d: %s", c.Loc.SysId, err)
	}
	if !galaxy.V3dValid(sys.Coos) {
		log.Tracef("system w/o coos: %d '%s'", sys.Id, sys.Name)
		sysResolve <- common.SysResolve{
			Names: []string{sys.Name},
		}
		return nil
	}
	if c.MissPath == nil {
		log.Debug("compute optimal mission path…")
		start := time.Now()
		c.MissPath, c.MissDist = mslv.solve(&sys.Coos)
		durtn := time.Since(start)
		if c.MissPath == nil {
			log.Debugf("no mission path after %s", durtn)
		} else {
			log.Debugf("mission path took %s", durtn)
		}
		if len(mslv.resolve) > 0 {
			var rq common.SysResolve
			for nm, _ := range mslv.resolve {
				rq.Names = append(rq.Names, nm)
			}
			sysResolve <- rq
		}
	}
	res = &tpcMissionData{Dist: c.MissDist}
	mslv.visit(c.MissPath, func(m *cmdr.Mission, dst int) {
		dsys, err := theGalaxy.GetSystem(m.Dests[dst])
		if err != nil {
			log.Panicf("cannot find mission destination system %d", m.Dests[dst])
		}
		res.Dests = append(res.Dests, &tpcMsnDest{
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
	if data == nil {
		data = &tpcMissionData{}
	}
	bt.BindGen(gxtMissions.TopicData, jsonContent(data))
	bt.Emit(w)
	CurrentTopic = UIMissions
}

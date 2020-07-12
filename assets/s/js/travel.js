Vue.component('trvlmap', {
    props: {
	size: Number,
	hbar: Number,
	data: Object
    },
    data: function() {return { anim: false }},
    template: '<canvas :width="(2*size)+hbar" :height="size" \
               v-on:click="anim=!anim;drawTrail(0)"></canvas>'
});

const trvlApp = new Vue({
    el: "main",
    data: {
	cfg: {
	    jblocks: [5,10,15,30]
	},
	jhist: theData.JumpHist ? theData.JumpHist : [],
	bookms: theData.Bookms,
	destbm: theData.DestBm,
	statsBlock: -1,
	tmapSys: {Coos:[]},
    },
    computed: {
	dest: function() {
	},
	speeds: function() { return [50, 200, 3000]; },
	tmap: function() {
	},
	jumps: function() {
	    var res = [];
	    if (this.jhist.length < 2) { return res; }
	    let loc = this.jhist[0].Coos;
	    for (let i=1; i < this.jhist.length; i++) {
	    	let a0 = this.jhist[i-1];
		if (a0.First) continue;
		let a1 = this.jhist[i];
	    	let t0 = new Date(a0.Time), t1 = new Date(a1.Time);
	    	let jump = {
	    	    dt: (t0-t1) / 1000.0,
	    	    jl: this.sysDist(a1.Coos, a0.Coos),
	    	    tl: this.sysDist(loc, a0.Coos)
	    	}
	    	res.push(jump);
	    }
	    return res;
	},
	ammBlk: function() {
	    var res = new Array(this.cfg.jblocks.length);
	    for (let i = 0; i < this.cfg.jblocks.length; i++) {
		let blen = this.cfg.jblocks[i];
		if (this.jumps.length < blen) {
		    res[i] = null;
		    continue;
		}
		res[i] = avgMinMax(this.jumps.slice(0, blen));
	    }
	    return res;
	},
	ammAll: function() {
	    return this.jumps.length == 0
		? null
		: avgMinMax(this.jumps);
	},
	trailLen: function() {
	    if (this.jhist.length == 0) return 0;
	    let hlen = this.statsBlock < 0
		? this.jhist.length
		: this.cfg.jblocks[this.statsBlock]+1;
	    if (hlen > this.jhist.length) { hlen = this.jhist.length; }
	    return hlen;
	}
    },
    methods: {
	onMsg: function(evt) {
	    if (evt.Cmd != "upd") return;
	    let jump = evt.P
	    if (jump) {
		if (this.jhist.length == 0) {
		    this.jhist.push(jump);
		} else if (this.jhist[0].Addr != jump.Addr) {
		    this.jhist.unshift(jump);
		    if (this.jhist.length > 51) {
			this.jhist.splice(51);
		    }
		}
		trlMem(this.jhist.length);
	    }
	    this.tmap.loc = this.hdrSystem();
	    this.$refs.tmap.paint();
	},
	sysDist: (l1, l2) => {
	    return Math.sqrt(sysDist2(l1, l2));
	},
	hdrSystem: function() {
	    if (!hdrData.Loc) return null;
	    if (hdrData.Loc['@type'] == 'system') return hdrData.Loc;
	    return hdrData.Loc.Sys;
	},
	tmpLoc: function(dl) {
	    this.tmapSys = dl;
	    this.$refs.tmap.paint();
	},
	reLoc: function() {
	    this.tmapSys = this.hdrSystem();
	    this.$refs.tmap.paint();
	},
	computeVic: function() {
	    switch (this.trailLen) {
	    case 0:
		return {c: [0, 0, 0], r: 400};
	    case 1:
		return {c: this.jhist[0].Coos, r: 400};
	    }
	    let dmax = 0, i0, i1;
	    for (let i=0; i+1 < this.trailLen; i++) {
		let p = this.jhist[i].Coos;
		for (let j=i+1; j < this.trailLen; j++) {
		    let d = sysDist2(p, this.jhist[j].Coos);
		    if (d > dmax) {
			i0 = i; i1 = j; dmax = d;
		    }
		}
	    }
	    let p0 = this.jhist[i0].Coos, p1 = this.jhist[i1].Coos;
	    let res = {
		c: [(p0[0]+p1[0])/2, (p0[1]+p1[1])/2, (p0[2]+p1[2])/2],
		r: Math.sqrt(dmax) / 1.7
	    };
	    return res;
	},
	sortJHist: function() {
	    if (theData.JumpHist) {
		theData.JumpHist.sort((l, r) => {
		    var ld = new Date(l.Time), rd = new Date(r.Time);
		    return rd.valueOf() - ld.valueOf();
		});
	    }
	}
    },
    mounted: function() {
	const app = this;
	//apiGetJSON("/travel", function(data) {
	//    hdrData.Name = data.Hdr.Cmdr;
	//    hdrData.Ship = data.Hdr.Ship;
	//    hdrData.Loc = data.Hdr.Loc;
	//    app.jhist = data.JumpHist;
	//    app.sortJHist();
	//});
	wsMsgCalls.push(this.onMsg);
    }
});

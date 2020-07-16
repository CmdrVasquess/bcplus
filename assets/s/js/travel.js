function sysDist2(coos1, coos2) {
    const dx = coos1[0]-coos2[0];
    const dy = coos1[1]-coos2[1];
    const dz = coos1[2]-coos2[2];
    return dx*dx + dy*dy + dz*dz;
}
function sysDist(coos1, coos2) { return Math.sqrt(sysDist2(coos1, coos2)); }

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
	jhist: []
    },
    computed: {
	jumps: function() {
	    if (this.jhist.length < 2) return [];
	    let res = [], last = this.jhist[0], t0 = new Date(last.Time);
	    for (let i=1; i < this.jhist.length; i++) {
		let now = this.jhist[i];
		let t1 = new Date(now.Time);
		res.push({
		    len: this.sysDist(last.Coos, now.Coos),
		    dur: t0 - t1
		});
		last = now;
		t0 = t1;
	    }
	    return res;
	}
    },
    methods: {
	onMsg(evt) {
	    switch (evt.Cmd) {
	    case 'upd':
		if (evt.Data.JumpHist) {
		    this.sortJHist( evt.Data.JumpHist);
		    this.jhist =  evt.Data.JumpHist;
		}
		break;
	    default:
		console.log("unkonwn command: "+evt.Cmd);
	    }
	    //this.$refs.tmap.paint();
	},
	sysDist(c1, c2) { return sysDist(c1,c2); },
	sortJHist(jh) {
	    if (jh) {
		jh.sort((l, r) => {
		    var ld = new Date(l.Time), rd = new Date(r.Time);
		    return rd.valueOf() - ld.valueOf();
		});
	    }
	}
    },
    mounted: function() {
	const app = this;
	wsMsgCalls.push(this.onMsg);
    }
});

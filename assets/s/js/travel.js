function sysDist2(coos1, coos2) {
    const dx = coos1[0]-coos2[0];
    const dy = coos1[1]-coos2[1];
    const dz = coos1[2]-coos2[2];
    return dx*dx + dy*dy + dz*dz;
}
function sysDist(coos1, coos2) { return Math.sqrt(sysDist2(coos1, coos2)); }

const trvlApp = new Vue({
    el: "main",
    data: {
	cfg: {
	    jblocks: [5,10,15,30]
	},
	jhist: [],
	travel: { width: 0, height: 0 }
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
	},
	tmapDim: function() {
	    const bar = 60;
	    let q = this.travel.width - bar;
	    if (q > this.travel.height)
		q = this.travel.height;
	    return {
		bar: bar,
		quad: q
	    };
	},
	tmapStyle: function() {
	    let sz = this.tmapDim.quad+"px "+this.tmapDim.quad+"px";
	    return {
		"background-size": sz
	    };
	}
    },
    methods: {
	onMsg(evt) {
	    switch (evt.Cmd) {
	    case 'upd':
		if (!evt.Data) return;
		if (evt.Data.JumpHist) {
		    this.sortJHist( evt.Data.JumpHist);
		    this.jhist =  evt.Data.JumpHist;
		} else if (evt.Data.Jump) {
		    this.jhist.unshift(evt.Data.Jump);
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
	},
	updTravelDim() {
// https://developer.mozilla.org/en-US/docs/Web/API/CSS_Object_Model/Determining_the_dimensions_of_elements	
	    let elm = document.getElementById('travel');
	    this.travel.width = elm.clientWidth - 30;
	    this.travel.height = window.innerHeight;
	    elm = document.getElementById('vue-app');
	    this.travel.height -= elm.offsetHeight + 30;
	}
    },
    mounted: function() {
	const app = this;
	wsMsgCalls.push(this.onMsg);
	window.addEventListener('resize', () => {
	    this.updTravelDim();
	});
	this.updTravelDim();
    }
});

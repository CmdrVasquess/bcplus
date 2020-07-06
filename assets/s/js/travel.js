const trl = {
	rot: [0, 0, 0, 0],
	sloc: [0, 0, 0], // screen x, y, z
	scrn: null,
	zord: null
};
function trlMem(n) {
	trl.scrn = new Array(3*n);
	trl.zord = new Int32Array(n);
}
trlMem(theData.JumpHist.length);
function trlZSort(n) {
    if (!trl.zord || trl.zord.length != n) {
	trl.zord = new Int32Array(n);
    }
    for (let i=0; i < n; i++) trl.zord[i] = i;
    trl.zord.sort((a, b) => { return trl.scrn[3*b+2] - trl.scrn[3*a+2];	});
    return trl.zord;
}

function sysDist2(l1, l2) {
    let dx = l1[0]-l2[0], dy = l1[1]-l2[1], dz = l1[2]-l2[2];
    return dx*dx + dy*dy + dz*dz;
}

function rot2d(w) {
    let s = Math.sin(w), c = Math.cos(w);
	trl.rot[0] = c; trl.rot[1] = -s;
	trl.rot[2] = s; trl.rot[3] = c;
    return trl.rot;
}
function rotate2d(R, p0, p1) {
    let r0 = R[0]*p0 + R[1]*p1;
    let r1 = R[2]*p0 + R[3]*p1;
    return [r0, r1];
}

Vue.component('trvlmap', {
    props: {
	size: Number,
	hbar: Number,
	data: Object
    },
    data: function() { return {
	anim: false
    }},
    template: '<canvas :width="(2*size)+hbar" :height="size" \
               v-on:click="anim=!anim;drawTrail(0)"></canvas>',
    computed: {
	width: function() { return 2 * this.size + this.hbar; },
	sd2: function() { return this.size / 2; },
	gxyScale: function() { return this.size / 94000; },
	gxySol: function() {
	    return { x: this.sd2, y: this.sd2, z: 0.765*this.size };
	},
	localScale: function() { return this.size / (2*this.data.vic.r); },
	g2: function() {
	    var canvas = this.$el;
	    return canvas.getContext("2d");
	}
    },
    mounted: function() { this.paint(); },
    watch: {
	"data.speeds": function() { this.paint(); }
    },
    methods: {
	paint: function() {
	    this.clear();
	    this.drawLoc();
	    this.drawDest();
	    this.drawTrail(0);
	},
	lyColor: (ly, sat, a) => {
	    let f = 4*Math.atan(ly/800)/Math.PI, r = 0, g = 0, b = 0;
	    if (f > 1) { r = 1; g = 2 - f; }
	    else if (f > 0) { r = f; g = 1; }
	    else if (f > -1) { g = 1; b = -f; }
	    else { g = 2 + f; b = 1; }
	    r = Math.round(sat*r); g = Math.round(sat*g); b = Math.round(sat*b);
	    return "rgba("+r+","+g+","+b+","+a+")";
	},
	screenZ: function(ly) {
	    return this.sd2 - this.size * Math.atan(ly/800) / Math.PI;
	},
	localXY: function(R, p, res, off) {
	    const vic = this.data.vic
	    let y = p[1] - vic.c[1];
	    let r = rotate2d(R, p[0] - vic.c[0], p[2] - vic.c[2]);
	    let z = 1 + r[1] * this.localScale / this.size;
		res[off] = this.size+this.hbar+this.sd2 + r[0] * this.localScale;
		res[++off] = this.sd2 - y/z  * this.localScale;
		res[++off] = z;
	},
	gxyXYZ: function(lx, ly, lz) {
	    let res = [
		this.gxySol.x + this.gxyScale * lx,
		this.gxySol.z - this.gxyScale * lz,
		this.screenZ(ly)
	    ];
	    return res;
	},
	clear: function() {
	    const g2 = this.g2;
     	    g2.clearRect(0, 0, this.size+this.hbar, this.size);
	    g2.font = "14px Arial";
	    g2.save();
	    let hbarfill = g2.createLinearGradient(0, 0, 0, this.size);
	    hbarfill.addColorStop(0, "#ff0000");
	    hbarfill.addColorStop(.25, "#ffff00");
	    hbarfill.addColorStop(.5, "#00ff00");
	    hbarfill.addColorStop(.75, "#00ffff");
	    hbarfill.addColorStop(1, "#0000ff");
	    g2.fillStyle = hbarfill;
	    g2.fillRect(this.size, 0, this.hbar, this.size);
	    g2.restore();
	},
	drawScale: function() {
	    const g2 = this.g2;
	    g2.save();
	    g2.lineWidth = 1;
	    g2.strokeStyle = "black";
	    for (let y = -4000; y <= 4000; y += 1000) {
		let z = this.screenZ(y);
		g2.beginPath();
		g2.moveTo(this.size, z);
		g2.lineTo(this.size + this.hbar, z);
		g2.stroke();
      	    }
	    for (let y = -1500; y <= 1500; y += 1000) {
		let z = this.screenZ(y);
		g2.beginPath();
		g2.moveTo(this.size + .25*this.hbar, z);
		g2.lineTo(this.size + .75*this.hbar, z);
		g2.stroke();
	    }
	    g2.restore();
	},
	drawMarker: function(scr) {
	    const g2 = this.g2;
	    g2.beginPath();
	    g2.arc(scr[0], scr[1], 3, 0, 2*Math.PI);
	    g2.fill();
	    g2.beginPath();
	    g2.moveTo(scr[0], scr[1]);
	    g2.lineTo(this.size, scr[2]);
	    g2.stroke();
	    return scr;
	},
	drawDest: function() {
	    if (!this.data.dest) { return; }
	    const g2 = this.g2;
	    g2.save();
	    g2.shadowColor = "black";
	    g2.shadowBlur = 5;
	    g2.lineWidth = 1.2;
	    let scr = this.gxyXYZ(this.data.dest.Coos[0],
				  this.data.dest.Coos[1],
				  this.data.dest.Coos[2]);
	    let loc = this.gxyXYZ(this.data.loc.Coos[0],
				  this.data.loc.Coos[1],
				  this.data.loc.Coos[2]);
	    g2.beginPath();
	    g2.strokeStyle = "#D25C00";
	    g2.setLineDash([5, 3]);
	    g2.moveTo(loc[0], loc[1]);
	    g2.lineTo(scr[0], scr[1]);
	    g2.stroke();
	    g2.setLineDash([]);
	    g2.fillStyle = "#ff7000";
	    this.drawMarker(scr);
	    g2.strokeStyle = "#ff7000";
	    g2.moveTo(this.size, scr[2]);
	    g2.lineTo(this.size + this.hbar, scr[2]);
	    g2.stroke();
	    g2.restore();
	},
	drawLoc: function() {
	    const g2 = this.g2;
	    g2.save();
	    let scr = this.gxyXYZ(this.data.loc.Coos[0],
				  this.data.loc.Coos[1],
				  this.data.loc.Coos[2]);
	    let bar = this.hbar / 3;
	    g2.strokeStyle = "#00A6EA";
	    g2.fillStyle = "#00A6EABB";
	    g2.fillRect(this.size+bar, this.sd2, bar, -(this.sd2-scr[2]));
	    this.drawScale();
	    g2.shadowColor = "black";
	    g2.shadowBlur = 5;
	    g2.beginPath();
	    g2.fillStyle = "#00A6EA66";
	    let h1r = this.data.speeds[2] * this.gxyScale;
	    g2.arc(scr[0], scr[1], h1r, 0, 2*Math.PI);
	    g2.fill();
	    g2.save();
	    g2.globalCompositeOperation = 'destination-out';
	    g2.fillStyle = "white";
	    g2.shadowBlur = 0;
	    g2.beginPath();
	    h1r = this.data.speeds[0] * this.gxyScale;
	    g2.arc(scr[0], scr[1], h1r, 0, 2*Math.PI);
	    g2.fill();
	    g2.restore();
	    g2.beginPath();
	    g2.lineWidth = .9;
	    h1r = this.data.speeds[1] * this.gxyScale;
	    g2.arc(scr[0], scr[1], h1r, 0, 2*Math.PI);
	    g2.stroke();
	    g2.beginPath();
	    g2.strokeStyle = "#00A6EA";
	    g2.shadowBlur = 5;
	    g2.lineWidth = 1.5;
	    g2.moveTo(0, scr[1]);
	    g2.lineTo(scr[0], scr[1]);
	    g2.lineTo(scr[0], this.size);
	    g2.moveTo(scr[0], scr[1]);
	    g2.lineTo(this.size, scr[2]);
	    g2.lineTo(this.size+this.hbar, scr[2]);
	    g2.stroke();
	    g2.fillStyle = "white";
	    if (scr[0] > this.sd2) {
		lb = Math.round(this.data.loc.Coos[0]);
		txm = g2.measureText(lb);
		g2.fillText(lb, scr[0]-txm.width-3, this.size-3);
	    } else {
		g2.fillText(Math.round(this.data.loc.Coos[0]), scr[0]+3, this.size-3);
	    }
	    if (scr[1] > this.sd2) {
		g2.fillText(Math.round(this.data.loc.Coos[2]), 2, scr[1]-3);
	    } else {
		g2.fillText(Math.round(this.data.loc.Coos[2]), 2, scr[1]+14);
	    }
	    var lb = Math.round(this.data.loc.Coos[1]);
	    var txm = g2.measureText(lb);
	    g2.fillStyle = "black";
	    g2.shadowColor = "#ff7000";
	    if (scr[2] > this.sd2) {
		g2.fillText(lb, this.size+(this.hbar-txm.width)/2, scr[2]+14);
	    } else {
		g2.fillText(lb, this.size+(this.hbar-txm.width)/2, scr[2]-3);
	    }
	    g2.restore();
	},
	drawTrail: function(t) {
	    const jumps = this.data.jhist;
	    if (jumps.length == 0) { return; }
	    const w = t/2000, R = rot2d(-w);
	    for (let i=0; i < jumps.length; i++) {
		this.localXY(R, jumps[i].Coos, trl.scrn, 3*i);
	    }
	    const zord = trlZSort(jumps.length);
	    const g2 = this.g2;
   	    g2.save();
  	    g2.clearRect(this.size+this.hbar, 0, this.size, this.size);
	    g2.lineWidth = 1.2;
	    g2.lineCap = 'round';
	    g2.lineJoin = 'round';
	    const sin = Math.sin(w), cos = Math.cos(w);
	    const cx = 1.5*this.size+this.hbar, cy = this.size/2;
	    g2.beginPath();
	    g2.strokeStyle = "white";
	    g2.moveTo(cx+this.size/2.1*sin, cy+this.size/2.1*cos);
	    g2.lineTo(cx+this.size/2.2*sin, cy+this.size/2.2*cos);
	    g2.stroke();
	    g2.beginPath();
	    g2.strokeStyle = "#ff7000";
	    g2.moveTo(trl.scrn[0], trl.scrn[1]);
	    for (let i=1; i < jumps.length; i++) {
		let off = 3*i;
		g2.lineTo(trl.scrn[off], trl.scrn[off+1]);
	    }
	    g2.stroke();
	    g2.strokeStyle = "black";
	    g2.lineWidth = .9;
	    g2.fillStyle = "#ff7000";
	    for (let i=0; i < jumps.length; i++) {
		let off = 3*zord[i];
		g2.fillStyle = this.lyColor(jumps[zord[i]].Coos[1], 255, 1);
		g2.beginPath();
		g2.arc(trl.scrn[off], trl.scrn[off+1], 3, 0, 2*Math.PI);
		g2.fill();
		g2.stroke();
	    }		
	    this.localXY(R, this.data.loc.Coos, trl.sloc, 0);
	    let scr = this.gxyXYZ(this.data.loc.Coos[0],
				this.data.loc.Coos[1],
				this.data.loc.Coos[2]);
	    g2.beginPath();
	    g2.lineWidth = 1.5;
	    g2.strokeStyle = "#00A6EA";
	    g2.fillStyle = "#00A6EA";
	    g2.moveTo(this.size+this.hbar, scr[2]);
	    g2.lineTo(trl.sloc[0], trl.sloc[1]);
	    g2.stroke();
	    g2.beginPath();
	    g2.restore();
	    if (this.anim)
		window.requestAnimationFrame(this.drawTrail);
	}
    }
});

function lyph(l, t) { return 3600 * l / t; }

function avgMinMax(jumps) {
    if (jumps.length == 0) { return null; }
    let j = jumps[0], speed = lyph(j.jl, j.dt);
    let res = {
	tAvg: j.dt,	tMin: j.dt, tMax: j.dt,
	jAvg: j.jl,	jMin: j.jl, jMax: j.jl,
	sAvg: speed, sMin: speed, sMax: speed
    };
    for (let i=1; i < jumps.length; i++) {
	j = jumps[i];
	speed = lyph(j.jl, j.dt);
	res.tAvg += j.dt;
	if (j.dt < res.tMin) { res.tMin = j.dt; }
	else if (j.dt > res.tMax) { res.tMax = j.dt; }
	res.jAvg += j.jl;
	if (j.jl < res.jMin) { res.jMin = j.jl; }
	else if (j.jl > res.jMax) { res.jMax = j.jl; }
	res.sAvg += speed;
	if (speed < res.sMin) { res.sMin = speed; }
	else if (speed > res.sMax) { res.sMax = speed; }
    }
    res.tAvg /= jumps.length;
    res.jAvg /= jumps.length;
    res.sAvg /= jumps.length;
    return res;
}

var trvlApp = new Vue({
    el: "main",
    data: {
	cfg: {
	    jblocks: [5,10,15,30]
	},
	jhist: theData.JumpHist ? theData.JumpHist : [],
	bookms: theData.Bookms,
	destbm: theData.DestBm,
	statsBlock: -1,
	tmapLoc: hdrData.Loc.Sys
    },
    computed: {
	dest: function() {
	    if (this.bookms && this.destbm >= 0) {
		return this.bookms[this.destbm];
	    }
	    return null
	},
	speeds: function() {
	    var block;
	    if (this.statsBlock >= 0) {
		block = this.ammBlk[this.statsBlock];
	    } else {
		block = this.ammAll;
	    }
	    if (block) {
		return [block.sMin, block.sAvg, block.sMax];
	    }
	    return [420, 3000, 6400];
	},
	tmap: function() {
	    let res = {
		vic: this.computeVic(),
		loc: this.tmapLoc,
		speeds: this.speeds,
		selSpd: 1,
		dest: this.dest,
		statsBlock: this.statsBlock,
		jhist: this.jhist.slice(0, this.trailLen)
	    };
	    return res;
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
	    if (this.jhist.length == 0) { return 0; }
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
	    this.tmap.loc = hdrData.Loc.Sys;
	    this.$refs.tmap.paint();
	},
	sysDist: (l1, l2) => {
	    return Math.sqrt(sysDist2(l1, l2));
	},
	tmpLoc: function(dl) {
	    this.tmapLoc = dl;
	    this.$refs.tmap.paint();
	},
	reLoc: function() {
	    this.tmapLoc = hdrData.Loc.Sys;
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
	}
    },
    beforeCreate: () => {
	if (theData.JumpHist) {
	    theData.JumpHist.sort((l, r) => {
		var ld = new Date(l.Time), rd = new Date(r.Time);
		return rd.valueOf() - ld.valueOf();
	    });
	}
    },
    mounted: function() {
		apiGetJSON("/travel", function(data) {
			hdrData.Name = data.Hdr.Cmdr;
			hdrData.Ship = data.Hdr.Ship;
			hdrData.Loc = data.Hdr.Loc;
			this.jhist = data.JumpHist;
		});
		wspgmsg.push(this.onMsg);
		console.log("added travel callback");
    }
});

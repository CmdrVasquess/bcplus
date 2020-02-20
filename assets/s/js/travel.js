Vue.component('trvlmap', {
    props: {
		size: Number,
		hbar: Number,
		data: Object
    },
    template: '<canvas :width="(2*size)+hbar" :height="size"></canvas>',
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
			this.drawDest();
			this.drawLoc();
			this.drawTrail();
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
		localXY: function(lx, lz) {
			let vic = this.data.vic;
			let x = this.size+this.hbar+this.sd2 + (lx-vic.cx) * this.localScale;
			let y = this.sd2 - (lz-vic.cz) * this.localScale;
			return [x, y];
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
     	    g2.clearRect(0, 0, this.width, this.size);
			g2.font = "14px Arial";
			g2.save();
			g2.fillStyle = this.lyColor(this.data.loc.Coos[1], 71, 1);
			g2.fillRect(this.size+this.hbar, 0, this.size, this.size);
			let hbarfill = g2.createLinearGradient(0, 0, 0, this.size);
			hbarfill.addColorStop(0, "#ff0000");
			hbarfill.addColorStop(.25, "#ffff00");
			hbarfill.addColorStop(.5, "#00ff00");
			hbarfill.addColorStop(.75, "#00ffff");
			hbarfill.addColorStop(1, "#0000ff");
			g2.fillStyle = hbarfill;
			g2.fillRect(this.size, 0, this.hbar, this.size);
			if (this.data.vic) {
				let vc = this.gxyXYZ(this.data.vic.cx, this.data.vic.cz, 0);
				let vs = this.data.vic.r * this.gxyScale;
				g2.strokeStyle = "#FF7000";
				g2.lineWidth = .8;
				g2.shadowColor = "black";
				g2.shadowBlur = 5;
				g2.strokeRect(vc[0]-vs, vc[1]-vs, 2*vs, 2*vs);
			}
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
			let h1r = this.data.speeds[2] * this.gxyScale;
			g2.arc(scr[0], scr[1], h1r, 0, 2*Math.PI);
			g2.fill();
			g2.save();
			g2.globalCompositeOperation = 'destination-out';
			g2.shadowBlur = 0;
			g2.beginPath();
			h1r = this.data.speeds[0] * this.gxyScale;
			g2.arc(scr[0], scr[1], h1r, 0, 2*Math.PI);
			g2.fill();
			g2.restore();
			g2.beginPath();
			g2.lineWidth = 1;
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
			let lxy = this.localXY(this.data.loc.Coos[0], this.data.loc.Coos[2]);
			g2.beginPath();
			g2.shadowBlur = 0;
			g2.moveTo(this.size+this.hbar, scr[2]);
			g2.lineTo(lxy[0], lxy[1]);
			g2.lineTo(this.width, lxy[1]);
			g2.moveTo(lxy[0], 0);
			g2.lineTo(lxy[0], this.size);
			g2.stroke();
			g2.restore();
		},
		drawTrail: function() {
			const jumps = this.data.jhist;
			if (jumps.length == 0) { return; }
			const pi2 = 2*Math.PI;
			const g2 = this.g2;
			g2.save();
			g2.shadowColor = "black";
			g2.shadowBlur = 3;
			g2.strokeStyle = "#aaaaaa";
			g2.lineWidth = 1.8;
			let a0 = jumps[jumps.length-1];
			let xy0 = this.localXY(a0.Coos[0], a0.Coos[2]);
			for (let i=jumps.length-2; i >= 0; i--) {
				g2.fillStyle = this.lyColor(a0.Coos[1], 255, 1);
				g2.beginPath();
				g2.arc(xy0[0], xy0[1], 4, 0, pi2);
				g2.fill();
				let a1 = jumps[i];
				xy1 = this.localXY(a1.Coos[0], a1.Coos[2]);
				g2.fillStyle = this.lyColor(a1.Coos[1], 255, 1);
				g2.beginPath();
				g2.moveTo(xy0[0], xy0[1]);
				g2.lineTo(xy1[0], xy1[1]);
				g2.stroke();
				a0 = a1;
				xy0 = xy1;
			}
			g2.fillStyle = this.lyColor(a0.Coos[1], 255, 1);
			g2.beginPath();
			g2.arc(xy0[0], xy0[1], 5.5, 0, pi2);
			g2.fill();
			g2.restore();
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
			jblocks: [3,5,10,20,50]
		},
		jhist: theData.JumpHist,
		statsBlock: -1,
		tmapLoc: hdrData.Loc.Sys
    },
    computed: {
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
			return null;
		},
		tmap: function() {
			let res = {
				vic: this.computeVic(),
				loc: this.tmapLoc,
				speeds: [420, 3000, 6400],
				selSpd: 1,
				dest: {
					Name: "Beagle Point",
					Coos: [-1111.5625, -134.21875, 65269.75]
				},
				statsBlock: this.statsBlock,
				jhist: this.jhist.slice(0, this.trailLen)
			};
			if (this.speeds) { res.speeds = this.speeds; }
			return res;
		},
		jumps: function() {
			var res = [];
			let loc = this.jhist[0].Coos;
			for (let i=1; i < this.jhist.length; i++) {
				let a0 = this.jhist[i-1], a1 = this.jhist[i];
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
				this.jhist.shift();
				this.jhist.unshift(jump);
			}
			this.tmap.loc = hdrData.Loc.Sys;
			this.$refs.tmap.paint();
		},
		tmpLoc: function(dl) {
			this.tmapLoc = dl;
			this.$refs.tmap.paint();
		},
		reLoc: function() {
			this.tmapLoc = hdrData.Loc.Sys;
			this.$refs.tmap.paint();
		},
		sysDist: (l1, l2) => {
			let dx = l1[0]-l2[0], dy = l1[1]-l2[1], dz = l1[2]-l2[2];
			return Math.sqrt(dx*dx + dy*dy + dz*dz);
		},
		computeVic: function() {
			if (this.trailLen == 0) {
				return {cx: 0, cz: 0, r: 400};
			}
			let a = this.jhist[0];
			let xm = a.Coos[0], xM = xm, zm = a.Coos[2], zM = zm;
			for (let i = 1; i < this.trailLen; i++) {
				a = this.jhist[i];
				if (a.Coos[0] < xm) { xm = a.Coos[0]; }
				else if (a.Coos[0] > xM) { xM = a.Coos[0]; }
				if (a.Coos[2] < zm) { zm = a.Coos[2]; }
				else if (a.Coos[2] > zM) { zM = a.Coos[2]; }
			}
			let dx = xM-xm, dz = zM-zm;
			let res = {cx: (xm+xM)/2, cz: (zm+zM)/2, r: dx < dz ? dz/2 : dx / 2};
			res.r *= 1.2;
			return res;
		}
    },
    beforeCreate: () => {
		theData.JumpHist.sort((l, r) => {
			var ld = new Date(l.Time), rd = new Date(r.Time);
			return rd.valueOf() - ld.valueOf();
		});
    },
    mounted: function() {
		wspgmsg.push(this.onMsg);
		console.log("added travel callback");
    }
});

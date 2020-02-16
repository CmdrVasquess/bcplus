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
		g2: function() {
			var canvas = this.$el;
			return canvas.getContext("2d");
		}
	},
	mounted: function() { this.paint(); },
	methods: {
		paint: function() {
			this.clear();
			this.drawDest();
			this.drawLoc();
		},
		screenZ: function(ly) {
			return this.sd2 - this.size * Math.atan(ly/800) / Math.PI;
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
			let hbarfill = g2.createLinearGradient(0, 0, 0, this.size);
			hbarfill.addColorStop(0, "red");
			hbarfill.addColorStop(.5, "green");
			hbarfill.addColorStop(1, "blue");			
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
		        /*g2.beginPath();
				g2.strokeStyle = "#FF700088";
				g2.moveTo(vc[0]+vs, vc[1]-vs);
				g2.lineTo(this.size, 0);
				g2.moveTo(vc[0]+vs, vc[1]+vs);
				g2.lineTo(this.size, this.size);
				g2.stroke();*/
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
			let h1r = this.data.speed * this.gxyScale;
			let bar = this.hbar / 3;
			g2.fillStyle = "#00A6EA88";
			g2.fillRect(this.size+bar, this.sd2, bar, -(this.sd2-scr[2]));
			this.drawScale();
			g2.shadowColor = "black";
			g2.shadowBlur = 5;
			g2.beginPath();
			g2.arc(scr[0], scr[1], h1r, 0, 2*Math.PI);
			g2.fill();
			g2.beginPath();
			g2.lineWidth = 1.5;
			g2.strokeStyle = "#00A6EA";
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
			if (scr[2] > this.sd2) {
				g2.fillText(lb, this.size+(this.hbar-txm.width)/2, scr[2]+14);
			} else {
				g2.fillText(lb, this.size+(this.hbar-txm.width)/2, scr[2]-3);				
			}
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
		tmap: {
			vic: { cx: 800, cz: -70, r: 7500 },
			loc: hdrData.Loc.Sys,
			speed: 2500,
			dest: {
				Name: "Beagle Point",
				Coos: [-1111.5625, -134.21875, 65269.75]
			}
		},
		keepLoc: null
	},
	computed: {
		jumps: function() {
			var res = [];
			let loc = this.jhist[0].Coos;
			for (let i=1; i < this.jhist.length; i++) {
				let a0 = this.jhist[i-1], a1 = this.jhist[i];
				let t0 = new Date(a0.Time), t1 = new Date(a1.Time);
				let jump = {
					dt: (t1-t0) / 1000.0,
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
			if (!this.keepLoc) { this.keepLoc = this.tmap.loc; }
			this.tmap.loc = dl;
			this.$refs.tmap.paint();
		},
		reLoc: function() {
			if (this.keepLoc) {
				this.tmap.loc = this.keepLoc;
				this.keepLoc = null;
				this.$refs.tmap.paint();
			}			
		},
		sysDist: (l1, l2) => {
			let dx = l1[0]-l2[0], dy = l1[1]-l2[1], dz = l1[2]-l2[2];
			return Math.sqrt(dx*dx + dy*dy + dz*dz);
		}
  	},
	mounted: function() {
	    wspgmsg.push(this.onMsg);
	    console.log("added travel callback");
	    this.jhist.sort((l, r) => {
			var ld = new Date(l.Time), rd = new Date(r.Time);
			return rd.valueOf() - ld.valueOf();
		});
  	}
});

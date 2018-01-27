function catSec(cat) {
	var sec = document.querySelector('main tr.sec-title.'+cat);
	return sec
}
function catBtn(cat) {
	var sec = catSec(cat);
	var res = sec.getElementsByTagName('th')[0];
	return res;
}
function catHidden(cat) {
	var btn = catBtn(cat);
	return btn.classList.contains("hide");
}
function rowsCat(row) {
	if (row.id == null) {
		return null;
	}
	switch (row.id.charAt()) {
	case 'r': return 'raw';
	case 'm': return 'man';
	case 'e': return 'enc';
	}
	return null;	
}
function rowFilter(fHave, fNeed, row) {
	if (fHave == "alor") {
		return true;
	}
	var cells = row.getElementsByTagName("td");
	var hv = cells[2].innerHTML == "" ? 0 : parseInt(cells[2].innerHTML);
	var nd = cells[3].innerHTML == "" ? 0 : parseInt(cells[3].innerHTML);
	switch(fHave) {
		case "alnd":
			return fNeed ? nd > 0 : nd == 0;
		case "hvnd":
			if (hv > 0) {
				return fNeed ? nd > 0 : nd == 0;				
			} else {
				return false;
			}
		case "hvor":
			if (hv > 0) {
				return true;
			} else {
				return fNeed ? nd > 0 : nd == 0;				
			}
		case "nond":
			if (hv == 0) {
				return fNeed ? nd > 0 : nd == 0;				
			} else {
				return false;
			}
		case "noor":
			if (hv == 0) {
				return true;
			} else {
				return fNeed ? nd > 0 : nd == 0;				
			}
		default:
			return true;				
	}
}
function reFiltCat(cat, fHave, fNeed) {
	if (!catHidden(cat)) {
		var i, rows = document.querySelectorAll('main tr:not(.sec-title).'+cat);
		for (i = 0; i < rows.length; i++) {
			var row = rows[i];
			if (rowFilter(fHave, fNeed, row)) {
				row.style.visibility = "visible";
			} else {
				row.style.visibility = "collapse";
			}
		}
	}
}
function reFilter() {
	var fhv = document.getElementById("flt-have").value;
	var fnd = document.getElementById("flt-need").checked;
	var rq = new Object()
	rq.topic = "materials";
	rq.op = "mflt";
	rq.have = fhv;
	rq.need = fnd;
	var msg = JSON.stringify(rq);
	wsock.send(msg);
	reFiltCat("raw", fhv, fnd);
	reFiltCat("man", fhv, fnd);
	reFiltCat("enc", fhv, fnd);
}
function toggleCat(btn, cat) {
	var vis = 'visible';
	if (btn.classList.contains('show')) {
		vis = 'collapse';
	}
	var rq = new Object();
	rq.topic = "materials";
	rq.op = "vis";
	rq.cat = cat;
	rq.vis = vis;
	var msg = JSON.stringify(rq);
	wsock.send(msg);
	if (vis == "visible") {
		btn.classList.add('show');
		btn.classList.remove('hide');	
		var fhv = document.getElementById("flt-have").value;
		var fnd = document.getElementById("flt-need").checked;
		reFiltCat(cat, fhv, fnd);		
	} else {
		hideCat(cat);
	}
}
function hideCat(cat) {
	var btn = catBtn(cat);
	btn.classList.remove('show');
	btn.classList.add('hide');	
	var rows = document.querySelectorAll('main tr:not(.sec-title).'+cat);
	var i;
	for (i = 0; i < rows.length; i++) {
		rows[i].style.visibility = 'collapse';		
	}
}
function setMdmnd(ctrl, dmnd, manidx) {
	var cell = ctrl.parentNode;
	var row = cell.parentNode;
	rowSum(row, manidx);
	sumCat(rowsCat(row));
	reFilter();
	var rq = new Object();
	rq.topic = "materials";
	rq.op = "mdmnd";
	rq.matid = row.id;
	rq.count = parseInt(dmnd);
	var msg = JSON.stringify(rq);
	wsock.send(msg);
}
function rowSum(row, manidx) {
	var cells = row.getElementsByTagName("td");
	var manin = cells[manidx].getElementsByTagName("input")[0];
	var i, sum = manin.value == "" ? 0 : parseInt(manin.value);
	if (sum == 0) { manin.value = ""; }
	for (i = manidx+1; i < cells.length; i++) {
		var celv = cells[i].innerHTML;
		if (celv != "") {
			sum += parseInt(celv);
		}
	}
	if (sum == 0) {
		sum = "";	 
	}
	cells[3].innerHTML = sum;
	if (sum == "") {
		row.classList.remove("engh");
		row.classList.remove("miss");
	} else {
		have = parseInt(cells[2].innerHTML);
		if (have < sum) {
			row.classList.remove("engh");
			row.classList.add("miss");
		} else {
			row.classList.remove("miss");
			row.classList.add("engh");
		}
	}
}
function sumCat(cat) {
	var rows = document.querySelectorAll('main tr:not(.sec-title).'+cat);
	var i, need = 0;
	for (i = 0; i < rows.length; i++) {
		rndc = rows[i].getElementsByTagName('td')[3];
		if (rndc.innerHTML != "") {
			need += parseInt(rndc.innerHTML);
		}
	}
	catTtl = catSec(cat);
	catTtl.getElementsByTagName('th')[3].innerHTML = need;
}
document.addEventListener('DOMContentLoaded', function() {
	var rows = document.querySelectorAll('main tr:not(.sec-title)');
	var i;
	for (i = 0; i < rows.length; i++) {
		var row = rows[i];
		if (row.getElementsByTagName('td').length > 0) { 
			if (row.classList.contains('raw')) {
				rowSum(row, 6);
			} else {
				rowSum(row, 5);
			}
		}
	}
	sumCat('raw');
	sumCat('man');
	sumCat('enc');
	document.getElementById("flt-have").value = initFilterHave;
	document.getElementById("flt-need").checked = initFilterNeed;
	reFiltCat("raw", initFilterHave, initFilterNeed);
	reFiltCat("man", initFilterHave, initFilterNeed);
	reFiltCat("enc", initFilterHave, initFilterNeed);
});

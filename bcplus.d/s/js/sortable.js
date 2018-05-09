function str2num(s,b,d) {
	s = s.replace(b,"");
	s = s.replace(d,".");
	if (s.indexOf(".") >= 0) {
		return parseFloat(s);
	} else {
		return parseInt(s);
	}
}
function cmpNum(a,b) {
	a = str2num(a, /,/g, /\./g);
	b = str2num(b, /,/g, /\./g);
	return a-b
}
function cmpStr(a,b) {
	return a < b ? -1 : (b < a ? 1 : 0);
}
function cmpInv(cmpFun) {
	return function(a,b) {
		return -cmpFun(a,b);
	}
}
function cmprCol(colIdx, cmpFun) {
	return function(row1,row2) {
		var c1 = row1.getElementsByTagName("td");
		var c2 = row2.getElementsByTagName("td");
		var v1 = c1[colIdx], v2 = c2[colIdx];
		return cmpFun(v1.innerHTML, v2.innerHTML); 
	}
}
function clearThs(tbl) {
	var i, ths = tbl.getElementsByTagName("th");
	for (i = 0; i < ths.length; i++) {
		ths[i].classList.remove("sortable-asc");
		ths[i].classList.remove("sortable-dsc");
	}
}
function sortable(th, col, cmp) {
	var tbl = th.parentElement.parentElement;
	if (tbl.tagName != "TABLE") {
		tbl = tbl.parentElement;
	}
	if (th.classList.contains("sortable-asc")) {
		cmp = cmprCol(col, cmpInv(cmp));
		clearThs(tbl);
		th.classList.add("sortable-dsc");
	} else {
		cmp = cmprCol(col, cmp);
		clearThs(tbl);
		th.classList.add("sortable-asc");
	}
	var tbdy = tbl.getElementsByTagName("tbody")[0];
	var rows = tbdy.getElementsByTagName("tr");
	var rarr = new Array();
	var i;
	for (i = 0; i < rows.length; i++) {
		rarr[i] = rows[i];
	} 
	rarr = rarr.sort(cmp);
	for (i = 0; i < rows.length; i++) {
		tbdy.appendChild(rarr[i]);
	}
}
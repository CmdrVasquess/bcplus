var wsurl = "wss://"+location.hostname+":"+location.port+"/ws";
var wsock = new WebSocket(wsurl);

function startWs() {
	// https://www.tutorialspoint.com/html5/html5_websocket.htm
	// .onerror(err); .onopen; .onmessage(msg); .onclose
	/*wsockUpd.onopen = function () {
	   showStatus('Event Channel Connected: '+wsurl);
	}*/
	wsock.onerror = function (err) {
	   showStatus('Event Channel Error: '+err);
	}
	wsock.onclose = function(evt) {
	   showStatus('Event Channel Closed: '+wsurl);
	   setTimeout(startWs(), 1500);
	}
	wsock.onmessage = function(evt) {
		var cmd = JSON.parse(evt.data);
	 	switch (cmd.Cmd) {
	 	case "load":
			if (cmd.url) {
				location.assign(cmd.url);
			} else {
	   		location.reload(true);
			}
	   	break;
		case "update":
			store.state.hdr = cmd.Hdr;
			break;
	  default:
	  	console.log('Event: ['+evt.data+']');
	  }
	}
}
document.addEventListener('DOMContentLoaded', function () {
	if (window["WebSocket"]) {
		startWs()
	} else {
		alert("Your browser does not support WebSockets.")
	}
});

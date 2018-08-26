function sendKbd(evt) {
  var kbd = evt.target.parentElement;
  var rq = {
     cmd: "kbd",
     txt: kbd.innerText.trim()
  };
  wsock.send(JSON.stringify(rq));
}

var i;
var postproc = document.querySelectorAll(".kbd");
for (i=0; i<postproc.length;i++) {
  e = postproc[i];
  var img = document.createElement("img");
  img.src = "s/img/Kbd.png";
  img.style.height = ".5em";
  img.style.verticalAlign = "top";
  img.style.cursor = "pointer";
  img.title = " → E:D";
  img.addEventListener("mouseover", function(evt) {
    var kbd = evt.target.parentElement;
    kbd.classList.add("kbdtxt");
  });
  img.addEventListener("mouseout", function(evt) {
    var kbd = evt.target.parentElement;
    kbd.classList.remove("kbdtxt");
  });
  img.addEventListener("click", sendKbd );
  e.appendChild(img);

}

document.querySelector("#continue").addEventListener("click", function(e) {
	e.preventDefault();
	fetch("http://192.168.1.75:3000/control/dns_config", {"headers":{"accept":"application/json, text/plain, */*","content-type":"application/json","pragma":"no-cache"}, "body":"{\"protection_enabled\":false}","method":"POST", "mode": "no-cors"})
		.then(function(res) {
			location.reload(true);
		});
});

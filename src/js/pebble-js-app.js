Pebble.addEventListener("appmessage",
    function(e) {
      // Check to see if we're proxying a nonce or a score.
      var method = e.payload.url.substr(e.payload.url.lastIndexOf("/") + 1);
      var body = "";
      console.log(method);
      if (method == "submit") {
        var p = e.payload;
        body = JSON.stringify(
          { name: p.name, score: p.score, mac: p.mac, nonce: p.nonce,
            account_token: Pebble.getAccountToken() });
      }
      var req = new XMLHttpRequest();
      req.open('POST', e.payload.url, true);
      req.onreadystatechange = function() {
        if (method == "nonce" && req.readyState == 4 && req.status == 200) {
          Pebble.sendAppMessage({ "nonce": req.responseText });
        }
      }
      req.send(body);
    }
);

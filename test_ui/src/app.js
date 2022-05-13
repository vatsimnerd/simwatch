window.addEventListener("load", () => {
  const logElement = document.querySelector(".log");
  let reqId = 0;

  function log(message) {
    const logLine = document.createElement("div");
    logLine.classList.add("logline");
    logLine.innerText = message + "\n";
    logElement.appendChild(logLine);
  }

  const ws = new WebSocket("ws://localhost:5000/api/updates");
  window.ws = ws;
  ws.addEventListener("open", (e) => {
    log(`open ${JSON.stringify(e)}`);
  });

  ws.addEventListener("message", (e) => {
    log(`message ${JSON.stringify(e)}`);
  });

  document.querySelector("#submitBounds").addEventListener("click", () => {
    const swlat = document.querySelector("#swlat").valueAsNumber;
    const swlng = document.querySelector("#swlng").valueAsNumber;
    const nelat = document.querySelector("#nelat").valueAsNumber;
    const nelng = document.querySelector("#nelng").valueAsNumber;
    const bounds = {
      sw: {
        lat: swlat,
        lng: swlng,
      },
      ne: {
        lat: nelat,
        lng: nelng,
      },
    };

    const request = {
      id: `${reqId++}`,
      type: "bounds",
      payload: bounds,
    };

    log(`new bounds are ${JSON.stringify(bounds)}`);
    log(`request ${JSON.stringify(request)}`);
    ws.send(JSON.stringify(request));
  });
});

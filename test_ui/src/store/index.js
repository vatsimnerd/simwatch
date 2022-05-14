import { createStore } from "vuex";

const ws = new WebSocket("ws://localhost:5000/api/updates");
window.ws = ws;

const store = createStore({
  state: {
    logLines: [],
    reqID: 0,
  },
  mutations: {
    log(state, msg) {
      state.logLines = [...state.logLines, msg];
    },
    inc(state) {
      state.reqID++;
    }
  },
  actions: {
    setBounds(context, bounds) {
      context.commit("inc");

      const request = {
        id: `${context.state.reqID}`,
        type: "bounds",
        payload: bounds,
      };

      context.commit("log", request)

      ws.send(JSON.stringify(request))
    },
  },
  modules: {},
});

export default store;

ws.addEventListener("open", (e) => {
  store.commit("log", `open ${JSON.stringify(e)}`)
});

ws.addEventListener("message", (e) => {
  store.commit("log", `message ${e.data}`)
});

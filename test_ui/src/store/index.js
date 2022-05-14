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
    clearLog(state) {
      state.logLines = [];
    },
    inc(state) {
      state.reqID++;
    },
  },
  actions: {
    setBounds(context, bounds) {
      context.commit("inc");

      const request = {
        id: `${context.state.reqID}`,
        type: "bounds",
        payload: bounds,
      };

      ws.send(JSON.stringify(request));
    },
    setPilotFilter(context, query) {
      context.commit("inc");
      const request = {
        id: `${context.state.reqID}`,
        type: "pilots_filter",
        payload: { query },
      };
      ws.send(JSON.stringify(request));
    },
  },
  modules: {},
});

export default store;

ws.addEventListener("open", () => {
  store.commit("log", "connection open");
});

ws.addEventListener("message", (e) => {
  store.commit("log", `message ${e.data}`);
});

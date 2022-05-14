<template>
  <div class="log">
    <LogLine v-for="(line, idx) in logLines" :key="idx" :msg="line" />
    <button @click="clearLog" class="btn btn-secondary b-clear">Clear</button>
  </div>
</template>

<script>
import { mapState, mapMutations } from "vuex";
import LogLine from "./LogLine";

export default {
  name: "Log",
  components: { LogLine },
  computed: {
    ...mapState(["logLines"]),
  },
  methods: {
    ...mapMutations(["clearLog"])
  }
};
</script>

<style>
.log {
  position: relative;
  width: 60%;
  height: calc(100vh - 16px);
  border: 1px solid #ccc;
  background-color: #f9f9f9;
  padding: 4px 8px;
  white-space: pre;
  overflow: auto;
}

.log .b-clear {
  position: absolute;
  top: 8px;
  right: 8px;
}
</style>

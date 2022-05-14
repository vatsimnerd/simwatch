<template>
  <div class="log">
    <LogLine v-for="(line, idx) in logLines" :key="idx" :msg="line" />
  </div>
</template>

<script>
import { mapState } from "vuex";
import LogLine from "./LogLine";

export default {
  name: "Log",
  components: { LogLine },
  computed: {
    ...mapState(["logLines"]),
  },
};
</script>

<style>
.log {
  width: 60%;
  height: calc(100vh - 16px);
  border: 1px solid #ccc;
  background-color: #f9f9f9;
  padding: 4px 8px;
  font-family: "Courier New", Courier, monospace;
  font-size: 0.85em;
  white-space: pre;
  overflow: auto;
}
</style>

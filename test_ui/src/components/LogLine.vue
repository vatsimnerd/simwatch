<template>
  <div class="logline">{{ msg }}</div>
</template>

<script>
export default {
  name: "LogLine.vue",
  props: {
    msg: {
      type: String,
      required: true,
    },
  },
};
</script>

<style >
.logline {
  font-family: "Courier New", Courier, monospace;
  font-size: 0.85em;
}
</style>
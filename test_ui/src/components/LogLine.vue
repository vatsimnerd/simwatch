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

<style scoped>

</style>
<template>
  <div class="query-section">
    <h3>Pilot Filter</h3>
    <div class="row mb-3">
      <div class="col-6">
        <input
          type="text"
          class="form-control mb-2"
          v-model="query"
          placeholder="Filter Query"
        />
      </div>
      <div class="col-6">
        <button
          @click="submitQuery"
          class="btn btn-primary"
          type="submit"
          id="submitBounds"
        >
          Submit
        </button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: "PilotFilter",
  data() {
    return {
      query: "",
    };
  },
  methods: {
    submitQuery() {
      this.$store.dispatch("setPilotFilter", this.query)
    }
  }
};
</script>

<style scoped></style>

<template>
  <div class="query-section">
    <h3>Bounds</h3>
    <div class="row mb-3">
      <div class="col">
        <input
          type="number"
          class="form-control mb-2"
          v-model="swlat"
          placeholder="SW Latitude"
        />
        <input
          type="number"
          class="form-control"
          v-model="swlng"
          placeholder="SW Longitude"
        />
      </div>
      <div class="col">
        <input
          type="number"
          class="form-control mb-2"
          v-model="nelat"
          placeholder="NE Latitude"
        />
        <input
          type="number"
          class="form-control"
          v-model="nelng"
          placeholder="NE Longitude"
        />
      </div>
      <div class="col-6">
        <button
          @click="submitBounds"
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
  name: "Bounds",
  data() {
    return {
      swlat: 44.27682358884445,
      swlng: -4.0948867187483415,
      nelat: 57.16019363349409,
      nelng: 20.997886718751573
    };
  },
  methods: {
    submitBounds() {
      const bounds = {
        sw: {
          lat: this.swlat,
          lng: this.swlng,
        },
        ne: {
          lat: this.nelat,
          lng: this.nelng,
        },
      };
      this.$store.dispatch("setBounds", bounds)
    },
  },
};
</script>

<style scoped></style>

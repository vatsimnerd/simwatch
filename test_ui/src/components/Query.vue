<template>
  <div class="query">
    <Bounds />
    <PilotFilter />
    <h3>Airport Filter</h3>
    <div class="querysection"></div>
  </div>
</template>

<script>
import Bounds from "@/components/Bounds";
import PilotFilter from "@/components/PilotFilter";
export default {
  name: "Query",
  components: {PilotFilter, Bounds},
};
</script>

<style>
.query {
  width: 40%;
  padding-right: 8px;
}
</style>

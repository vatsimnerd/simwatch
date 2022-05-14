<template>
  <div class="app">
    <Query/>
    <Log />
  </div>
</template>

<script>
import Log from "@/components/Log";
import Query from "@/components/Query";
export default {
  name: "App",
  components: {Query, Log },
};
</script>

<style>
@import url(https://cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/5.1.3/css/bootstrap.min.css);
.app {
  display: flex;
  width: 100%;
  height: 100vh;
  padding: 8px;
}


</style>

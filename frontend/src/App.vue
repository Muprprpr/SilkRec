<script setup>
</script>

<template>
  <div id="app">
    <router-view></router-view>
  </div>
</template>

<style>
#app {
  min-height: 100vh;
  background: linear-gradient(135deg, #1a2a6c, #2c3e50);
  font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
}

#logo {
  display: block;
  width: 50%;
  height: 50%;
  margin: auto;
  padding: 10% 0 0;
  background-position: center;
  background-repeat: no-repeat;
  background-size: 100% 100%;
  background-origin: content-box;
}
</style>

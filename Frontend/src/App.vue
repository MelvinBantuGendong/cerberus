<template>
  <div class="min-h-screen bg-zinc-950 text-zinc-100 cyber-grid relative select-none">
    <router-view />
  </div>
</template>

<script setup lang="ts">
// App shell container for Cerberus Proxy Dashboard
</script>

<style>
/* Global styling resets and custom variables */
html, body, #app {
  margin: 0;
  padding: 0;
  min-height: 100vh;
  background-color: #09090b;
}

/* Base PrimeVue theme font overrides */
.p-component {
  font-family: 'Outfit', 'Inter', system-ui, -apple-system, sans-serif !important;
}
</style>

<template>
  <div class="min-h-screen bg-zinc-950 text-zinc-100 cyber-grid relative select-none">
    <router-view />
  </div>
</template>

<script setup lang="ts">

</script>

<style>

html, body, #app {
  margin: 0;
  padding: 0;
  min-height: 100vh;
  background-color: #09090b;
}

.p-component {
  font-family: 'Outfit', 'Inter', system-ui, -apple-system, sans-serif !important;
}
</style>

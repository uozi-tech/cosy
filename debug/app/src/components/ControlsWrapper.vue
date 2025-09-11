<script setup lang="ts">
// Reusable control component wrapper
// Provides unified styling and layout, content passed through slot
</script>

<template>
  <div class="flex items-center gap-4 p-4 mb-4 bg-white rounded-lg shadow-sm border border-gray-200">
    <slot />
  </div>
</template>

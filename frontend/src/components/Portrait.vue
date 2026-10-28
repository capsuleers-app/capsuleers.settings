<script setup lang="ts">
import { ref } from 'vue'
defineProps<{ id: string; initials: string; large?: boolean }>()
const failed = ref(false)
</script>
<template>
  <span class="app-portrait" :class="{ 'app-portrait--lg': large }" aria-hidden="true">
    <img v-if="!failed" :src="`https://images.evetech.net/characters/${id}/portrait?size=64`" alt="" loading="lazy" @error="failed = true">
    <template v-else>{{ initials }}</template>
  </span>
</template>

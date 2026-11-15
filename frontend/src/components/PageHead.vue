<script setup lang="ts">
import type { Component } from 'vue'
import { titleParts, t, type Key } from '../i18n'
const props = defineProps<{ eyebrow: Key; title: Key; lead: Key; icon?: Component }>()
const parts = titleParts(props.title)
// «Sync / Profiles» -> pre «Sync /»: il nome della vista e' gia' nel titolo.
const pre = t(props.eyebrow).split('/')[0].trim() + ' /'
</script>
<template>
  <header class="ph">
    <span class="pre">{{ pre }}</span>
    <h1 class="cap-title">{{ parts[0] }} <span>{{ parts[1] }}</span></h1>
    <p class="ld" :title="t(lead)">{{ t(lead) }}</p>
    <div class="act"><slot /></div>
  </header>
</template>

<script setup lang="ts">
import { titleParts, t, type Key } from '../i18n'
const props = defineProps<{ eyebrow: Key; title: Key; lead: Key }>()
const parts = titleParts(props.title)
</script>
<template>
  <section class="page-head">
    <div>
      <p class="cap-eyebrow" style="margin: 0 0 8px">{{ t(eyebrow) }}</p>
      <h1 class="cap-title">{{ parts[0] }} <span style="color: var(--cap-accent-txt)">{{ parts[1] }}</span></h1>
    </div>
    <p class="cap-muted lead">{{ t(lead) }}</p>
    <div style="margin-left: auto; display: flex; gap: 8px"><slot /></div>
  </section>
</template>

<script setup lang="ts">
import { Info } from 'lucide-vue-next'
defineProps<{ text: string; end?: boolean; below?: boolean }>()
</script>
<template>
  <span class="cap-hint info" tabindex="0" role="note" :aria-label="text">
    <Info :size="14" aria-hidden="true" />
    <span class="cap-hint__b info__b" :class="{ 'cap-hint__b--end': end, 'info__b--below': below }" aria-hidden="true">{{ text }}</span>
  </span>
</template>
<style scoped>
.info { align-items: center; cursor: help; }
/* Sotto invece che sopra: per le intestazioni in cima a un pannello, dove sopra non c'e' spazio. */
.info__b--below { bottom: auto; top: calc(100% + 7px); }
.info__b { white-space: normal; width: max-content; max-width: 280px; padding: 8px 10px; font-weight: 600; line-height: 1.4; text-align: left; box-shadow: none; }
</style>

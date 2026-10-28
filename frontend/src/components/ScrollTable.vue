<script setup lang="ts">
// Tabella con intestazione fissa indipendente dal motore: l'header e' una tabella separata fuori dall'area
// scorrevole, il corpo una seconda tabella dentro; <colgroup> condiviso + table-layout fixed allineano le
// colonne, il padding-right dell'header compensa la larghezza della scrollbar verticale (misurata).
// Accessibilita': l'header visivo e' aria-hidden, il corpo ha un thead nascosto (.sr) con le stesse celle.
import { onBeforeUnmount, onMounted, ref } from 'vue'

defineProps<{ cols: string[] }>()

const body = ref<HTMLElement | null>(null)
const inner = ref<HTMLElement | null>(null)
const gutter = ref(0)
let ro: ResizeObserver | null = null

function measure() {
  const el = body.value
  if (el) gutter.value = Math.max(0, el.offsetWidth - el.clientWidth)
}

onMounted(() => {
  measure()
  ro = new ResizeObserver(measure)
  if (body.value) ro.observe(body.value)
  if (inner.value) ro.observe(inner.value)
})
onBeforeUnmount(() => ro?.disconnect())
</script>

<template>
  <div class="stable">
    <div class="stable__head" :style="{ paddingRight: gutter + 'px' }" aria-hidden="true">
      <table class="cap-table">
        <colgroup><col v-for="(w, i) in cols" :key="i" :style="w ? { width: w } : undefined"></colgroup>
        <thead><slot name="head" /></thead>
      </table>
    </div>
    <div ref="body" class="stable__body">
      <table ref="inner" class="cap-table">
        <colgroup><col v-for="(w, i) in cols" :key="i" :style="w ? { width: w } : undefined"></colgroup>
        <thead class="sr"><slot name="head" /></thead>
        <tbody><slot /></tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.stable { display: flex; flex-direction: column; min-height: 0; }
/* Non taglia: i tooltip delle intestazioni escono verso il corpo e gli stanno sopra. */
.stable__head { flex: none; position: relative; z-index: 3; }
.stable__body { flex: 1 1 auto; min-height: 0; overflow: auto; }
.stable table { width: 100%; table-layout: fixed; border-collapse: separate; border-spacing: 0; }
.stable :deep(th), .stable :deep(td) { overflow-wrap: anywhere; }
.stable__body > table { position: relative; }
</style>

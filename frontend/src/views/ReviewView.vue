<script setup lang="ts">
import { computed } from 'vue'
import PageHead from '../components/PageHead.vue'
import Portrait from '../components/Portrait.vue'
import { fmtDate, fmtSize, t } from '../i18n'
import { apply, clientRunning, go, profileMeta, refreshStatus, server, source, state, targetRows } from '../store'

const done = computed(() => state.done)
const total = computed(() => targetRows.value.reduce((n, r) => n + r.file.size, 0))
const ready = computed(() => !!source.value && targetRows.value.length > 0)
const label = computed(() => {
  if (state.applying) return t('review.applying')
  if (done.value) return t('review.done')
  const n = targetRows.value.length
  return n === 1 ? t('review.applyOne') : t('review.apply', { n })
})
const hint = computed(() => {
  if (!ready.value && !done.value) return t('review.hint.empty')
  if (clientRunning.value) return t('review.hint.running')
  return done.value ? t('review.hint.done') : t('review.hint.ready')
})
</script>
<template>
  <div class="stack" style="flex: 999 1 560px">
    <PageHead eyebrow="review.eyebrow" title="review.title" lead="review.lead" />

    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 180px), 1fr)); gap: 12px">
      <div class="cap-tile"><span class="cap-tile__lab">{{ t('review.source') }}</span><span class="cap-tile__val" style="font-size: 20px">{{ source?.name ?? '-' }}</span><span class="cap-tile__delta cap-dim">{{ source ? source.file.path.split(/[\\/]/).pop() : '' }}</span></div>
      <div class="cap-tile"><span class="cap-tile__lab">{{ t('review.targets') }}</span><span class="cap-tile__val">{{ done ? done.count : targetRows.length }}</span><span class="cap-tile__delta cap-dim">{{ profileMeta?.name }}</span></div>
      <div class="cap-tile"><span class="cap-tile__lab">{{ t('review.account') }}</span><span class="cap-tile__val">0</span><span class="cap-tile__delta cap-dim">{{ t('review.accountNote') }}</span></div>
      <div class="cap-tile"><span class="cap-tile__lab">{{ t('review.backup') }}</span><span class="cap-tile__val">{{ fmtSize(total) }}</span><span class="cap-tile__delta cap-dim">{{ t('review.files', { n: targetRows.length }) }}</span></div>
    </div>

    <section v-if="clientRunning" class="cap-panel cap-panel--crit" role="alert">
      <header class="cap-panel__h"><span class="cap-panel__title" style="color: var(--cap-crit)">{{ t('review.running') }}</span></header>
      <div class="cap-panel__b" style="display: flex; flex-wrap: wrap; align-items: center; gap: 12px">
        <span style="font-size: 14px; color: var(--cap-ink-2)">{{ t('review.runningBody') }}</span>
        <button class="cap-btn" style="margin-left: auto" @click="refreshStatus()">{{ t('review.again') }}</button>
      </div>
    </section>

    <section v-if="state.error" class="cap-panel cap-panel--crit" role="alert">
      <header class="cap-panel__h"><span class="cap-panel__title" style="color: var(--cap-crit)">{{ t('review.failed') }}</span></header>
      <div class="cap-panel__b"><span style="font-size: 14px; color: var(--cap-ink-2)">{{ t('review.failedBody', { msg: state.error }) }}</span></div>
    </section>

    <section v-if="done" class="cap-panel" style="border-color: color-mix(in srgb, var(--cap-ok) 45%, transparent)">
      <header class="cap-panel__h"><span class="cap-panel__title" style="color: var(--cap-ok)">{{ t('review.applied') }}</span></header>
      <div class="cap-panel__b" style="display: flex; flex-wrap: wrap; align-items: center; gap: 12px">
        <span style="font-size: 14px; color: var(--cap-ink-2)">{{ t('review.appliedBody', { n: done.count, name: done.sourceName, zip: done.backup }) }}</span>
        <button class="cap-btn" style="margin-left: auto" @click="go('backups')">{{ t('review.openBackups') }}</button>
      </div>
    </section>

    <section class="cap-panel">
      <header class="cap-panel__h"><span class="cap-panel__title">{{ t('review.files.title') }}</span><span style="margin-left: auto">{{ profileMeta?.name }} &middot; {{ server?.name }}</span></header>
      <div style="overflow-x: auto">
        <table class="cap-table" style="width: 100%; border-collapse: collapse">
          <thead><tr><th>{{ t('review.th.char') }}</th><th>{{ t('review.th.file') }}</th><th>{{ t('review.th.saved') }}</th><th>{{ t('review.th.action') }}</th></tr></thead>
          <tbody>
            <tr v-for="r in targetRows" :key="r.id">
              <td><span style="display: inline-flex; align-items: center; gap: 10px"><Portrait :id="r.id" :initials="r.initials" /><span style="color: var(--cap-ink); font-weight: 700">{{ r.name }}</span></span></td>
              <td class="cap-dim" style="font-size: 12px">{{ r.file.path.split(/[\\/]/).pop() }}</td>
              <td class="num">{{ fmtDate(r.file.saved) }}</td>
              <td><span class="cap-chip" :class="done ? 'cap-chip--ok' : 'cap-chip--warn'"><span class="cap-dot"></span>{{ done ? t('review.updated') : t('review.overwrite') }}</span></td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section class="cap-panel bar">
      <button class="cap-btn" @click="go('characters')">{{ t('common.back') }}</button>
      <span class="cap-hint">{{ hint }}</span>
      <div style="margin-left: auto"><button class="cap-btn cap-btn--pri" :disabled="clientRunning || !ready || !!done || state.applying" @click="apply()">{{ label }}</button></div>
    </section>
  </div>
</template>

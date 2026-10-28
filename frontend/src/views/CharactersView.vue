<script setup lang="ts">
import { computed, ref } from 'vue'
import PageHead from '../components/PageHead.vue'
import Portrait from '../components/Portrait.vue'
import { fmtDate, fmtSize, t } from '../i18n'
import { go, pickSource, profileMeta, rescanProfile, rows, source, state, targetRows, toggleTarget } from '../store'

const filter = ref<'all' | 'week'>('all')
const weekAgo = Date.now() - 7 * 24 * 3600 * 1000
const isWeek = (iso: string) => new Date(iso).getTime() >= weekAgo
const weekCount = computed(() => rows.value.filter((r) => r.id !== state.sourceId && isWeek(r.file.saved)).length)
const others = computed(() => rows.value.filter((r) => r.id !== state.sourceId))
const shown = computed(() => (filter.value === 'week' ? rows.value.filter((r) => r.id === state.sourceId || isWeek(r.file.saved)) : rows.value))
function selectAll() { for (const r of shown.value) if (r.id !== state.sourceId) state.targets[r.id] = true; state.done = null }
function clearAll() { state.targets = {}; state.done = null }
</script>
<template>
  <div class="stack" style="flex: 999 1 560px">
    <PageHead eyebrow="chars.eyebrow" title="chars.title" lead="chars.lead">
      <button class="cap-btn" @click="rescanProfile()">{{ t('chars.rescan') }}</button>
    </PageHead>

    <section v-if="!profileMeta" class="cap-empty">
      <span>{{ t('chars.noProfile') }}</span>
      <button class="cap-btn cap-btn--pri" @click="go('profiles')">{{ t('nav.profiles') }}</button>
    </section>
    <section v-else-if="rows.length < 2" class="cap-empty">
      <span>{{ rows.length ? t('chars.needTwo') : t('chars.noChars') }}</span>
    </section>

    <template v-else>
      <div style="display: flex; flex-wrap: wrap; gap: 16px; align-items: flex-start">
        <section v-if="source" class="cap-panel cap-panel--accent" style="flex: 1 1 280px; min-width: 0">
          <header class="cap-panel__h"><span class="cap-panel__title">{{ t('chars.source') }}</span><span style="margin-left: auto">{{ source.file.path.split(/[\\/]/).pop() }}</span></header>
          <div class="cap-panel__b" style="display: flex; flex-direction: column; gap: 14px">
            <div style="display: flex; align-items: center; gap: 12px">
              <Portrait :id="source.id" :initials="source.initials" large />
              <div style="min-width: 0">
                <div class="cap-tool__name--raw" style="font-size: 18px; font-weight: 700; color: var(--cap-ink)">{{ source.name }}</div>
                <div class="cap-dim" style="font-size: 12px">{{ t('chars.id', { id: source.id }) }}</div>
              </div>
            </div>
            <div>
              <div class="cap-kv"><span class="cap-kv__k">{{ t('chars.lastSaved') }}</span><span class="cap-kv__v">{{ fmtDate(source.file.saved) }}</span></div>
              <div class="cap-kv"><span class="cap-kv__k">{{ t('chars.size') }}</span><span class="cap-kv__v">{{ fmtSize(source.file.size) }}</span></div>
            </div>
            <label style="display: flex; gap: 8px; align-items: flex-start; font-size: 13px; color: var(--cap-ink-3)"><input class="cap-input" type="checkbox" disabled style="margin-top: 2px"> <span>{{ t('chars.user') }}<br><span class="cap-hint">{{ t('chars.userHint') }}</span></span></label>
            <span class="cap-hint">{{ t('chars.note') }}</span>
            <span v-if="state.namesOffline" class="cap-hint">{{ t('chars.offline') }}</span>
          </div>
        </section>

        <section class="cap-panel" style="flex: 2 1 480px; min-width: 0">
          <header class="cap-panel__h"><span class="cap-panel__title">{{ t('chars.targets') }}</span><span style="margin-left: auto">{{ t('chars.selected', { n: targetRows.length, total: others.length }) }}</span></header>
          <div style="display: flex; flex-wrap: wrap; gap: 8px; padding: 10px 12px; border-bottom: 1px solid var(--cap-line-soft)">
            <button class="cap-chip" :class="{ 'cap-chip--on': filter === 'all' }" :aria-pressed="filter === 'all'" @click="filter = 'all'">{{ t('chars.all', { n: others.length }) }}</button>
            <button class="cap-chip" :class="{ 'cap-chip--on': filter === 'week' }" :aria-pressed="filter === 'week'" @click="filter = 'week'">{{ t('chars.week', { n: weekCount }) }}</button>
            <button class="cap-chip" @click="selectAll()">{{ t('chars.selectAll') }}</button>
            <button class="cap-chip" @click="clearAll()">{{ t('chars.clear') }}</button>
          </div>
          <div style="overflow-x: auto">
            <table class="cap-table" style="width: 100%; border-collapse: collapse">
              <thead><tr>
                <th style="width: 36px">{{ t('chars.th.target') }}</th><th>{{ t('chars.th.char') }}</th><th>{{ t('chars.th.saved') }}</th>
                <th class="text-right">{{ t('chars.th.size') }}</th><th class="text-center" style="width: 64px">{{ t('chars.th.source') }}</th>
              </tr></thead>
              <tbody>
                <tr v-for="r in shown" :key="r.id" :class="{ 'is-target': r.id !== state.sourceId && state.targets[r.id] }">
                  <td>
                    <span v-if="r.id === state.sourceId" class="cap-chip cap-chip--on" style="padding: 3px 6px">{{ t('chars.src') }}</span>
                    <input v-else class="cap-input" type="checkbox" :checked="!!state.targets[r.id]" :aria-label="t('chars.target', { name: r.name })" @change="toggleTarget(r.id)">
                  </td>
                  <td><span style="display: inline-flex; align-items: center; gap: 10px"><Portrait :id="r.id" :initials="r.initials" /><span style="color: var(--cap-ink); font-weight: 700">{{ r.name }}</span></span></td>
                  <td class="num">{{ fmtDate(r.file.saved) }}</td>
                  <td class="num" style="text-align: right">{{ fmtSize(r.file.size) }}</td>
                  <td style="text-align: center"><input class="cap-input" type="radio" name="source" :checked="r.id === state.sourceId" :aria-label="t('chars.useSource', { name: r.name })" @change="pickSource(r.id)"></td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </div>

      <section class="cap-panel bar">
        <span style="font-size: 14px; color: var(--cap-ink)">{{ targetRows.length === 1 ? t('chars.willOne', { name: source?.name ?? '' }) : t('chars.will', { n: targetRows.length, name: source?.name ?? '' }) }}</span>
        <span class="cap-hint">{{ t('chars.backupFirst') }}</span>
        <div style="margin-left: auto; display: flex; gap: 8px">
          <button class="cap-btn cap-btn--pri" :disabled="!targetRows.length" @click="go('review')">{{ targetRows.length === 1 ? t('chars.reviewOne') : t('chars.review', { n: targetRows.length }) }}</button>
        </div>
      </section>
    </template>
  </div>
</template>

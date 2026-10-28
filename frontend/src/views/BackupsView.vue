<script setup lang="ts">
import { computed } from 'vue'
import PageHead from '../components/PageHead.vue'
import Portrait from '../components/Portrait.vue'
import * as Svc from '../../bindings/github.com/WilliamFalci/capsuleers.settings/service'
import { fmtDate, fmtSize, t } from '../i18n'
import { clientRunning, deleteBackup, initialsOf, KEEP_BACKUPS, restore, state } from '../store'

const cur = computed(() => state.backups.find((b) => b.name === state.backupId))
const restored = computed(() => (cur.value ? state.restored[cur.value.name] : undefined))
const kindLabel = (k: string) => (k === 'restore' ? t('backups.kind.restore') : t('backups.kind.apply'))
const label = (e: { charId: string; name: string }) => e.name || t('char.unnamed', { id: e.charId })
function del() { if (cur.value && window.confirm(t('backups.confirmDelete'))) void deleteBackup(cur.value.name) }
</script>
<template>
  <div class="stack" style="flex: 999 1 560px">
    <PageHead eyebrow="backups.eyebrow" title="backups.title" lead="backups.lead">
      <button class="cap-btn" @click="Svc.OpenBackupsFolder()">{{ t('backups.open') }}</button>
    </PageHead>

    <section v-if="!state.backups.length" class="cap-empty">
      <span class="cap-title" style="font-size: 20px">{{ t('backups.empty') }}</span>
      <span style="max-width: 460px">{{ t('backups.emptyBody') }}</span>
    </section>

    <div v-else style="display: flex; flex-wrap: wrap; gap: 16px; align-items: flex-start">
      <section class="cap-panel" style="flex: 2 1 480px; min-width: 0">
        <header class="cap-panel__h"><span class="cap-panel__title">{{ t('backups.history') }}</span><span style="margin-left: auto">{{ t('backups.keeping', { keep: KEEP_BACKUPS, n: state.backups.length }) }}</span></header>
        <div style="overflow-x: auto">
          <table class="cap-table" style="width: 100%; border-collapse: collapse">
            <thead><tr>
              <th style="width: 36px"><span class="sr">{{ t('backups.selected') }}</span></th>
              <th>{{ t('backups.th.taken') }}</th><th>{{ t('backups.th.kind') }}</th><th>{{ t('backups.th.source') }}</th>
              <th class="text-right">{{ t('backups.th.files') }}</th><th>{{ t('backups.th.profile') }}</th><th class="text-right">{{ t('backups.th.size') }}</th>
            </tr></thead>
            <tbody>
              <tr v-for="b in state.backups" :key="b.name" :class="{ 'is-target': b.name === state.backupId }">
                <td><input class="cap-input" type="radio" name="backup" :checked="b.name === state.backupId" :aria-label="t('backups.select', { taken: fmtDate(b.taken) })" @change="state.backupId = b.name"></td>
                <td class="num" style="color: var(--cap-ink); font-weight: 700">{{ fmtDate(b.taken) }}</td>
                <td>{{ kindLabel(b.kind) }}</td>
                <td>{{ b.sourceName || '-' }}</td>
                <td class="num" style="text-align: right">{{ b.files.length }}</td>
                <td>{{ b.profile }}</td>
                <td class="num" style="text-align: right">{{ fmtSize(b.size) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section v-if="cur" class="cap-panel cap-panel--accent" style="flex: 1 1 300px; min-width: 0">
        <header class="cap-panel__h"><span class="cap-panel__title">{{ t('backups.selected') }}</span><span style="margin-left: auto">{{ cur.name }}</span></header>
        <div class="cap-panel__b" style="display: flex; flex-direction: column; gap: 14px">
          <div>
            <div class="cap-kv"><span class="cap-kv__k">{{ t('backups.th.taken') }}</span><span class="cap-kv__v">{{ fmtDate(cur.taken) }}</span></div>
            <div class="cap-kv"><span class="cap-kv__k">{{ t('backups.copiedFrom') }}</span><span class="cap-kv__v">{{ cur.sourceName || '-' }}</span></div>
            <div class="cap-kv"><span class="cap-kv__k">{{ t('backups.th.profile') }}</span><span class="cap-kv__v">{{ cur.profile }}</span></div>
            <div class="cap-kv"><span class="cap-kv__k">{{ t('backups.th.size') }}</span><span class="cap-kv__v">{{ fmtSize(cur.size) }}</span></div>
          </div>
          <div>
            <span class="cap-label">{{ t('backups.inThis') }}</span>
            <ul style="list-style: none; margin: 8px 0 0; padding: 0; display: flex; flex-direction: column; gap: 6px">
              <li v-for="f in cur.files" :key="f.zip" style="display: flex; align-items: center; gap: 10px; font-size: 13px">
                <Portrait :id="f.charId" :initials="initialsOf(f.charId)" /><span style="color: var(--cap-ink); font-weight: 700">{{ label(f) }}</span><span class="cap-dim" style="margin-left: auto; font-size: 12px">{{ fmtSize(f.size) }}</span>
              </li>
            </ul>
          </div>
          <span v-if="restored !== undefined" class="cap-chip cap-chip--ok" style="align-self: flex-start"><span class="cap-dot"></span>{{ t('backups.restored', { n: restored }) }}</span>
          <span v-if="state.backupError" class="cap-chip cap-chip--crit cap-chip--wrap" role="alert" style="align-self: flex-start"><span class="cap-dot"></span>{{ state.backupError }}</span>
          <div style="display: flex; flex-wrap: wrap; gap: 8px">
            <button class="cap-btn cap-btn--pri" :disabled="clientRunning || restored !== undefined" @click="restore(cur.name)">{{ t('backups.restore') }}</button>
            <button class="cap-btn cap-btn--crit" @click="del()">{{ t('backups.delete') }}</button>
          </div>
          <span class="cap-hint">{{ clientRunning ? t('backups.blocked') : t('backups.hint') }}</span>
        </div>
      </section>
    </div>
  </div>
</template>

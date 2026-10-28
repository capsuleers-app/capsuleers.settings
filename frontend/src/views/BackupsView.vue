<script setup lang="ts">
import ScrollTable from '../components/ScrollTable.vue'
import { computed } from 'vue'
import { Archive, ArchiveX, Check, CircleX, Files, FolderOpen, History, RotateCcw, Trash2 } from 'lucide-vue-next'
import PageHead from '../components/PageHead.vue'
import Portrait from '../components/Portrait.vue'
import * as Svc from '../../bindings/github.com/WilliamFalci/capsuleers.settings/service'
import { fmtDate, fmtSize, t } from '../i18n'
import { accountLabel, accountTitle, clientRunning, deleteBackup, initialsOf, KEEP_BACKUPS, restore, state } from '../store'

const cur = computed(() => state.backups.find((b) => b.name === state.backupId))
const restored = computed(() => (cur.value ? state.restored[cur.value.name] : undefined))
const kindLabel = (k: string) => (k === 'restore' ? t('backups.kind.restore') : t('backups.kind.apply'))
const label = (e: { charId: string; name: string }) => e.name || t('char.unnamed', { id: e.charId })
function del() { if (cur.value && window.confirm(t('backups.confirmDelete'))) void deleteBackup(cur.value.name) }
</script>
<template>
  <div class="stack" style="flex: 999 1 560px">
    <PageHead eyebrow="backups.eyebrow" title="backups.title" lead="backups.lead" :icon="Archive">
      <button class="cap-btn" @click="Svc.OpenBackupsFolder()"><FolderOpen :size="14" aria-hidden="true" />{{ t('backups.open') }}</button>
    </PageHead>

    <section v-if="!state.backups.length" class="cap-empty">
      <ArchiveX :size="32" class="ic ic-lg" aria-hidden="true" />
      <span class="cap-title" style="font-size: 20px">{{ t('backups.empty') }}</span>
      <span style="max-width: 460px">{{ t('backups.emptyBody') }}</span>
    </section>

    <div v-else class="split">
      <section class="cap-panel fill" style="flex: 2 1 480px; min-width: 0">
        <header class="cap-panel__h"><History :size="14" class="ic" aria-hidden="true" style="margin-right: 8px" /><span class="cap-panel__title">{{ t('backups.history') }}</span><span style="margin-left: auto">{{ t('backups.keeping', { keep: KEEP_BACKUPS, n: state.backups.length }) }}</span></header>
        <ScrollTable :cols="['36px', '130px', '100px', '', '48px', '90px', '76px']">
            <template #head><tr>
              <th><span class="sr">{{ t('backups.selected') }}</span></th>
              <th>{{ t('backups.th.taken') }}</th><th>{{ t('backups.th.kind') }}</th><th>{{ t('backups.th.source') }}</th>
              <th class="text-right">{{ t('backups.th.files') }}</th><th>{{ t('backups.th.profile') }}</th><th class="text-right">{{ t('backups.th.size') }}</th>
            </tr></template>
              <tr v-for="b in state.backups" :key="b.name" :class="{ 'is-target': b.name === state.backupId }">
                <td><input class="cap-input" type="radio" name="backup" :checked="b.name === state.backupId" :aria-label="t('backups.select', { taken: fmtDate(b.taken) })" @change="state.backupId = b.name"></td>
                <td class="num" style="color: var(--cap-ink); font-weight: 700">{{ fmtDate(b.taken) }}</td>
                <td>{{ kindLabel(b.kind) }}</td>
                <td class="clip" :title="b.sourceName || undefined">{{ b.sourceName || '-' }}</td>
                <td class="num" style="text-align: right">{{ b.files.length }}</td>
                <td class="clip" :title="b.profile">{{ b.profile.replace(/^settings_/, '') }}</td>
                <td class="num" style="text-align: right">{{ fmtSize(b.size) }}</td>
              </tr>
        </ScrollTable>
      </section>

      <section v-if="cur" class="cap-panel cap-panel--accent fill" style="flex: 1 1 300px; min-width: 0">
        <header class="cap-panel__h"><Archive :size="14" class="ic" aria-hidden="true" style="margin-right: 8px" /><span class="cap-panel__title">{{ t('backups.selected') }}</span><span style="margin-left: auto">{{ cur.name }}</span></header>
        <div class="cap-panel__b" style="display: flex; flex-direction: column; gap: 14px; border-bottom: 1px solid var(--cap-line)">
          <div>
            <div class="cap-kv"><span class="cap-kv__k">{{ t('backups.th.taken') }}</span><span class="cap-kv__v">{{ fmtDate(cur.taken) }}</span></div>
            <div class="cap-kv"><span class="cap-kv__k">{{ t('backups.copiedFrom') }}</span><span class="cap-kv__v">{{ cur.sourceName || '-' }}</span></div>
            <div class="cap-kv"><span class="cap-kv__k">{{ t('backups.th.profile') }}</span><span class="cap-kv__v">{{ cur.profile }}</span></div>
            <div class="cap-kv"><span class="cap-kv__k">{{ t('backups.th.size') }}</span><span class="cap-kv__v">{{ fmtSize(cur.size) }}</span></div>
          </div>
          <span class="cap-label">{{ t('backups.inThis') }}</span>
        </div>
        <div class="scroll" style="padding: 12px">
            <ul style="list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 6px">
              <li v-for="f in cur.files" :key="f.zip" style="display: flex; align-items: center; gap: 10px; font-size: 13px">
                <template v-if="f.kind === 'user'"><span class="app-portrait" aria-hidden="true"><Files :size="14" /></span><span :title="accountTitle(f.userId)" style="color: var(--cap-ink); font-weight: 700">{{ accountLabel(f.userId) }}</span><span v-if="!accountLabel(f.userId).includes(f.userId)" class="cap-dim num" style="font-size: 11px">{{ f.userId }}</span></template>
                <template v-else><Portrait :id="f.charId" :initials="initialsOf(f.charId)" /><span style="color: var(--cap-ink); font-weight: 700">{{ label(f) }}</span></template><span class="cap-dim" style="margin-left: auto; font-size: 12px">{{ fmtSize(f.size) }}</span>
              </li>
            </ul>
        </div>
        <div class="cap-panel__b" style="display: flex; flex-direction: column; gap: 14px; border-top: 1px solid var(--cap-line)">
          <span v-if="restored !== undefined" class="cap-chip cap-chip--ok" style="align-self: flex-start"><Check :size="12" aria-hidden="true" />{{ t('backups.restored', { n: restored }) }}</span>
          <span v-if="state.backupError" class="cap-chip cap-chip--crit cap-chip--wrap" role="alert" style="align-self: flex-start"><CircleX :size="12" aria-hidden="true" />{{ state.backupError }}</span>
          <div style="display: flex; gap: 8px">
            <button class="cap-btn cap-btn--pri sbtn" :disabled="clientRunning || restored !== undefined" @click="restore(cur.name)"><RotateCcw :size="14" aria-hidden="true" />{{ t('backups.restore') }}</button>
            <button class="cap-btn cap-btn--crit sbtn" @click="del()"><Trash2 :size="14" aria-hidden="true" />{{ t('backups.delete') }}</button>
          </div>
          <span class="cap-hint">{{ clientRunning ? t('backups.blocked') : t('backups.hint') }}</span>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
/* Long names stay on one line; the full text is in the tooltip. */
/* Restore + delete on one row, compact enough for the narrow detail panel. */
.sbtn { padding: 6px 10px; font-size: 11px; white-space: nowrap; }
.clip { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
</style>

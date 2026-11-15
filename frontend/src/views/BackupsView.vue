<script setup lang="ts">
import ScrollTable from '../components/ScrollTable.vue'
import { computed } from 'vue'
import { Archive, ArchiveX, Check, CircleX, FolderOpen, RotateCcw, Trash2 } from 'lucide-vue-next'
import PageHead from '../components/PageHead.vue'
import * as Svc from '../../bindings/github.com/capsuleers-app/capsuleers.settings/service'
import { fmtDate, fmtSize, t } from '../i18n'
import { accountLabel, accountTitle, clientRunning, deleteBackup, KEEP_BACKUPS, restore, state } from '../store'

const cur = computed(() => state.backups.find((b) => b.name === state.backupId))
const restored = computed(() => (cur.value ? state.restored[cur.value.name] : undefined))
const kindLabel = (k: string) => (k === 'restore' ? t('backups.kind.restore') : t('backups.kind.apply'))
const fileName = (f: { kind: string; charId: string; userId: string }) => (f.kind === 'user' ? `core_user_${f.userId}.dat` : `core_char_${f.charId}.dat`)
const label = (e: { charId: string; name: string }) => e.name || t('char.unnamed', { id: e.charId })
function del() { if (cur.value && window.confirm(t('backups.confirmDelete'))) void deleteBackup(cur.value.name) }
</script>
<template>
  <div class="stack">
    <PageHead eyebrow="backups.eyebrow" title="backups.title" lead="backups.lead" :icon="Archive">
      <button class="cap-btn" @click="Svc.OpenBackupsFolder()"><FolderOpen :size="14" aria-hidden="true" />{{ t('backups.open') }}</button>
    </PageHead>

    <section v-if="!state.backups.length" class="cap-empty">
      <ArchiveX :size="32" class="ic ic-lg" aria-hidden="true" />
      <span class="cap-title" style="font-size: 20px">{{ t('backups.empty') }}</span>
      <span style="max-width: 460px">{{ t('backups.emptyBody') }}</span>
    </section>

    <div v-else class="sc" style="overflow: hidden">
      <div class="bkw">
        <section class="bkl" :aria-label="t('backups.history')">
          <div class="sh" style="margin-bottom: 10px"><h2>{{ t('backups.history') }}</h2><span>{{ t('backups.keeping', { keep: KEEP_BACKUPS, n: state.backups.length }) }}</span></div>
          <button v-for="b in state.backups" :key="b.name" class="bkr" :class="{ on: b.name === state.backupId }" :aria-pressed="b.name === state.backupId" :aria-label="t('backups.select', { taken: fmtDate(b.taken) })" @click="state.backupId = b.name">
            <span class="tag bk-tag" :class="b.kind === 'restore' ? 'warn' : 'ok'">{{ kindLabel(b.kind) }}</span>
            <span class="bk-info" style="display: flex; flex-direction: column; gap: 2px; min-width: 0"><b class="num" style="color: var(--cap-ink); white-space: nowrap">{{ fmtDate(b.taken) }}</b><span class="mut clip" style="font-size: 12px" :title="b.sourceName || undefined">{{ [b.sourceName, b.profile.replace(/^settings_/, ''), b.files.length].filter(Boolean).join(' · ') }}</span></span>
            <span class="num mut bk-size">{{ fmtSize(b.size) }}</span>
          </button>
        </section>

        <section v-if="cur" class="bkd">
          <div class="sh"><h2>{{ kindLabel(cur.kind) }}</h2><span class="num">{{ fmtDate(cur.taken) }} &middot; {{ cur.files.length }} &middot; {{ fmtSize(cur.size) }}</span></div>
          <p style="font-size: 13px; color: var(--cap-ink-2); margin: 0 0 14px">{{ t('backups.copiedFrom') }} <b style="color: var(--cap-ink)">{{ cur.sourceName || '-' }}</b> &middot; {{ t('backups.th.profile') }} {{ cur.profile }}</p>
          <span class="lab" style="margin-bottom: 6px">{{ t('backups.inThis') }}</span>
          <ScrollTable style="flex: 0 1 auto" :cols="['', '110px', '90px']">
            <template #head><tr><th>{{ t('review.th.file') }}</th><th>{{ t('backups.th.source') }}</th><th class="r">{{ t('backups.th.size') }}</th></tr></template>
            <tr v-for="f in cur.files" :key="f.zip">
              <td class="num" style="color: var(--cap-ink); font-size: 12.5px">{{ fileName(f) }}</td>
              <td v-if="f.kind === 'user'" class="mut" :title="accountTitle(f.userId)">{{ accountLabel(f.userId) }}</td>
              <td v-else class="mut">{{ label(f) }}</td>
              <td class="num r mut">{{ fmtSize(f.size) }}</td>
            </tr>
          </ScrollTable>
          <div style="flex: none; padding: 14px 0 16px; display: flex; flex-direction: column; gap: 10px">
            <span v-if="restored !== undefined" class="tag ok" style="align-self: flex-start"><Check :size="12" aria-hidden="true" style="margin-right: 4px" />{{ t('backups.restored', { n: restored }) }}</span>
            <span v-if="state.backupError" class="tag crit" role="alert" style="align-self: flex-start; height: auto; padding: 4px 7px; white-space: normal"><CircleX :size="12" aria-hidden="true" style="margin-right: 4px; flex: none" />{{ state.backupError }}</span>
            <div style="display: flex; gap: 8px; flex-wrap: wrap">
              <button class="cap-btn cap-btn--pri" :disabled="clientRunning || restored !== undefined" @click="restore(cur.name)"><RotateCcw :size="14" aria-hidden="true" />{{ t('backups.restore') }}</button>
              <button class="cap-btn cap-btn--crit" style="margin-left: auto" @click="del()"><Trash2 :size="14" aria-hidden="true" />{{ t('backups.delete') }}</button>
            </div>
            <span class="cap-hint">{{ clientRunning ? t('backups.blocked') : t('backups.hint') }}</span>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<style scoped>
.clip { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
</style>

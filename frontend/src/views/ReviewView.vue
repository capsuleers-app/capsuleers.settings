<script setup lang="ts">
import ScrollTable from '../components/ScrollTable.vue'
import { computed } from 'vue'
import { Archive, ArrowLeft, Check, CircleCheck, CircleX, ClipboardCheck, FileText, Files, HardDrive, Minus, Play, RefreshCw, Target, TriangleAlert, UserRound } from 'lucide-vue-next'
import PageHead from '../components/PageHead.vue'
import Portrait from '../components/Portrait.vue'
import { fmtDate, fmtSize, t } from '../i18n'
import { accountLabel, accountTitle, apply, charsOnAccount, clientRunning, copyUserOn, go, nameOf, planReady, profileMeta, refreshStatus, server, source, state, targetRows } from '../store'

const done = computed(() => state.done)
// Account part of the plan, only while the account copy is on.
const plan = computed(() => (copyUserOn.value ? state.plan : null))
const userSize = (id: string) => state.profile?.users.find((u) => u.id === id)?.size ?? 0
const acctFiles = computed(() => plan.value?.writes.length ?? 0)
const total = computed(() => targetRows.value.reduce((n, r) => n + r.file.size, 0) + (plan.value?.writes ?? []).reduce((n, w) => n + userSize(w.userId), 0))
const blocked = computed(() => copyUserOn.value && !planReady.value)
const ready = computed(() => !!source.value && targetRows.value.length > 0 && !blocked.value)
const names = (ids: string[]) => ids.map(nameOf).join(', ')
const label = computed(() => {
  if (state.applying) return t('review.applying')
  if (done.value) return t('review.done')
  const n = targetRows.value.length
  return n === 1 ? t('review.applyOne') : t('review.apply', { n })
})
const hint = computed(() => {
  if (blocked.value && !done.value && targetRows.value.length) return t('review.hint.unknown')
  if (!ready.value && !done.value) return t('review.hint.empty')
  if (clientRunning.value) return t('review.hint.running')
  return done.value ? t('review.hint.done') : t('review.hint.ready')
})
</script>
<template>
  <div class="stack" style="flex: 999 1 560px">
    <PageHead eyebrow="review.eyebrow" title="review.title" lead="review.lead" :icon="ClipboardCheck" />

    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 180px), 1fr)); gap: 12px">
      <div class="cap-tile"><span class="cap-tile__lab" style="display: inline-flex; align-items: center; gap: 6px"><UserRound :size="14" class="ic" aria-hidden="true" />{{ t('review.source') }}</span><span class="cap-tile__val" style="font-size: 20px">{{ source?.name ?? '-' }}</span><span class="cap-tile__delta cap-dim">{{ source ? source.file.path.split(/[\\/]/).pop() : '' }}</span></div>
      <div class="cap-tile"><span class="cap-tile__lab" style="display: inline-flex; align-items: center; gap: 6px"><Target :size="14" class="ic" aria-hidden="true" />{{ t('review.targets') }}</span><span class="cap-tile__val">{{ done ? done.count : targetRows.length }}</span><span class="cap-tile__delta cap-dim">{{ profileMeta?.name }}</span></div>
      <div class="cap-tile"><span class="cap-tile__lab" style="display: inline-flex; align-items: center; gap: 6px"><Files :size="14" class="ic" aria-hidden="true" />{{ t('review.account') }}</span><span class="cap-tile__val">{{ done ? done.accounts : acctFiles }}</span><span class="cap-tile__delta cap-dim">{{ copyUserOn ? t('review.accountOn') : t('review.accountNote') }}</span></div>
      <div class="cap-tile"><span class="cap-tile__lab" style="display: inline-flex; align-items: center; gap: 6px"><HardDrive :size="14" class="ic" aria-hidden="true" />{{ t('review.backup') }}</span><span class="cap-tile__val">{{ fmtSize(total) }}</span><span class="cap-tile__delta cap-dim">{{ t('review.files', { n: targetRows.length + acctFiles }) }}</span></div>
    </div>

    <section v-if="clientRunning" class="cap-panel cap-panel--crit" role="alert">
      <header class="cap-panel__h"><TriangleAlert :size="14" class="ic" aria-hidden="true" style="margin-right: 8px; color: var(--cap-crit)" /><span class="cap-panel__title" style="color: var(--cap-crit)">{{ t('review.running') }}</span></header>
      <div class="cap-panel__b" style="display: flex; flex-wrap: wrap; align-items: center; gap: 12px">
        <span style="font-size: 14px; color: var(--cap-ink-2)">{{ t('review.runningBody') }}</span>
        <button class="cap-btn" style="margin-left: auto" @click="refreshStatus()"><RefreshCw :size="14" aria-hidden="true" />{{ t('review.again') }}</button>
      </div>
    </section>

    <section v-if="state.error" class="cap-panel cap-panel--crit" role="alert">
      <header class="cap-panel__h"><CircleX :size="14" class="ic" aria-hidden="true" style="margin-right: 8px; color: var(--cap-crit)" /><span class="cap-panel__title" style="color: var(--cap-crit)">{{ t('review.failed') }}</span></header>
      <div class="cap-panel__b"><span style="font-size: 14px; color: var(--cap-ink-2)">{{ t('review.failedBody', { msg: state.error }) }}</span></div>
    </section>

    <section v-if="done" class="cap-panel" style="border-color: color-mix(in srgb, var(--cap-ok) 45%, transparent)">
      <header class="cap-panel__h"><CircleCheck :size="14" class="ic" aria-hidden="true" style="margin-right: 8px; color: var(--cap-ok)" /><span class="cap-panel__title" style="color: var(--cap-ok)">{{ t('review.applied') }}</span></header>
      <div class="cap-panel__b" style="display: flex; flex-wrap: wrap; align-items: center; gap: 12px">
        <span style="font-size: 14px; color: var(--cap-ink-2)">{{ done.accounts ? t('review.appliedAcct', { n: done.count, m: done.accounts, name: done.sourceName, zip: done.backup }) : t('review.appliedBody', { n: done.count, name: done.sourceName, zip: done.backup }) }}</span>
        <button class="cap-btn" style="margin-left: auto" @click="go('backups')"><Archive :size="14" aria-hidden="true" />{{ t('review.openBackups') }}</button>
      </div>
    </section>

    <section class="cap-panel fill">
      <header class="cap-panel__h"><FileText :size="14" class="ic" aria-hidden="true" style="margin-right: 8px" /><span class="cap-panel__title">{{ t('review.files.title') }}</span><span style="margin-left: auto">{{ profileMeta?.name }} &middot; {{ server?.name }}</span></header>
      <ScrollTable :cols="['', '220px', '150px', '150px']">
        <template #head><tr><th>{{ t('review.th.char') }}</th><th>{{ t('review.th.file') }}</th><th>{{ t('review.th.saved') }}</th><th>{{ t('review.th.action') }}</th></tr></template>
            <tr v-for="r in targetRows" :key="r.id">
              <td><span style="display: inline-flex; align-items: center; gap: 10px"><Portrait :id="r.id" :initials="r.initials" /><span style="color: var(--cap-ink); font-weight: 700">{{ r.name }}</span></span></td>
              <td class="cap-dim" style="font-size: 12px">{{ r.file.path.split(/[\\/]/).pop() }}</td>
              <td class="num">{{ fmtDate(r.file.saved) }}</td>
              <td><span class="cap-chip" :class="done ? 'cap-chip--ok' : 'cap-chip--warn'"><component :is="done ? Check : TriangleAlert" :size="12" aria-hidden="true" />{{ done ? t('review.updated') : t('review.overwrite') }}</span></td>
            </tr>
      </ScrollTable>
    </section>

    <section v-if="plan && plan.unknown.length" class="cap-panel cap-panel--crit" role="alert">
      <header class="cap-panel__h"><TriangleAlert :size="14" class="ic" aria-hidden="true" style="margin-right: 8px; color: var(--cap-crit)" /><span class="cap-panel__title" style="color: var(--cap-crit)">{{ t('review.acct.unknown') }}</span></header>
      <div class="cap-panel__b"><span style="font-size: 14px; color: var(--cap-ink-2)">{{ t('review.acct.unknownBody', { names: names(plan.unknown) }) }}</span></div>
    </section>

    <section v-if="plan && (plan.writes.length || plan.same.length)" class="cap-panel">
      <header class="cap-panel__h"><Files :size="14" class="ic" aria-hidden="true" style="margin-right: 8px" /><span class="cap-panel__title">{{ t('review.acct.title') }}</span><span style="margin-left: auto">{{ accountLabel(plan.sourceUser) }} &rarr; {{ plan.writes.length }}</span></header>
      <div class="cap-panel__b" style="padding-bottom: 0"><span class="cap-hint" style="align-items: center; gap: 6px; color: var(--cap-warn)"><TriangleAlert :size="14" aria-hidden="true" />{{ t('review.acct.warn') }}</span></div>
      <table class="cap-table" style="width: 100%; border-collapse: separate; border-spacing: 0">
        <thead><tr><th>{{ t('review.th.account') }}</th><th>{{ t('review.th.chars') }}</th><th>{{ t('review.th.file') }}</th><th>{{ t('review.th.action') }}</th></tr></thead>
        <tbody>
          <tr v-for="w in plan.writes" :key="w.userId">
            <td :title="accountTitle(w.userId)" style="color: var(--cap-ink); font-weight: 700">{{ accountLabel(w.userId) }}<span class="cap-dim num" style="display: block; font-size: 11px; font-weight: 400">{{ w.userId }}</span></td>
            <td style="font-size: 13px; color: var(--cap-ink-2)">{{ names(w.chars.length ? w.chars : charsOnAccount(w.userId)) }}</td>
            <td class="cap-dim" style="font-size: 12px">core_user_{{ w.userId }}.dat</td>
            <td><span class="cap-chip" :class="done ? 'cap-chip--ok' : 'cap-chip--warn'"><component :is="done ? Check : TriangleAlert" :size="12" aria-hidden="true" />{{ done ? t('review.updated') : t('review.overwrite') }}</span></td>
          </tr>
          <tr v-if="plan.same.length">
            <td :title="accountTitle(plan.sourceUser)" style="color: var(--cap-ink-3)">{{ accountLabel(plan.sourceUser) }}<span class="cap-dim num" style="display: block; font-size: 11px">{{ plan.sourceUser }}</span></td>
            <td style="font-size: 13px; color: var(--cap-ink-3)">{{ names(plan.same) }}</td>
            <td class="cap-dim" style="font-size: 12px">core_user_{{ plan.sourceUser }}.dat</td>
            <td><span class="cap-chip"><Minus :size="12" aria-hidden="true" />{{ t('review.acct.same') }}</span></td>
          </tr>
        </tbody>
      </table>
    </section>

    <section class="cap-panel bar">
      <button class="cap-btn" @click="go('characters')"><ArrowLeft :size="14" aria-hidden="true" />{{ t('common.back') }}</button>
      <span class="cap-hint" style="align-items: center; gap: 6px"><Archive :size="14" class="ic" aria-hidden="true" />{{ hint }}</span>
      <div style="margin-left: auto"><button class="cap-btn cap-btn--pri" :disabled="clientRunning || !ready || !!done || state.applying" @click="apply()"><component :is="done ? Check : Play" :size="14" aria-hidden="true" />{{ label }}</button></div>
    </section>
  </div>
</template>

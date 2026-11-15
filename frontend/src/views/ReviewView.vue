<script setup lang="ts">
import { computed } from 'vue'
import { Archive, ArrowLeft, Check, CircleX, Lock, Minus, Play, RefreshCw, TriangleAlert } from 'lucide-vue-next'
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
  <div class="stack">
    <PageHead eyebrow="review.eyebrow" title="review.title" lead="review.lead" />

    <div class="sc">
      <div v-if="clientRunning" class="bn crit" role="alert">
        <Lock :size="18" class="ic" aria-hidden="true" />
        <div><b>{{ t('review.running') }}</b><p>{{ t('review.runningBody') }}</p></div>
        <button class="cap-btn btn-r" @click="refreshStatus()"><RefreshCw :size="14" aria-hidden="true" />{{ t('review.again') }}</button>
      </div>

      <div v-if="state.error" class="bn crit" role="alert">
        <CircleX :size="18" class="ic" aria-hidden="true" />
        <div><b>{{ t('review.failed') }}</b><p>{{ t('review.failedBody', { msg: state.error }) }}</p></div>
      </div>

      <div v-if="done" class="bn" role="status">
        <Check :size="20" class="ic" aria-hidden="true" />
        <div><b>{{ t('review.applied') }}</b><p>{{ done.accounts ? t('review.appliedAcct', { n: done.count, m: done.accounts, name: done.sourceName, zip: done.backup }) : t('review.appliedBody', { n: done.count, name: done.sourceName, zip: done.backup }) }}</p></div>
        <button class="cap-btn btn-r" @click="go('backups')"><Archive :size="14" aria-hidden="true" />{{ t('review.openBackups') }}</button>
      </div>

      <div v-if="plan && plan.unknown.length" class="bn crit" role="alert">
        <TriangleAlert :size="18" class="ic" aria-hidden="true" />
        <div><b>{{ t('review.acct.unknown') }}</b><p>{{ t('review.acct.unknownBody', { names: names(plan.unknown) }) }}</p></div>
      </div>

      <div class="sec first" :style="clientRunning || state.error || done || (plan && plan.unknown.length) ? 'padding-top: 20px' : ''">
        <div class="big">
          <div class="k"><span style="display: flex; align-items: center; gap: 10px; letter-spacing: 0; text-transform: none"><Portrait v-if="source" :id="source.id" :initials="source.initials" style="width: 34px; height: 34px" /><b style="font-size: 20px; color: var(--cap-ink)">{{ source?.name ?? '-' }}</b></span><span>{{ t('review.source') }}</span></div>
          <div class="k"><b class="num" style="font-size: 30px">{{ done ? done.count : targetRows.length }}</b><span>{{ t('review.targets') }}</span></div>
          <div class="k"><b class="num" style="font-size: 30px">{{ done ? done.accounts : acctFiles }}</b><span>{{ t('review.account') }}</span></div>
          <div class="k"><b class="num" style="font-size: 30px">{{ fmtSize(total) }}</b><span>{{ t('review.backup') }}</span></div>
        </div>
      </div>

      <div class="sec" style="margin-top: 22px">
        <div class="sh"><h2>{{ t('review.files.title') }}</h2><span>{{ profileMeta?.name }} &middot; {{ server?.name }}</span></div>
        <table class="cap-table t">
          <thead><tr><th>{{ t('review.th.char') }}</th><th>{{ t('review.th.file') }}</th><th>{{ t('review.th.saved') }}</th><th class="r">{{ t('review.th.action') }}</th></tr></thead>
          <tbody>
            <tr v-for="r in targetRows" :key="r.id">
              <td><span class="nm"><Portrait :id="r.id" :initials="r.initials" />{{ r.name }}</span></td>
              <td class="mut num" style="font-size: 12px">{{ r.file.path.split(/[\\/]/).pop() }}</td>
              <td class="mut num">{{ fmtDate(r.file.saved) }}</td>
              <td class="r"><span class="tag" :class="{ ok: done }">{{ done ? t('review.updated') : t('review.overwrite') }}</span></td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="plan && (plan.writes.length || plan.same.length)" class="sec" style="margin-top: 22px; margin-bottom: 20px">
        <div class="sh"><h2>{{ t('review.acct.title') }}</h2><span>{{ accountLabel(plan.sourceUser) }} &rarr; {{ plan.writes.length }}</span></div>
        <p style="font-size: 13px; color: var(--cap-warn); margin: 0 0 12px; display: flex; align-items: center; gap: 6px"><TriangleAlert :size="14" class="ic" aria-hidden="true" />{{ t('review.acct.warn') }}</p>
        <table class="cap-table t">
          <thead><tr><th>{{ t('review.th.account') }}</th><th>{{ t('review.th.chars') }}</th><th>{{ t('review.th.file') }}</th><th class="r">{{ t('review.th.action') }}</th></tr></thead>
          <tbody>
            <tr v-for="w in plan.writes" :key="w.userId">
              <td :title="accountTitle(w.userId)" style="color: var(--cap-ink); font-weight: 700">{{ accountLabel(w.userId) }}</td>
              <td style="font-size: 13px">{{ names(w.chars.length ? w.chars : charsOnAccount(w.userId)) }}</td>
              <td class="mut num" style="font-size: 12px">core_user_{{ w.userId }}.dat</td>
              <td class="r"><span class="tag" :class="{ ok: done }">{{ done ? t('review.updated') : t('review.overwrite') }}</span></td>
            </tr>
            <tr v-if="plan.same.length">
              <td :title="accountTitle(plan.sourceUser)" class="mut">{{ accountLabel(plan.sourceUser) }}</td>
              <td class="mut" style="font-size: 13px">{{ names(plan.same) }}</td>
              <td class="mut num" style="font-size: 12px">core_user_{{ plan.sourceUser }}.dat</td>
              <td class="r"><span class="tag"><Minus :size="12" aria-hidden="true" style="margin-right: 4px" />{{ t('review.acct.same') }}</span></td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="fb">
      <button class="cap-btn g" @click="go('characters')"><ArrowLeft :size="14" aria-hidden="true" />{{ t('common.back') }}</button>
      <span>{{ hint }}</span>
      <div class="sp"><button class="cap-btn cap-btn--pri" :disabled="clientRunning || !ready || !!done || state.applying" @click="apply()"><component :is="done ? Check : Play" :size="14" aria-hidden="true" />{{ label }}</button></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import ScrollTable from '../components/ScrollTable.vue'
import { computed, nextTick, onBeforeUnmount, ref } from 'vue'
import { Archive, ArrowRight, Check, Eye, FolderOpen, Pencil, RefreshCw, Search, Target, UserRound, Users, UserSearch, UserX, X } from 'lucide-vue-next'
import IdentifyPanel from '../components/IdentifyPanel.vue'
import InfoTip from '../components/InfoTip.vue'
import PageHead from '../components/PageHead.vue'
import Portrait from '../components/Portrait.vue'
import { fmtDate, fmtSize, t } from '../i18n'
import { accountAlias, accountGroups, accountLabel, accountTitle, copyUserOn, go, openPreview, pickSource, profileMeta, rescanProfile, rows, setAccountAlias, setCopyUser, source, sourceAccount, startIdentify, state, stopIdentify, targetRows, toggleTarget } from '../store'

const others = computed(() => rows.value.filter((r) => r.id !== state.sourceId))
const query = ref('')
// Cerca per nome o ID; la sorgente resta sempre in elenco.
const matches = (r: { id: string; name: string }) => {
  const q = query.value.trim().toLowerCase()
  return !q || r.name.toLowerCase().includes(q) || r.id.includes(q)
}
const shown = computed(() => rows.value.filter((r) => r.id === state.sourceId || matches(r)))
const shownTargets = computed(() => shown.value.filter((r) => r.id !== state.sourceId))
// Groups with at least one row in the (searched) list; the source stays visible.
const shownIds = computed(() => new Set(shown.value.map((r) => r.id)))
const shownGroups = computed(() => accountGroups.value.groups
  .map((g) => ({ ...g, shownRows: g.rows.filter((r) => shownIds.value.has(r.id)) }))
  .filter((g) => g.shownRows.length))
const shownUnknown = computed(() => accountGroups.value.unknown.filter((r) => shownIds.value.has(r.id)))
const loaded = computed(() => !!state.accounts)
const countText = (n: number) => (n === 1 ? t('acct.countOne') : t('acct.count', { n }))
// Inline rename of an account (pencil in the group header).
const editing = ref('')
const draft = ref('')
const field = ref<HTMLInputElement[] | HTMLInputElement | null>(null)
async function startEdit(userId: string) {
  draft.value = accountAlias(userId)
  editing.value = userId
  await nextTick()
  const el = Array.isArray(field.value) ? field.value[0] : field.value
  el?.focus()
  el?.select()
}
async function saveEdit() {
  const u = editing.value
  if (!u) return
  editing.value = ''
  await setAccountAlias(u, draft.value.trim())
}
onBeforeUnmount(() => void stopIdentify())
const noMatch = computed(() => !!query.value.trim() && !shownTargets.value.length)
// Un solo pulsante: seleziona i personaggi in elenco, o li deseleziona se lo sono gia' tutti.
const allSelected = computed(() => shownTargets.value.length > 0 && shownTargets.value.every((r) => state.targets[r.id]))
function toggleAll() {
  const on = !allSelected.value
  for (const r of shownTargets.value) {
    if (on) state.targets[r.id] = true
    else delete state.targets[r.id]
  }
  state.done = null
}
</script>
<template>
  <div class="stack" style="flex: 999 1 560px">
    <PageHead eyebrow="chars.eyebrow" title="chars.title" lead="chars.lead" :icon="Users">
      <button class="cap-btn" @click="rescanProfile()"><RefreshCw :size="14" aria-hidden="true" />{{ t('chars.rescan') }}</button>
    </PageHead>

    <section v-if="!profileMeta" class="cap-empty">
      <FolderOpen :size="32" class="ic ic-lg" aria-hidden="true" />
      <span>{{ t('chars.noProfile') }}</span>
      <button class="cap-btn cap-btn--pri" @click="go('profiles')"><FolderOpen :size="14" aria-hidden="true" />{{ t('nav.profiles') }}</button>
    </section>
    <section v-else-if="rows.length < 2" class="cap-empty">
      <UserX :size="32" class="ic ic-lg" aria-hidden="true" />
      <span>{{ rows.length ? t('chars.needTwo') : t('chars.noChars') }}</span>
    </section>

    <template v-else>
      <div class="split">
        <section v-if="source" class="cap-panel cap-panel--accent" style="flex: 1 1 280px; min-width: 0">
          <header class="cap-panel__h"><Target :size="14" class="ic" aria-hidden="true" style="margin-right: 8px" /><span class="cap-panel__title">{{ t('chars.source') }}</span><span style="margin-left: auto">{{ source.file.path.split(/[\\/]/).pop() }}</span></header>
          <div class="cap-panel__b" style="display: flex; flex-direction: column; gap: 14px">
            <div style="display: flex; align-items: center; gap: 12px">
              <Portrait :id="source.id" :initials="source.initials" large />
              <div style="min-width: 0">
                <div class="cap-tool__name--raw" style="font-size: 18px; font-weight: 700; color: var(--cap-ink)">{{ source.name }}</div>
                <div class="cap-dim" style="font-size: 12px; display: flex; align-items: center; gap: 6px"><UserRound :size="14" class="ic" aria-hidden="true" />{{ t('chars.id', { id: source.id }) }}</div>
              </div>
            </div>
            <div>
              <div class="cap-kv"><span class="cap-kv__k">{{ t('chars.lastSaved') }}</span><span class="cap-kv__v">{{ fmtDate(source.file.saved) }}</span></div>
              <div class="cap-kv"><span class="cap-kv__k">{{ t('chars.size') }}</span><span class="cap-kv__v">{{ fmtSize(source.file.size) }}</span></div>
              <div class="cap-kv"><span class="cap-kv__k">{{ t('chars.account') }}</span><span class="cap-kv__v">{{ sourceAccount ? accountLabel(sourceAccount) : t('acct.unknown') }}</span></div>
            </div>
            <label style="display: flex; gap: 8px; align-items: flex-start; font-size: 13px" :style="{ color: sourceAccount ? 'var(--cap-ink-2)' : 'var(--cap-ink-3)' }"><input class="cap-input" type="checkbox" :disabled="!sourceAccount" :checked="copyUserOn" style="margin-top: 2px" @change="setCopyUser(($event.target as HTMLInputElement).checked)"> <span>{{ t('chars.user') }} <InfoTip :text="sourceAccount ? t('chars.userHint') : t('chars.userNeed')" /></span></label>
            <span class="cap-hint" style="align-items: center; gap: 6px">{{ t('chars.noteShort') }} <InfoTip :text="t('chars.note')" /></span>
            <span v-if="state.namesOffline" class="cap-hint">{{ t('chars.offline') }}</span>
          </div>
        </section>

        <section class="cap-panel fill" style="flex: 2 1 480px; min-width: 0">
          <header class="cap-panel__h"><Users :size="14" class="ic" aria-hidden="true" style="margin-right: 8px" /><span class="cap-panel__title">{{ t('chars.targets') }}</span><span style="margin-left: auto">{{ t('chars.selected', { n: targetRows.length, total: others.length }) }}</span></header>
          <div style="display: flex; flex-wrap: wrap; gap: 8px; padding: 10px 12px; border-bottom: 1px solid var(--cap-line-soft)">
            <button class="cap-chip cap-chip--on" aria-pressed="true">{{ t('chars.all', { n: others.length }) }}</button>
            <button class="cap-chip" :disabled="!shownTargets.length" @click="toggleAll()">{{ allSelected ? t('chars.deselectAll') : t('chars.selectAll') }}</button>
            <span class="targets__search-w"><Search :size="14" class="ic" aria-hidden="true" /><input v-model="query" class="cap-input targets__search" type="search" :placeholder="t('chars.search')" :aria-label="t('chars.search')"></span>
          </div>
          <ScrollTable :cols="['80px', '', '150px', '90px', '64px']">
            <template #head><tr>
                <th class="text-center">{{ t('chars.th.source') }}</th><th>{{ t('chars.th.char') }}</th><th>{{ t('chars.th.saved') }}</th>
                <th class="text-right">{{ t('chars.th.size') }}</th>
                <th class="text-center">{{ t('chars.th.target') }}</th>
              </tr></template>
            <template v-for="g in shownGroups" :key="g.userId">
              <tr class="grp"><td colspan="5">
                <span v-if="editing === g.userId" class="grp__row">
                  <input ref="field" v-model="draft" class="cap-input grp__edit" maxlength="40" :aria-label="t('acct.rename')" :placeholder="t('acct.aliasPh')" @keydown.enter.prevent="saveEdit" @keydown.esc.prevent="editing = ''">
                  <button class="grp__ib" :title="t('acct.save')" :aria-label="t('acct.save')" @click="saveEdit"><Check :size="14" aria-hidden="true" /></button>
                  <button class="grp__ib" :title="t('acct.cancel')" :aria-label="t('acct.cancel')" @click="editing = ''"><X :size="14" aria-hidden="true" /></button>
                </span>
                <span v-else class="grp__row" :title="accountTitle(g.userId)">
                  <span class="grp__name">{{ accountLabel(g.userId) }}</span>
                  <button class="grp__ib" :title="t('acct.rename')" :aria-label="t('acct.rename')" @click="startEdit(g.userId)"><Pencil :size="14" aria-hidden="true" /></button>
                  <span class="cap-dim grp__n">{{ countText(g.rows.length) }}</span>
                </span>
              </td></tr>
              <tr v-for="r in g.shownRows" :key="r.id" :class="{ 'is-target': r.id !== state.sourceId && state.targets[r.id] }">
                <td style="text-align: center"><input class="cap-input" type="radio" name="source" :checked="r.id === state.sourceId" :aria-label="t('chars.useSource', { name: r.name })" @change="pickSource(r.id)"></td>
                <td><span style="display: inline-flex; align-items: center; gap: 10px"><Portrait :id="r.id" :initials="r.initials" /><span style="color: var(--cap-ink); font-weight: 700">{{ r.name }}</span></span></td>
                <td class="num">{{ fmtDate(r.file.saved) }}</td>
                <td class="num" style="text-align: right">{{ fmtSize(r.file.size) }}</td>
                <td style="text-align: center">
                  <span v-if="r.id === state.sourceId" class="cap-chip cap-chip--on" style="padding: 3px 6px">{{ t('chars.src') }}</span>
                  <input v-else class="cap-input" type="checkbox" :checked="!!state.targets[r.id]" :aria-label="t('chars.target', { name: r.name })" @change="toggleTarget(r.id)">
                </td>
              </tr>
            </template>
            <template v-if="loaded && shownUnknown.length">
              <tr class="grp"><td colspan="5"><span class="grp__row">
                <span class="grp__name">{{ t('acct.toIdentify') }}</span><InfoTip :text="t('acct.toIdentifyHint')" below />
                <span class="cap-dim grp__n">{{ countText(accountGroups.unknown.length) }}</span>
                <button class="cap-btn grp__btn" :disabled="state.identify.open" @click="startIdentify()"><UserSearch :size="14" aria-hidden="true" />{{ t('ident.btn') }}</button>
              </span></td></tr>
              <tr v-for="r in shownUnknown" :key="r.id" :class="{ 'is-target': r.id !== state.sourceId && state.targets[r.id] }">
                <td style="text-align: center"><input class="cap-input" type="radio" name="source" :checked="r.id === state.sourceId" :aria-label="t('chars.useSource', { name: r.name })" @change="pickSource(r.id)"></td>
                <td><span style="display: inline-flex; align-items: center; gap: 10px"><Portrait :id="r.id" :initials="r.initials" /><span style="color: var(--cap-ink); font-weight: 700">{{ r.name }}</span></span></td>
                <td class="num">{{ fmtDate(r.file.saved) }}</td>
                <td class="num" style="text-align: right">{{ fmtSize(r.file.size) }}</td>
                <td style="text-align: center">
                  <span v-if="r.id === state.sourceId" class="cap-chip cap-chip--on" style="padding: 3px 6px">{{ t('chars.src') }}</span>
                  <input v-else class="cap-input" type="checkbox" :checked="!!state.targets[r.id]" :aria-label="t('chars.target', { name: r.name })" @change="toggleTarget(r.id)">
                </td>
              </tr>
            </template>
            <tr v-if="noMatch"><td colspan="5" class="cap-dim" style="text-align: center; padding: 16px">{{ t('chars.noMatch', { q: query.trim() }) }}</td></tr>
          </ScrollTable>
        </section>
      </div>

      <IdentifyPanel v-if="state.identify.open" />

      <section class="cap-panel bar">
        <span style="font-size: 14px; color: var(--cap-ink)">{{ targetRows.length === 1 ? t('chars.willOne', { name: source?.name ?? '' }) : t('chars.will', { n: targetRows.length, name: source?.name ?? '' }) }}</span>
        <span class="cap-hint" style="align-items: center; gap: 6px"><Archive :size="14" class="ic" aria-hidden="true" />{{ t('chars.backupFirst') }}</span>
        <div style="margin-left: auto; display: flex; gap: 8px">
          <button class="cap-btn" :disabled="!source" @click="openPreview()"><Eye :size="14" aria-hidden="true" />{{ t('chars.preview') }}</button>
          <button class="cap-btn cap-btn--pri" :disabled="!targetRows.length" @click="go('review')">{{ targetRows.length === 1 ? t('chars.reviewOne') : t('chars.review', { n: targetRows.length }) }}<ArrowRight :size="14" aria-hidden="true" /></button>
        </div>
      </section>
    </template>
  </div>
</template>

<style scoped>
/* `input.cap-input` per battere `.cap-input:not(...):not(...)` del bundle (piu' specifico). */
.targets__search-w { margin-left: auto; display: inline-flex; align-items: center; gap: 8px; flex: 0 1 240px; color: var(--cap-ink-3); }
input.cap-input.targets__search { flex: 1; width: auto; min-width: 120px; height: 28px; padding: 0 10px; font-size: 13px; }
tr.grp > td { background: var(--cap-panel-2); border-top: 1px solid var(--cap-line); padding: 6px 12px; }
.grp__row { display: flex; align-items: center; gap: 8px; min-width: 0; }
.grp__name { color: var(--cap-ink); font-weight: 700; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.grp__n { margin-left: auto; font-size: 12px; white-space: nowrap; }
.grp__btn { flex: none; height: 28px; }
input.cap-input.grp__edit { flex: 1 1 0; min-width: 0; width: auto; height: 28px; padding: 0 8px; font-size: 13px; }
.grp__ib { display: inline-flex; align-items: center; justify-content: center; flex: none; width: 24px; height: 28px; padding: 0; background: transparent; border: 0; color: var(--cap-ink-3); cursor: pointer; }
.grp__ib:hover, .grp__ib:focus-visible { color: var(--cap-ink); }
</style>

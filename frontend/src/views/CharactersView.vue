<script setup lang="ts">
import ScrollTable from '../components/ScrollTable.vue'
import { computed, nextTick, onBeforeUnmount, ref } from 'vue'
import { Archive, ArrowRight, Check, Eye, FolderOpen, Pencil, RefreshCw, Search, Users, UserSearch, UserX, X } from 'lucide-vue-next'
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
  <div class="stack">
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
      <div class="sc" style="padding-bottom: 0; overflow: hidden">
        <div class="cw">
          <section v-if="source" class="srcp">
            <span class="lab">{{ t('chars.source') }}</span>
            <Portrait :id="source.id" :initials="source.initials" square />
            <div style="min-width: 0">
              <h2 class="cap-title" style="font-size: 20px; margin: 0; font-weight: 700; word-break: break-word; text-transform: none">{{ source.name }}</h2>
              <p class="mut num" style="font-size: 12px; margin: 3px 0 0; word-break: break-all">{{ source.file.path.split(/[\\/]/).pop() }}</p>
            </div>
            <div>
              <div class="kvr">{{ t('chars.lastSaved') }}<b class="num">{{ fmtDate(source.file.saved) }}</b></div>
              <div class="kvr">{{ t('chars.size') }}<b class="num">{{ fmtSize(source.file.size) }}</b></div>
              <div class="kvr">{{ t('chars.account') }}<b>{{ sourceAccount ? accountLabel(sourceAccount) : t('acct.unknown') }}</b></div>
            </div>
            <label style="display: flex; gap: 10px; align-items: flex-start; font-size: 13px" :style="{ color: sourceAccount ? 'var(--cap-ink-2)' : 'var(--cap-ink-3)' }"><input class="cap-input" type="checkbox" :disabled="!sourceAccount" :checked="copyUserOn" style="margin-top: 2px" @change="setCopyUser(($event.target as HTMLInputElement).checked)"> <span>{{ t('chars.user') }} <InfoTip :text="sourceAccount ? t('chars.userHint') : t('chars.userNeed')" end /></span></label>
            <span class="cap-hint" style="align-items: center; gap: 6px">{{ t('chars.noteShort') }} <InfoTip :text="t('chars.note')" end /></span>
            <span v-if="state.namesOffline" class="cap-hint">{{ t('chars.offline') }}</span>
          </section>

          <section class="dest">
            <div class="sh" style="margin-bottom: 12px">
              <h2>{{ t('chars.targets') }}</h2><span class="num"><b style="color: var(--cap-accent-txt)">{{ targetRows.length }}</b> / {{ others.length }}</span>
              <span class="r" style="display: flex; gap: 8px; align-items: center">
                <button class="cap-btn s" :disabled="!shownTargets.length" @click="toggleAll()">{{ allSelected ? t('chars.deselectAll') : t('chars.selectAll') }}</button>
                <label class="inp" style="width: 220px"><Search :size="14" class="ic" aria-hidden="true" /><input v-model="query" type="search" :placeholder="t('chars.search')" :aria-label="t('chars.search')"></label>
              </span>
            </div>
            <ScrollTable :cols="['56px', '', '160px', '110px', '100px']">
              <template #head><tr>
                <th>{{ t('chars.th.source') }}</th><th>{{ t('chars.th.char') }}</th><th>{{ t('chars.th.saved') }}</th>
                <th class="r">{{ t('chars.th.size') }}</th>
                <th class="r">{{ t('chars.th.target') }}</th>
              </tr></template>
              <template v-for="g in shownGroups" :key="g.userId">
                <tr class="gh"><td colspan="5">
                  <span v-if="editing === g.userId" class="grp__row">
                    <input ref="field" v-model="draft" class="cap-input grp__edit" maxlength="40" :aria-label="t('acct.rename')" :placeholder="t('acct.aliasPh')" @keydown.enter.prevent="saveEdit" @keydown.esc.prevent="editing = ''">
                    <button class="grp__ib" :title="t('acct.save')" :aria-label="t('acct.save')" @click="saveEdit"><Check :size="14" aria-hidden="true" /></button>
                    <button class="grp__ib" :title="t('acct.cancel')" :aria-label="t('acct.cancel')" @click="editing = ''"><X :size="14" aria-hidden="true" /></button>
                  </span>
                  <span v-else class="grp__row" :title="accountTitle(g.userId)">
                    <span class="grp__name">{{ accountLabel(g.userId) }}</span>
                    <button class="grp__ib" :title="t('acct.rename')" :aria-label="t('acct.rename')" @click="startEdit(g.userId)"><Pencil :size="14" aria-hidden="true" /></button>
                    <span class="mut grp__n">{{ countText(g.rows.length) }}</span>
                  </span>
                </td></tr>
                <tr v-for="r in g.shownRows" :key="r.id" :class="{ 'is-target': r.id !== state.sourceId && state.targets[r.id] }">
                <td><input class="cap-input" type="radio" name="source" :checked="r.id === state.sourceId" :aria-label="t('chars.useSource', { name: r.name })" @change="pickSource(r.id)"></td>
                <td><span class="nm"><Portrait :id="r.id" :initials="r.initials" />{{ r.name }}</span></td>
                <td class="num mut">{{ fmtDate(r.file.saved) }}</td>
                <td class="num r mut">{{ fmtSize(r.file.size) }}</td>
                <td class="r">
                  <span v-if="r.id === state.sourceId" class="tag acc">{{ t('chars.src') }}</span>
                  <input v-else class="cap-input" type="checkbox" :checked="!!state.targets[r.id]" :aria-label="t('chars.target', { name: r.name })" @change="toggleTarget(r.id)">
                </td>
              </tr>
              </template>
              <template v-if="loaded && shownUnknown.length">
                <tr class="gh"><td colspan="5"><span class="grp__row">
                  <span class="grp__name">{{ t('acct.toIdentify') }}</span><InfoTip :text="t('acct.toIdentifyHint')" below />
                  <span class="mut grp__n">{{ countText(accountGroups.unknown.length) }}</span>
                  <button class="cap-btn s grp__btn" :disabled="state.identify.open" @click="startIdentify()"><UserSearch :size="14" aria-hidden="true" />{{ t('ident.btn') }}</button>
                </span></td></tr>
                <tr v-for="r in shownUnknown" :key="r.id" :class="{ 'is-target': r.id !== state.sourceId && state.targets[r.id] }">
                <td><input class="cap-input" type="radio" name="source" :checked="r.id === state.sourceId" :aria-label="t('chars.useSource', { name: r.name })" @change="pickSource(r.id)"></td>
                <td><span class="nm"><Portrait :id="r.id" :initials="r.initials" />{{ r.name }}</span></td>
                <td class="num mut">{{ fmtDate(r.file.saved) }}</td>
                <td class="num r mut">{{ fmtSize(r.file.size) }}</td>
                <td class="r">
                  <span v-if="r.id === state.sourceId" class="tag acc">{{ t('chars.src') }}</span>
                  <input v-else class="cap-input" type="checkbox" :checked="!!state.targets[r.id]" :aria-label="t('chars.target', { name: r.name })" @change="toggleTarget(r.id)">
                </td>
              </tr>
              </template>
              <tr v-if="noMatch"><td colspan="5" class="mut" style="text-align: center">{{ t('chars.noMatch', { q: query.trim() }) }}</td></tr>
            </ScrollTable>
          </section>
        </div>
      </div>

      <div class="fb">
        <span><b>{{ targetRows.length === 1 ? t('chars.willOne', { name: source?.name ?? '' }) : t('chars.will', { n: targetRows.length, name: source?.name ?? '' }) }}</b> <Archive :size="14" class="ic" aria-hidden="true" style="display: inline; vertical-align: -2px" /> {{ t('chars.backupFirst') }}</span>
        <div class="sp">
          <button class="cap-btn" :disabled="!source" @click="openPreview()"><Eye :size="14" aria-hidden="true" />{{ t('chars.preview') }}</button>
          <button class="cap-btn cap-btn--pri" :disabled="!targetRows.length" @click="go('review')">{{ targetRows.length === 1 ? t('chars.reviewOne') : t('chars.review', { n: targetRows.length }) }}<ArrowRight :size="14" aria-hidden="true" /></button>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.grp__row { display: flex; align-items: center; gap: 8px; min-width: 0; }
.grp__name { color: var(--cap-ink); font-weight: 700; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.grp__n { margin-left: auto; font-size: 12px; font-weight: 400; white-space: nowrap; }
.grp__btn { flex: none; }
input.cap-input.grp__edit { flex: 1 1 0; min-width: 0; width: auto; height: 28px; padding: 0 8px; font-size: 13px; }
.grp__ib { display: inline-flex; align-items: center; justify-content: center; flex: none; width: 26px; height: 28px; padding: 0; background: transparent; border: 0; color: var(--cap-ink-3); cursor: pointer; }
.grp__ib:hover, .grp__ib:focus-visible { color: var(--cap-ink); }
</style>

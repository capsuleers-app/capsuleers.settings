<script setup lang="ts">
import { computed, ref } from 'vue'
import type { PreviewWindow } from '../../bindings/github.com/WilliamFalci/capsuleers.settings/models'
import { t, titleParts, type Key } from '../i18n'
import { go, profileMeta, rows, setPreviewSource, setPreviewTarget, state, type PreviewMode, type PreviewTab } from '../store'

const KEY_CAP = 400
const filter = ref('')
const title = titleParts('preview.title')

const p = computed(() => state.preview)
const src = computed(() => p.value?.source)
const tgt = computed(() => p.value?.target ?? null)
const hasTarget = computed(() => !!tgt.value?.readable)
const items = computed(() => p.value?.items ?? [])
const mode = computed<PreviewMode>(() => (hasTarget.value ? state.previewMode : 'source'))
const others = computed(() => rows.value.filter((r) => r.id !== state.sourceId))

const statusClass = (s: string) => ({ add: 'cap-chip cap-chip--ok', remove: 'cap-chip cap-chip--crit', move: 'cap-chip cap-chip--warn', same: 'cap-chip' })[s] ?? 'cap-chip'
const statusText = (s: string) => t(`preview.st.${s || 'none'}` as Key)

// Every window is placed on the screen it was saved on, so windows saved at different UI scales line up.
const rectStyle = (w: PreviewWindow) =>
  `left:${(w.left / w.screenW) * 100}%;top:${(w.top / w.screenH) * 100}%;width:${(w.width / w.screenW) * 100}%;height:${(w.height / w.screenH) * 100}%;`
const base = 'position:absolute;box-sizing:border-box;padding:0;margin:0;font:inherit;cursor:pointer;overflow:hidden;'

const solid = computed(() =>
  items.value
    .map((i) => ({ i, w: mode.value === 'target' ? i.target : i.source }))
    .filter((x): x is { i: (typeof items.value)[number]; w: PreviewWindow } => !!x.w)
    .map(({ i, w }) => ({
      id: i.id,
      label: i.label,
      style: base + rectStyle(w) + (i.id === state.previewSel
        ? 'background:var(--cap-accent-well);border:1px solid var(--cap-accent);z-index:2;'
        : 'background:var(--cap-panel);border:1px solid var(--cap-line);'),
    })),
)
const ghost = computed(() =>
  mode.value !== 'overlay'
    ? []
    : items.value.filter((i) => i.target && i.status !== 'same').map((i) => ({
        id: i.id,
        style: 'position:absolute;box-sizing:border-box;pointer-events:none;border:1px dashed var(--cap-ink-3);z-index:3;' + rectStyle(i.target!),
      })),
)
const side = computed(() => (mode.value === 'target' ? tgt.value : src.value))
const aspect = computed(() => (side.value && side.value.screenW > 0 && side.value.screenH > 0 ? `${side.value.screenW} / ${side.value.screenH}` : '16 / 9'))

const diffs = computed(() => items.value.filter((i) => i.status && i.status !== 'same'))
const sel = computed(() => items.value.find((i) => i.id === state.previewSel) ?? null)
const selWin = computed(() => sel.value?.source ?? sel.value?.target ?? null)
const stateText = (w: PreviewWindow | null | undefined) => {
  if (!w) return '-'
  const f = [w.locked && 'locked', w.collapsed && 'collapsed', w.compact && 'compact', w.minimized && 'minimized'].filter(Boolean) as string[]
  return (f.length ? f : ['open']).map((k) => t(`preview.s.${k}` as Key)).join(', ')
}
const rectText = (w: PreviewWindow) => `${w.left}, ${w.top} · ${w.width} × ${w.height}`
const selTarget = computed(() => {
  if (!hasTarget.value) return t('preview.pickTarget')
  return sel.value?.target ? rectText(sel.value.target) : t('preview.notOpen')
})

const scaleText = computed(() => {
  const d = p.value?.display
  if (!d?.known || (!d.scaleFullscreen && !d.scaleWindowed)) return t('preview.t.screenOnly')
  const f = Math.round((d.scaleFullscreen || d.scaleWindowed) * 100)
  const w = Math.round((d.scaleWindowed || d.scaleFullscreen) * 100)
  return f === w ? t('preview.t.scale', { n: f }) : t('preview.t.scaleBoth', { a: f, b: w })
})
const chatText = computed(() => {
  const ch = p.value?.channels.filter((c) => c.status !== 'remove') ?? []
  if (!ch.length) return t('preview.chat.empty')
  const names = ch.slice(0, 3).map((c) => c.label).join(', ')
  return ch.length > 3 ? t('preview.t.more', { names, n: ch.length - 3 }) : names
})
const chatTotal = computed(() => src.value?.channels ?? 0)

const keys = computed(() => {
  const q = filter.value.trim().toLowerCase()
  const all = src.value?.keys ?? []
  return q ? all.filter((k) => `${k.group} ${k.name}`.toLowerCase().includes(q)) : all
})

const tabs: [PreviewTab, Key][] = [['layout', 'preview.tab.layout'], ['chat', 'preview.tab.chat'], ['keys', 'preview.tab.keys']]
const modes: [PreviewMode, Key][] = [['source', 'preview.show.source'], ['target', 'preview.show.target'], ['overlay', 'preview.show.overlay']]
const chip = (m: PreviewMode) => (mode.value === m ? 'cap-chip cap-chip--on' : 'cap-chip')

function copyLayout() {
  if (state.previewTarget) state.targets[state.previewTarget] = true
  state.done = null
  go('review')
}
</script>
<template>
  <div class="stack" style="flex: 999 1 560px">
    <section style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 24px; padding: 20px 24px; background: var(--cap-panel); border: 1px solid var(--cap-line)">
      <div>
        <p class="cap-eyebrow" style="margin: 0 0 8px">{{ t('preview.eyebrow') }}</p>
        <h1 class="cap-title" style="font-size: 36px; margin: 0">{{ title[0] }} <span style="color: var(--cap-accent-txt)">{{ title[1] }}</span></h1>
      </div>
      <div v-if="rows.length" style="display: flex; flex-wrap: wrap; gap: 12px; align-items: flex-end">
        <label class="cap-field" style="min-width: 180px"><span class="cap-label">{{ t('preview.source') }}</span>
          <select class="cap-select" :value="state.sourceId" @change="setPreviewSource(($event.target as HTMLSelectElement).value)">
            <option v-for="r in rows" :key="r.id" :value="r.id">{{ r.name }}</option>
          </select>
        </label>
        <label class="cap-field" style="min-width: 180px"><span class="cap-label">{{ t('preview.compare') }}</span>
          <select class="cap-select" :value="state.previewTarget" @change="setPreviewTarget(($event.target as HTMLSelectElement).value)">
            <option v-if="!others.length" value="">{{ t('preview.noTarget') }}</option>
            <option v-for="r in others" :key="r.id" :value="r.id">{{ r.name }}</option>
          </select>
        </label>
      </div>
      <div style="margin-left: auto; display: flex; gap: 8px">
        <button v-for="[m, k] in modes" :key="m" :class="chip(m)" :disabled="m !== 'source' && !hasTarget" :aria-pressed="mode === m" @click="state.previewMode = m">{{ t(k) }}</button>
      </div>
    </section>

    <section v-if="!profileMeta" class="cap-empty">
      <span>{{ t('chars.noProfile') }}</span>
      <button class="cap-btn cap-btn--pri" @click="go('profiles')">{{ t('nav.profiles') }}</button>
    </section>
    <section v-else-if="!rows.length" class="cap-empty"><span>{{ t('chars.noChars') }}</span></section>
    <section v-else-if="!p" class="cap-empty"><span>{{ t('preview.reading') }}</span></section>
    <section v-else-if="!src?.readable" class="cap-empty" role="status">
      <span class="cap-title" style="font-size: 20px">{{ t('preview.err.title') }}</span>
      <span style="max-width: 480px">{{ src?.error === 'missing' ? t('preview.err.missing') : t('preview.err.body') }}</span>
    </section>

    <template v-else>
      <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 160px), 1fr)); gap: 12px">
        <div class="cap-tile"><span class="cap-tile__lab">{{ t('preview.t.windows') }}</span><span class="cap-tile__val">{{ src.windows }}</span><span class="cap-tile__delta cap-dim">{{ t('preview.t.stacks', { n: src.stacks, m: src.locked }) }}</span></div>
        <div class="cap-tile"><span class="cap-tile__lab">{{ t('preview.t.saved') }}</span><span class="cap-tile__val">{{ src.screenW ? `${src.screenW} × ${src.screenH}` : '-' }}</span><span class="cap-tile__delta cap-dim">{{ scaleText }}</span></div>
        <div class="cap-tile"><span class="cap-tile__lab">{{ t('preview.t.chat') }}</span><span class="cap-tile__val">{{ chatTotal }}</span><span class="cap-tile__delta cap-dim">{{ chatText }}</span></div>
        <div class="cap-tile"><span class="cap-tile__lab">{{ t('preview.t.changes') }}</span><span class="cap-tile__val">{{ hasTarget ? diffs.length : '-' }}</span><span class="cap-tile__delta cap-dim">{{ hasTarget ? t('preview.t.of', { n: items.length }) : t('preview.t.pickTarget') }}</span></div>
      </div>

      <section v-if="tgt && !tgt.readable" class="cap-panel cap-panel--crit" role="status">
        <div class="cap-panel__b" style="font-size: 13px; color: var(--cap-ink-2)">{{ tgt.error === 'missing' ? t('preview.err.missing') : t('preview.err.target') }}</div>
      </section>

      <div style="display: flex; flex-wrap: wrap; gap: 16px; align-items: flex-start">
        <section class="cap-panel" style="flex: 3 1 560px; min-width: 0">
          <nav style="display: flex; flex-wrap: wrap; border-bottom: 1px solid var(--cap-line)" :aria-label="t('preview.tabs')">
            <a v-for="[k, lab] in tabs" :key="k" class="cap-surface-tab" :class="{ 'is-active': state.previewTab === k }" href="#" :aria-current="state.previewTab === k ? 'page' : undefined" @click.prevent="state.previewTab = k">{{ t(lab) }}</a>
          </nav>

          <div v-if="state.previewTab === 'layout'" class="cap-panel__b">
            <p v-if="!items.length" class="cap-hint" style="margin: 0">{{ t('preview.noWindows') }}</p>
            <template v-else>
              <div :style="`position: relative; width: 100%; aspect-ratio: ${aspect}; background: var(--cap-panel-2); border: 1px solid var(--cap-line); overflow: hidden`">
                <button v-for="w in solid" :key="w.id" :style="w.style" :aria-label="w.label" :aria-pressed="w.id === state.previewSel" @click="state.previewSel = w.id">
                  <span style="display: block; padding: 3px 6px; border-bottom: 1px solid var(--cap-line-soft); font-size: 10px; font-weight: 800; letter-spacing: 0.08em; color: var(--cap-ink-2); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; text-align: left">{{ w.label }}</span>
                </button>
                <span v-for="g in ghost" :key="'g' + g.id" :style="g.style" aria-hidden="true"></span>
              </div>
              <div style="display: flex; flex-wrap: wrap; gap: 16px; margin-top: 10px; font-size: 12px; color: var(--cap-ink-3)">
                <span style="display: inline-flex; align-items: center; gap: 6px"><span style="width: 14px; height: 10px; border: 1px solid var(--cap-line); background: var(--cap-panel)"></span>{{ t('preview.lg.window') }}</span>
                <span v-if="mode === 'overlay'" style="display: inline-flex; align-items: center; gap: 6px"><span style="width: 14px; height: 10px; border: 1px dashed var(--cap-ink-3)"></span>{{ t('preview.lg.ghost') }}</span>
                <span style="display: inline-flex; align-items: center; gap: 6px"><span style="width: 14px; height: 10px; border: 1px solid var(--cap-accent)"></span>{{ t('preview.lg.sel') }}</span>
                <span v-if="side?.screenW" style="margin-left: auto">{{ t('preview.lg.scaled', { w: side.screenW, h: side.screenH }) }}</span>
              </div>
            </template>
          </div>

          <div v-else-if="state.previewTab === 'chat'" style="overflow-x: auto">
            <p v-if="!p.channels.length" class="cap-hint" style="margin: 12px">{{ t('preview.chat.empty') }}</p>
            <table v-else class="cap-table" style="width: 100%; border-collapse: collapse">
              <thead><tr><th>{{ t('preview.chat.th.channel') }}</th><th>{{ t('preview.chat.th.type') }}</th><th>{{ t('preview.chat.th.status') }}</th></tr></thead>
              <tbody>
                <tr v-for="c in p.channels" :key="c.id">
                  <td style="color: var(--cap-ink); font-weight: 700">{{ c.label }}</td>
                  <td>{{ t(`preview.kind.${c.kind}` as Key) }}</td>
                  <td><span :class="statusClass(c.status)"><span class="cap-dot"></span>{{ statusText(c.status) }}</span></td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-else>
            <div style="display: flex; flex-wrap: wrap; align-items: center; gap: 12px; padding: 10px 12px; border-bottom: 1px solid var(--cap-line-soft)">
              <input v-model="filter" class="cap-input" type="search" :placeholder="t('preview.keys.filter')" :aria-label="t('preview.keys.filter')" style="flex: 1 1 200px">
              <span class="cap-hint">{{ t('preview.keys.count', { n: keys.length }) }}</span>
            </div>
            <div style="overflow: auto; max-height: 420px">
              <table class="cap-table" style="width: 100%; border-collapse: collapse">
                <thead><tr><th>{{ t('preview.keys.th.group') }}</th><th>{{ t('preview.keys.th.key') }}</th><th>{{ t('preview.keys.th.type') }}</th><th class="text-right">{{ t('preview.keys.th.size') }}</th></tr></thead>
                <tbody>
                  <tr v-for="k in keys.slice(0, KEY_CAP)" :key="k.group + '/' + k.name">
                    <td>{{ k.group }}</td><td style="word-break: break-all">{{ k.name }}</td><td>{{ t(`preview.type.${k.type}` as Key) }}</td><td class="num" style="text-align: right">{{ k.type === 'list' || k.type === 'map' || k.type === 'tuple' || k.type === 'text' ? k.len : '' }}</td>
                  </tr>
                </tbody>
              </table>
              <p v-if="!keys.length" class="cap-hint" style="margin: 12px">{{ t('preview.keys.none') }}</p>
            </div>
            <p v-if="keys.length > KEY_CAP" class="cap-hint" style="margin: 0; padding: 8px 12px">{{ t('preview.keys.cap', { n: KEY_CAP }) }}</p>
          </div>
        </section>

        <div style="flex: 1 1 260px; min-width: 0; display: flex; flex-direction: column; gap: 16px">
          <section class="cap-panel cap-panel--accent">
            <header class="cap-panel__h"><span class="cap-panel__title">{{ t('preview.win') }}</span><span style="margin-left: auto; word-break: break-all">{{ sel?.id ?? '' }}</span></header>
            <div v-if="sel && selWin" class="cap-panel__b" style="display: flex; flex-direction: column; gap: 10px">
              <div class="cap-tool__name--raw" style="font-size: 16px; font-weight: 700; color: var(--cap-ink)">{{ sel.label }}</div>
              <span :class="statusClass(sel.status)" style="align-self: flex-start"><span class="cap-dot"></span>{{ statusText(sel.status) }}</span>
              <div>
                <div class="cap-kv"><span class="cap-kv__k">{{ t('preview.pos') }}</span><span class="cap-kv__v">{{ sel.source ? `${sel.source.left}, ${sel.source.top}` : t('preview.notInSource') }}</span></div>
                <div class="cap-kv"><span class="cap-kv__k">{{ t('preview.size') }}</span><span class="cap-kv__v">{{ selWin.width }} × {{ selWin.height }}</span></div>
                <div class="cap-kv"><span class="cap-kv__k">{{ t('preview.tgtToday') }}</span><span class="cap-kv__v">{{ selTarget }}</span></div>
                <div class="cap-kv"><span class="cap-kv__k">{{ t('preview.stack') }}</span><span class="cap-kv__v">{{ selWin.members.length ? selWin.members.join(', ') : '-' }}</span></div>
                <div class="cap-kv"><span class="cap-kv__k">{{ t('preview.state') }}</span><span class="cap-kv__v">{{ stateText(selWin) }}</span></div>
              </div>
            </div>
            <div v-else class="cap-panel__b cap-hint">{{ t('preview.noWindows') }}</div>
          </section>

          <section class="cap-panel">
            <header class="cap-panel__h"><span class="cap-panel__title">{{ t('preview.applies') }}</span></header>
            <p v-if="!hasTarget" class="cap-hint" style="margin: 0; padding: 12px">{{ t('preview.pickTarget') }}</p>
            <p v-else-if="!diffs.length" class="cap-hint" style="margin: 0; padding: 12px">{{ t('preview.noChanges') }}</p>
            <ul v-else style="list-style: none; margin: 0; padding: 6px 0; max-height: 320px; overflow: auto">
              <li v-for="d in diffs" :key="d.id">
                <button style="width: 100%; display: flex; align-items: center; gap: 8px; padding: 6px 12px; background: none; border: 0; color: var(--cap-ink-2); font: inherit; font-size: 13px; cursor: pointer; text-align: left" @click="state.previewSel = d.id; state.previewTab = 'layout'">
                  <span style="flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis">{{ d.label }}</span>
                  <span :class="statusClass(d.status)"><span class="cap-dot"></span>{{ statusText(d.status) }}</span>
                </button>
              </li>
            </ul>
          </section>
        </div>
      </div>

      <section class="cap-panel bar">
        <span class="cap-hint">{{ t('preview.readonly') }}</span>
        <div style="margin-left: auto"><button class="cap-btn cap-btn--pri" :disabled="!state.previewTarget" @click="copyLayout()">{{ t('preview.copy') }}</button></div>
      </section>
    </template>
  </div>
</template>

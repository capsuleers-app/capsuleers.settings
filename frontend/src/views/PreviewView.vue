<script setup lang="ts">
import ScrollTable from '../components/ScrollTable.vue'
import { computed, ref, watch } from 'vue'
import type { PreviewWindow } from '../../bindings/github.com/WilliamFalci/capsuleers.settings/models'
import { AppWindow, Clock, Compass, Eye, FileWarning, FolderOpen, KeyRound, LayoutGrid, LayoutTemplate, MessageSquare, Monitor, Rocket, Shapes, Store, Swords, Users } from 'lucide-vue-next'
import type { Component } from 'vue'
import { t, titleParts, type Key } from '../i18n'
import Portrait from '../components/Portrait.vue'
import { go, profileMeta, rows, source, state, type PreviewTab } from '../store'

const KEY_CAP = 400
const CAT_ICON: Record<string, Component> = { combat: Swords, chat: MessageSquare, fleet: Users, ship: Rocket, market: Store, navigation: Compass, other: Shapes }
const TAB_ICON: Record<PreviewTab, Component> = { layout: LayoutTemplate, chat: MessageSquare, keys: KeyRound }
const filter = ref('')
const title = titleParts('preview.title')

const p = computed(() => state.preview)
const src = computed(() => p.value?.source)
const items = computed(() => p.value?.windows ?? [])
// Every window is placed on the screen it was saved on, so windows saved at different UI scales line up.
const rectStyle = (w: PreviewWindow) =>
  `left:${(w.left / w.screenW) * 100}%;top:${(w.top / w.screenH) * 100}%;width:${(w.width / w.screenW) * 100}%;height:${(w.height / w.screenH) * 100}%;`
const base = 'position:absolute;box-sizing:border-box;padding:0;margin:0;font:inherit;cursor:pointer;overflow:hidden;'

const CATS = ['combat', 'chat', 'fleet', 'ship', 'market', 'navigation', 'other']
const catCounts = computed(() => {
  const m: Record<string, number> = {}
  for (const i of items.value) m[i.category] = (m[i.category] ?? 0) + 1
  return m
})
const catList = computed(() => CATS.filter((c) => catCounts.value[c]))
const activeCats = computed(() => state.previewCats.filter((c) => catCounts.value[c]))
const filtered = computed(() => activeCats.value.length > 0)
const visible = computed(() => (filtered.value ? items.value.filter((i) => activeCats.value.includes(i.category)) : items.value))
const toggleCat = (c: string | null) => {
  if (c === null) { state.previewCats = []; return }
  state.previewCats = activeCats.value.includes(c) ? activeCats.value.filter((x) => x !== c) : [...activeCats.value, c]
}
const catChip = (on: boolean) => 'cap-chip' + (on ? ' cat-on' : '')
watch(visible, (v) => { if (state.previewSel && !v.some((i) => i.id === state.previewSel)) state.previewSel = '' })

const solid = computed(() =>
  visible.value.map((w) => ({
    id: w.id,
    label: w.label,
    style: base + rectStyle(w) + (w.id === state.previewSel
      ? 'background:var(--cap-accent-well);border:1px solid var(--cap-accent);z-index:2;'
      : 'background:var(--cap-panel);border:1px solid var(--cap-line);'),
  })),
)
const ratio = computed(() => (src.value && src.value.screenW > 0 && src.value.screenH > 0 ? src.value.screenW / src.value.screenH : 16 / 9))
// Il canvas sta tutto nel riquadro disponibile (come object-fit: contain), senza scorrere.
const stage = ref<HTMLElement | null>(null)
const stageSize = ref({ w: 0, h: 0 })
let ro: ResizeObserver | null = null
watch(stage, (el, _old, onCleanup) => {
  ro?.disconnect()
  if (!el) return
  ro = new ResizeObserver(([e]) => { stageSize.value = { w: e.contentRect.width, h: e.contentRect.height } })
  ro.observe(el)
  onCleanup(() => ro?.disconnect())
})
const canvasStyle = computed(() => {
  const { w, h } = stageSize.value
  if (!w || !h) return 'width: 100%; aspect-ratio: ' + ratio.value
  const cw = Math.min(w, h * ratio.value)
  return `width: ${Math.floor(cw)}px; height: ${Math.floor(cw / ratio.value)}px`
})

const sel = computed(() => items.value.find((i) => i.id === state.previewSel) ?? null)
const stateText = (w: PreviewWindow) => {
  const f = [w.locked && 'locked', w.collapsed && 'collapsed', w.compact && 'compact', w.minimized && 'minimized'].filter(Boolean) as string[]
  return (f.length ? f : ['open']).map((k) => t(`preview.s.${k}` as Key)).join(', ')
}
const scaleText = computed(() => {
  const d = p.value?.display
  if (!d?.known || (!d.scaleFullscreen && !d.scaleWindowed)) return t('preview.t.screenOnly')
  const f = Math.round((d.scaleFullscreen || d.scaleWindowed) * 100)
  const w = Math.round((d.scaleWindowed || d.scaleFullscreen) * 100)
  return f === w ? t('preview.t.scale', { n: f }) : t('preview.t.scaleBoth', { a: f, b: w })
})
const chatText = computed(() => {
  const ch = p.value?.channels ?? []
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
</script>
<template>
  <div class="stack" style="flex: 999 1 560px">
    <section style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 24px; padding: 20px 24px; background: var(--cap-panel); border: 1px solid var(--cap-line)">
      <div>
        <p class="cap-eyebrow" style="margin: 0 0 8px; display: flex; align-items: center; gap: 8px"><Eye :size="14" aria-hidden="true" />{{ t('preview.eyebrow') }}</p>
        <h1 class="cap-title" style="font-size: 36px; margin: 0">{{ title[0] }} <span style="color: var(--cap-accent-txt)">{{ title[1] }}</span></h1>
      </div>
      <div v-if="source" style="display: flex; align-items: center; gap: 12px; margin-left: auto">
        <Portrait :id="source.id" :initials="source.initials" large />
        <div style="min-width: 0">
          <div class="cap-tool__name--raw" style="font-size: 18px; font-weight: 700; color: var(--cap-ink)">{{ source.name }}</div>
          <div class="cap-dim" style="font-size: 12px">{{ t('chars.id', { id: source.id }) }} &middot; {{ source.file.path.split(/[\/]/).pop() }}</div>
        </div>
      </div>
    </section>

    <section v-if="!profileMeta" class="cap-empty">
      <FolderOpen :size="32" class="ic ic-lg" aria-hidden="true" />
      <span>{{ t('chars.noProfile') }}</span>
      <button class="cap-btn cap-btn--pri" @click="go('profiles')"><FolderOpen :size="14" aria-hidden="true" />{{ t('nav.profiles') }}</button>
    </section>
    <section v-else-if="!rows.length || !source" class="cap-empty">
      <Users :size="32" class="ic ic-lg" aria-hidden="true" />
      <span>{{ rows.length ? t('preview.noSource') : t('chars.noChars') }}</span>
      <button class="cap-btn cap-btn--pri" @click="go('characters')"><Users :size="14" aria-hidden="true" />{{ t('nav.characters') }}</button>
    </section>
    <section v-else-if="!p" class="cap-empty"><Clock :size="32" class="ic ic-lg" aria-hidden="true" /><span>{{ t('preview.reading') }}</span></section>
    <section v-else-if="!src?.readable" class="cap-empty" role="status">
      <FileWarning :size="32" class="ic ic-lg" aria-hidden="true" />
      <span class="cap-title" style="font-size: 20px">{{ t('preview.err.title') }}</span>
      <span style="max-width: 480px">{{ src?.error === 'missing' ? t('preview.err.missing') : t('preview.err.body') }}</span>
    </section>

    <template v-else>
      <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 160px), 1fr)); gap: 12px">
        <div class="cap-tile"><span class="cap-tile__lab" style="display: inline-flex; align-items: center; gap: 6px"><AppWindow :size="14" class="ic" aria-hidden="true" />{{ t('preview.t.windows') }}</span><span class="cap-tile__val">{{ src.windows }}</span><span class="cap-tile__delta cap-dim">{{ t('preview.t.stacks', { n: src.stacks, m: src.locked }) }}</span></div>
        <div class="cap-tile"><span class="cap-tile__lab" style="display: inline-flex; align-items: center; gap: 6px"><Monitor :size="14" class="ic" aria-hidden="true" />{{ t('preview.t.saved') }}</span><span class="cap-tile__val">{{ src.screenW ? `${src.screenW} × ${src.screenH}` : '-' }}</span><span class="cap-tile__delta cap-dim">{{ scaleText }}</span></div>
        <div class="cap-tile"><span class="cap-tile__lab" style="display: inline-flex; align-items: center; gap: 6px"><MessageSquare :size="14" class="ic" aria-hidden="true" />{{ t('preview.t.chat') }}</span><span class="cap-tile__val">{{ chatTotal }}</span><span class="cap-tile__delta cap-dim">{{ chatText }}</span></div>
      </div>

      <div class="split">
        <section class="cap-panel fill" style="flex: 3 1 560px; min-width: 0">
          <nav style="display: flex; flex-wrap: wrap; border-bottom: 1px solid var(--cap-line)" :aria-label="t('preview.tabs')">
            <a v-for="[k, lab] in tabs" :key="k" class="cap-surface-tab" :class="{ 'is-active': state.previewTab === k }" href="#" :aria-current="state.previewTab === k ? 'page' : undefined" @click.prevent="state.previewTab = k"><component :is="TAB_ICON[k]" :size="14" class="ic" aria-hidden="true" style="margin-right: 6px; vertical-align: -2px" />{{ t(lab) }}</a>
          </nav>

          <div v-if="state.previewTab === 'layout'" class="cap-panel__b fill layout-tab">
            <p v-if="!items.length" class="cap-hint" style="margin: 0">{{ t('preview.noWindows') }}</p>
            <template v-else>
              <div role="group" :aria-label="t('preview.cat.filter')" style="display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 12px">
                <button :class="catChip(!filtered)" :aria-pressed="!filtered" @click="toggleCat(null)"><LayoutGrid :size="12" aria-hidden="true" />{{ t('preview.cat.all') }} · {{ items.length }}<template v-if="!filtered"> · {{ t('preview.cat.on') }}</template></button>
                <button v-for="c in catList" :key="c" :class="catChip(activeCats.includes(c))" :aria-pressed="activeCats.includes(c)" @click="toggleCat(c)"><component :is="CAT_ICON[c]" :size="12" aria-hidden="true" />{{ t(`preview.cat.${c}` as Key) }} · {{ catCounts[c] }}<template v-if="activeCats.includes(c)"> · {{ t('preview.cat.on') }}</template></button>
              </div>
              <div ref="stage" class="layout-stage">
              <div :style="`position: relative; ${canvasStyle}; box-sizing: border-box; background: var(--cap-panel-2); border: 1px solid var(--cap-line); overflow: hidden`">
                <button v-for="w in solid" :key="w.id" :style="w.style" :aria-label="w.label" :aria-pressed="w.id === state.previewSel" @click="state.previewSel = w.id">
                  <span style="display: block; padding: 3px 6px; border-bottom: 1px solid var(--cap-line-soft); font-size: 10px; font-weight: 800; letter-spacing: 0.08em; color: var(--cap-ink-2); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; text-align: left">{{ w.label }}</span>
                </button>
              </div>
              </div>
              <div style="display: flex; flex-wrap: wrap; gap: 16px; margin-top: 10px; font-size: 12px; color: var(--cap-ink-3)">
                <span style="display: inline-flex; align-items: center; gap: 6px"><span style="width: 14px; height: 10px; border: 1px solid var(--cap-line); background: var(--cap-panel)"></span>{{ t('preview.lg.window') }}</span>
                <span style="display: inline-flex; align-items: center; gap: 6px"><span style="width: 14px; height: 10px; border: 1px solid var(--cap-accent)"></span>{{ t('preview.lg.sel') }}</span>
                <span v-if="filtered">{{ t('preview.lg.shown', { n: visible.length, m: items.length }) }}</span>
                <span v-if="src.screenW" style="margin-left: auto">{{ t('preview.lg.scaled', { w: src.screenW, h: src.screenH }) }}</span>
              </div>
            </template>
          </div>

          <template v-else-if="state.previewTab === 'chat'">
            <p v-if="!p.channels.length" class="cap-hint" style="margin: 12px">{{ t('preview.chat.empty') }}</p>
            <ScrollTable v-else :cols="['', '200px']">
              <template #head><tr><th>{{ t('preview.chat.th.channel') }}</th><th>{{ t('preview.chat.th.type') }}</th></tr></template>
                <tr v-for="c in p.channels" :key="c.id">
                  <td style="color: var(--cap-ink); font-weight: 700">{{ c.label }}</td>
                  <td>{{ t(`preview.kind.${c.kind}` as Key) }}</td>
                </tr>
            </ScrollTable>
          </template>

          <div v-else class="fill">
            <div style="display: flex; flex-wrap: wrap; align-items: center; gap: 12px; padding: 10px 12px; border-bottom: 1px solid var(--cap-line-soft)">
              <input v-model="filter" class="cap-input" type="search" :placeholder="t('preview.keys.filter')" :aria-label="t('preview.keys.filter')" style="flex: 1 1 200px">
              <span class="cap-hint">{{ t('preview.keys.count', { n: keys.length }) }}</span>
            </div>
            <ScrollTable v-if="keys.length" :cols="['180px', '', '110px', '90px']">
                <template #head><tr><th>{{ t('preview.keys.th.group') }}</th><th>{{ t('preview.keys.th.key') }}</th><th>{{ t('preview.keys.th.type') }}</th><th class="text-right">{{ t('preview.keys.th.size') }}</th></tr></template>
                  <tr v-for="k in keys.slice(0, KEY_CAP)" :key="k.group + '/' + k.name">
                    <td>{{ k.group }}</td><td style="word-break: break-all">{{ k.name }}</td><td>{{ t(`preview.type.${k.type}` as Key) }}</td><td class="num" style="text-align: right">{{ k.type === 'list' || k.type === 'map' || k.type === 'tuple' || k.type === 'text' ? k.len : '' }}</td>
                  </tr>
            </ScrollTable>
            <p v-else class="cap-hint" style="margin: 12px">{{ t('preview.keys.none') }}</p>
            <p v-if="keys.length > KEY_CAP" class="cap-hint" style="margin: 0; padding: 8px 12px">{{ t('preview.keys.cap', { n: KEY_CAP }) }}</p>
          </div>
        </section>

        <div style="flex: 1 1 260px; min-width: 0">
          <section class="cap-panel cap-panel--accent">
            <header class="cap-panel__h"><AppWindow :size="14" class="ic" aria-hidden="true" style="margin-right: 8px" /><span class="cap-panel__title">{{ t('preview.win') }}</span><span style="margin-left: auto; word-break: break-all">{{ sel?.id ?? '' }}</span></header>
            <div v-if="sel" class="cap-panel__b" style="display: flex; flex-direction: column; gap: 10px">
              <div class="cap-tool__name--raw" style="font-size: 16px; font-weight: 700; color: var(--cap-ink)">{{ sel.label }}</div>
              <div>
                <div class="cap-kv"><span class="cap-kv__k">{{ t('preview.pos') }}</span><span class="cap-kv__v">{{ `${sel.left}, ${sel.top}` }}</span></div>
                <div class="cap-kv"><span class="cap-kv__k">{{ t('preview.size') }}</span><span class="cap-kv__v">{{ sel.width }} × {{ sel.height }}</span></div>
                <div class="cap-kv"><span class="cap-kv__k">{{ t('preview.stack') }}</span><span class="cap-kv__v">{{ sel.members.length ? sel.members.join(', ') : '-' }}</span></div>
                <div class="cap-kv"><span class="cap-kv__k">{{ t('preview.state') }}</span><span class="cap-kv__v">{{ stateText(sel) }}</span></div>
              </div>
            </div>
            <div v-else class="cap-panel__b cap-hint">{{ t('preview.noWindows') }}</div>
          </section>

        </div>
      </div>

      <section class="cap-panel bar">
        <span class="cap-hint" style="align-items: center; gap: 6px"><Eye :size="14" class="ic" aria-hidden="true" />{{ t('preview.readonly') }}</span>
      </section>
    </template>
  </div>
</template>
<style scoped>
/* Il canvas riempie lo spazio restante e ci sta tutto (scala proporzionale), senza scorrere. */
.layout-tab { display: flex; flex-direction: column; min-height: 0; overflow: hidden; }
.layout-stage { flex: 1 1 0; min-height: 160px; display: flex; align-items: center; justify-content: center; overflow: hidden; }
.cat-on { color: var(--cap-ink); background: var(--cap-accent-well); border-color: var(--cap-accent-line); }
</style>

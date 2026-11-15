<script setup lang="ts">
import ScrollTable from '../components/ScrollTable.vue'
import { computed, ref, watch } from 'vue'
import type { PreviewWindow } from '../../bindings/github.com/capsuleers-app/capsuleers.settings/models'
import { Clock, Search, Compass, FileWarning, FolderOpen, KeyRound, LayoutGrid, LayoutTemplate, MessageSquare, Rocket, Shapes, Store, Swords, Users } from 'lucide-vue-next'
import type { Component } from 'vue'
import { t, type Key } from '../i18n'
import PageHead from '../components/PageHead.vue'
import Portrait from '../components/Portrait.vue'
import { go, profileMeta, rows, source, state, type PreviewTab } from '../store'

const KEY_CAP = 400
const CAT_ICON: Record<string, Component> = { combat: Swords, chat: MessageSquare, fleet: Users, ship: Rocket, market: Store, navigation: Compass, other: Shapes }
const TAB_ICON: Record<PreviewTab, Component> = { layout: LayoutTemplate, chat: MessageSquare, keys: KeyRound }
const filter = ref('')

const p = computed(() => state.preview)
const src = computed(() => p.value?.source)
const items = computed(() => p.value?.windows ?? [])
// Every window is placed on the screen it was saved on, so windows saved at different UI scales line up.
const rectStyle = (w: PreviewWindow) =>
  `left:${(w.left / w.screenW) * 100}%;top:${(w.top / w.screenH) * 100}%;width:${(w.width / w.screenW) * 100}%;height:${(w.height / w.screenH) * 100}%;`
const base = 'position:absolute;box-sizing:border-box;padding:0;margin:0;font:inherit;cursor:pointer;overflow:hidden;text-align:left;'

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
const catChip = (on: boolean) => 'fchip' + (on ? ' on' : '')
watch(visible, (v) => { if (state.previewSel && !v.some((i) => i.id === state.previewSel)) state.previewSel = '' })

const solid = computed(() =>
  visible.value.map((w) => ({
    id: w.id,
    label: w.label,
    style: base + rectStyle(w) + (w.id === state.previewSel
      ? 'background:var(--cap-accent-well);border:1px solid var(--cap-accent);z-index:2;color:var(--cap-ink);'
      : 'background:var(--cap-panel-2);border:1px solid var(--cap-line);color:var(--cap-ink-2);'),
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
  <div class="stack">
    <PageHead eyebrow="preview.eyebrow" title="preview.title" lead="preview.readonly" />

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

    <div v-else class="sc" style="overflow: hidden">
      <div class="sec first" style="padding-bottom: 0">
        <div class="ins" style="gap: 14px">
          <Portrait :id="source.id" :initials="source.initials" large />
          <div class="who"><b style="font-size: 16px">{{ source.name }}</b><p class="num">{{ source.file.path.split(/[\\/]/).pop() }}</p></div>
          <div class="big" style="margin-left: auto">
            <div class="k"><b class="num">{{ src.windows }}</b><span>{{ t('preview.t.windows') }} &middot; {{ t('preview.t.stacks', { n: src.stacks, m: src.locked }) }}</span></div>
            <div class="k"><b class="num">{{ src.screenW ? `${src.screenW} × ${src.screenH}` : '-' }}</b><span>{{ t('preview.t.saved') }} &middot; {{ scaleText }}</span></div>
            <div class="k" :title="chatText"><b class="num">{{ chatTotal }}</b><span>{{ t('preview.t.chat') }}</span></div>
          </div>
        </div>
      </div>

      <div class="sec grow" style="margin-top: 22px; padding-top: 0; border-top: 0; display: flex; flex-direction: column; min-height: 0">
        <nav class="tbs" :aria-label="t('preview.tabs')">
          <a v-for="[k, lab] in tabs" :key="k" :class="{ on: state.previewTab === k }" href="#" :aria-current="state.previewTab === k ? 'page' : undefined" @click.prevent="state.previewTab = k"><component :is="TAB_ICON[k]" :size="14" class="ic" aria-hidden="true" />{{ t(lab) }}</a>
        </nav>

        <div v-if="state.previewTab === 'layout'" class="lay-tab">
          <p v-if="!items.length" class="cap-hint" style="margin: 14px 0">{{ t('preview.noWindows') }}</p>
          <template v-else>
            <div role="group" :aria-label="t('preview.cat.filter')" style="display: flex; flex-wrap: wrap; gap: 6px; margin: 14px 0; flex: none">
              <button :class="catChip(!filtered)" :aria-pressed="!filtered" @click="toggleCat(null)"><LayoutGrid :size="12" aria-hidden="true" />{{ t('preview.cat.all') }} &middot; {{ items.length }}</button>
              <button v-for="c in catList" :key="c" :class="catChip(activeCats.includes(c))" :aria-pressed="activeCats.includes(c)" @click="toggleCat(c)"><component :is="CAT_ICON[c]" :size="12" aria-hidden="true" />{{ t(`preview.cat.${c}` as Key) }} &middot; {{ catCounts[c] }}</button>
            </div>
            <div class="lay">
              <div class="lay-l">
                <div ref="stage" class="scr" role="img" :aria-label="t('preview.lg.scaled', { w: src.screenW, h: src.screenH })">
                  <div :style="`position: relative; ${canvasStyle}; box-sizing: border-box; background: var(--cap-panel); border: 1px solid var(--cap-line); overflow: hidden`">
                    <button v-for="w in solid" :key="w.id" :style="w.style" :aria-label="w.label" :aria-pressed="w.id === state.previewSel" @click="state.previewSel = w.id">
                      <span class="wl">{{ w.label }}</span>
                    </button>
                  </div>
                </div>
                <div style="display: flex; flex-wrap: wrap; gap: 16px; margin-top: 8px; font-size: 12px; color: var(--cap-ink-3); flex: none">
                  <span style="display: inline-flex; align-items: center; gap: 6px"><span style="width: 14px; height: 10px; border: 1px solid var(--cap-line); background: var(--cap-panel-2)"></span>{{ t('preview.lg.window') }}</span>
                  <span style="display: inline-flex; align-items: center; gap: 6px"><span style="width: 14px; height: 10px; border: 1px solid var(--cap-accent); background: var(--cap-accent-well)"></span>{{ t('preview.lg.sel') }}</span>
                  <span v-if="filtered">{{ t('preview.lg.shown', { n: visible.length, m: items.length }) }}</span>
                  <span v-if="src.screenW" style="margin-left: auto">{{ t('preview.lg.scaled', { w: src.screenW, h: src.screenH }) }}</span>
                </div>
              </div>
              <aside class="wd">
                <span class="lab">{{ t('preview.win') }}</span>
                <template v-if="sel">
                  <h3 class="cap-title" style="text-transform: none">{{ sel.label }}</h3>
                  <p class="mut num" style="font-size: 12px; margin: 0; word-break: break-all">{{ sel.id }}</p>
                  <div style="margin-top: 12px">
                    <div class="kvr">{{ t('preview.pos') }}<b class="num">{{ `${sel.left}, ${sel.top}` }}</b></div>
                    <div class="kvr">{{ t('preview.size') }}<b class="num">{{ sel.width }} × {{ sel.height }}</b></div>
                    <div class="kvr">{{ t('preview.stack') }}<b :title="sel.members.join(', ')">{{ sel.members.length ? sel.members.join(', ') : '-' }}</b></div>
                    <div class="kvr">{{ t('preview.state') }}<b>{{ stateText(sel) }}</b></div>
                  </div>
                </template>
                <p v-else class="cap-hint" style="margin: 8px 0 0">{{ t('preview.noWindows') }}</p>
              </aside>
            </div>
          </template>
        </div>

        <template v-else-if="state.previewTab === 'chat'">
          <p v-if="!p.channels.length" class="cap-hint" style="margin: 14px 0">{{ t('preview.chat.empty') }}</p>
          <ScrollTable v-else style="flex: 1 1 0; min-height: 0; margin-top: 14px" :cols="['', '200px']">
            <template #head><tr><th>{{ t('preview.chat.th.channel') }}</th><th>{{ t('preview.chat.th.type') }}</th></tr></template>
              <tr v-for="c in p.channels" :key="c.id">
                <td style="color: var(--cap-ink); font-weight: 700">{{ c.label }}</td>
                <td class="mut">{{ t(`preview.kind.${c.kind}` as Key) }}</td>
              </tr>
          </ScrollTable>
        </template>

        <div v-else class="keys-tab">
          <div style="display: flex; align-items: center; gap: 12px; margin: 14px 0; flex: none">
            <label class="inp" style="width: 340px"><Search :size="14" class="ic" aria-hidden="true" /><input v-model="filter" type="search" :placeholder="t('preview.keys.filter')" :aria-label="t('preview.keys.filter')"></label>
            <span class="cap-hint">{{ t('preview.keys.count', { n: keys.length }) }}</span>
          </div>
          <ScrollTable v-if="keys.length" style="flex: 1 1 0; min-height: 0" :cols="['180px', '', '110px', '90px']">
              <template #head><tr><th>{{ t('preview.keys.th.group') }}</th><th>{{ t('preview.keys.th.key') }}</th><th>{{ t('preview.keys.th.type') }}</th><th class="r">{{ t('preview.keys.th.size') }}</th></tr></template>
                <tr v-for="k in keys.slice(0, KEY_CAP)" :key="k.group + '/' + k.name">
                  <td class="mut">{{ k.group }}</td><td style="word-break: break-all; color: var(--cap-ink)">{{ k.name }}</td><td class="mut">{{ t(`preview.type.${k.type}` as Key) }}</td><td class="num r mut">{{ k.type === 'list' || k.type === 'map' || k.type === 'tuple' || k.type === 'text' ? k.len : '' }}</td>
                </tr>
          </ScrollTable>
          <p v-else class="cap-hint" style="margin: 12px 0">{{ t('preview.keys.none') }}</p>
          <p v-if="keys.length > KEY_CAP" class="cap-hint" style="margin: 0; padding: 8px 0; flex: none">{{ t('preview.keys.cap', { n: KEY_CAP }) }}</p>
        </div>
      </div>
    </div>
  </div>
</template>
<style scoped>
.lay-tab, .keys-tab { flex: 1 1 0; min-height: 0; display: flex; flex-direction: column; }
.wl { display: block; padding: 6px 7px; font-size: 11px; font-weight: 700; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; text-align: left; }
</style>

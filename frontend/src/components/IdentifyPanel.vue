<script setup lang="ts">
// Guided identification, shown as a right-hand drawer: the backend watches the profile files (read-only)
// while the user logs characters in and out; every character it pairs with an account shows up live.
import { Check, TriangleAlert, X } from 'lucide-vue-next'
import InfoTip from './InfoTip.vue'
import Portrait from './Portrait.vue'
import { t } from '../i18n'
import { accountLabel, initialsOf, nameOf, state, stopIdentify } from '../store'
</script>
<template>
  <div class="scrim" @click.self="stopIdentify()">
    <aside class="dr" role="dialog" aria-live="polite" :aria-label="t('ident.title')">
      <div class="dr-h">
        <div style="flex: 1; min-width: 0"><span class="lab">{{ t('chars.account') }}</span><h2 class="cap-title">{{ t('ident.title') }}</h2></div>
        <button class="ib" :aria-label="t('ident.done')" @click="stopIdentify()"><X :size="14" aria-hidden="true" /></button>
      </div>
      <div class="dr-s">
        <div style="display: flex; flex-direction: column; gap: 14px">
          <div class="stp"><span class="n">1</span><span>{{ t('ident.step1') }}</span></div>
          <div class="stp"><span class="n">2</span><span>{{ t('ident.step2') }} <InfoTip :text="t('ident.hint')" /></span></div>
        </div>
        <div style="display: flex; align-items: center; gap: 8px; padding: 14px; border: 1px solid var(--cap-line)">
          <span class="pulse" /><div><div style="font-size: 14px; font-weight: 700; text-transform: uppercase">{{ t('ident.listening') }}</div></div>
        </div>
        <span v-if="state.identify.ambiguous" class="cap-hint" style="align-items: center; gap: 6px; color: var(--cap-warn)"><TriangleAlert :size="14" aria-hidden="true" />{{ t('ident.ambiguous') }}</span>
        <span v-if="!state.identify.found.length" class="mut" style="font-size: 13px">{{ t('ident.empty') }}</span>
        <ul v-else style="list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column">
          <li v-for="f in state.identify.found" :key="f.charId" style="display: flex; align-items: center; gap: 10px; flex-wrap: wrap; min-height: 44px; border-bottom: 1px solid var(--cap-line-soft)">
            <Portrait :id="f.charId" :initials="initialsOf(f.charId)" />
            <span style="color: var(--cap-ink); font-weight: 700">{{ nameOf(f.charId) }}</span>
            <span class="mut" style="font-size: 13px">{{ accountLabel(f.userId) }}</span>
            <span class="tag" :class="f.remapped ? 'warn' : 'ok'" style="margin-left: auto"><Check :size="12" aria-hidden="true" style="margin-right: 4px" />{{ f.remapped ? t('ident.changed') : t('ident.found') }}</span>
          </li>
        </ul>
      </div>
      <div class="dr-f"><button class="cap-btn" style="flex: 1; justify-content: center" @click="stopIdentify()">{{ t('ident.done') }}</button></div>
    </aside>
  </div>
</template>

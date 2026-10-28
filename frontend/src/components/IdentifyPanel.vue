<script setup lang="ts">
// Guided identification: the backend watches the profile files (read-only) while the user logs
// characters in and out; every character it pairs with an account shows up here live.
import { Check, LogIn, TriangleAlert, UserSearch } from 'lucide-vue-next'
import InfoTip from './InfoTip.vue'
import Portrait from './Portrait.vue'
import { t } from '../i18n'
import { accountLabel, initialsOf, nameOf, state, stopIdentify } from '../store'
</script>
<template>
  <section class="cap-panel" aria-live="polite">
    <header class="cap-panel__h"><UserSearch :size="14" class="ic" aria-hidden="true" style="margin-right: 8px" /><span class="cap-panel__title">{{ t('ident.title') }}</span><span class="cap-chip" style="margin-left: auto"><LogIn :size="12" aria-hidden="true" />{{ t('ident.listening') }}</span></header>
    <div class="cap-panel__b" style="display: flex; flex-direction: column; gap: 12px">
      <ol class="cap-dim" style="margin: 0; padding-left: 20px; font-size: 14px; color: var(--cap-ink-2); display: flex; flex-direction: column; gap: 4px">
        <li>{{ t('ident.step1') }}</li>
        <li>{{ t('ident.step2') }} <InfoTip :text="t('ident.hint')" /></li>
      </ol>
      <span v-if="state.identify.ambiguous" class="cap-hint" style="align-items: center; gap: 6px; color: var(--cap-warn)"><TriangleAlert :size="14" aria-hidden="true" />{{ t('ident.ambiguous') }}</span>
      <span v-if="!state.identify.found.length" class="cap-dim" style="font-size: 13px">{{ t('ident.empty') }}</span>
      <ul v-else style="list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 8px">
        <li v-for="f in state.identify.found" :key="f.charId" style="display: flex; align-items: center; gap: 10px; flex-wrap: wrap">
          <Portrait :id="f.charId" :initials="initialsOf(f.charId)" />
          <span style="color: var(--cap-ink); font-weight: 700">{{ nameOf(f.charId) }}</span>
          <span class="cap-dim" style="font-size: 13px">{{ accountLabel(f.userId) }}</span>
          <span class="cap-chip" :class="{ 'cap-chip--warn': f.remapped }"><Check :size="12" aria-hidden="true" />{{ f.remapped ? t('ident.changed') : t('ident.found') }}</span>
        </li>
      </ul>
      <div style="display: flex; justify-content: flex-end"><button class="cap-btn" @click="stopIdentify()">{{ t('ident.done') }}</button></div>
    </div>
  </section>
</template>

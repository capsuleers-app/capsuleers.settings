<script setup lang="ts">
import { Window } from '@wailsio/runtime'
import emblem from '../assets/emblem.png'
import { t } from '../i18n'
import { clientRunning, profileMeta, server, state } from '../store'
</script>
<template>
  <header class="titlebar">
    <img :src="emblem" alt="" style="width: 22px; height: 25px; object-fit: contain">
    <span class="cap-title" style="font-size: 15px">CAPSULEERS<span style="color: var(--cap-accent)">.SETTINGS</span></span>
    <div class="titlebar__ctl">
      <span v-if="clientRunning" class="cap-chip cap-chip--crit"><span class="cap-dot"></span>{{ t('client.running') }}</span>
      <span v-else class="cap-chip cap-chip--ok"><span class="cap-dot"></span>{{ t('client.closed') }}</span>
      <span v-if="state.view !== 'profiles' && profileMeta" class="cap-chip">{{ server?.name }} &middot; {{ profileMeta.name }}</span>
      <button class="cap-icon-btn" :aria-label="t('win.min')" @click="Window.Minimise()"><svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M2 9h8" /></svg></button>
      <button class="cap-icon-btn" :aria-label="t('win.max')" @click="Window.ToggleMaximise()"><svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><rect x="2.5" y="2.5" width="7" height="7" /></svg></button>
      <button class="cap-icon-btn" :aria-label="t('win.close')" @click="Window.Close()"><svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M3 3l6 6M9 3l-6 6" /></svg></button>
    </div>
  </header>
</template>

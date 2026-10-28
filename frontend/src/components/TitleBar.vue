<script setup lang="ts">
import { Window } from '@wailsio/runtime'
import { CircleCheck, Minus, Square, TriangleAlert, X } from 'lucide-vue-next'
import emblem from '../assets/emblem.png'
import { t } from '../i18n'
import { clientRunning, profileMeta, server, state } from '../store'
</script>
<template>
  <header class="titlebar">
    <img :src="emblem" alt="" style="width: 22px; height: 25px; object-fit: contain">
    <span class="cap-title" style="font-size: 15px">CAPSULEERS<span style="color: var(--cap-accent)">.SETTINGS</span></span>
    <div class="titlebar__ctl">
      <span v-if="clientRunning" class="cap-chip cap-chip--crit"><TriangleAlert :size="12" aria-hidden="true" />{{ t('client.running') }}</span>
      <span v-else class="cap-chip cap-chip--ok"><CircleCheck :size="12" aria-hidden="true" />{{ t('client.closed') }}</span>
      <span v-if="state.view !== 'profiles' && profileMeta" class="cap-chip">{{ server?.name }} &middot; {{ profileMeta.name }}</span>
      <button class="cap-icon-btn" :aria-label="t('win.min')" @click="Window.Minimise()"><Minus :size="14" aria-hidden="true" /></button>
      <button class="cap-icon-btn" :aria-label="t('win.max')" @click="Window.ToggleMaximise()"><Square :size="12" aria-hidden="true" /></button>
      <button class="cap-icon-btn" :aria-label="t('win.close')" @click="Window.Close()"><X :size="14" aria-hidden="true" /></button>
    </div>
  </header>
</template>

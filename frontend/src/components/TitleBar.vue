<script setup lang="ts">
import { Window } from '@wailsio/runtime'
import { Minus, Square, X } from 'lucide-vue-next'
import { t } from '../i18n'
import { clientRunning, profileMeta, server } from '../store'
</script>
<template>
  <header class="titlebar">
    <!-- Emblema Settings (docs/brand/emblem-mark-settings.svg) ricolorato dall'accento via currentColor. -->
    <svg class="mark" width="20" height="20" viewBox="-24 20 560 560" aria-hidden="true"><path d="M176 40 L336 40 L408 150 L388 184 L418 184 L418 376 L256 560 L94 376 L94 184 L124 184 L104 150 Z" fill="#202027" /><g fill="none" stroke="currentColor" stroke-linejoin="round" stroke-linecap="round"><path stroke-width="13" d="M212 110 L300 110 L336 198 L300 284 L212 284 L176 198 Z M176 40 L212 110 M336 40 L300 110 M176 198 L124 184 M336 198 L388 184 M124 184 L182 322 M388 184 L330 322 M212 284 L182 322 M300 284 L330 322 M182 322 L330 322 M182 322 L94 376 M330 322 L418 376 M226 322 L256 560 M286 322 L256 560 M182 322 L212 510 M330 322 L300 510" /><path stroke-width="26" d="M176 40 L336 40 L408 150 L388 184 L418 184 L418 376 L256 560 L94 376 L94 184 L124 184 L104 150 Z" /></g></svg>
    <span class="titlebar__wm">CAPSULEERS<span>.SETTINGS</span></span>
    <span class="titlebar__sep" />
    <span class="titlebar__st">
      <span v-if="clientRunning" class="chip2 crit"><i class="d" />{{ t('client.running') }}</span>
      <span v-else class="chip2 ok"><i class="d" />{{ t('client.closed') }}</span>
      <span v-if="profileMeta" class="chip2">{{ server?.name }} &middot; {{ profileMeta.name.replace(/^settings_/, '') }}</span>
    </span>
    <div class="titlebar__ctl">
      <button :aria-label="t('win.min')" @click="Window.Minimise()"><Minus :size="16" aria-hidden="true" /></button>
      <button :aria-label="t('win.max')" @click="Window.ToggleMaximise()"><Square :size="14" aria-hidden="true" /></button>
      <button :aria-label="t('win.close')" @click="Window.Close()"><X :size="16" aria-hidden="true" /></button>
    </div>
  </header>
</template>

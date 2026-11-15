<script setup lang="ts">
import { Archive, ClipboardCheck, Eye, FolderOpen, Server, SlidersHorizontal, Users } from 'lucide-vue-next'
import type { Component } from 'vue'
import { t } from '../i18n'
import { go, profileMeta, server, state, type View } from '../store'
const sync: [View, 'nav.profiles' | 'nav.characters' | 'nav.preview' | 'nav.review', Component][] = [['profiles', 'nav.profiles', FolderOpen], ['characters', 'nav.characters', Users], ['preview', 'nav.preview', Eye], ['review', 'nav.review', ClipboardCheck]]
</script>
<template>
  <nav class="rail" :aria-label="t('nav.sections')">
    <span class="rail__grp">{{ t('nav.sync') }}</span>
    <a v-for="[v, k, ic] in sync" :key="v" class="rail__link" :class="{ 'is-active': state.view === v }" href="#" :aria-current="state.view === v ? 'page' : undefined" @click.prevent="go(v)"><component :is="ic" :size="16" class="ic" aria-hidden="true" />{{ t(k) }}</a>
    <span class="rail__grp">{{ t('nav.safety') }}</span>
    <a class="rail__link" :class="{ 'is-active': state.view === 'backups' }" href="#" :aria-current="state.view === 'backups' ? 'page' : undefined" @click.prevent="go('backups')"><Archive :size="16" class="ic" aria-hidden="true" />{{ t('nav.backups') }}<b v-if="state.backups.length" class="num">{{ state.backups.length }}</b></a>
    <div v-if="profileMeta" class="rail__foot">
      <span class="h">{{ t('nav.active') }}</span>
      <div class="rail__kv"><Server :size="14" class="ic" aria-hidden="true" />{{ t('kv.server') }}<b>{{ server?.name }}</b></div>
      <div class="rail__kv"><SlidersHorizontal :size="14" class="ic" aria-hidden="true" />{{ t('kv.profile') }}<b>{{ profileMeta.name.replace(/^settings_/, '') }}</b></div>
      <div class="rail__kv"><Users :size="14" class="ic" aria-hidden="true" />{{ t('kv.characters') }}<b>{{ profileMeta.chars.length }}</b></div>
    </div>
  </nav>
</template>

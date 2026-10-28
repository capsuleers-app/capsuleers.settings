<script setup lang="ts">
import { Archive, ClipboardCheck, Eye, FolderOpen, Server, SlidersHorizontal, Users } from 'lucide-vue-next'
import type { Component } from 'vue'
import { t } from '../i18n'
import { go, profileMeta, server, state, type View } from '../store'
const sync: [View, 'nav.profiles' | 'nav.characters' | 'nav.preview' | 'nav.review', Component][] = [['profiles', 'nav.profiles', FolderOpen], ['characters', 'nav.characters', Users], ['preview', 'nav.preview', Eye], ['review', 'nav.review', ClipboardCheck]]
</script>
<template>
  <nav class="cap-rail shell__rail" style="flex: 1 1 200px; max-width: 240px" :aria-label="t('nav.sections')">
    <span class="cap-rail__grp">{{ t('nav.sync') }}</span>
    <a v-for="[v, k, ic] in sync" :key="v" class="cap-rail__link" :class="{ 'is-active': state.view === v }" href="#" :aria-current="state.view === v ? 'page' : undefined" @click.prevent="go(v)"><component :is="ic" :size="16" class="ic" aria-hidden="true" />{{ t(k) }}</a>
    <span class="cap-rail__grp">{{ t('nav.safety') }}</span>
    <a class="cap-rail__link" :class="{ 'is-active': state.view === 'backups' }" href="#" :aria-current="state.view === 'backups' ? 'page' : undefined" @click.prevent="go('backups')"><Archive :size="16" class="ic" aria-hidden="true" />{{ t('nav.backups') }}</a>
    <div v-if="profileMeta" class="rail__foot">
      <span class="cap-rail__grp">{{ t('nav.active') }}</span>
      <div style="padding: 0 8px">
        <div class="cap-kv"><span class="cap-kv__k kvk"><Server :size="14" class="ic" aria-hidden="true" />{{ t('kv.server') }}</span><span class="cap-kv__v">{{ server?.name }}</span></div>
        <div class="cap-kv"><span class="cap-kv__k kvk"><SlidersHorizontal :size="14" class="ic" aria-hidden="true" />{{ t('kv.profile') }}</span><span class="cap-kv__v">{{ profileMeta.name.replace(/^settings_/, '') }}</span></div>
        <div class="cap-kv"><span class="cap-kv__k kvk"><Users :size="14" class="ic" aria-hidden="true" />{{ t('kv.characters') }}</span><span class="cap-kv__v">{{ profileMeta.chars.length }}</span></div>
      </div>
    </div>
  </nav>
</template>

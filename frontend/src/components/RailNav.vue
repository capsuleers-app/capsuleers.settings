<script setup lang="ts">
import { t } from '../i18n'
import { go, profileMeta, server, state, type View } from '../store'
const sync: [View, 'nav.profiles' | 'nav.characters' | 'nav.review'][] = [['profiles', 'nav.profiles'], ['characters', 'nav.characters'], ['review', 'nav.review']]
</script>
<template>
  <nav class="cap-rail" style="flex: 1 1 200px; max-width: 240px" :aria-label="t('nav.sections')">
    <span class="cap-rail__grp">{{ t('nav.sync') }}</span>
    <a v-for="[v, k] in sync" :key="v" class="cap-rail__link" :class="{ 'is-active': state.view === v }" href="#" :aria-current="state.view === v ? 'page' : undefined" @click.prevent="go(v)">{{ t(k) }}</a>
    <span class="cap-rail__grp">{{ t('nav.safety') }}</span>
    <a class="cap-rail__link" :class="{ 'is-active': state.view === 'backups' }" href="#" :aria-current="state.view === 'backups' ? 'page' : undefined" @click.prevent="go('backups')">{{ t('nav.backups') }}</a>
    <template v-if="profileMeta">
      <span class="cap-rail__grp">{{ t('nav.active') }}</span>
      <div style="padding: 0 8px">
        <div class="cap-kv"><span class="cap-kv__k">{{ t('kv.server') }}</span><span class="cap-kv__v">{{ server?.name }}</span></div>
        <div class="cap-kv"><span class="cap-kv__k">{{ t('kv.profile') }}</span><span class="cap-kv__v">{{ profileMeta.name.replace(/^settings_/, '') }}</span></div>
        <div class="cap-kv"><span class="cap-kv__k">{{ t('kv.characters') }}</span><span class="cap-kv__v">{{ profileMeta.chars.length }}</span></div>
      </div>
    </template>
  </nav>
</template>

<script setup lang="ts">
import PageHead from '../components/PageHead.vue'
import { fmtDate, t } from '../i18n'
import { addFolder, go, install, profileMeta, removeFolder, scan, selectInstall, selectProfile, selectServer, server, state } from '../store'
</script>
<template>
  <div class="stack" style="flex: 999 1 560px">
    <PageHead eyebrow="profiles.eyebrow" title="profiles.title" lead="profiles.lead">
      <button class="cap-btn" @click="scan()">{{ t('profiles.rescan') }}</button>
      <button class="cap-btn" @click="addFolder()">{{ t('profiles.add') }}</button>
    </PageHead>

    <section v-if="!state.installs.length && !state.loading" class="cap-empty">
      <span class="cap-title" style="font-size: 20px">{{ t('profiles.emptyTitle') }}</span>
      <span style="max-width: 460px">{{ t('profiles.emptyBody') }}</span>
      <button class="cap-btn cap-btn--pri" @click="addFolder()">{{ t('profiles.add') }}</button>
    </section>

    <template v-else>
      <section class="cap-panel">
        <header class="cap-panel__h"><span class="cap-panel__title">{{ t('profiles.installs') }}</span><span style="margin-left: auto">{{ t('profiles.found', { n: state.installs.length }) }}</span></header>
        <div class="cap-panel__b" style="display: grid; grid-template-columns: repeat(auto-fill, minmax(min(100%, 260px), 1fr)); gap: 12px">
          <div v-for="it in state.installs" :key="it.id" style="display: flex; flex-direction: column">
            <button class="cap-tool" :class="{ 'is-sel': it.id === state.installId }" style="text-align: left; cursor: pointer; font: inherit; color: inherit; flex: 1" :aria-pressed="it.id === state.installId" @click="selectInstall(it.id)">
              <span style="display: flex; align-items: center; gap: 8px"><span class="cap-tool__name">{{ it.name }}</span><span class="cap-chip" style="margin-left: auto; padding: 3px 6px">{{ it.kind }}</span></span>
              <span class="cap-tool__desc" style="font-size: 12px; color: var(--cap-ink-3); word-break: break-all">{{ it.path }}</span>
              <span class="cap-dim" style="font-size: 12px">{{ t('profiles.meta', { servers: it.servers.length, profiles: it.servers.reduce((n, s) => n + s.profiles.length, 0) }) }}</span>
            </button>
            <button v-if="it.manual" class="cap-btn" style="margin-top: 4px" @click="removeFolder(it.path)">{{ t('profiles.removeFolder') }}</button>
          </div>
        </div>
      </section>

      <section v-if="install" class="cap-panel">
        <nav style="display: flex; border-bottom: 1px solid var(--cap-line)" :aria-label="t('profiles.server')">
          <a v-for="s in install.servers" :key="s.dir" class="cap-surface-tab" :class="{ 'is-active': s.dir === state.serverDir }" href="#" @click.prevent="selectServer(s.dir)">{{ s.name }}</a>
        </nav>
        <div style="overflow-x: auto">
          <table class="cap-table" style="width: 100%; border-collapse: collapse">
            <thead><tr>
              <th style="width: 36px"><span class="sr">{{ t('profiles.th.selected') }}</span></th>
              <th>{{ t('profiles.th.profile') }}</th><th class="text-right">{{ t('profiles.th.chars') }}</th><th class="text-right">{{ t('profiles.th.accounts') }}</th>
              <th>{{ t('profiles.th.saved') }}</th><th>{{ t('profiles.th.folder') }}</th>
            </tr></thead>
            <tbody>
              <tr v-for="p in server?.profiles" :key="p.dir" :class="{ 'is-target': p.dir === state.profileDir }">
                <td><input class="cap-input" type="radio" name="profile" :checked="p.dir === state.profileDir" :aria-label="t('profiles.use', { name: p.name })" @change="selectProfile(p.dir)"></td>
                <td style="color: var(--cap-ink); font-weight: 700">{{ p.name }}</td>
                <td class="num" style="text-align: right">{{ p.chars.length }}</td>
                <td class="num" style="text-align: right">{{ p.users.length }}</td>
                <td class="num">{{ fmtDate(p.saved) }}</td>
                <td class="cap-dim" style="font-size: 12px; word-break: break-all">{{ server?.dir.split(/[\\/]/).pop() }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section v-if="profileMeta" class="cap-panel bar">
        <span style="font-size: 14px; color: var(--cap-ink)">{{ t('profiles.using', { profile: profileMeta.name, install: install?.name ?? '' }) }}</span>
        <span class="cap-hint">{{ t('profiles.hint') }}</span>
        <div style="margin-left: auto"><button class="cap-btn cap-btn--pri" @click="go('characters')">{{ t('profiles.open') }}</button></div>
      </section>
    </template>
  </div>
</template>

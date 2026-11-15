<script setup lang="ts">
import ScrollTable from '../components/ScrollTable.vue'
import { FolderOpen, FolderPlus, FolderSearch, FolderX, HardDrive, RefreshCw, Users } from 'lucide-vue-next'
import InfoTip from '../components/InfoTip.vue'
import PageHead from '../components/PageHead.vue'
import { fmtDate, t } from '../i18n'
import { addFolder, go, install, profileMeta, removeFolder, scan, selectInstall, selectProfile, selectServer, server, state } from '../store'
</script>
<template>
  <div class="stack">
    <PageHead eyebrow="profiles.eyebrow" title="profiles.title" lead="profiles.lead" :icon="FolderOpen">
      <button v-if="state.installs.length" class="cap-btn" @click="addFolder()"><FolderPlus :size="14" aria-hidden="true" />{{ t('profiles.add') }}</button>
      <button v-if="state.installs.length" class="cap-btn" @click="scan()"><RefreshCw :size="14" aria-hidden="true" />{{ t('profiles.rescan') }}</button>
    </PageHead>

    <section v-if="!state.installs.length && !state.loading" class="cap-empty">
      <FolderSearch :size="32" class="ic ic-lg" aria-hidden="true" />
      <span class="cap-title" style="font-size: 20px">{{ t('profiles.emptyTitle') }}</span>
      <span style="max-width: 460px">{{ t('profiles.emptyBody') }}</span>
      <button class="cap-btn cap-btn--pri" @click="addFolder()"><FolderPlus :size="14" aria-hidden="true" />{{ t('profiles.add') }}</button>
    </section>

    <template v-else>
      <div class="sc">
        <div class="sec first" :aria-label="t('profiles.installs')">
          <div v-for="it in state.installs" :key="it.id" class="inst" :class="{ on: it.id === state.installId }">
            <button class="ins" :aria-pressed="it.id === state.installId" @click="selectInstall(it.id)">
              <span class="iic"><HardDrive :size="22" class="ic" aria-hidden="true" /></span>
              <span class="who"><b>{{ it.name }} <span class="tag" style="margin-left: 6px">{{ it.kind }}</span></b><p>{{ it.path }}</p></span>
              <span class="big" style="margin-left: auto; gap: 32px"><span class="k"><b class="num">{{ it.servers.length }}</b><span>{{ t('profiles.server') }}</span></span><span class="k"><b class="num">{{ it.servers.reduce((n, s) => n + s.profiles.length, 0) }}</b><span>{{ t('profiles.th.profile') }}</span></span></span>
            </button>
            <button v-if="it.manual" class="ib" :aria-label="t('profiles.removeFolder')" :title="t('profiles.removeFolder')" @click="removeFolder(it.path)"><FolderX :size="14" aria-hidden="true" /></button>
          </div>
        </div>

        <div v-if="install" class="sec grow" style="margin-top: 24px; padding-top: 0; border-top: 0; display: flex; flex-direction: column; min-height: 0">
          <nav class="tbs" :aria-label="t('profiles.server')">
            <a v-for="s in install.servers" :key="s.dir" :class="{ on: s.dir === state.serverDir }" href="#" :aria-current="s.dir === state.serverDir ? 'page' : undefined" @click.prevent="selectServer(s.dir)">{{ s.name }} <b class="num">{{ s.profiles.length }}</b></a>
          </nav>
          <ScrollTable style="flex: 1 1 0; min-height: 0; margin-top: 14px" :cols="['40px', '', '140px', '140px', '190px', '180px']">
            <template #head><tr>
              <th><span class="sr">{{ t('profiles.th.selected') }}</span></th>
              <th>{{ t('profiles.th.profile') }}</th><th class="r">{{ t('profiles.th.chars') }}</th><th class="r">{{ t('profiles.th.accounts') }}</th>
              <th>{{ t('profiles.th.saved') }}</th><th>{{ t('profiles.th.folder') }}</th>
            </tr></template>
              <tr v-for="p in server?.profiles" :key="p.dir" :class="{ 'is-target': p.dir === state.profileDir }">
                <td><input class="cap-input" type="radio" name="profile" :checked="p.dir === state.profileDir" :aria-label="t('profiles.use', { name: p.name })" @change="selectProfile(p.dir)"></td>
                <td style="color: var(--cap-ink); font-weight: 700">{{ p.name.replace(/^settings_/, '') }}</td>
                <td class="num r">{{ p.chars.length }}</td>
                <td class="num r">{{ p.users.length }}</td>
                <td class="num mut">{{ fmtDate(p.saved) }}</td>
                <td class="mut" style="font-size: 12px; word-break: break-all">{{ p.name }}</td>
              </tr>
          </ScrollTable>
        </div>
      </div>

      <div v-if="profileMeta" class="fb">
        <span>{{ t('profiles.using', { profile: profileMeta.name, install: install?.name ?? '' }) }}</span>
        <InfoTip :text="t('profiles.hint')" />
        <div class="sp"><button class="cap-btn cap-btn--pri" @click="go('characters')">{{ t('profiles.open') }}<Users :size="14" aria-hidden="true" /></button></div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.inst { display: flex; align-items: center; gap: 8px; border-left: 2px solid transparent; padding: 8px 0 8px 12px; }
.inst + .inst { border-top: 1px solid var(--cap-line-soft); }
.inst.on { border-left-color: var(--cap-accent); }
.inst .ins { flex: 1; min-width: 0; background: transparent; border: 0; padding: 0; color: inherit; font: inherit; text-align: left; cursor: pointer; }
.inst .ins:focus-visible { outline: 2px solid var(--cap-accent); outline-offset: 4px; }
</style>

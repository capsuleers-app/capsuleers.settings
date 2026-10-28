<script setup lang="ts">
import { onMounted } from 'vue'
import RailNav from './components/RailNav.vue'
import TitleBar from './components/TitleBar.vue'
import BackupsView from './views/BackupsView.vue'
import CharactersView from './views/CharactersView.vue'
import PreviewView from './views/PreviewView.vue'
import ProfilesView from './views/ProfilesView.vue'
import ReviewView from './views/ReviewView.vue'
import { init, state } from './store'

onMounted(init)
</script>
<template>
  <div class="shell">
    <TitleBar />
    <div class="shell__body">
      <RailNav />
      <ProfilesView v-if="state.view === 'profiles'" />
      <CharactersView v-else-if="state.view === 'characters'" />
      <PreviewView v-else-if="state.view === 'preview'" />
      <ReviewView v-else-if="state.view === 'review'" />
      <BackupsView v-else />
    </div>
  </div>
</template>

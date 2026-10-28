import { computed, reactive } from 'vue'
import * as Svc from '../bindings/github.com/WilliamFalci/capsuleers.settings/service'
import { ApplyRequest } from '../bindings/github.com/WilliamFalci/capsuleers.settings/internal/backup/models'
import type { Info as BackupInfo } from '../bindings/github.com/WilliamFalci/capsuleers.settings/internal/backup/models'
import type { Install, Profile, File as EveFile, Server } from '../bindings/github.com/WilliamFalci/capsuleers.settings/internal/eve/models'
import { t } from './i18n'

export type View = 'profiles' | 'characters' | 'review' | 'backups'
export const KEEP_BACKUPS = 20

export interface Row {
  id: string
  name: string
  initials: string
  file: EveFile
}

export const state = reactive({
  view: 'profiles' as View,
  loading: true,
  installs: [] as Install[],
  installId: '',
  serverDir: '',
  profileDir: '',
  profile: null as Profile | null,
  names: {} as Record<string, string>,
  namesOffline: false,
  sourceId: '',
  targets: {} as Record<string, boolean>,
  copyUser: false,
  running: 0,
  applying: false,
  done: null as { backup: string; count: number; sourceName: string } | null,
  error: '',
  backups: [] as BackupInfo[],
  backupId: '',
  restored: {} as Record<string, number>,
  backupError: '',
})

export const install = computed<Install | undefined>(() => state.installs.find((i) => i.id === state.installId))
export const server = computed<Server | undefined>(() => install.value?.servers.find((s) => s.dir === state.serverDir))
export const profileMeta = computed(() => server.value?.profiles.find((p) => p.dir === state.profileDir))

export function nameOf(id: string): string {
  return state.names[id] || t('char.unnamed', { id })
}
export function initialsOf(id: string): string {
  const n = state.names[id]
  if (!n) return id.slice(-2)
  const w = n.split(/\s+/).filter(Boolean)
  return (w.length > 1 ? w[0][0] + w[1][0] : n.slice(0, 2)).toUpperCase()
}

export const rows = computed<Row[]>(() =>
  (state.profile?.chars ?? []).map((f) => ({ id: f.id, name: nameOf(f.id), initials: initialsOf(f.id), file: f })),
)
export const source = computed(() => rows.value.find((r) => r.id === state.sourceId))
export const targetRows = computed(() => rows.value.filter((r) => r.id !== state.sourceId && state.targets[r.id]))
export const clientRunning = computed(() => state.running > 0)

async function loadNames(ids: string[]) {
  if (!ids.length) return
  Object.assign(state.names, await Svc.CachedNames(ids))
  const got = await Svc.ResolveNames(ids)
  Object.assign(state.names, got)
  state.namesOffline = ids.some((id) => !state.names[id])
}

function resetSelection() {
  state.targets = {}
  state.done = null
  state.error = ''
  const chars = state.profile?.chars ?? []
  state.sourceId = chars.length ? [...chars].sort((a, b) => b.saved.localeCompare(a.saved))[0].id : ''
}

export async function selectProfile(dir: string) {
  state.profileDir = dir
  state.profile = dir ? await Svc.Profile(dir) : null
  resetSelection()
  void loadNames((state.profile?.chars ?? []).map((c) => c.id))
}

export async function selectInstall(id: string, keepServer = false) {
  state.installId = id
  const inst = install.value
  if (!keepServer || !inst?.servers.some((s) => s.dir === state.serverDir)) state.serverDir = inst?.servers[0]?.dir ?? ''
  const prof = server.value?.profiles.find((p) => p.dir === state.profileDir) ?? server.value?.profiles[0]
  await selectProfile(prof?.dir ?? '')
}

export async function selectServer(dir: string) {
  state.serverDir = dir
  await selectProfile(server.value?.profiles[0]?.dir ?? '')
}

export async function scan() {
  state.loading = true
  try {
    state.installs = await Svc.Installs()
    const keep = state.installs.some((i) => i.id === state.installId)
    await selectInstall(keep ? state.installId : (state.installs[0]?.id ?? ''), keep)
  } finally {
    state.loading = false
  }
}

export async function rescanProfile() {
  if (!state.profileDir) return
  const keepSource = state.sourceId
  const keepTargets = { ...state.targets }
  state.profile = await Svc.Profile(state.profileDir)
  state.sourceId = state.profile.chars.some((c) => c.id === keepSource) ? keepSource : (state.profile.chars[0]?.id ?? '')
  state.targets = keepTargets
  void loadNames(state.profile.chars.map((c) => c.id))
}

export async function addFolder() {
  const p = await Svc.AddFolder()
  if (p) await scan()
}
export async function removeFolder(path: string) {
  await Svc.RemoveFolder(path)
  await scan()
}

export function pickSource(id: string) {
  state.sourceId = id
  delete state.targets[id]
  state.done = null
}
export function toggleTarget(id: string) {
  if (id === state.sourceId) return
  state.targets[id] = !state.targets[id]
  state.done = null
}

export async function refreshStatus() {
  try {
    state.running = await Svc.ClientStatus()
  } catch {
    /* keep last value */
  }
}

export function errText(code: string, msg: string): string {
  const key = `err.${code}` as Parameters<typeof t>[0]
  const base = ['client_running', 'bad_request', 'bad_backup', 'io'].includes(code) ? t(key) : code
  return msg && code !== 'client_running' ? `${base} (${msg})` : base
}

export async function apply() {
  if (state.applying || clientRunning.value || !source.value || !targetRows.value.length) return
  state.applying = true
  state.error = ''
  try {
    const res = await Svc.Apply(new ApplyRequest({
      profileDir: state.profileDir,
      sourceId: state.sourceId,
      targetIds: targetRows.value.map((r) => r.id),
    }))
    if (res.ok) {
      state.done = { backup: res.backup, count: res.updated.length, sourceName: source.value.name }
      await rescanProfile()
    } else {
      state.error = errText(res.code, res.message)
      if (res.code === 'client_running') await refreshStatus()
    }
  } catch (e) {
    state.error = String(e)
  } finally {
    state.applying = false
  }
}

export async function loadBackups() {
  state.backups = (await Svc.Backups()) ?? []
  if (!state.backups.some((b) => b.name === state.backupId)) state.backupId = state.backups[0]?.name ?? ''
}

export async function restore(name: string) {
  state.backupError = ''
  const res = await Svc.Restore(name)
  if (res.ok) {
    state.restored[name] = res.updated.length
    await loadBackups()
    await rescanProfile()
  } else {
    state.backupError = errText(res.code, res.message)
    if (res.code === 'client_running') await refreshStatus()
  }
}

export async function deleteBackup(name: string) {
  await Svc.DeleteBackup(name)
  await loadBackups()
}

export async function go(v: View) {
  state.view = v
  if (v === 'backups') await loadBackups()
}

export async function init() {
  await Promise.all([scan(), refreshStatus()])
  setInterval(refreshStatus, 3000)
}

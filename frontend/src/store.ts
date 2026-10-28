import { computed, reactive } from 'vue'
import * as Svc from '../bindings/github.com/capsuleers-app/capsuleers.settings/service'
import { ApplyRequest } from '../bindings/github.com/capsuleers-app/capsuleers.settings/internal/backup/models'
import type { Info as BackupInfo } from '../bindings/github.com/capsuleers-app/capsuleers.settings/internal/backup/models'
import type { Install, Profile, File as EveFile, Server } from '../bindings/github.com/capsuleers-app/capsuleers.settings/internal/eve/models'
import type { AccountRow, AccountsResult, IdentFound, PreviewResult } from '../bindings/github.com/capsuleers-app/capsuleers.settings/models'
import type { Plan as AccountPlan } from '../bindings/github.com/capsuleers-app/capsuleers.settings/internal/accounts/models'
import { locale, t } from './i18n'

export type View = 'profiles' | 'characters' | 'preview' | 'review' | 'backups'
export type PreviewTab = 'layout' | 'chat' | 'keys'
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
  copyUser: true,
  accounts: null as AccountsResult | null,
  plan: null as AccountPlan | null,
  identify: { open: false, found: [] as IdentFound[], ambiguous: 0, seq: 0, error: '' },
  running: 0,
  applying: false,
  done: null as { backup: string; count: number; accounts: number; sourceName: string } | null,
  error: '',
  backups: [] as BackupInfo[],
  backupId: '',
  restored: {} as Record<string, number>,
  backupError: '',
  previewTab: 'layout' as PreviewTab,
  previewSel: '',
  previewCats: [] as string[], // selected window categories; empty = all (session only)
  preview: null as PreviewResult | null,
  previewLoading: false,
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

/** Account (core_user id) of each character, from certain evidence only; '' = to identify. */
export const accountRows = computed<Record<string, AccountRow>>(() =>
  Object.fromEntries((state.accounts?.chars ?? []).map((r) => [r.charId, r])),
)
export const accountOf = (id: string) => accountRows.value[id]?.userId ?? ''
export const sourceAccount = computed(() => accountOf(state.sourceId))
/** The account copy is on only while the source's account is known. */
export const copyUserOn = computed(() => state.copyUser && !!sourceAccount.value)
/** Characters sharing one account file, for the review. */
export function charsOnAccount(userId: string): string[] {
  return (state.accounts?.chars ?? []).filter((r) => r.userId === userId).map((r) => r.charId)
}

/** Characters on an account (confirmed or high-confidence), source first then alphabetical. */
export function accountCharNames(userId: string): string[] {
  const ids = charsOnAccount(userId)
  return ids
    .map((id) => ({ id, name: nameOf(id) }))
    .sort((a, b) => (a.id === state.sourceId ? -1 : b.id === state.sourceId ? 1 : a.name.localeCompare(b.name)))
    .map((x) => x.name)
}
/** "A, B +N": at most two names. */
function shortNames(names: string[]): string {
  return names.slice(0, 2).join(', ') + (names.length > 2 ? ` +${names.length - 2}` : '')
}
export const accountAlias = (userId: string): string => state.accounts?.aliases?.[userId] ?? ''
/** Display label: alias, else the characters' names, else "Account <id>". */
export function accountLabel(userId: string): string {
  const alias = accountAlias(userId)
  if (alias) return alias
  const names = accountCharNames(userId)
  return names.length ? shortNames(names) : t('acct.label', { id: userId })
}
/** Tooltip: label, characters (when an alias hides them) and the numeric id. */
export function accountTitle(userId: string): string {
  const names = accountCharNames(userId)
  const parts = [accountLabel(userId)]
  if (accountAlias(userId) && names.length) parts.push(names.join(', '))
  parts.push(`ID ${userId}`)
  return parts.join('\n')
}

/** Name an account ('' removes the name). Stored in the app config only. */
export async function setAccountAlias(userId: string, alias: string) {
  await Svc.SetAccountAlias(state.profileDir, userId, alias)
  await loadAccounts()
}

async function loadAccounts() {
  const dir = state.profileDir
  const res = dir ? await Svc.Accounts(dir) : null
  if (dir === state.profileDir) state.accounts = res
}

export async function loadPlan() {
  if (!copyUserOn.value || !state.profileDir || !targetRows.value.length) {
    state.plan = null
    return
  }
  const req = new ApplyRequest({ profileDir: state.profileDir, sourceId: state.sourceId, targetIds: targetRows.value.map((r) => r.id), copyUser: true })
  state.plan = await Svc.AccountPlan(req)
}

export interface AccountGroup {
  userId: string
  label: string
  rows: Row[]
}
/** Characters grouped by account (sorted by label); `unknown` = no certain account yet. */
export const accountGroups = computed(() => {
  const by = new Map<string, Row[]>()
  const unknown: Row[] = []
  for (const r of rows.value) {
    const u = accountOf(r.id)
    if (!u) unknown.push(r)
    else by.set(u, [...(by.get(u) ?? []), r])
  }
  const groups: AccountGroup[] = [...by].map(([userId, rs]) => ({ userId, label: accountLabel(userId), rows: rs }))
  groups.sort((a, b) => a.label.localeCompare(b.label))
  return { groups, unknown }
})

let identTimer: ReturnType<typeof setInterval> | undefined
async function pollIdentify() {
  if (!state.identify.open) return
  try {
    const st = await Svc.IdentifyPoll()
    if (!state.identify.open) return
    state.identify.found = st.found
    state.identify.ambiguous = st.ambiguous
    if (st.seq !== state.identify.seq) {
      state.identify.seq = st.seq
      await loadAccounts()
      state.done = null
    }
  } catch {
    /* keep last state */
  }
}
/** Open the guided identification: the backend watches the profile files (read-only) and saves what it learns. */
export async function startIdentify() {
  if (!state.profileDir || state.identify.open) return
  state.identify.error = ''
  try {
    await Svc.IdentifyStart(state.profileDir)
  } catch (e) {
    state.identify.error = String(e)
    return
  }
  Object.assign(state.identify, { open: true, found: [], ambiguous: 0, seq: 0 })
  identTimer = setInterval(pollIdentify, 1000)
}
/** Stop it (panel closed, DONE, view or profile changed). */
export async function stopIdentify() {
  clearInterval(identTimer)
  if (!state.identify.open) return
  state.identify.open = false
  await Svc.IdentifyStop()
  await loadAccounts()
}

export function setCopyUser(on: boolean) {
  state.copyUser = on
  state.done = null
}

async function loadNames(ids: string[]) {
  if (!ids.length) return
  Object.assign(state.names, await Svc.CachedNames(ids))
  const got = await Svc.ResolveNames(ids)
  Object.assign(state.names, got)
  state.namesOffline = ids.some((id) => !state.names[id])
}

function resetSelection() {
  state.previewSel = ''
  state.preview = null
  state.targets = {}
  state.done = null
  state.error = ''
  state.copyUser = true
  state.plan = null
  const chars = state.profile?.chars ?? []
  state.sourceId = chars.length ? [...chars].sort((a, b) => b.saved.localeCompare(a.saved))[0].id : ''
}

export async function selectProfile(dir: string) {
  await stopIdentify()
  state.profileDir = dir
  state.profile = dir ? await Svc.Profile(dir) : null
  state.accounts = null
  resetSelection()
  // The last source picked in this profile wins over the newest-saved default.
  const last = dir ? await Svc.LastSource(dir) : ''
  if (last && state.profileDir === dir) state.sourceId = last
  void loadAccounts()
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
  await loadAccounts()
  if (state.view === 'preview') void loadPreview()
  if (state.view === 'review') await loadPlan()
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
  if (state.profileDir) void Svc.SetLastSource(state.profileDir, id).catch(() => {})
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
  const base = ['client_running', 'bad_request', 'bad_backup', 'io', 'unknown_account'].includes(code) ? t(key) : code
  return msg && code !== 'client_running' ? `${base} (${msg})` : base
}

export async function apply() {
  if (state.applying || clientRunning.value || !source.value || !targetRows.value.length) return
  if (copyUserOn.value && !planReady.value) return
  state.applying = true
  state.error = ''
  try {
    const res = await Svc.Apply(new ApplyRequest({
      profileDir: state.profileDir,
      sourceId: state.sourceId,
      targetIds: targetRows.value.map((r) => r.id),
      copyUser: copyUserOn.value,
    }))
    if (res.ok) {
      state.restored = {}
      state.done = { backup: res.backup, count: res.updated.length, accounts: res.accounts.length, sourceName: source.value.name }
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
    state.restored = { [name]: res.updated.length + res.accounts.length }
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

let previewSeq = 0
/** Read-only decode of the source; a stale answer never overwrites a newer one. */
export async function loadPreview() {
  const seq = ++previewSeq
  if (!state.profileDir || !state.sourceId) {
    state.preview = null
    return
  }
  state.previewLoading = true
  try {
    const res = await Svc.Preview(state.profileDir, state.sourceId, locale)
    if (seq !== previewSeq) return
    state.preview = res
    if (!res.windows.some((w) => w.id === state.previewSel)) state.previewSel = res.windows[0]?.id ?? ''
  } catch {
    if (seq === previewSeq) state.preview = null
  } finally {
    if (seq === previewSeq) state.previewLoading = false
  }
}

/** Open the preview of the current source. */
export async function openPreview() {
  await go('preview')
}

/** Account copy can run: plan loaded, source account known, every target on a known account. */
export const planReady = computed(() => !!state.plan && !!state.plan.sourceUser && state.plan.unknown.length === 0)

export async function go(v: View) {
  if (v !== 'characters') await stopIdentify()
  state.view = v
  if (v === 'review') await loadPlan()
  if (v === 'backups') await loadBackups()
  if (v === 'preview') await loadPreview()
}

export async function init() {
  await Promise.all([scan(), refreshStatus()])
  setInterval(refreshStatus, 3000)
}

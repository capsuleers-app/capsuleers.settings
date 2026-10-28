// Demo mode only (npm run demo): in-memory stand-in for the generated Wails service, used to take
// README screenshots. Every character, account and path below is invented; nothing is read from disk.
const NOW = Date.now()
const ago = (min: number) => new Date(NOW - min * 60_000).toISOString()
const PRE = '/home/pilot/.local/share/Steam/steamapps/compatdata/8500/pfx/drive_c/users/steamuser/AppData/Local/CCP/EVE'

// userId, alias, [charId, name, minutes since saved]
const ACCOUNTS: [string, string, [string, string, number][]][] = [
  ['810001', 'Main', [['9100000001', 'Vael Korrin', 12], ['9100000002', 'Sera Kovanth', 190], ['9100000003', 'Dren Alvaris', 2900]]],
  ['810002', 'Industry', [['9100000004', 'Kaelith Orne', 640], ['9100000005', 'Tobrix Maelen', 1500], ['9100000006', 'Jessa Varnholt', 4300]]],
  ['810003', 'Scouts', [['9100000007', 'Ryn Ostrava', 330], ['9100000008', 'Lio Thessaly', 7200]]],
]
const UNKNOWN: [string, string, number][] = [['9100000009', 'Marek Duvall', 980], ['9100000010', 'Nyx Talvera', 5100]]

const NAMES: Record<string, string> = {}
const userOf: Record<string, string> = {}
const aliases: Record<string, string> = {}
const chars: any[] = []
const users: any[] = []
const addChar = (id: string, name: string, min: number) => {
  NAMES[id] = name
  chars.push({ id, path: `${PRE}/c_ccp_eve_tq_tranquility/settings_Default/core_char_${id}.dat`, size: 41000 + (+id.slice(-2) * 2371) % 9000, saved: ago(min) })
}
for (const [uid, alias, cs] of ACCOUNTS) {
  aliases[uid] = alias
  users.push({ id: uid, path: `${PRE}/c_ccp_eve_tq_tranquility/settings_Default/core_user_${uid}.dat`, size: 118000 + +uid.slice(-1) * 5200, saved: ago(cs[0][2]) })
  for (const [id, name, min] of cs) { addChar(id, name, min); userOf[id] = uid }
}
for (const [id, name, min] of UNKNOWN) addChar(id, name, min)

const profile = (name: string, cs: any[], us: any[]) => ({ name, dir: `${PRE}/${name}`, chars: cs, users: us, saved: cs[0]?.saved ?? '' })
const mainProfile = { name: 'settings_Default', dir: `${PRE}/c_ccp_eve_tq_tranquility/settings_Default`, chars, users, saved: ago(12) }
const sisiChars = chars.slice(0, 2).map((c) => ({ ...c, path: c.path.replace('tranquility', 'singularity'), saved: ago(9000) }))
const sisiProfile = profile('settings_Default', sisiChars, [])
sisiProfile.dir = `${PRE}/c_ccp_eve_tq_singularity/settings_Default`
const installs = [{
  id: `${PRE}`, name: 'Steam', kind: 'Proton', path: PRE, manual: false,
  servers: [
    { name: 'Tranquility', dir: `${PRE}/c_ccp_eve_tq_tranquility`, profiles: [mainProfile] },
    { name: 'Singularity', dir: `${PRE}/c_ccp_eve_tq_singularity`, profiles: [sisiProfile] },
  ],
}]

const win = (id: string, label: string, category: string, left: number, top: number, width: number, height: number, extra: any = {}) =>
  ({ id, label, left, top, width, height, screenW: 2560, screenH: 1440, category, members: [], locked: false, collapsed: false, compact: false, minimized: false, ...extra })
const windows = [
  win('overview', 'Overview', 'combat', 1900, 60, 640, 700, { locked: true }),
  win('selecteditemview', 'Selected Item', 'combat', 1900, 770, 640, 320, { locked: true }),
  win('droneview', 'Drones', 'combat', 1900, 1100, 640, 300),
  win('directionalScannerWindow', 'Directional Scanner', 'combat', 1400, 740, 480, 260),
  win('probeScannerWindow', 'Probe Scanner', 'combat', 1400, 1020, 480, 380),
  win('chatchannel_local', 'Local', 'chat', 20, 1000, 620, 400),
  win('ChatWindowStack', 'Chat', 'chat', 650, 1000, 620, 400, { members: ['chatchannel_corp', 'chatchannel_alliance', 'chatchannel_fleet'] }),
  win('fleetwindow', 'Fleet', 'fleet', 20, 60, 420, 520),
  win('watchlistpanel', 'Watch List', 'fleet', 450, 60, 300, 260, { compact: true }),
  win('inventory', 'Inventory', 'ship', 20, 600, 720, 380),
  win('fitting', 'Fitting', 'ship', 760, 740, 620, 240, { collapsed: true }),
  win('marketbuyaction', 'Buy', 'market', 800, 200, 560, 520),
  win('wallet', 'Wallet', 'market', 450, 340, 330, 240),
  win('primary_map_panel', 'Map', 'navigation', 1400, 200, 480, 500),
].sort((a, b) => a.label.toLowerCase().localeCompare(b.label.toLowerCase()))
const channels = [
  { id: 'chatchannel_local', kind: 'local', label: 'Local' },
  { id: 'chatchannel_corp', kind: 'corp', label: 'Corp' },
  { id: 'chatchannel_alliance', kind: 'alliance', label: 'Alliance' },
  { id: 'chatchannel_fleet', kind: 'fleet', label: 'Fleet' },
  { id: 'chatchannel_player1', kind: 'player', label: 'Industry Chatter' },
  { id: 'chatchannel_player2', kind: 'player', label: 'Scout Intel' },
]
const keys = [
  ['windows', 'windowSizesAndPositions_1', 'map', 14], ['windows', 'openWindows', 'list', 14], ['windows', 'lockedWindows', 'list', 3],
  ['windows', 'stacksWindows', 'map', 4], ['windows', 'collapsedWindows', 'list', 1], ['ui', 'chatchannels', 'list', 6],
  ['ui', 'overviewScrollPos', 'number', 0], ['ui', 'neocomWidth', 'number', 0], ['ui', 'autoRejectInvitations', 'flag', 0],
  ['cmd', 'customCmds', 'map', 12], ['audio', 'masterVolume', 'number', 0], ['audio', 'uiEventsEnabled', 'flag', 0],
].map(([group, name, type, len]) => ({ group, name, type, len }))

const preview = (id: string) => ({
  ok: true, code: '',
  source: { charId: id, readable: true, error: '', windows: windows.length, stacks: 1, locked: 2, channels: channels.length, screenW: 2560, screenH: 1440, keys },
  display: { known: true, screenW: 2560, screenH: 1440, scaleFullscreen: 1, scaleWindowed: 1 },
  windows, channels,
})

const backups = [
  { name: 'apply', taken: ago(95), kind: 'apply', sourceName: 'Vael Korrin', profile: 'settings_Default', size: 548_000,
    entries: [['char', '9100000004'], ['char', '9100000005'], ['char', '9100000006'], ['user', '810002']] },
  { name: 'restore', taken: ago(3300), kind: 'restore', sourceName: '', profile: 'settings_Default', size: 233_000,
    entries: [['char', '9100000007'], ['char', '9100000008'], ['user', '810003']] },
  { name: 'apply', taken: ago(4800), kind: 'apply', sourceName: 'Sera Kovanth', profile: 'settings_Default', size: 162_000,
    entries: [['char', '9100000003'], ['user', '810001']] },
].map((b) => ({
  name: `${b.name}-${b.taken.slice(0, 19).replace(/[-:]/g, '').replace('T', '-')}.zip`, taken: b.taken, kind: b.kind, sourceName: b.sourceName, profile: b.profile, profileDir: mainProfile.dir, size: b.size,
  files: b.entries.map(([kind, id], i) => ({
    zip: `${kind}_${id}.dat`, original: `${mainProfile.dir}/core_${kind}_${id}.dat`, kind, charId: kind === 'char' ? id : '', userId: kind === 'user' ? id : '',
    name: kind === 'char' ? NAMES[id] : '', size: Math.round(b.size / b.entries.length) - i * 700,
  })),
}))

let lastSource = ''
let identPolls = 0
let identOn = false
const done = <T>(v: T): Promise<T> => Promise.resolve(JSON.parse(JSON.stringify(v ?? null)))

export const Meta = () => done({ version: 'demo', repo: 'https://github.com/WilliamFalci/capsuleers.settings' })
export const Installs = () => done(installs)
export const Profile = (dir: string) => done(installs.flatMap((i) => i.servers.flatMap((s) => s.profiles)).find((p) => p.dir === dir) ?? mainProfile)
export const AddFolder = () => done('')
export const RemoveFolder = (_p: string) => done(undefined)
export const CachedNames = (ids: string[]) => done(Object.fromEntries(ids.filter((i) => NAMES[i]).map((i) => [i, NAMES[i]])))
export const ResolveNames = (ids: string[]) => CachedNames(ids)
export const ClientStatus = () => done(0)
export const LastSource = (_d: string) => done(lastSource)
export const SetLastSource = (_d: string, id: string) => { lastSource = id; return done(undefined) }
export const Accounts = (_d: string) => done({
  users: Object.keys(aliases), aliases,
  chars: Object.entries(userOf).map(([charId, userId]) => ({ charId, userId, source: charId.endsWith('4') ? 'guided' : 'launcher' })),
})
export const SetAccountAlias = (_d: string, uid: string, alias: string) => { if (alias) aliases[uid] = alias; else delete aliases[uid]; return done(undefined) }
export const AccountPlan = (req: any) => {
  const sourceUser = userOf[req.sourceId] ?? ''
  const writes: any[] = []
  const same: string[] = []
  const unknown: string[] = []
  for (const t of req.targetIds as string[]) {
    const u = userOf[t]
    if (!u) unknown.push(t)
    else if (u === sourceUser) same.push(t)
    else {
      let w = writes.find((x) => x.userId === u)
      if (!w) writes.push((w = { userId: u, targets: [], chars: Object.keys(userOf).filter((c) => userOf[c] === u).sort() }))
      w.targets.push(t)
    }
  }
  writes.sort((a, b) => a.userId.localeCompare(b.userId))
  return done({ sourceUser, writes, same, unknown })
}
export const Apply = (req: any) => done({ ok: true, code: '', message: '', backup: 'apply-demo.zip', updated: req.targetIds, accounts: [], taken: ago(0) })
export const Backups = () => done(backups)
export const Restore = (_n: string) => done({ ok: true, code: '', message: '', backup: '', updated: [], accounts: [], taken: ago(0) })
export const DeleteBackup = (_n: string) => done(undefined)
export const OpenBackupsFolder = () => done(undefined)
export const Preview = (_d: string, id: string, _l: string) => done(preview(id))
export const IdentifyStart = (_d: string) => { identOn = true; identPolls = 0; return done(undefined) }
export const IdentifyStop = () => { identOn = false; return done(undefined) }
export const IdentifyPoll = () => {
  if (!identOn) return done({ active: false, seq: 0, ambiguous: 0, found: [] })
  identPolls++
  if (identPolls === 2) { userOf['9100000009'] = '810003' } // pairs one unknown character, leaves the other
  return done({
    active: true, seq: identPolls >= 2 ? 1 : 0, ambiguous: 0,
    found: identPolls >= 2 ? [{ charId: '9100000009', userId: '810003', remapped: false }] : [],
  })
}

// EN + IT, chosen once from the system locale (no in-app switch).
const en = {
  'nav.sync': 'Sync', 'nav.profiles': 'Profiles', 'nav.characters': 'Characters', 'nav.review': 'Review and apply',
  'nav.safety': 'Safety', 'nav.backups': 'Backups', 'nav.active': 'Active profile', 'nav.sections': 'Sections',
  'kv.server': 'Server', 'kv.profile': 'Profile', 'kv.characters': 'Characters',
  'win.min': 'Minimize', 'win.max': 'Maximize', 'win.close': 'Close',
  'client.closed': 'EVE client closed', 'client.running': 'EVE client running',
  'common.back': 'Back', 'common.loading': 'Loading', 'common.never': 'unknown',
  'profiles.eyebrow': 'Sync / Profiles', 'profiles.title': 'Settings <em>folders</em>',
  'profiles.lead': 'EVE installs found on this computer. Choose the one you play on, then the settings profile to work with.',
  'profiles.add': 'Add folder', 'profiles.rescan': 'Rescan', 'profiles.installs': 'Installs', 'profiles.found': '{n} found',
  'profiles.removeFolder': 'Remove folder', 'profiles.meta': '{servers} servers, {profiles} profiles',
  'profiles.th.profile': 'Profile', 'profiles.th.chars': 'Characters', 'profiles.th.accounts': 'Account files',
  'profiles.th.saved': 'Last saved', 'profiles.th.folder': 'Folder', 'profiles.th.selected': 'Selected', 'profiles.use': 'Use profile {name}',
  'profiles.using': 'Profile {profile} on {install}',
  'profiles.hint': 'Character names come from ESI; files are read, never changed, until you apply.',
  'profiles.open': 'Open characters', 'profiles.server': 'Server',
  'profiles.emptyTitle': 'No EVE settings found', 'profiles.emptyBody': 'Start the client once so it writes its settings, or add the folder that holds your settings_* profiles.',
  'chars.eyebrow': 'Sync / Characters', 'chars.title': 'Copy <em>settings</em>', 'chars.rescan': 'Rescan folder',
  'chars.lead': 'Pick one character as the source, then tick the characters that should receive its client settings.',
  'chars.source': 'Source', 'chars.targets': 'Targets', 'chars.selected': '{n} of {total} selected',
  'chars.id': 'Character ID {id}', 'chars.lastSaved': 'Last saved', 'chars.size': 'File size',
  'chars.user': 'Also copy account settings (core_user)', 'chars.userHint': 'Account mapping comes in a later version.',
  'chars.note': 'Everything stored in the character file is copied: window layout, chat channels, shortcuts. Settings CCP keeps on the server are not touched.',
  'chars.all': 'All {n}', 'chars.week': 'Saved this week {n}', 'chars.selectAll': 'Select all', 'chars.clear': 'Clear',
  'chars.th.target': 'Target', 'chars.th.char': 'Character', 'chars.th.saved': 'Last saved', 'chars.th.size': 'Size', 'chars.th.source': 'Source',
  'chars.src': 'Src', 'chars.target': 'Target {name}', 'chars.useSource': 'Use {name} as source',
  'chars.will': '{n} character files will be overwritten with {name}’s settings.', 'chars.willOne': 'One character file will be overwritten with {name}’s settings.',
  'chars.backupFirst': 'A backup of every target is taken first.', 'chars.review': 'Review {n} changes', 'chars.reviewOne': 'Review 1 change',
  'chars.noProfile': 'Choose a profile first', 'chars.noChars': 'This profile has no character files yet.', 'chars.needTwo': 'At least two characters are needed to copy settings.',
  'chars.offline': 'Names unavailable offline; showing character IDs.',
  'review.eyebrow': 'Sync / Review', 'review.title': 'Review <em>changes</em>',
  'review.lead': 'Nothing has been written yet. Check the list, then apply: every target is backed up first.',
  'review.source': 'Source', 'review.targets': 'Targets', 'review.account': 'Account files', 'review.accountNote': 'core_user not copied',
  'review.backup': 'Backup size', 'review.files': '{n} files, before zip',
  'review.running': 'Client running',
  'review.runningBody': 'An EVE client is open. The client rewrites its settings on exit and would undo the copy. Close every client, then apply.',
  'review.again': 'Check again', 'review.applied': 'Applied',
  'review.appliedBody': '{n} characters now use {name}’s settings. Backup saved as {zip}.', 'review.openBackups': 'Open backups',
  'review.files.title': 'Files', 'review.th.char': 'Character', 'review.th.file': 'File', 'review.th.saved': 'Last saved', 'review.th.action': 'Action',
  'review.overwrite': 'Overwrite', 'review.updated': 'Updated',
  'review.apply': 'Apply to {n} characters', 'review.applyOne': 'Apply to 1 character', 'review.done': 'Applied', 'review.applying': 'Applying',
  'review.hint.running': 'Close every EVE client to continue.', 'review.hint.done': 'Start the client to see the new layout.', 'review.hint.ready': 'Backup first, then copy. Takes under a second.',
  'review.hint.empty': 'Choose a source and at least one target first.', 'review.pick': 'Open characters',
  'review.failed': 'Apply failed', 'review.failedBody': 'Nothing more was written. {msg}',
  'backups.eyebrow': 'Safety / Backups', 'backups.title': 'Settings <em>backups</em>',
  'backups.lead': 'A copy of every overwritten file, taken before each apply. Restore puts those files back exactly as they were.',
  'backups.open': 'Open folder', 'backups.history': 'History', 'backups.keeping': 'Keeping the last {keep} · {n} stored',
  'backups.th.taken': 'Taken', 'backups.th.source': 'Source', 'backups.th.files': 'Files', 'backups.th.profile': 'Profile', 'backups.th.size': 'Size', 'backups.th.kind': 'Type',
  'backups.select': 'Select backup {taken}', 'backups.selected': 'Selected', 'backups.copiedFrom': 'Copied from', 'backups.inThis': 'Files in this backup',
  'backups.restore': 'Restore this backup', 'backups.delete': 'Delete', 'backups.confirmDelete': 'Delete this backup permanently?',
  'backups.hint': 'Restoring takes a new backup of the current files first, so it can be undone too.',
  'backups.restored': 'Restored · {n} files', 'backups.kind.apply': 'Apply', 'backups.kind.restore': 'Before restore',
  'backups.empty': 'No backups yet', 'backups.emptyBody': 'A backup is created automatically the first time you apply settings.',
  'backups.blocked': 'Close every EVE client to restore.',
  'err.client_running': 'Close every EVE client first.', 'err.bad_request': 'The request was not valid.', 'err.bad_backup': 'This backup is damaged or does not match a settings folder.', 'err.io': 'A file could not be written.',
  'nav.preview': 'Preview', 'chars.preview': 'Preview layout',
  'preview.eyebrow': 'Sync / Preview', 'preview.title': 'Layout <em>preview</em>',
  'preview.source': 'Source', 'preview.compare': 'Compare with target', 'preview.noTarget': 'No target',
  'preview.show.source': 'Source', 'preview.show.target': 'Target now', 'preview.show.overlay': 'Overlay',
  'preview.t.windows': 'Windows', 'preview.t.stacks': 'Stacks {n} · locked {m}', 'preview.t.saved': 'Saved at',
  'preview.t.scale': 'UI scale {n}%', 'preview.t.scaleBoth': 'UI scale {a}% full screen, {b}% windowed', 'preview.t.screenOnly': 'Screen size when saved',
  'preview.t.chat': 'Chat channels', 'preview.t.more': '{names} + {n}', 'preview.t.changes': 'Changes on target',
  'preview.t.of': 'of {n} windows', 'preview.t.pickTarget': 'No target selected',
  'preview.tab.layout': 'Window layout', 'preview.tab.chat': 'Chat channels', 'preview.tab.keys': 'All keys', 'preview.tabs': 'Preview views',
  'preview.lg.window': 'Window', 'preview.lg.ghost': 'Target today (overlay)', 'preview.lg.sel': 'Selected', 'preview.lg.scaled': 'Scaled from the saved {w} × {h} screen',
  'preview.win': 'Window', 'preview.pos': 'Position', 'preview.size': 'Size', 'preview.tgtToday': 'Target today', 'preview.stack': 'Stack', 'preview.state': 'State',
  'preview.notInSource': 'Not in source', 'preview.notOpen': 'Not open', 'preview.noneVal': 'None', 'preview.noWindows': 'No windows saved in this file.',
  'preview.s.open': 'Open', 'preview.s.locked': 'Locked', 'preview.s.collapsed': 'Collapsed', 'preview.s.compact': 'Compact', 'preview.s.minimized': 'Minimized',
  'preview.st.same': 'Unchanged', 'preview.st.move': 'Will move', 'preview.st.add': 'Will be added', 'preview.st.remove': 'Will be removed', 'preview.st.none': 'In source',
  'preview.applies': 'What Apply changes', 'preview.noChanges': 'Nothing would change.', 'preview.pickTarget': 'Choose a target to see what changes.',
  'preview.readonly': 'Read only: the preview decodes the settings file and never writes it. If the file cannot be decoded, copying still works.',
  'preview.copy': 'Copy this layout',
  'preview.chat.th.channel': 'Channel', 'preview.chat.th.type': 'Type', 'preview.chat.th.status': 'On apply', 'preview.chat.empty': 'No chat channels saved in this file.',
  'preview.kind.local': 'Local', 'preview.kind.corp': 'Corp', 'preview.kind.alliance': 'Alliance', 'preview.kind.fleet': 'Fleet',
  'preview.kind.private': 'Private', 'preview.kind.player': 'Player', 'preview.kind.other': 'Other',
  'preview.keys.filter': 'Filter keys', 'preview.keys.count': 'Keys: {n}', 'preview.keys.cap': 'Showing the first {n}; narrow the filter to see more.',
  'preview.keys.th.group': 'Group', 'preview.keys.th.key': 'Key', 'preview.keys.th.type': 'Type', 'preview.keys.th.size': 'Items', 'preview.keys.none': 'No keys match.',
  'preview.type.flag': 'flag', 'preview.type.number': 'number', 'preview.type.text': 'text', 'preview.type.list': 'list', 'preview.type.map': 'map',
  'preview.type.tuple': 'tuple', 'preview.type.empty': 'empty', 'preview.type.object': 'object',
  'preview.err.title': 'Can’t read this file', 'preview.err.body': 'The settings file could not be decoded, so there is nothing to preview. Copying it to other characters still works.',
  'preview.err.missing': 'This character has no settings file in the profile.', 'preview.err.target': 'The target file can’t be read, so only the source is shown.',
  'preview.reading': 'Reading settings',
  'char.unnamed': 'Character {id}',
} as const

export type Key = keyof typeof en

const it: Record<Key, string> = {
  'nav.sync': 'Sincronizza', 'nav.profiles': 'Profili', 'nav.characters': 'Personaggi', 'nav.review': 'Revisione e applica',
  'nav.safety': 'Sicurezza', 'nav.backups': 'Backup', 'nav.active': 'Profilo attivo', 'nav.sections': 'Sezioni',
  'kv.server': 'Server', 'kv.profile': 'Profilo', 'kv.characters': 'Personaggi',
  'win.min': 'Riduci a icona', 'win.max': 'Ingrandisci', 'win.close': 'Chiudi',
  'client.closed': 'Client EVE chiuso', 'client.running': 'Client EVE in esecuzione',
  'common.back': 'Indietro', 'common.loading': 'Caricamento', 'common.never': 'sconosciuto',
  'profiles.eyebrow': 'Sincronizza / Profili', 'profiles.title': 'Cartelle <em>impostazioni</em>',
  'profiles.lead': 'Le installazioni di EVE trovate su questo computer. Scegli quella che usi, poi il profilo impostazioni su cui lavorare.',
  'profiles.add': 'Aggiungi cartella', 'profiles.rescan': 'Riscansiona', 'profiles.installs': 'Installazioni', 'profiles.found': '{n} trovate',
  'profiles.removeFolder': 'Rimuovi cartella', 'profiles.meta': '{servers} server, {profiles} profili',
  'profiles.th.profile': 'Profilo', 'profiles.th.chars': 'Personaggi', 'profiles.th.accounts': 'File account',
  'profiles.th.saved': 'Ultimo salvataggio', 'profiles.th.folder': 'Cartella', 'profiles.th.selected': 'Selezionato', 'profiles.use': 'Usa il profilo {name}',
  'profiles.using': 'Profilo {profile} su {install}',
  'profiles.hint': 'I nomi dei personaggi vengono da ESI; i file sono solo letti, mai modificati, finché non applichi.',
  'profiles.open': 'Apri personaggi', 'profiles.server': 'Server',
  'profiles.emptyTitle': 'Nessuna impostazione EVE trovata', 'profiles.emptyBody': 'Avvia il client una volta così scrive le sue impostazioni, oppure aggiungi la cartella che contiene i profili settings_*.',
  'chars.eyebrow': 'Sincronizza / Personaggi', 'chars.title': 'Copia <em>impostazioni</em>', 'chars.rescan': 'Riscansiona cartella',
  'chars.lead': 'Scegli un personaggio come sorgente, poi spunta quelli che devono ricevere le sue impostazioni del client.',
  'chars.source': 'Sorgente', 'chars.targets': 'Destinazioni', 'chars.selected': '{n} di {total} selezionati',
  'chars.id': 'ID personaggio {id}', 'chars.lastSaved': 'Ultimo salvataggio', 'chars.size': 'Dimensione file',
  'chars.user': 'Copia anche le impostazioni account (core_user)', 'chars.userHint': 'La mappatura degli account arriva in una versione successiva.',
  'chars.note': 'Viene copiato tutto il contenuto del file del personaggio: layout delle finestre, canali chat, scorciatoie. Le impostazioni che CCP tiene sul server non vengono toccate.',
  'chars.all': 'Tutti {n}', 'chars.week': 'Salvati questa settimana {n}', 'chars.selectAll': 'Seleziona tutti', 'chars.clear': 'Svuota',
  'chars.th.target': 'Dest.', 'chars.th.char': 'Personaggio', 'chars.th.saved': 'Ultimo salvataggio', 'chars.th.size': 'Dimensione', 'chars.th.source': 'Sorgente',
  'chars.src': 'Sorg.', 'chars.target': 'Destinazione {name}', 'chars.useSource': 'Usa {name} come sorgente',
  'chars.will': '{n} file personaggio verranno sovrascritti con le impostazioni di {name}.', 'chars.willOne': 'Un file personaggio verrà sovrascritto con le impostazioni di {name}.',
  'chars.backupFirst': 'Prima viene fatto un backup di ogni destinazione.', 'chars.review': 'Rivedi {n} modifiche', 'chars.reviewOne': 'Rivedi 1 modifica',
  'chars.noProfile': 'Scegli prima un profilo', 'chars.noChars': 'Questo profilo non ha ancora file personaggio.', 'chars.needTwo': 'Servono almeno due personaggi per copiare le impostazioni.',
  'chars.offline': 'Nomi non disponibili offline; mostro gli ID dei personaggi.',
  'review.eyebrow': 'Sincronizza / Revisione', 'review.title': 'Rivedi <em>modifiche</em>',
  'review.lead': 'Non è stato scritto ancora nulla. Controlla l’elenco, poi applica: ogni destinazione viene prima salvata in backup.',
  'review.source': 'Sorgente', 'review.targets': 'Destinazioni', 'review.account': 'File account', 'review.accountNote': 'core_user non copiato',
  'review.backup': 'Dimensione backup', 'review.files': '{n} file, prima dello zip',
  'review.running': 'Client in esecuzione',
  'review.runningBody': 'Un client EVE è aperto. Il client riscrive le impostazioni in uscita e annullerebbe la copia. Chiudi ogni client, poi applica.',
  'review.again': 'Controlla di nuovo', 'review.applied': 'Applicato',
  'review.appliedBody': '{n} personaggi ora usano le impostazioni di {name}. Backup salvato come {zip}.', 'review.openBackups': 'Apri backup',
  'review.files.title': 'File', 'review.th.char': 'Personaggio', 'review.th.file': 'File', 'review.th.saved': 'Ultimo salvataggio', 'review.th.action': 'Azione',
  'review.overwrite': 'Sovrascrivi', 'review.updated': 'Aggiornato',
  'review.apply': 'Applica a {n} personaggi', 'review.applyOne': 'Applica a 1 personaggio', 'review.done': 'Applicato', 'review.applying': 'Applico',
  'review.hint.running': 'Chiudi ogni client EVE per continuare.', 'review.hint.done': 'Avvia il client per vedere il nuovo layout.', 'review.hint.ready': 'Prima il backup, poi la copia. Meno di un secondo.',
  'review.hint.empty': 'Scegli prima una sorgente e almeno una destinazione.', 'review.pick': 'Apri personaggi',
  'review.failed': 'Applicazione fallita', 'review.failedBody': 'Non è stato scritto altro. {msg}',
  'backups.eyebrow': 'Sicurezza / Backup', 'backups.title': 'Backup <em>impostazioni</em>',
  'backups.lead': 'Una copia di ogni file sovrascritto, fatta prima di ogni applicazione. Il ripristino rimette quei file esattamente com’erano.',
  'backups.open': 'Apri cartella', 'backups.history': 'Cronologia', 'backups.keeping': 'Conservo gli ultimi {keep} · {n} presenti',
  'backups.th.taken': 'Eseguito', 'backups.th.source': 'Sorgente', 'backups.th.files': 'File', 'backups.th.profile': 'Profilo', 'backups.th.size': 'Dimensione', 'backups.th.kind': 'Tipo',
  'backups.select': 'Seleziona il backup {taken}', 'backups.selected': 'Selezionato', 'backups.copiedFrom': 'Copiato da', 'backups.inThis': 'File in questo backup',
  'backups.restore': 'Ripristina questo backup', 'backups.delete': 'Elimina', 'backups.confirmDelete': 'Eliminare definitivamente questo backup?',
  'backups.hint': 'Il ripristino fa prima un nuovo backup dei file attuali, così si può annullare anche quello.',
  'backups.restored': 'Ripristinato · {n} file', 'backups.kind.apply': 'Applicazione', 'backups.kind.restore': 'Prima del ripristino',
  'backups.empty': 'Ancora nessun backup', 'backups.emptyBody': 'Un backup viene creato automaticamente la prima volta che applichi le impostazioni.',
  'backups.blocked': 'Chiudi ogni client EVE per ripristinare.',
  'err.client_running': 'Chiudi prima ogni client EVE.', 'err.bad_request': 'La richiesta non è valida.', 'err.bad_backup': 'Questo backup è danneggiato o non corrisponde a una cartella impostazioni.', 'err.io': 'Non è stato possibile scrivere un file.',
  'nav.preview': 'Anteprima', 'chars.preview': 'Anteprima layout',
  'preview.eyebrow': 'Sincronizza / Anteprima', 'preview.title': 'Anteprima <em>layout</em>',
  'preview.source': 'Sorgente', 'preview.compare': 'Confronta con la destinazione', 'preview.noTarget': 'Nessuna destinazione',
  'preview.show.source': 'Sorgente', 'preview.show.target': 'Destinazione ora', 'preview.show.overlay': 'Sovrapposto',
  'preview.t.windows': 'Finestre', 'preview.t.stacks': 'Gruppi {n} · bloccate {m}', 'preview.t.saved': 'Salvato a',
  'preview.t.scale': 'Scala UI {n}%', 'preview.t.scaleBoth': 'Scala UI {a}% a schermo intero, {b}% in finestra', 'preview.t.screenOnly': 'Dimensione schermo al salvataggio',
  'preview.t.chat': 'Canali chat', 'preview.t.more': '{names} + {n}', 'preview.t.changes': 'Modifiche sulla destinazione',
  'preview.t.of': 'su {n} finestre', 'preview.t.pickTarget': 'Nessuna destinazione scelta',
  'preview.tab.layout': 'Layout finestre', 'preview.tab.chat': 'Canali chat', 'preview.tab.keys': 'Tutte le chiavi', 'preview.tabs': 'Viste anteprima',
  'preview.lg.window': 'Finestra', 'preview.lg.ghost': 'Destinazione oggi (sovrapposto)', 'preview.lg.sel': 'Selezionata', 'preview.lg.scaled': 'Scalato dallo schermo salvato {w} × {h}',
  'preview.win': 'Finestra', 'preview.pos': 'Posizione', 'preview.size': 'Dimensione', 'preview.tgtToday': 'Destinazione oggi', 'preview.stack': 'Gruppo', 'preview.state': 'Stato',
  'preview.notInSource': 'Non nella sorgente', 'preview.notOpen': 'Non aperta', 'preview.noneVal': 'Nessuno', 'preview.noWindows': 'Nessuna finestra salvata in questo file.',
  'preview.s.open': 'Aperta', 'preview.s.locked': 'Bloccata', 'preview.s.collapsed': 'Raccolta', 'preview.s.compact': 'Compatta', 'preview.s.minimized': 'Ridotta',
  'preview.st.same': 'Invariata', 'preview.st.move': 'Verrà spostata', 'preview.st.add': 'Verrà aggiunta', 'preview.st.remove': 'Verrà rimossa', 'preview.st.none': 'Nella sorgente',
  'preview.applies': 'Cosa cambia con Applica', 'preview.noChanges': 'Non cambierebbe nulla.', 'preview.pickTarget': 'Scegli una destinazione per vedere cosa cambia.',
  'preview.readonly': 'Sola lettura: l’anteprima decodifica il file delle impostazioni e non lo scrive mai. Se il file non si decodifica, la copia funziona comunque.',
  'preview.copy': 'Copia questo layout',
  'preview.chat.th.channel': 'Canale', 'preview.chat.th.type': 'Tipo', 'preview.chat.th.status': 'Con Applica', 'preview.chat.empty': 'Nessun canale chat salvato in questo file.',
  'preview.kind.local': 'Locale', 'preview.kind.corp': 'Corp', 'preview.kind.alliance': 'Alleanza', 'preview.kind.fleet': 'Flotta',
  'preview.kind.private': 'Privato', 'preview.kind.player': 'Giocatore', 'preview.kind.other': 'Altro',
  'preview.keys.filter': 'Filtra chiavi', 'preview.keys.count': 'Chiavi: {n}', 'preview.keys.cap': 'Mostro le prime {n}; restringi il filtro per vederne altre.',
  'preview.keys.th.group': 'Gruppo', 'preview.keys.th.key': 'Chiave', 'preview.keys.th.type': 'Tipo', 'preview.keys.th.size': 'Elementi', 'preview.keys.none': 'Nessuna chiave corrisponde.',
  'preview.type.flag': 'flag', 'preview.type.number': 'numero', 'preview.type.text': 'testo', 'preview.type.list': 'lista', 'preview.type.map': 'mappa',
  'preview.type.tuple': 'tupla', 'preview.type.empty': 'vuoto', 'preview.type.object': 'oggetto',
  'preview.err.title': 'Impossibile leggere questo file', 'preview.err.body': 'Il file delle impostazioni non si è potuto decodificare, quindi non c’è nulla da mostrare. Copiarlo sugli altri personaggi funziona comunque.',
  'preview.err.missing': 'Questo personaggio non ha un file di impostazioni nel profilo.', 'preview.err.target': 'Il file della destinazione non è leggibile, quindi vedi solo la sorgente.',
  'preview.reading': 'Lettura impostazioni',
  'char.unnamed': 'Personaggio {id}',
}

export const locale: 'en' | 'it' = (navigator.language || 'en').toLowerCase().startsWith('it') ? 'it' : 'en'

export function applyLocale() {
  document.documentElement.lang = locale
}

export function t(key: Key, params: Record<string, string | number> = {}): string {
  const s = (locale === 'it' ? it : en)[key] as string
  return s.replace(/\{(\w+)\}/g, (_, k) => String(params[k] ?? ''))
}

/** Titles carry one accented word as <em>; render it with the accent colour, escaping the rest. */
export function titleParts(key: Key): [string, string] {
  const m = /^(.*?)<em>(.*)<\/em>(.*)$/.exec(t(key))
  return m ? [m[1] + m[3], m[2]] : [t(key), '']
}

/** Local time as `YYYY-MM-DD HH:mm`, locale neutral like the reference screens. */
export function fmtDate(iso: string | undefined): string {
  if (!iso) return t('common.never')
  const d = new Date(iso)
  if (isNaN(d.getTime())) return t('common.never')
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}

export function fmtSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${Math.max(1, Math.round(bytes / 1024))} KB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

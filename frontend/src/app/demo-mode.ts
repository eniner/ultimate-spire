import {demoAdminBody} from "@/app/demo-admin"

export function isDemoMode() {
  return process.env.VUE_APP_DEMO === "true"
}

let tables: any = null
let systems: any = null
let loadPromise: Promise<any> | null = null

function loadTables() {
  if (tables && systems) {
    return Promise.resolve(tables)
  }
  if (!loadPromise) {
    const base = process.env.BASE_URL || "/"
    loadPromise = Promise.all([
      fetch(base + "demo/peq.json").then((r) => r.json()),
      fetch(base + "demo/systems.json").then((r) => r.json()),
    ]).then(([peq, sys]) => {
      tables = peq
      systems = sys
      return peq
    })
  }
  return loadPromise
}

function field(row: any, name: string) {
  if (row[name] !== undefined) {
    return row[name]
  }
  const lower = String(name || "").toLowerCase()
  if (row[lower] !== undefined) {
    return row[lower]
  }
  if (lower === "name" && row.Name !== undefined) {
    return row.Name
  }
  return undefined
}

function parseWhere(raw: string) {
  if (!raw) {
    return []
  }
  return String(raw).split(".").filter(Boolean).map((part) => {
    const ops = [
      ["_like_", "like"],
      ["_notlike_", "notlike"],
      ["_ne_", "ne"],
      ["_gte_", "gte"],
      ["_lte_", "lte"],
      ["_gt_", "gt"],
      ["_lt_", "lt"],
      ["_bitwiseand_", "band"],
      ["__", "eq"],
    ]
    for (let i = 0; i < ops.length; i++) {
      const token = ops[i][0]
      const at = part.toLowerCase().indexOf(token)
      if (at > 0) {
        return {field: part.slice(0, at), op: ops[i][1], value: part.slice(at + token.length)}
      }
    }
    return {field: part, op: "eq", value: ""}
  })
}

function matchRow(row: any, clause: any) {
  const actual = field(row, clause.field)
  const expect = String(clause.value == null ? "" : clause.value)
  if (clause.op === "like") {
    const needle = expect.replace(/%/g, "").toLowerCase()
    return String(actual == null ? "" : actual).toLowerCase().indexOf(needle) !== -1
  }
  if (clause.op === "notlike") {
    const needle = expect.replace(/%/g, "").toLowerCase()
    return String(actual == null ? "" : actual).toLowerCase().indexOf(needle) === -1
  }
  if (clause.op === "ne") {
    return String(actual) !== expect
  }
  if (clause.op === "band") {
    return (Number(actual) & Number(expect)) !== 0
  }
  if (clause.op === "gte") {
    return Number(actual) >= Number(expect)
  }
  if (clause.op === "lte") {
    return Number(actual) <= Number(expect)
  }
  if (clause.op === "gt") {
    return Number(actual) > Number(expect)
  }
  if (clause.op === "lt") {
    return Number(actual) < Number(expect)
  }
  return String(actual) === expect
}

function queryTable(name: string, params: any) {
  const rows = (tables && tables[name]) || []
  const where = parseWhere(params.where)
  const whereOr = parseWhere(params.whereOr)
  let out = rows.filter((row) => {
    const andOk = where.every((c) => matchRow(row, c))
    const orOk = !whereOr.length || whereOr.some((c) => matchRow(row, c))
    return andOk && orOk
  })
  if (params.orderBy) {
    const keys = String(params.orderBy).split(".")
    out = out.slice().sort((a, b) => {
      for (let i = 0; i < keys.length; i++) {
        const av = field(a, keys[i])
        const bv = field(b, keys[i])
        if (av < bv) {
          return -1
        }
        if (av > bv) {
          return 1
        }
      }
      return 0
    })
  }
  const limit = Number(params.limit || 0)
  const page = Math.max(1, Number(params.page || 1))
  if (limit > 0) {
    const start = (page - 1) * limit
    out = out.slice(start, start + limit)
  }
  if (name === "inventory") {
    out = out.map((row) => {
      const item = ((tables && tables.item) || []).find((n) => String(n.id) === String(row.item_id))
      return Object.assign({}, row, {item: item || null})
    })
  }
  if (name === "sharedbank") {
    out = out.map((row) => {
      const item = ((tables && tables.item) || []).find((n) => String(n.id) === String(row.item_id))
      return Object.assign({}, row, {item: item || null})
    })
  }
  if (name === "character_parcels") {
    out = out.map((row) => {
      const item = ((tables && tables.item) || []).find((n) => String(n.id) === String(row.item_id))
      return Object.assign({}, row, {item: item || null})
    })
  }
  if (name === "spawn2") {
    out = out.map((row) => {
      const entries = ((tables && tables.spawnentry) || []).filter((e) => String(e.spawngroupID) === String(row.spawngroupID)).map((e) => {
        const npc = ((tables && tables.npc_types) || []).find((n) => String(n.id) === String(e.npcID))
        return Object.assign({}, e, {npc_type: npc || {id: e.npcID, name: "NPC " + e.npcID}})
      })
      return Object.assign({}, row, {spawnentries: entries})
    })
  }
  return out
}

function pathOf(config: any) {
  const url = String((config && config.url) || "")
  const base = String((config && config.baseURL) || "")
  let raw = url || base
  if (base && url && !/^https?:/i.test(url)) {
    raw = base.replace(/\/$/, "") + "/" + url.replace(/^\//, "")
  }
  try {
    if (/^https?:/i.test(raw)) {
      return new URL(raw).pathname
    }
  } catch (_err) {
    // fall through
  }
  return raw.indexOf("/") === 0 ? raw : "/" + raw
}

function paramsOf(config: any) {
  const out: any = Object.assign({}, config.params || {})
  const raw = String((config && config.url) || "")
  const q = raw.indexOf("?") >= 0 ? raw.slice(raw.indexOf("?") + 1) : ""
  if (q) {
    const usp = new URLSearchParams(q)
    usp.forEach((value, key) => {
      out[key] = value
    })
  }
  let data = config && config.data
  if (typeof data === "string") {
    try {
      data = JSON.parse(data)
    } catch (_err) {
      data = null
    }
  }
  if (data && typeof data === "object" && !Array.isArray(data)) {
    Object.assign(out, data)
    if (data.body && typeof data.body === "object") {
      Object.assign(out, data.body)
    }
  }
  return out
}

function ok(data: any, config: any) {
  return {
    data,
    status: 200,
    statusText: "OK",
    headers: {"content-type": "application/json"},
    config,
    request: {},
  }
}

function demoBody(config: any) {
  const path = pathOf(config)
  const method = String(config.method || "get").toLowerCase()
  const params = paramsOf(config)

  if (path.indexOf("/app/env") !== -1) {
    return {
      data: {
        os: "demo",
        env: "local",
        version: "4.24.5-demo",
        features: {github_auth_enabled: false},
        settings: [{setting: "AUTH_ENABLED", value: "false"}],
        is_spire_initialized: true,
      },
    }
  }
  if (path.indexOf("/me") !== -1) {
    return {id: 0, user_name: "demo", is_admin: true}
  }

  const admin = demoAdminBody(path, method, params, tables)
  if (admin !== undefined) {
    return admin
  }

  const seeded = systemsBody(path, method, params)
  if (seeded !== undefined) {
    return seeded
  }

  const raw = peqRawBody(path, method, params)
  if (raw !== undefined) {
    return raw
  }

  if (method !== "get" && path.indexOf("/api/v1/") !== -1) {
    if (path.indexOf("/bulk") !== -1) {
      const name = tableFromPath(path.replace(/\/bulk.*/, ""))
      const ids = [].concat(params.ids || []).map(String)
      return ((tables && tables[name]) || []).filter((row) => ids.indexOf(String(row.id)) !== -1)
    }
    if (path.indexOf("/count") !== -1) {
      const countParams = Object.assign({}, params, {limit: 0, page: 1})
      return {count: queryTable(tableFromPath(path.replace(/\/count.*/, "")), countParams).length}
    }
    return {ok: false, dryRun: true, error: "This public demo is read-only. Use the local exe against your own peq."}
  }
  if (path.indexOf("/count") !== -1) {
    const countParams = Object.assign({}, params, {limit: 0, page: 1})
    return {count: queryTable(tableFromPath(path.replace(/\/count.*/, "")), countParams).length}
  }

  const name = tableFromPath(path)
  if (name && tables && tables[name]) {
    return queryTable(name, params)
  }
  if (path.indexOf("/api/v1/") !== -1 && method === "get") {
    return []
  }
  return []
}

function peqRawBody(path: string, method: string, params: any) {
  if (path.indexOf("/peq-raw/") === -1) {
    return undefined
  }
  if (path.indexOf("/peq-raw/tables") !== -1) {
    const names = Object.keys(tables || {}).filter((k) => k !== "source")
    return {tables: names}
  }
  const clean = path.replace(/\/$/, "").split("?")[0]
  const parts = clean.split("/").filter(Boolean)
  const last = parts[parts.length - 1] || ""
  const prev = parts[parts.length - 2] || ""
  const tableName = last === "count" ? decodeURIComponent(prev) : decodeURIComponent(last)
  const resolved = resolveTableName(tableName)
  const rows = ((tables && (tables[resolved] || tables[tableName])) || []).slice()
  const search = String(params.search || "").toLowerCase()
  const filtered = !search ? rows : rows.filter((row) => JSON.stringify(row).toLowerCase().indexOf(search) !== -1)
  if (last === "count") {
    return {count: filtered.length}
  }
  if (method !== "get") {
    return {ok: false, dryRun: true, error: "This public demo is read-only."}
  }
  const limit = Number(params.limit || 100)
  const page = Math.max(1, Number(params.page || 1))
  const start = (page - 1) * limit
  const pageRows = filtered.slice(start, start + limit)
  const columns = pageRows[0] ? Object.keys(pageRows[0]) : (filtered[0] ? Object.keys(filtered[0]) : ["id"])
  return {
    rows: pageRows,
    columns,
    primaryKey: columns.indexOf("id") !== -1 ? ["id"] : [columns[0]],
  }
}

function lastSegment(path: string) {
  const parts = path.replace(/\/$/, "").split("/").filter(Boolean)
  return parts[parts.length - 1] || ""
}

function systemsBody(path: string, method: string, params: any) {
  if (!systems) {
    return undefined
  }
  if (path.indexOf("/admin/zone-controller/status") !== -1) {
    return systems.zcStatus
  }
  if (path.indexOf("/admin/zone-controller/systems") !== -1) {
    return systems.zcSystems
  }
  if (path.indexOf("/admin/zone-controller/zones/") !== -1) {
    const id = lastSegment(path)
    return (systems.zcZones && systems.zcZones[id]) || {error: "zone not in demo seed"}
  }
  if (path.indexOf("/admin/zone-controller/zones") !== -1) {
    return systems.zcIndex
  }
  if (path.indexOf("/admin/ultimate-systems/status") !== -1) {
    return systems.ultStatus
  }
  if (path.indexOf("/admin/ultimate-systems/catalog") !== -1) {
    return systems.catalog
  }
  if (path.indexOf("/admin/ultimate-systems/talents/") !== -1) {
    const id = decodeURIComponent(lastSegment(path))
    const row = ((systems.catalog && systems.catalog.talents) || []).find((t: any) => t.talentId === id)
    return row || {error: "talent not found"}
  }
  if (path.indexOf("/admin/ultimate-systems/traits") !== -1) {
    return systems.traits
  }
  if (path.indexOf("/admin/ultimate-systems/runewords") !== -1) {
    return systems.runewords
  }
  if (path.indexOf("/admin/server-files/status") !== -1 || path.indexOf("/admin/server-files/connect") !== -1) {
    return {
      ok: true,
      connected: true,
      source: "demo",
      root: "quests",
      writable: false,
      detected: {questsDir: "quests", suggestions: ["quests"]},
    }
  }
  if (path.indexOf("/admin/zone-controller/create") !== -1 || path.indexOf("/admin/zone-controller/tier") !== -1 || path.indexOf("/admin/zone-controller/mobs") !== -1 || path.indexOf("/admin/zone-controller/factory") !== -1 || path.indexOf("/admin/zone-controller/classify") !== -1 || path.indexOf("/admin/zone-controller/recipes") !== -1) {
    if (path.indexOf("/admin/zone-controller/recipes") !== -1 && method === "GET") {
      return {
        ok: true,
        recipes: [{
          id: "classic-t1",
          name: "Classic T1",
          group: "classic-t1",
          step: 1.35,
          prefix: "T1",
          basedata: {
            trash: {level: "71", max_hp: "500000", min_hit: "800", max_hit: "1200"},
            boss: {level: "75", max_hp: "2000000", min_hit: "1400", max_hit: "2200"},
            raid: {level: "80", max_hp: "8000000", min_hit: "2200", max_hit: "3600"},
          },
          kit: {sourceZone: 17, items: [], tables: []},
        }],
      }
    }
    return {
      ok: true,
      dryRun: true,
      writes: [],
      steps: [],
      warnings: ["Public demo is read-only. This is a preview of the write plan."],
    }
  }
  if (path.indexOf("/admin/zone-controller/validate") !== -1) {
    return {ok: true, zones: 0, errors: 0, warnings: 0, issues: []}
  }
  if (path.indexOf("/admin/zone-controller/live") !== -1) {
    return {
      ok: true,
      worldOk: false,
      worldNote: "Public demo is read-only.",
      zones: [{zoneId: 17, short: "blackburrow", long: "Blackburrow", popped: false, players: 0}],
    }
  }
  if (path.indexOf("/admin/zone-controller/apply") !== -1) {
    return {
      ok: true,
      commands: ["!initdata 17"],
      reloaded: [],
      rebooted: [],
      queued: [],
      live: [],
      worldOk: false,
      worldNote: "Public demo is read-only.",
    }
  }
  if (path.indexOf("/admin/content-factory/status") !== -1) {
    return {ok: true, reserved: 800000, recipeRel: "global/ultimatedata/_spire_recipes/recipes.json", runRel: "global/ultimatedata/_spire_runs"}
  }
  if (path.indexOf("/admin/content-factory/ids") !== -1) {
    return {
      ok: true,
      reserved: 800000,
      rows: [
        {key: "items", table: "items", column: "id", floor: 800000, maxId: 800120, nextId: 800121, reservedUsed: 12, note: "Demo reserved gear."},
        {key: "spells_new", table: "spells_new", column: "id", floor: 800000, maxId: 45200, nextId: 800000, reservedUsed: 0, note: "Item click/proc clones."},
        {key: "loottable", table: "loottable", column: "id", floor: 800000, maxId: 1200, nextId: 800000, reservedUsed: 0, note: "Loot table headers."},
      ],
    }
  }
  if (path.indexOf("/admin/content-factory/meta") !== -1) {
    return {slots: [], classes: [], npcSpellIdCap: 65535, reserved: 800000}
  }
  if (path.indexOf("/admin/content-factory/search") !== -1) {
    return {ok: true, kind: params.kind || "items", query: params.q || "", rows: [{id: 1001, name: "Demo Helm"}]}
  }
  if (path.indexOf("/admin/content-factory/runs") !== -1) {
    return {ok: true, runs: []}
  }
  if (path.indexOf("/admin/content-factory/census") !== -1) {
    return {
      ok: true,
      zoneId: 17,
      short: "blackburrow",
      long: "Blackburrow",
      counts: {npcs: 1, items: 1, merchants: 1, spellSets: 1, trash: 1, boss: 0, raid: 0, ignore: 0, groundSpawns: 2, forages: 3},
      npcs: [{id: 5036, name: "a gnoll", level: 5, hp: 120, pops: 8, role: "trash", faction: "Clan", loottableId: 120, npcSpellsId: 2, merchantId: 0, spellSet: "Default", spells: [{id: 13, name: "Complete Heal"}]}],
      items: [{id: 1001, name: "Cloth Cap", slot: "Head", class: "WAR", click: 0, proc: 0}],
      merchants: [{id: 1, npcName: "Merchant", npcId: 1, stock: 3, nextSlot: 4}],
      spellSets: [{id: 2, name: "Default", npcs: 1, spells: [{id: 13, name: "Complete Heal"}]}],
      neighbors: [{id: 2, short: "qeytoqrg", long: "Qeynos Hills"}],
      groundSpawns: 2,
      forages: 3,
    }
  }
  if (path.indexOf("/admin/content-factory/draft") !== -1) {
    return {ok: true, relPath: "global/ultimatedata/_spire_drafts/content-factory.json", body: {}}
  }
  if (path.indexOf("/admin/content-factory/spell-set") !== -1) {
    return {ok: true, id: 2, name: "Default", entries: [{spellId: 13, name: "Complete Heal", minLevel: 1, maxLevel: 255, type: 1, manacost: -1, recastDelay: 0, priority: 0}]}
  }
  if (path.indexOf("/admin/content-factory/catalog") !== -1) {
    return {ok: true, kind: params.kind || "items", floor: 800000, rows: [{id: 800001, name: "Demo Helm", extra: "click 13", useNote: "unused"}]}
  }
  if (path.indexOf("/admin/content-factory/merchant") !== -1) {
    return {ok: true, merchantId: 1, npcName: "Demo Merchant", npcId: 1, count: 1, nextSlot: 2, replaceNote: "Append uses the next free slot.", stock: [{slot: 1, itemId: 1001, name: "Cloth Cap", price: 5, probability: 100}]}
  }
  if (path.indexOf("/admin/content-factory/import-zone") !== -1) {
    return {ok: true, zoneId: 17, short: "blackburrow", filled: 1, cells: [{slot: "Head", class: "WAR", sourceId: 1001, name: "Cloth Cap"}]}
  }
  if (path.indexOf("/admin/content-factory/") !== -1) {
    return {
      ok: true,
      dryRun: true,
      kind: "demo",
      clones: [],
      dbWrites: [],
      writes: [],
      warnings: ["Public demo is read-only. This is a preview of the write plan."],
    }
  }
  if (path.indexOf("/admin/ultimate-systems/write") !== -1) {
    return {
      ok: true,
      dryRun: true,
      kind: params.kind || "talent",
      action: params.action || "preview",
      relPath: "global/talentdata/talent_catalog.json",
      key: params.talent_id || params.key || "",
      warnings: ["Public demo is read-only. This is a preview of the write plan."],
    }
  }
  if (path.indexOf("/admin/server-files/list") !== -1) {
    const key = String(params.path || "").replace(/^\/+|\/+$/g, "")
    const tree = (systems.serverFiles && systems.serverFiles.tree) || {}
    return {entries: tree[key] || [], writable: false, root: "quests", source: "demo"}
  }
  if (path.indexOf("/admin/server-files/file") !== -1) {
    const filePath = String(params.path || "")
    const files = (systems.serverFiles && systems.serverFiles.files) || {}
    if (method === "get") {
      return {content: files[filePath] || "", writable: false, path: filePath, hash: "demo"}
    }
    return {ok: false, dryRun: true, error: "This public demo is read-only."}
  }
  if (path.indexOf("/admin/server-files/search") !== -1) {
    const q = String(params.q || "").toLowerCase()
    const files = (systems.serverFiles && systems.serverFiles.files) || {}
    const entries = Object.keys(files).filter((p) => !q || p.toLowerCase().indexOf(q) !== -1).map((p) => ({
      name: p.split("/").pop(),
      path: p,
      isDir: false,
      size: String(files[p] || "").length,
      modified: 0,
      editable: true,
    }))
    return {entries}
  }
  return undefined
}

function resolveTableName(last: string) {
  const aliases: any = {
    zone: "zone",
    zones: "zone",
    item: "item",
    items: "item",
    task: "task",
    tasks: "task",
    items_evolving_details: "items_evolving_details",
    items_evolving_detail: "items_evolving_details",
    npc: "npc_types",
    npc_type: "npc_types",
    npc_types: "npc_types",
    spawn2: "spawn2",
    spawn_2: "spawn2",
    spawn: "spawn2",
    spawn2s: "spawn2",
    spawn_2s: "spawn2",
    spells_new: "spells_new",
    spells_news: "spells_new",
    spell: "spells_new",
    spells: "spells_new",
    loottable: "loottable",
    loottables: "loottable",
    lootdrop: "lootdrop",
    lootdrops: "lootdrop",
    lootdrop_entry: "lootdrop_entries",
    lootdrop_entries: "lootdrop_entries",
    loottable_entry: "loottable_entries",
    loottable_entries: "loottable_entries",
    merchantlist: "merchantlist",
    merchantlists: "merchantlist",
    faction_list: "faction_list",
    faction_lists: "faction_list",
    door: "doors",
    doors: "doors",
    tradeskill_recipe: "tradeskill_recipe",
    tradeskill_recipes: "tradeskill_recipe",
    aa_ability: "aa_ability",
    aa_abilities: "aa_ability",
    grid: "grid",
    grids: "grid",
    grid_entry: "grid_entries",
    grid_entries: "grid_entries",
    object: "object",
    objects: "object",
    character_data: "character_data",
    character_datum: "character_data",
    inventory: "inventory",
    inventories: "inventory",
    sharedbank: "sharedbank",
    sharedbanks: "sharedbank",
    guild: "guilds",
    guilds: "guilds",
    npc_emote: "npc_emotes",
    npc_emotes: "npc_emotes",
    npc_spell: "npc_spells",
    npc_spells: "npc_spells",
    npc_spells_entry: "npc_spells_entries",
    npc_spells_entries: "npc_spells_entries",
    spawnentry: "spawnentry",
    spawnentries: "spawnentry",
    spawngroup: "spawngroup",
    spawngroups: "spawngroup",
    db_str: "db_str",
    db_strs: "db_str",
    account: "account",
    accounts: "account",
    mail: "mail",
    variable: "variables",
    variables: "variables",
    rule_value: "rule_values",
    rule_values: "rule_values",
    logsys_category: "logsys_categories",
    logsys_categories: "logsys_categories",
    discord_webhook: "discord_webhooks",
    discord_webhooks: "discord_webhooks",
    player_event_log: "player_event_logs",
    player_event_logs: "player_event_logs",
    player_event_log_setting: "player_event_log_settings",
    player_event_log_settings: "player_event_log_settings",
    trap: "traps",
    traps: "traps",
    ground_spawn: "ground_spawns",
    ground_spawns: "ground_spawns",
    forage: "forage",
    forages: "forage",
    fishing: "fishing",
    fishings: "fishing",
    graveyard: "graveyard",
    graveyards: "graveyard",
    blocked_spell: "blocked_spells",
    blocked_spells: "blocked_spells",
    data_bucket: "data_buckets",
    data_buckets: "data_buckets",
    content_flag: "content_flags",
    content_flags: "content_flags",
    adventure_template: "adventure_template",
    adventure_templates: "adventure_template",
    alternate_currency: "alternate_currency",
    alternate_currencies: "alternate_currency",
    title: "titles",
    titles: "titles",
    aura: "auras",
    auras: "auras",
    quest_global: "quest_globals",
    quest_globals: "quest_globals",
    keyring: "keyring",
    keyrings: "keyring",
    character_parcel: "character_parcels",
    character_parcels: "character_parcels",
    raid_member: "raid_members",
    raid_members: "raid_members",
    dynamic_zone_template: "dynamic_zone_templates",
    dynamic_zone_templates: "dynamic_zone_templates",
    shared_task: "shared_tasks",
    shared_tasks: "shared_tasks",
    chatchannel: "chatchannel",
    chatchannels: "chatchannel",
    merc_template: "merc_templates",
    merc_templates: "merc_templates",
    merc_type: "merc_types",
    merc_types: "merc_types",
  }
  if (aliases[last]) {
    return aliases[last]
  }
  if (tables && tables[last]) {
    return last
  }
  if (tables && last.endsWith("s") && tables[last.slice(0, -1)]) {
    return last.slice(0, -1)
  }
  const compact = last.replace(/_/g, "")
  if (tables) {
    const keys = Object.keys(tables)
    for (let i = 0; i < keys.length; i++) {
      if (keys[i].replace(/_/g, "") === compact) {
        return keys[i]
      }
    }
  }
  return last
}

function tableFromPath(path: string) {
  const clean = path.replace(/\/$/, "").split("?")[0]
  const parts = clean.split("/").filter(Boolean)
  const last = parts[parts.length - 1] || ""
  const prev = parts[parts.length - 2] || ""
  if (/^\d+$/.test(last) && prev && prev.indexOf("admin") === -1) {
    const resolved = resolveTableName(prev)
    if (resolved && tables && tables[resolved]) {
      return resolved
    }
  }
  return resolveTableName(last)
}

export function createDemoAdapter() {
  return async function demoAdapter(config: any) {
    await loadTables()
    const id = pathOf(config)
    if (/^\d+$/.test(id.split("/").pop() || "") && tables && id.indexOf("/admin/") === -1) {
      const name = tableFromPath(id)
      const key = id.split("/").pop()
      const rows = (tables[name] || []).filter((row) => String(row.id) === String(key) || String(row.zoneidnumber) === String(key))
      if (rows.length) {
        const row = Object.assign({}, rows[0])
        if (name === "npc_spells") {
          row.npc_spells_entries = (tables.npc_spells_entries || []).filter((e) => String(e.npc_spells_id) === String(key))
        }
        return ok(row, config)
      }
    }
    return ok(demoBody(config), config)
  }
}

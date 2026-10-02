import {SpireApi} from "@/app/api/spire-api"

export type ZoneControllerFile = {
  kind: string
  relPath: string
  exists: boolean
  size: number
}

export type ZoneControllerSummary = {
  zoneId: number
  name?: string
  objective?: string
  tip?: string
  respawn?: string
  customCount: number
  ignoreCount: number
  depopCount: number
  lootTableCount: number
  itemCount: number
  files?: { [k: string]: ZoneControllerFile }
}

export type ZoneControllerLootRow = { chance: string; id: string }

export type ZoneControllerCustomMob = {
  name: string
  type: string
  mobId: string
  static: string
  mods: { [k: string]: string }
  loot: ZoneControllerLootRow[]
}

export type ZoneControllerLootTable = { id: string; count: number; items: string[] }

export type ZoneControllerItem = {
  name: string
  itemId: string
  spellId: string
  globalKey: string
  isGlobal: string
  description: string
}

export type ZoneControllerDetail = ZoneControllerSummary & {
  basedata?: { [kind: string]: { [stat: string]: string } }
  custom?: ZoneControllerCustomMob[]
  ignore?: string[]
  depop?: string[]
  loot?: ZoneControllerLootTable[]
  items?: ZoneControllerItem[]
}

export type ZoneControllerIndex = {
  ok: boolean
  error?: string
  questsDir?: string
  dataDir?: string
  dataRel?: string
  scriptRel?: string
  zones: ZoneControllerSummary[]
}

export type ZoneControllerTalent = {
  key: string
  classFamily: string
  bucketSuffix: string
  legacyGlobalKey: string
  valueType: string
  stateType: string
}

export type ZoneControllerUnlock = {
  key: string
  spellId: string
  spellName: string
  bucketSuffix: string
  expectedValue: string
  legacyGlobalKey: string
  notes: string
}

export type ZoneControllerSystems = {
  talents: ZoneControllerTalent[]
  unlocks: ZoneControllerUnlock[]
  talentMeta?: any
  unlockMeta?: any
}

export const ZONE_CONTROLLER_BROWSE = [
  {id: "mobs", label: "View Mob List", command: "viewallmobs"},
  {id: "loot", label: "View All Loot Tables", command: "viewallloottables"},
  {id: "items", label: "View All Items", command: "viewalllootdrops_0"},
  {id: "info", label: "View Zone Info", command: "viewzoneinfo"},
  {id: "ignore", label: "View Ignored Mobs", command: "viewignoredmobs"},
  {id: "depop", label: "View Depop Mobs", command: "viewdepopmobs"},
  {id: "diag", label: "View ZC Runtime", command: "viewzcdiag"},
  {id: "commands", label: "View Remote Commands", command: "viewremotecommands"},
]

export const ZONE_CONTROLLER_LIVE = [
  {label: "Apply Updates", command: "applyallchanges"},
  {label: "Reset Updates", command: "resetallchanges"},
  {label: "Save Updates", command: "saveallchanges"},
  {label: "Repop Bosses", command: "repopallstaticbosses"},
  {label: "Depop Bosses", command: "depopallstaticbosses"},
  {label: "Rebuff Zone", command: "rebuffzone"},
  {label: "Refresh Zone Data", command: "refreshzonedata"},
]

export const ZONE_CONTROLLER_WORLD = [
  {label: "Reload quests", action: "reload"},
  {label: "Reboot zone", action: "reboot"},
  {label: "Boot zone", action: "boot"},
]

export const ZONE_CONTROLLER_REMOTE = [
  {cmd: "!initdata <zoneid>", info: "Reload JSON into the popped zone controller from any zone."},
  {cmd: "!repop <zoneid>", info: "Repop static named mobs in that zone."},
  {cmd: "!depop <zoneid>", info: "Depop static named mobs in that zone."},
  {cmd: "!reloadzone <zoneid>", info: "Reload live decoded zone controller data."},
  {cmd: "!remoteaddignore <mob name>", info: "Add current target to the ignore list."},
  {cmd: "!remoteadddepop <mob name>", info: "Add current target to the depop list."},
  {cmd: "!remoteaddcustomtypetrash <mob name>", info: "Add current target as custom trash."},
  {cmd: "!remoteaddcustomtypeboss <mob name>", info: "Add current target as custom boss."},
  {cmd: "!remoteaddcustomtyperaid <mob name>", info: "Add current target as custom raid."},
  {cmd: "!remoteadditem <itemid[,itemid]>", info: "Add item id(s) to the loot item list."},
  {cmd: "!remoteaddtable <table[,table]>", info: "Add loot table name(s)."},
  {cmd: "!remotebatchassociation <table,itemid,...>", info: "Batch attach existing items to existing tables."},
  {cmd: "!remotesave", info: "Write in-memory changes back to JSON."},
  {cmd: "!remoteupdatetimer <seconds>", info: "Set respawn on the selected mob's spawngroup."},
  {cmd: "!remoteupdatealltimer <seconds>", info: "Set respawn for all spawngroups of the selected name."},
  {cmd: "!togglevis", info: "Hide or unhide the zone controller NPC."},
  {cmd: "!showloot", info: "Print item varlinks in say."},
  {cmd: "zcdiag", info: "Show live spawn/buff counters while hailing the controller."},
  {cmd: "!zc <command> / #zc <command>", info: "Direct GM command mode without hailing (example: !zc viewallmobs)."},
]

export class ZoneControllerApi {
  static async status() {
    const r = await SpireApi.v1().get("/admin/zone-controller/status")
    return r.data
  }

  static async list(): Promise<ZoneControllerIndex> {
    const r = await SpireApi.v1().get("/admin/zone-controller/zones")
    return r.data
  }

  static async zone(id: number): Promise<ZoneControllerDetail> {
    const r = await SpireApi.v1().get("/admin/zone-controller/zones/" + id)
    return r.data
  }

  static async systems(): Promise<ZoneControllerSystems> {
    const r = await SpireApi.v1().get("/admin/zone-controller/systems")
    return r.data
  }

  static async create(body: any): Promise<ZoneControllerWritePlan> {
    const r = await SpireApi.v1().post("/admin/zone-controller/create", body)
    return r.data
  }

  static async applyTier(body: any): Promise<ZoneControllerWritePlan> {
    const r = await SpireApi.v1().post("/admin/zone-controller/tier", body)
    return r.data
  }

  static async addMobs(body: any): Promise<ZoneControllerWritePlan> {
    const r = await SpireApi.v1().post("/admin/zone-controller/mobs", body)
    return r.data
  }

  static async recipes() {
    const r = await SpireApi.v1().get("/admin/zone-controller/recipes")
    return r.data
  }

  static async saveRecipe(body: any): Promise<ZoneControllerWritePlan> {
    const r = await SpireApi.v1().post("/admin/zone-controller/recipes", body)
    return r.data
  }

  static async deleteRecipe(id: string): Promise<ZoneControllerWritePlan> {
    const r = await SpireApi.v1().delete("/admin/zone-controller/recipes/" + encodeURIComponent(id))
    return r.data
  }

  static async recipeFromZone(id: number) {
    const r = await SpireApi.v1().get("/admin/zone-controller/recipes/from/" + id)
    return r.data
  }

  static async factory(body: any) {
    const r = await SpireApi.v1().post("/admin/zone-controller/factory", body)
    return r.data
  }

  static async classify(body: any) {
    const r = await SpireApi.v1().post("/admin/zone-controller/classify", body)
    return r.data
  }

  static async validate(zoneIds?: number[]) {
    const ids = zoneIds || []
    const q = ids.length ? "?zones=" + ids.join(",") : ""
    const r = await SpireApi.v1().get("/admin/zone-controller/validate" + q)
    return r.data
  }

  static async apply(body: any) {
    const r = await SpireApi.v1().post("/admin/zone-controller/apply", body)
    return r.data
  }

  static async live(zoneIds?: number[]) {
    const ids = zoneIds || []
    const q = ids.length ? "?zones=" + ids.join(",") : ""
    const r = await SpireApi.v1().get("/admin/zone-controller/live" + q)
    return r.data
  }
}

export type ZoneControllerWriteFile = {
  zoneId: number
  kind: string
  relPath: string
  action: string
}

export type ZoneControllerWritePlan = {
  ok: boolean
  dryRun: boolean
  writes?: ZoneControllerWriteFile[]
  backups?: string[]
  warnings?: string[]
  error?: string
}

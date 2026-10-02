import {SpireApi} from "@/app/api/spire-api"

export class ContentFactoryApi {
  static async status() {
    const r = await SpireApi.v1().get("/admin/content-factory/status")
    return r.data
  }

  static async ids() {
    const r = await SpireApi.v1().get("/admin/content-factory/ids")
    return r.data
  }

  static async meta() {
    const r = await SpireApi.v1().get("/admin/content-factory/meta")
    return r.data
  }

  static async search(kind: string, q: string, limit = 20) {
    const r = await SpireApi.v1().get("/admin/content-factory/search", {params: {kind, q, limit}})
    return r.data
  }

  static async runs() {
    const r = await SpireApi.v1().get("/admin/content-factory/runs")
    return r.data
  }

  static async census(zone: string) {
    const r = await SpireApi.v1().get("/admin/content-factory/census", {params: {zone}})
    return r.data
  }

  static async catalog(kind: string, limit = 80) {
    const r = await SpireApi.v1().get("/admin/content-factory/catalog", {params: {kind, limit}})
    return r.data
  }

  static async merchant(id: number) {
    const r = await SpireApi.v1().get("/admin/content-factory/merchant", {params: {id}})
    return r.data
  }

  static async importZone(zone: string) {
    const r = await SpireApi.v1().get("/admin/content-factory/import-zone", {params: {zone}})
    return r.data
  }

  static async draft() {
    const r = await SpireApi.v1().get("/admin/content-factory/draft")
    return r.data
  }

  static async saveDraft(body: any) {
    const r = await SpireApi.v1().put("/admin/content-factory/draft", body)
    return r.data
  }

  static async spellSet(id: number) {
    const r = await SpireApi.v1().get("/admin/content-factory/spell-set", {params: {id}})
    return r.data
  }

  static async saveSpellSet(body: any) {
    const r = await SpireApi.v1().post("/admin/content-factory/spell-set", body)
    return r.data
  }

  static async spells(body: any) {
    const r = await SpireApi.v1().post("/admin/content-factory/spells", body)
    return r.data
  }

  static async items(body: any) {
    const r = await SpireApi.v1().post("/admin/content-factory/items", body)
    return r.data
  }

  static async loot(body: any) {
    const r = await SpireApi.v1().post("/admin/content-factory/loot", body)
    return r.data
  }

  static async attach(body: any) {
    const r = await SpireApi.v1().post("/admin/content-factory/attach", body)
    return r.data
  }

  static async npcSpells(body: any) {
    const r = await SpireApi.v1().post("/admin/content-factory/npc-spells", body)
    return r.data
  }

  static async npcs(body: any) {
    const r = await SpireApi.v1().post("/admin/content-factory/npcs", body)
    return r.data
  }

  static async pawn(body: any) {
    const r = await SpireApi.v1().post("/admin/content-factory/pawn", body)
    return r.data
  }

  static async give(body: any) {
    const r = await SpireApi.v1().post("/admin/content-factory/give", body)
    return r.data
  }

  static async probe(body: any) {
    const r = await SpireApi.v1().post("/admin/content-factory/probe", body)
    return r.data
  }

  static async pipeline(body: any) {
    const r = await SpireApi.v1().post("/admin/content-factory/pipeline", body)
    return r.data
  }

  static async exportClient(dbstr = false) {
    const r = await SpireApi.v1().post("/admin/content-factory/export", {dbstr})
    return r.data
  }

  static async reload(zoneIds: number[]) {
    const r = await SpireApi.v1().post("/admin/content-factory/reload", {zoneIds})
    return r.data
  }

  static async syncZone(zoneIds: number[], dryRun = false) {
    const r = await SpireApi.v1().post("/admin/content-factory/sync-zone", {zoneIds, dryRun})
    return r.data
  }

  static async undo(runId: string) {
    const r = await SpireApi.v1().post("/admin/content-factory/undo", {runId})
    return r.data
  }
}

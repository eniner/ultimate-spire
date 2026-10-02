import {SpireApi} from "@/app/api/spire-api"

export const ULT_CLASS_NAMES = [
  "", "Warrior", "Cleric", "Paladin", "Ranger", "Shadow Knight", "Druid",
  "Monk", "Bard", "Rogue", "Shaman", "Necromancer", "Wizard", "Magician",
  "Enchanter", "Beastlord", "Berserker",
]

export const ULT_CLASS_SHORT = [
  "", "WAR", "CLR", "PAL", "RNG", "SK", "DRU", "MNK", "BRD", "ROG",
  "SHM", "NEC", "WIZ", "MAG", "ENC", "BST", "BER",
]

export type UltWritePlan = {
  ok: boolean
  dryRun: boolean
  kind?: string
  action?: string
  relPath?: string
  key?: string
  backups?: string[]
  warnings?: string[]
  error?: string
}

export class UltimateSystemsApi {
  static async status() {
    const r = await SpireApi.v1().get("/admin/ultimate-systems/status")
    return r.data
  }

  static async catalog() {
    const r = await SpireApi.v1().get("/admin/ultimate-systems/catalog")
    return r.data
  }

  static async talent(id: string) {
    const r = await SpireApi.v1().get("/admin/ultimate-systems/talents/" + encodeURIComponent(id))
    return r.data
  }

  static async traits() {
    const r = await SpireApi.v1().get("/admin/ultimate-systems/traits")
    return r.data
  }

  static async runewords() {
    const r = await SpireApi.v1().get("/admin/ultimate-systems/runewords")
    return r.data
  }

  static async write(body: any): Promise<UltWritePlan> {
    const r = await SpireApi.v1().post("/admin/ultimate-systems/write", body)
    return r.data
  }
}

export function classLabel(id: number): string {
  return ULT_CLASS_NAMES[id] || ("Class " + id)
}

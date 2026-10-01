import {DB_CLASSES_ICONS} from "./constants/eq-class-icon-constants"
import {DB_PLAYER_CLASSES} from "./constants/eq-classes-constants"
import {DB_PLAYER_RACES} from "./constants/eq-races-constants"
import {DB_DIETIES_FULL} from "./constants/eq-deities-constants"

const CLASS_BITS = [1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768]

export function classBitIcon(bit: number): number {
  const id = Math.round(Math.log2(bit)) + 1
  return Number(DB_CLASSES_ICONS[id]) || 0
}

export function raceBitIcon(bit: number): number {
  const recs = Object.keys(DB_PLAYER_RACES).map((key) => DB_PLAYER_RACES[key])
  for (let i = 0; i < recs.length; i++) {
    const rec = recs[i]
    if (rec && Number(rec.mask) === bit && rec.icon) {
      return Number(rec.icon)
    }
  }
  return 0
}

export function deityBitIcon(bit: number): number {
  const recs = Object.keys(DB_DIETIES_FULL).map((key) => DB_DIETIES_FULL[key])
  for (let i = 0; i < recs.length; i++) {
    const rec = recs[i]
    if (rec && rec.mask === bit && rec.icon) {
      return Number(rec.icon)
    }
  }
  return 0
}

export function itemClassIcons(classes: number): Array<{id: number, icon: number, name: string}> {
  if (!classes || classes >= 65535) {
    return []
  }
  const out: Array<{id: number, icon: number, name: string}> = []
  for (let i = 0; i < CLASS_BITS.length; i++) {
    const bit = CLASS_BITS[i]
    if ((classes & bit) !== bit) {
      continue
    }
    const id = i + 1
    out.push({
      id,
      icon: Number(DB_CLASSES_ICONS[id]) || 0,
      name: String(DB_PLAYER_CLASSES[String(id)] || ("Class " + id)),
    })
  }
  return out
}

export function itemRaceIcons(races: number): Array<{id: number, icon: number, name: string}> {
  if (!races || races >= 65535) {
    return []
  }
  const out: Array<{id: number, icon: number, name: string}> = []
  Object.keys(DB_PLAYER_RACES).forEach((key) => {
    const rec = (DB_PLAYER_RACES as any)[key]
    const mask = Number(rec && rec.mask)
    if (rec && rec.icon && mask && (races & mask) === mask) {
      out.push({id: Number(key), icon: Number(rec.icon), name: rec.race})
    }
  })
  return out
}

export function itemDeityIcons(deity: number): Array<{id: number, icon: number, name: string}> {
  if (!deity) {
    return []
  }
  const out: Array<{id: number, icon: number, name: string}> = []
  Object.keys(DB_DIETIES_FULL).forEach((key) => {
    const rec = (DB_DIETIES_FULL as any)[key]
    if (rec && rec.icon && rec.mask && (deity & rec.mask) === rec.mask) {
      out.push({id: Number(key), icon: Number(rec.icon), name: rec.name})
    }
  })
  return out
}

export function maskFieldIcon(field: string, bit: number): number {
  if (field === "classes") {
    return classBitIcon(bit)
  }
  if (field === "races") {
    return raceBitIcon(bit)
  }
  if (field === "deity") {
    return deityBitIcon(bit)
  }
  return 0
}

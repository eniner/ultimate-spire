import {PLAYER_INVENTORY_SLOTS} from "@/app/constants/eq-inventory-constants"

export const WORN_BEGIN = 0
export const WORN_END = 22
export const GENERAL_BEGIN = 23
export const GENERAL_END = 30
export const CURSOR_SLOT = 31
export const EXTRA_GENERAL_SLOT = 32

export const BANK_BEGIN = 2000
export const BANK_END = 2023
export const SHARED_BEGIN = 2500
export const SHARED_END = 2501

// This emu stores bag interiors in 200-wide blocks, not the old 10-slot 262 range.
export const INV_BAGS_BEGIN = 4010
export const INV_BAG_STRIDE = 200
export const INV_BAG_PARENTS = 10
export const BANK_BAGS_BEGIN = 6010
export const BANK_BAG_PARENTS = 24
export const SHARED_BAGS_BEGIN = 11010
export const SHARED_BAG_PARENTS = 2
export const MAX_BAG_SLOTS = 200

export const WORN_SLOTS = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22]
export const GENERAL_SLOTS = [23, 24, 25, 26, 27, 28, 29, 30]
export const EXTRA_SLOTS = [EXTRA_GENERAL_SLOT]
export const BANK_SLOTS = Array.from({length: BANK_BAG_PARENTS}, (_, i) => BANK_BEGIN + i)
export const SHARED_SLOTS = [SHARED_BEGIN, SHARED_END]

export const WORN_PAPERDOLL: Array<number | null> = [
  0, 2, 4,
  1, 3, 5,
  null, 17, 6,
  null, 7, 8,
  9, 12, 10,
  null, 18, 11,
  13, 19, 14,
  null, 20, null,
  15, 21, 16,
  null, 22, null,
]

const SHORT_LABELS: Record<number, string> = {
  0: "Charm",
  1: "Ear",
  2: "Head",
  3: "Face",
  4: "Ear",
  5: "Neck",
  6: "Shld",
  7: "Arms",
  8: "Back",
  9: "Wrist",
  10: "Wrist",
  11: "Range",
  12: "Hands",
  13: "Prim",
  14: "Off",
  15: "Ring",
  16: "Ring",
  17: "Chest",
  18: "Legs",
  19: "Feet",
  20: "Waist",
  21: "Power",
  22: "Ammo",
  31: "Cursor",
  32: "9",
}

export function shortSlotLabel(slotId: number): string {
  if (SHORT_LABELS[slotId]) {
    return SHORT_LABELS[slotId]
  }
  if (slotId >= GENERAL_BEGIN && slotId <= GENERAL_END) {
    return String(slotId - GENERAL_BEGIN + 1)
  }
  if (slotId >= BANK_BEGIN && slotId <= BANK_END) {
    return String(slotId - BANK_BEGIN + 1)
  }
  if (slotId >= SHARED_BEGIN && slotId <= SHARED_END) {
    return "Shared " + (slotId - SHARED_BEGIN + 1)
  }
  const parent = bagParentSlot(slotId)
  if (parent != null) {
    const begin = bagBegin(parent)
    return String(slotId - (begin || 0) + 1)
  }
  return String(slotId)
}

export function displayCharges(charges: number | null | undefined): number {
  const n = Number(charges)
  if (!n || n <= 1 || n >= 10000) {
    return 0
  }
  return n
}

export function wornName(slotId: number): string {
  const slot = PLAYER_INVENTORY_SLOTS[slotId]
  if (!slot) {
    return "Slot " + slotId
  }
  if (slotId === 1) {
    return "Ear 1"
  }
  if (slotId === 4) {
    return "Ear 2"
  }
  if (slotId === 9) {
    return "Bracer 1"
  }
  if (slotId === 10) {
    return "Bracer 2"
  }
  if (slotId === 15) {
    return "Ring 1"
  }
  if (slotId === 16) {
    return "Ring 2"
  }
  return slot.name
}

export function wornMask(slotId: number): number {
  const slot = PLAYER_INVENTORY_SLOTS[slotId]
  return slot ? slot.mask : 0
}

export function isWorn(slotId: number): boolean {
  return slotId >= WORN_BEGIN && slotId <= WORN_END
}

export function isGeneral(slotId: number): boolean {
  return slotId >= GENERAL_BEGIN && slotId <= GENERAL_END
}

export function isExtraGeneral(slotId: number): boolean {
  return slotId === EXTRA_GENERAL_SLOT
}

export function isCursor(slotId: number): boolean {
  return slotId === CURSOR_SLOT
}

export function isBank(slotId: number): boolean {
  return slotId >= BANK_BEGIN && slotId <= BANK_END
}

export function isShared(slotId: number): boolean {
  return slotId >= SHARED_BEGIN && slotId <= SHARED_END
}

export function isInventoryParent(slotId: number): boolean {
  return isGeneral(slotId) || isCursor(slotId) || isExtraGeneral(slotId)
}

export function bagBegin(parentSlot: number): number | null {
  if (isInventoryParent(parentSlot)) {
    return INV_BAGS_BEGIN + (parentSlot - GENERAL_BEGIN) * INV_BAG_STRIDE
  }
  if (isBank(parentSlot)) {
    return BANK_BAGS_BEGIN + (parentSlot - BANK_BEGIN) * INV_BAG_STRIDE
  }
  if (isShared(parentSlot)) {
    return SHARED_BAGS_BEGIN + (parentSlot - SHARED_BEGIN) * INV_BAG_STRIDE
  }
  return null
}

export function bagParentSlot(slotId: number): number | null {
  if (slotId >= INV_BAGS_BEGIN && slotId < INV_BAGS_BEGIN + INV_BAG_PARENTS * INV_BAG_STRIDE) {
    return GENERAL_BEGIN + Math.floor((slotId - INV_BAGS_BEGIN) / INV_BAG_STRIDE)
  }
  if (slotId >= BANK_BAGS_BEGIN && slotId < BANK_BAGS_BEGIN + BANK_BAG_PARENTS * INV_BAG_STRIDE) {
    return BANK_BEGIN + Math.floor((slotId - BANK_BAGS_BEGIN) / INV_BAG_STRIDE)
  }
  if (slotId >= SHARED_BAGS_BEGIN && slotId < SHARED_BAGS_BEGIN + SHARED_BAG_PARENTS * INV_BAG_STRIDE) {
    return SHARED_BEGIN + Math.floor((slotId - SHARED_BAGS_BEGIN) / INV_BAG_STRIDE)
  }
  return null
}

export function bagChildSlots(parentSlot: number, bagslots: number): number[] {
  const begin = bagBegin(parentSlot)
  if (begin == null) {
    return []
  }
  const count = Math.max(1, Math.min(Number(bagslots) || 10, MAX_BAG_SLOTS))
  const slots: number[] = []
  for (let i = 0; i < count; i++) {
    slots.push(begin + i)
  }
  return slots
}

export function bagGridClass(count: number): string {
  if (count > 16) {
    return "inv-grid-5"
  }
  return "inv-grid-4"
}

export function slotLabel(slotId: number): string {
  if (isWorn(slotId)) {
    return wornName(slotId)
  }
  if (isGeneral(slotId)) {
    return "General " + (slotId - GENERAL_BEGIN + 1)
  }
  if (isCursor(slotId)) {
    return "Cursor"
  }
  if (isExtraGeneral(slotId)) {
    return "General 9"
  }
  if (isBank(slotId)) {
    return "Bank " + (slotId - BANK_BEGIN + 1)
  }
  if (isShared(slotId)) {
    return "Shared " + (slotId - SHARED_BEGIN + 1)
  }
  const parent = bagParentSlot(slotId)
  if (parent != null) {
    const begin = bagBegin(parent)
    return slotLabel(parent) + " · " + (slotId - (begin || 0) + 1)
  }
  return "Slot " + slotId
}

export function isKnownSlot(slotId: number): boolean {
  if (slotId >= WORN_BEGIN && slotId <= EXTRA_GENERAL_SLOT) {
    return true
  }
  if (slotId >= BANK_BEGIN && slotId <= BANK_END) {
    return true
  }
  if (slotId >= SHARED_BEGIN && slotId <= SHARED_END) {
    return true
  }
  return bagParentSlot(slotId) != null
}

export function isSharedContext(slotId: number): boolean {
  return isShared(slotId) || isShared(bagParentSlot(slotId) || -1)
}

export function itemFitsWorn(item: {slots?: number} | null | undefined, slotId: number): boolean {
  if (!item || !isWorn(slotId)) {
    return true
  }
  const slots = Number(item.slots || 0)
  if (slots <= 0) {
    return false
  }
  const allWorn = 8388607
  if (slots === allWorn || slots === 65535) {
    return true
  }
  return (slots & wornMask(slotId)) !== 0
}

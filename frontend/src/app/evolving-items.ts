import {SpireApi}          from "@/app/api/spire-api";
import {SpireQueryBuilder} from "@/app/api/spire-query-builder";
import {ItemApi}           from "@/app/api";
import {DB_RACE_NAMES}     from "@/app/constants/eq-races-constants";

export const EVOLVING_TYPES = {
  1: "Experience",
  2: "Kills",
  3: "Race",
  4: "Zone",
  99: "UNK",
}

export const EVOLVING_EXP_SUBTYPES = {
  0: "All EXP",
  1: "Solo EXP",
  2: "Group EXP",
  3: "Raid EXP",
}

// select() takes database column names; the items name column is "Name"
export const EVOLVING_ITEM_FIELDS = ["id", "Name", "icon", "evoitem", "evoid", "evolvinglevel", "evomax"]

const NO_LIMIT = 1000000

export class EvolvingItems {

  static describeSubType(type, subType) {
    switch (Number(type)) {
      case 1:
        return EVOLVING_EXP_SUBTYPES[subType] || "UNK"
      case 2:
        return "N/A"
      case 3:
        return DB_RACE_NAMES[subType] || "UNK"
      case 4:
        return "Zone ID " + subType
      default:
        return "UNK"
    }
  }

  static async listDetails(evoId: number | null = null) {
    const q = (new SpireQueryBuilder()).limit(NO_LIMIT)
    if (evoId !== null) {
      q.where("item_evo_id", "=", evoId)
    }
    const r = await SpireApi.v1().get("/items_evolving_details", {params: q.get()})
    return r.data || []
  }

  static async createDetail(detail) {
    return SpireApi.v1().put("/items_evolving_detail", detail)
  }

  static async updateDetail(detail) {
    return SpireApi.v1().patch("/items_evolving_detail/" + detail.id, detail)
  }

  static async deleteDetail(id) {
    return SpireApi.v1().delete("/items_evolving_detail/" + id)
  }

  static async listCharacterItems() {
    const q = (new SpireQueryBuilder()).limit(NO_LIMIT)
    const r = await SpireApi.v1().get("/character_evolving_items", {params: q.get()})
    return r.data || []
  }

  // items flagged as evolving plus any item referenced by a detail row that is not flagged
  static async loadItemsFor(details) {
    const api     = new ItemApi(...SpireApi.cfg())
    const items: any = {}
    const flagged    = await api.listItems(
      // @ts-ignore
      (new SpireQueryBuilder()).select(EVOLVING_ITEM_FIELDS).where("evoid", ">", 0).limit(NO_LIMIT).get()
    )
    for (const i of (flagged.data || []) as any[]) {
      items[i.id] = i
    }

    const missing: any[] = [...new Set(details.map((d) => d.item_id))].filter((id: any) => !items[id])
    for (const id of missing) {
      const found = await this.loadItemRange(id, id)
      if (found[id]) {
        items[id] = found[id]
      }
    }

    return items
  }

  static async loadItemRange(fromId, toId) {
    const api = new ItemApi(...SpireApi.cfg())
    const r   = await api.listItems(
      // @ts-ignore
      (new SpireQueryBuilder())
        .select(EVOLVING_ITEM_FIELDS)
        .where("id", ">=", fromId)
        .where("id", "<=", toId)
        .limit(NO_LIMIT)
        .get()
    )
    const items: any = {}
    for (const i of (r.data || []) as any[]) {
      items[i.id] = i
    }
    return items
  }

  static groupChains(details) {
    const chains = {}
    for (const d of details) {
      if (!chains[d.item_evo_id]) {
        chains[d.item_evo_id] = []
      }
      chains[d.item_evo_id].push(d)
    }
    for (const id of Object.keys(chains)) {
      chains[id].sort((a, b) => a.item_evolve_level - b.item_evolve_level || a.id - b.id)
    }
    return chains
  }

  // mirrors how zone/common evolving_items.cpp resolves the next item:
  // next = row where item_evo_id == items.evoid and item_evolve_level == items.evolvinglevel + 1
  // the server caches rows keyed by item_id, so an item_id in more than one row keeps only one of them
  static analyzeChain(rows, items, allDetails) {
    const problems: string[] = []
    const levels    = rows.map((r) => r.item_evolve_level)
    const maxLevel  = Math.max(...levels)
    const rowIssues = {}

    const addRowIssue = (row, msg) => {
      rowIssues[row.id] = rowIssues[row.id] || []
      rowIssues[row.id].push(msg)
    }

    const missing: number[] = []
    for (let l = 1; l <= maxLevel; l++) {
      if (!levels.includes(l)) {
        missing.push(l)
      }
    }
    if (missing.length) {
      problems.push("Missing level " + missing.join(", ") + " - Evolving stops before each gap; XP transfers jump over it.")
    }

    const unknownTypes = [...new Set(rows.filter((r) => ![1, 2, 3, 4].includes(r.type)).map((r) => r.type))]
    if (unknownTypes.length) {
      problems.push("Type " + unknownTypes.join(", ") + " is not handled - Never gains progress from play; XP transfers treat every chain of this type as compatible.")
    }

    const zeroRequired = rows.filter((r) => r.item_evolve_level < maxLevel && !(r.required_amount > 0))
    if (zeroRequired.length) {
      problems.push("Required amount 0 on level " + zeroRequired.map((r) => r.item_evolve_level).join(", ") + " - Progress can't reach 100%; transferred XP skips to the last level.")
    }

    const seenLevels   = {}
    let mismatchedRows = 0
    for (const r of rows) {
      if (seenLevels[r.item_evolve_level]) {
        problems.push("Level " + r.item_evolve_level + " defined twice - The server picks one of them unpredictably.")
        addRowIssue(r, "Duplicate level")
        addRowIssue(seenLevels[r.item_evolve_level], "Duplicate level")
      }
      seenLevels[r.item_evolve_level] = r

      if (![1, 2, 3, 4].includes(r.type)) {
        addRowIssue(r, "Type " + r.type + " is ignored by the server")
      }
      if (r.item_evolve_level < maxLevel && !(r.required_amount > 0)) {
        addRowIssue(r, "Required amount must be above 0")
      }

      const inOtherRows = allDetails.filter((d) => d.item_id === r.item_id && d.id !== r.id)
      if (inOtherRows.length) {
        const where = inOtherRows.map((d) => "chain " + d.item_evo_id + " level " + d.item_evolve_level).join(", ")
        problems.push("Item " + r.item_id + " also in " + where + " - The server keeps only one of these rows.")
        addRowIssue(r, "Item also used in " + where)
      }

      const item = items[r.item_id]
      if (!item) {
        problems.push("Item " + r.item_id + " missing - Level " + r.item_evolve_level + " points to an item that isn't in the items table.")
        addRowIssue(r, "Item does not exist")
        continue
      }
      const columnIssues: string[] = []
      if (item.evoid !== r.item_evo_id) {
        columnIssues.push("items.evoid is " + item.evoid + ", should be " + r.item_evo_id)
      }
      if (item.evolvinglevel !== r.item_evolve_level) {
        columnIssues.push("items.evolvinglevel is " + item.evolvinglevel + ", should be " + r.item_evolve_level)
      }
      if (item.evomax !== maxLevel) {
        columnIssues.push("items.evomax is " + item.evomax + ", should be " + maxLevel)
      }
      if (item.evoitem !== 1) {
        columnIssues.push("items.evoitem is " + item.evoitem + ", should be 1")
      }
      if (columnIssues.length) {
        mismatchedRows++
        columnIssues.forEach((msg) => addRowIssue(r, msg))
      }
    }

    if (mismatchedRows) {
      problems.push(mismatchedRows + " item(s) out of sync - Their evoid / evolvinglevel / evomax / evoitem columns don't match this chain.")
    }

    return {problems: [...new Set(problems)], rowIssues, missing, maxLevel}
  }

  static splitProblem(p: string) {
    const i = p.indexOf(" - ")
    return i === -1 ? {title: p, detail: ""} : {title: p.slice(0, i), detail: p.slice(i + 3)}
  }
}

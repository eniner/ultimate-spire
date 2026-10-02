import {EXPANSION_ICONS_SMALL} from "@/app/constants/eq-expansion-icons";
import {App}                   from "@/constants/app";
import {EXPANSION_NAMES}       from "@/app/constants/eq-expansions";

export default class Expansions {
  // PEQ zone.expansion is 1-based (1=Classic, 2=Kunark).
  // Spire expansion IDs, icons, and min_expansion are 0-based (0=Classic, 1=Kunark).
  static ZONE_EXPANSION_OFFSET = 1

  static fromZoneTable(zoneExpansion) {
    const n = Number(zoneExpansion)
    return Number.isFinite(n) ? n - this.ZONE_EXPANSION_OFFSET : -1
  }

  static toZoneTable(expansionId) {
    return Number(expansionId) + this.ZONE_EXPANSION_OFFSET
  }

  static getExpansionIconUrlSmall(expansionId) {
    if (EXPANSION_ICONS_SMALL[expansionId]) {
      return App.ASSET_EXPANSION_ICON_SMALL_URL + EXPANSION_ICONS_SMALL[expansionId]
    }

    // return transparent base64 encoded image if nothing found
    return 'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7'
  }
  static getExpansionName(expansionId) {
    if (EXPANSION_NAMES[expansionId]) {
      return EXPANSION_NAMES[expansionId]
    }

    if (expansionId === -1) {
      return 'All'
    }

    // return unknown expansion if not found
    return 'Unknown'
  }
}

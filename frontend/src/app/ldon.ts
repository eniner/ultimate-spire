export const LDON_THEMES = [
  {id: 1, key: "guk", label: "Guk", prefix: "guk", points: "ldon_points_guk", wins: "guk_wins", losses: "guk_losses"},
  {id: 2, key: "mir", label: "Miragul", prefix: "mir", points: "ldon_points_mir", wins: "mir_wins", losses: "mir_losses"},
  {id: 3, key: "mmc", label: "Mistmoore", prefix: "mmc", points: "ldon_points_mmc", wins: "mmc_wins", losses: "mmc_losses"},
  {id: 4, key: "ruj", label: "Rujarkian", prefix: "ruj", points: "ldon_points_ruj", wins: "ruj_wins", losses: "ruj_losses"},
  {id: 5, key: "tak", label: "Takish", prefix: "tak", points: "ldon_points_tak", wins: "tak_wins", losses: "tak_losses"},
]

export const LDON_SEARCH_CATEGORIES = [
  {id: "zones", label: "Zones"},
  {id: "adventures", label: "Adventures"},
  {id: "characters", label: "Characters"},
]

export const LDON_ADVENTURE_TYPES = {
  1: "Collection",
  2: "Rescue",
  3: "Assassination",
  4: "Kill count",
}

export const CHAR_LDON_FIELDS = [
  "id", "account_id", "name", "last_name", "title", "suffix",
  "level", "class", "race", "gender", "deity", "zone_id", "zone_instance", "gm",
  "ldon_points_available", "ldon_points_guk", "ldon_points_mir",
  "ldon_points_mmc", "ldon_points_ruj", "ldon_points_tak",
]

const LDON_SHORT_NAME = /^(guk|mir|mmc|ruj|tak)[a-j]$/i

export function ldonShortName(value) {
  if (!value) {
    return ""
  }
  if (typeof value === "string") {
    return value.toLowerCase()
  }
  return String(value.short_name || value.zone || "").toLowerCase()
}

export function isLdonAdventureZone(value) {
  return LDON_SHORT_NAME.test(ldonShortName(value))
}

export function ldonThemeFromShortName(value) {
  const short = ldonShortName(value)
  return LDON_THEMES.find((theme) => short.indexOf(theme.prefix) === 0 && isLdonAdventureZone(short)) || null
}

export function ldonThemeFromId(id) {
  return LDON_THEMES.find((theme) => theme.id === Number(id)) || null
}

export function ldonAdventureType(type) {
  return LDON_ADVENTURE_TYPES[Number(type)] || ("Type " + type)
}

export function filterLdonZones(zones, themeKey, query) {
  const q = (query || "").trim().toLowerCase()
  return (zones || []).filter((zone) => {
    if (!isLdonAdventureZone(zone)) {
      return false
    }
    const theme = ldonThemeFromShortName(zone)
    if (themeKey && themeKey !== "all" && (!theme || theme.key !== themeKey)) {
      return false
    }
    if (!q) {
      return true
    }
    const hay = [
      zone.short_name,
      zone.long_name,
      zone.zoneidnumber,
      theme ? theme.label : "",
    ].join(" ").toLowerCase()
    return hay.indexOf(q) !== -1
  }).sort((a, b) => String(a.short_name).localeCompare(String(b.short_name)))
}

export function filterLdonAdventures(rows, themeKey, query) {
  const q = (query || "").trim().toLowerCase()
  return (rows || []).filter((row) => {
    if (!isLdonAdventureZone(row.zone)) {
      return false
    }
    const theme = ldonThemeFromId(row.theme) || ldonThemeFromShortName(row.zone)
    if (themeKey && themeKey !== "all" && (!theme || theme.key !== themeKey)) {
      return false
    }
    if (!q) {
      return true
    }
    const hay = [
      row.id,
      row.zone,
      row.text,
      row.min_level,
      row.max_level,
      theme ? theme.label : "",
      ldonAdventureType(row.type),
    ].join(" ").toLowerCase()
    return hay.indexOf(q) !== -1
  })
}

export type PeqEditorStatus = "dedicated" | "table" | "soon"

export type PeqEditorGroup = "content" | "zone" | "live"

export type PeqEditor = {
  id: string
  title: string
  php: string
  group: PeqEditorGroup
  status: PeqEditorStatus
  to: string
  blurb: string
}

export const PEQ_EDITORS: PeqEditor[] = [
  { id: "npcs", title: "NPCs", php: "npc", group: "content", status: "dedicated", to: "/zones", blurb: "Open a zone, then edit NPC types from the zone card or /npc/:id." },
  { id: "loot", title: "Loot", php: "loot", group: "content", status: "dedicated", to: "/loot", blurb: "Search loottables and preview drops. Full lootdrop tree is still thinner than PHP." },
  { id: "spawns", title: "Spawns", php: "spawn", group: "content", status: "table", to: "/editors/spawns", blurb: "spawn2 rows: zone, XYZ, heading, respawn, spawngroup. Map drag is on the zone page." },
  { id: "spawngroups", title: "Spawngroups", php: "spawn", group: "content", status: "table", to: "/editors/spawngroups", blurb: "Spawngroup limits, roam, and delay. Pair with Spawns." },
  { id: "merchants", title: "Merchants", php: "merchant", group: "content", status: "dedicated", to: "/merchants", blurb: "Merchant lists and sold items." },
  { id: "spells", title: "Spells", php: "spells", group: "content", status: "dedicated", to: "/spells", blurb: "Spell editor. NPC spell sets are under NPCs → Spells." },
  { id: "factions", title: "Factions", php: "faction", group: "content", status: "table", to: "/editors/factions", blurb: "faction_list name and base. Hit mods are a later pass." },
  { id: "tradeskills", title: "Tradeskills", php: "tradeskill", group: "content", status: "table", to: "/editors/tradeskills", blurb: "Recipe headers. Ingredient rows are a later pass." },
  { id: "zones", title: "Zones", php: "zone", group: "zone", status: "dedicated", to: "/zones", blurb: "Zone list, map, and spawn2 placement editor." },
  { id: "doors", title: "Doors", php: "misc", group: "zone", status: "table", to: "/editors/doors", blurb: "Doors and teleports. Also drawn on the zone map." },
  { id: "traps", title: "Traps", php: "misc", group: "zone", status: "table", to: "/editors/traps", blurb: "Zone traps: radius, effect, respawn." },
  { id: "ground-spawns", title: "Ground Spawns", php: "misc", group: "zone", status: "table", to: "/editors/ground-spawns", blurb: "Ground items by zoneid." },
  { id: "forage", title: "Forage", php: "misc", group: "zone", status: "table", to: "/editors/forage", blurb: "Forage table by zoneid." },
  { id: "fishing", title: "Fishing", php: "misc", group: "zone", status: "table", to: "/editors/fishing", blurb: "Fishing table by zoneid." },
  { id: "objects", title: "Objects", php: "misc", group: "zone", status: "table", to: "/editors/objects", blurb: "World objects / placeables." },
  { id: "graveyards", title: "Graveyards", php: "zone", group: "zone", status: "table", to: "/editors/graveyards", blurb: "Graveyard bind points." },
  { id: "blocked-spells", title: "Blocked Spells", php: "zone", group: "zone", status: "table", to: "/editors/blocked-spells", blurb: "Zone blocked spells." },
  { id: "grids", title: "Grids", php: "util", group: "zone", status: "table", to: "/editors/grids", blurb: "Path grids. PHP Utilities listed orphaned grids here." },
  { id: "tasks", title: "Tasks", php: "tasks", group: "content", status: "dedicated", to: "/tasks", blurb: "Task editor." },
  { id: "items", title: "Items", php: "items", group: "content", status: "dedicated", to: "/items", blurb: "Item editor plus evolving chains." },
  { id: "aas", title: "AAs", php: "aa", group: "content", status: "table", to: "/editors/aas", blurb: "aa_ability rows. Ranks/effects are a later pass." },
  { id: "adventures", title: "Adventures", php: "adventures", group: "content", status: "table", to: "/editors/adventures", blurb: "LDoN adventure templates." },
  { id: "altcur", title: "Alt Currency", php: "altcur", group: "content", status: "table", to: "/editors/altcur", blurb: "Alternate currency item definitions." },
  { id: "titles", title: "Titles", php: "titles", group: "content", status: "table", to: "/editors/titles", blurb: "Title prefixes and suffixes." },
  { id: "auras", title: "Auras", php: "auras", group: "content", status: "table", to: "/editors/auras", blurb: "Aura definitions keyed by type." },
  { id: "mercs", title: "Mercs", php: "mercs", group: "content", status: "table", to: "/editors/mercs", blurb: "Merc tables (templates, types, spells, stances, merchants)." },
  { id: "databuckets", title: "Data Buckets", php: "databuckets", group: "live", status: "table", to: "/editors/databuckets", blurb: "Runtime data buckets." },
  { id: "content-flags", title: "Content Flags", php: "content", group: "content", status: "table", to: "/editors/content-flags", blurb: "content_flags enable/disable." },
  { id: "quests", title: "Quests", php: "quest", group: "content", status: "dedicated", to: "/quest-api-explorer", blurb: "Quest API explorer. File quests still live on disk, not in PHP-style NPC attach UI." },
  { id: "qglobals", title: "QGlobals", php: "qglobal", group: "live", status: "table", to: "/editors/qglobals", blurb: "Quest globals. Save uses charid + npcid + zoneid + name." },
  { id: "players", title: "Players", php: "player", group: "live", status: "table", to: "/editors/players", blurb: "character_data sheets. Open Inventory from a player row for the paperdoll tool." },
  { id: "accounts", title: "Accounts", php: "account", group: "live", status: "table", to: "/editors/accounts", blurb: "Login accounts and status. Password is hidden." },
  { id: "guilds", title: "Guilds", php: "guild", group: "live", status: "table", to: "/editors/guilds", blurb: "Guilds. Members and bank are a later tree pass." },
  { id: "inventory", title: "Inventory", php: "inv", group: "live", status: "dedicated", to: "/editors/inventory", blurb: "Paperdoll, bags, bank, and item icons. Click a slot to place or clear items." },
  { id: "inventory-raw", title: "Inventory rows", php: "inv", group: "live", status: "table", to: "/editors/inventory-raw", blurb: "Raw inventory rows by character_id + slot_id." },
  { id: "keys", title: "Keys", php: "keys", group: "live", status: "table", to: "/editors/keys", blurb: "Character keyring." },
  { id: "mail", title: "Mail", php: "mail", group: "live", status: "table", to: "/editors/mail", blurb: "In-game mail." },
  { id: "parcels", title: "Parcels", php: "parcels", group: "live", status: "table", to: "/editors/parcels", blurb: "Character parcel / mailbox items." },
  { id: "pvp", title: "PVP", php: "pvp", group: "live", status: "table", to: "/editors/pvp", blurb: "Player PVP points, kills, and deaths. PHP leaderboard." },
  { id: "expeditions", title: "Expeditions", php: "expeditions", group: "live", status: "table", to: "/editors/expeditions", blurb: "Dynamic zone templates." },
  { id: "sharedtasks", title: "Shared Tasks", php: "sharedtasks", group: "live", status: "table", to: "/editors/sharedtasks", blurb: "Shared task runtime rows." },
  { id: "chat", title: "Chat", php: "chat", group: "live", status: "table", to: "/editors/chat", blurb: "Chat channels." },
  { id: "util", title: "Utilities", php: "util", group: "live", status: "dedicated", to: "/editors/util", blurb: "Backups, reload, grids, and the PHP utility shortcuts." },
  { id: "server", title: "Server", php: "server", group: "live", status: "dedicated", to: "/admin", blurb: "Server admin, rules, reload, backups, players online." },
]

export const PEQ_EDITOR_GROUPS: { id: PeqEditorGroup; title: string }[] = [
  { id: "content", title: "Game data" },
  { id: "zone", title: "Zone / misc" },
  { id: "live", title: "Live server" },
]

export function getPeqEditor(id: string): PeqEditor | undefined {
  return PEQ_EDITORS.find((e) => e.id === id)
}

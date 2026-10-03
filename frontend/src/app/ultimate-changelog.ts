export type ChangelogItem = {
  area: string
  text: string
}

export type ChangelogRelease = {
  version: string
  date: string
  items: ChangelogItem[]
}

export const ULTIMATE_CHANGELOG: ChangelogRelease[] = [
  {
    version: "Unreleased",
    date: "",
    items: [],
  },
  {
    version: "4.24.6",
    date: "10/3/2026",
    items: [
      {area: "Changelog", text: "In-app Ultimate notes at /changelog. Filter versions and generate markdown for CHANGELOG.md or a GitHub release. Vue serve no longer depends on a Go rebuild to show new notes."},
      {area: "Zone Guide", text: "In-app walkthrough for new zone kits, minting items, cloning NPCs, the pipeline, and Apply live. Linked from Zone Controller, Builder, Tier Factory, Content Factory, and Talents / Runewords."},
      {area: "Zone Controller pack", text: "Public starter zip with the live zone_controller.pl, blank ultimatedata templates, NPC 2000986 SQL, Perl requirements, and placement directions. No live item or NPC kits."},
      {area: "Demo", text: "Read-only GitHub Pages app plus landing-page directions for the pack, Sage, and Lantern. Sage / Lantern / 2D maps stay out of the zip."},
    ],
  },
  {
    version: "4.24.5",
    date: "10/2/2026",
    items: [
      {area: "Content Factory", text: "Census a zone, cherry-pick trash vs named, mint item kits and dual loot tables, clone NPC spell sets into the 50000–65535 band, then run one pipeline that writes clones at 800000+ instead of editing vanilla PEQ. Lint, draft, pawn, give, probe, undo."},
      {area: "Tier Factory", text: "Named recipes with trash / boss / raid numbers plus a gear kit. Stamp copies that kit onto many zones. Ladder steps each zone by the multiplier and mints a new item ID range so T2 is not wearing T1 loot."},
      {area: "LDoN", text: "Theme-first zone and adventure search (Guk, Miragul, Mistmoore, Rujarkian, Takish) plus character points and wins."},
      {area: "Keys / Flags", text: "Owned and known keys plus quest, bucket, zone, and account flags on one character page."},
      {area: "Zone Controller", text: "Live apply / classify / coverage / validate. Queues the same hail commands (refreshzonedata, rebuffzone, applyallchanges) through ultimatedata/_spire_commands."},
    ],
  },
  {
    version: "4.24.4",
    date: "10/1/2026",
    items: [
      {area: "Zone Controller", text: "Browse configured ultimatedata zones. Builder creates or clones blank kits, applies trash / boss / raid basedata, and batch-adds custom mobs."},
      {area: "Talents / Runewords", text: "Edit the talent catalog, rank buckets, unlock flags, class specializations, trait vendor list, and runeword combo recipes. Preview then write JSON with backups."},
      {area: "Tasks", text: "Searchable task table instead of the old 900-row native select."},
      {area: "Zones", text: "Expansion labels match PEQ (Classic is 1, Kunark is 2). Zone Controller badge jumps to configured kits."},
      {area: "Atlas", text: "3D Lantern mesh with spawn overlays and PEQ grid / grid_entries pathing. Sage connect stays on the EQ client folder."},
      {area: "Evolving items", text: "Targeted tables: chain details first, then item columns. Type 4 is zone-kill."},
    ],
  },
  {
    version: "4.24.3",
    date: "10/1/2026",
    items: [
      {area: "Release", text: "Drop-in Windows exe. Put it next to eqemu_config.json and run it. Auto-update reads eniner/ultimate-spire, not EQEmu/spire."},
      {area: "PEQ Editors", text: "Hub maps every PHP editor tab onto a local page or table editor."},
      {area: "Inventory", text: "Paperdoll, bags, bank, shared bank, and parcels with item icons."},
      {area: "Items", text: "Search with class / race / deity chips and item icons."},
      {area: "Server files", text: "Point Spire at a quests folder and edit perl / lua in place."},
      {area: "Website", text: "Discord user ↔ EQ account link and web_users roles (user / admin)."},
      {area: "Privacy", text: "Hide private details blurs names, hosts, keys, passwords, and paths for streams."},
      {area: "2D zones", text: "Map from first-launch eq-asset-preview. Spawn2 drag-save on the zone page."},
    ],
  },
]

export function formatUltimateChangelog(releases: ChangelogRelease[] = ULTIMATE_CHANGELOG) {
  return releases.filter((release) => release.items.length).map((release) => {
    const heading = release.date ? `## [${release.version}] ${release.date}` : `## [${release.version}]`
    const items = release.items.map((item) => `* **${item.area}** ${item.text}`).join("\n")
    return heading + "\n\n" + items
  }).join("\n\n") + "\n"
}

export function formatUltimateReleaseNotes(releases: ChangelogRelease[] = ULTIMATE_CHANGELOG) {
  const notes = formatUltimateChangelog(releases.filter((r) => r.version !== "Unreleased"))
  return [
    "Drop-in Windows build of Ultimate Spire.",
    "",
    "1. Download `spire-windows-amd64.exe.zip`",
    "2. Unzip `spire-windows-amd64.exe`",
    "3. Put the exe in the same folder as your emulator `eqemu_config.json`",
    "4. Double-click it",
    "",
    "Spire opens a browser on a free port between 8090 and 8099. Maps, icons, and 3D previews download on first launch.",
    "",
    "Use this release, not the official EQEmu/spire exe, or auto-update will replace Ultimate Spire with stock Spire.",
    "",
    notes.trim(),
    "",
    "Walkthrough: https://eniner.github.io/ultimate-spire-demo/",
    "",
  ].join("\n")
}

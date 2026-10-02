<template>
  <content-area>
    <eq-window :title="pageTitle">
      <b-alert show variant="danger" v-if="error">
        <i class="fa fa-warning"></i> {{ error }}
      </b-alert>

      <app-loader :is-loading="!loaded" class="mt-3 mb-3"/>

      <div v-if="loaded">
        <div class="ui-toolbar">
          <router-link v-if="zoneId || isSystems" class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER">
            All ZC zones
          </router-link>
          <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONES">Zones</router-link>
          <router-link
            v-if="peqZone && peqZone.short_name"
            class="btn btn-sm btn-dark"
            :to="'/zone/' + peqZone.short_name + '?v=' + (peqZone.version || 0)"
          >
            Zone map
          </router-link>
          <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER_BUILDER">Build zones / tiers</router-link>
          <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER_FACTORY">Tier Factory</router-link>
          <router-link class="btn btn-sm btn-dark" :to="ROUTE.CONTENT_FACTORY">Content Factory</router-link>
          <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER_SYSTEMS">Talents / Unlocks</router-link>
          <router-link class="btn btn-sm btn-dark" :to="ROUTE.ULTIMATE_SYSTEMS">Edit talents / runewords</router-link>
          <a v-if="scriptHref" class="btn btn-sm btn-dark" :href="scriptHref">zone_controller.pl</a>
          <div class="ui-stat-line ml-auto">
            <template v-if="index && !zoneId && !isSystems">
              <b>{{ (index.zones || []).length }}</b> configured zones
              <span class="zc-dot">·</span>
              JSON under <code>quests/global/ultimatedata</code>
            </template>
            <template v-else-if="detail">
              <b>{{ (detail.custom || []).length }}</b> custom
              <span class="zc-dot">·</span>
              <b>{{ (detail.loot || []).length }}</b> loot tables
              <span class="zc-dot">·</span>
              <b>{{ (detail.items || []).length }}</b> items
            </template>
          </div>
        </div>

        <div v-if="!zoneId && !isSystems">
          <eq-window title="Build faster" class="mb-3">
            <p class="zc-copy">
              In-game the controller is one zone at a time: hail, add mobs, set basedata, save JSON.
              The Tier Factory writes a named recipe onto many zones at once: basedata, cloned gear,
              spawn classify, then apply/reload.
            </p>
            <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER_BUILDER">Open builder</router-link>
            <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER_FACTORY">Tier Factory</router-link>
            <router-link class="btn btn-sm btn-dark" :to="ROUTE.CONTENT_FACTORY">Content Factory</router-link>
          </eq-window>

          <eq-window title="Systems" class="mb-3">
            <p class="zc-copy">
              Talent ranks and unlock flags live next to the zone JSON
              (<code>talent_rank_registry.json</code>, <code>unlock_state_registry.json</code>).
              They are bucket-backed player systems, not per-zone spawn data.
            </p>
            <div class="zc-system-links">
              <router-link class="btn btn-sm btn-dark" :to="{path: ROUTE.ZONE_CONTROLLER_SYSTEMS, query: {kind: 'talents'}}">
                Talent ranks
              </router-link>
              <router-link class="btn btn-sm btn-dark" :to="{path: ROUTE.ZONE_CONTROLLER_SYSTEMS, query: {kind: 'unlocks'}}">
                Unlock states
              </router-link>
              <router-link class="btn btn-sm btn-dark" :to="ROUTE.ULTIMATE_SYSTEMS">
                Create / edit talents, traits, runewords
              </router-link>
            </div>
          </eq-window>

          <eq-window title="Configured zones">
            <div class="ui-toolbar">
              <input
                v-model="zoneSearch"
                type="search"
                class="form-control form-control-sm"
                style="width: 280px"
                placeholder="Zone ID, short name, or long name"
              >
            </div>
            <table class="eq-table eq-highlight-rows" style="width: 100%">
              <thead>
              <tr>
                <th style="width: 70px">ID</th>
                <th>Zone</th>
                <th style="width: 90px">Live</th>
                <th style="width: 80px" class="text-right">Custom</th>
                <th style="width: 80px" class="text-right">Ignore</th>
                <th style="width: 80px" class="text-right">Depop</th>
                <th style="width: 90px" class="text-right">Loot</th>
                <th style="width: 90px" class="text-right">Items</th>
              </tr>
              </thead>
              <tbody>
              <tr
                v-for="z in filteredIndex"
                :key="z.zoneId"
                class="zc-row"
                @click="openZone(z.zoneId)"
              >
                <td class="tabular"><b>{{ z.zoneId }}</b></td>
                <td>
                  <b>{{ zoneLabel(z.zoneId) }}</b>
                  <div v-if="zoneShort(z.zoneId)" class="text-muted">{{ zoneShort(z.zoneId) }}</div>
                </td>
                <td>
                  <span v-if="liveById[z.zoneId] && liveById[z.zoneId].popped" class="zc-pill">
                    popped{{ liveById[z.zoneId].players ? " · " + liveById[z.zoneId].players : "" }}
                  </span>
                  <span v-else class="text-muted">down</span>
                </td>
                <td class="text-right tabular">{{ z.customCount }}</td>
                <td class="text-right tabular">{{ z.ignoreCount }}</td>
                <td class="text-right tabular">{{ z.depopCount }}</td>
                <td class="text-right tabular">{{ z.lootTableCount }}</td>
                <td class="text-right tabular">{{ z.itemCount }}</td>
              </tr>
              <tr v-if="!filteredIndex.length">
                <td colspan="8" class="text-muted">No ultimatedata folders matched.</td>
              </tr>
              </tbody>
            </table>
          </eq-window>
        </div>

        <div v-else-if="isSystems">
          <div class="ui-toolbar">
            <button
              type="button"
              class="btn btn-sm"
              :class="systemKind === 'talents' ? 'btn-primary' : 'btn-dark'"
              @click="setSystemKind('talents')"
            >
              Talent ranks ({{ (systems.talents || []).length }})
            </button>
            <button
              type="button"
              class="btn btn-sm"
              :class="systemKind === 'unlocks' ? 'btn-primary' : 'btn-dark'"
              @click="setSystemKind('unlocks')"
            >
              Unlock states ({{ (systems.unlocks || []).length }})
            </button>
            <input
              v-model="systemSearch"
              type="search"
              class="form-control form-control-sm"
              style="width: 280px"
              placeholder="Key, class, spell, or bucket"
            >
          </div>

          <eq-window v-if="systemKind === 'talents'" title="Talent rank registry">
            <p class="zc-copy">
              {{ talentPolicy }}
            </p>
            <table class="eq-table" style="width: 100%">
              <thead>
              <tr>
                <th>Key</th>
                <th>Class</th>
                <th>Bucket suffix</th>
                <th>Legacy global</th>
                <th>Value</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="row in filteredTalents" :key="row.key">
                <td><code>{{ row.key }}</code></td>
                <td>{{ row.classFamily }}</td>
                <td class="text-muted">{{ row.bucketSuffix }}</td>
                <td class="text-muted">{{ row.legacyGlobalKey }}</td>
                <td>{{ row.valueType }}</td>
              </tr>
              </tbody>
            </table>
          </eq-window>

          <eq-window v-else title="Unlock state registry">
            <p class="zc-copy">
              {{ unlockPolicy }}
            </p>
            <table class="eq-table" style="width: 100%">
              <thead>
              <tr>
                <th>Key</th>
                <th>Spell</th>
                <th style="width: 80px">ID</th>
                <th>Bucket suffix</th>
                <th>Expected</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="row in filteredUnlocks" :key="row.key">
                <td><code>{{ row.key }}</code></td>
                <td>
                  <router-link v-if="usableId(row.spellId)" :to="spellHref(row.spellId)">
                    {{ row.spellName || row.spellId }}
                  </router-link>
                  <span v-else>{{ row.spellName || "—" }}</span>
                </td>
                <td class="tabular">{{ row.spellId }}</td>
                <td class="text-muted">{{ row.bucketSuffix }}</td>
                <td>{{ row.expectedValue }}</td>
              </tr>
              </tbody>
            </table>
          </eq-window>
        </div>

        <div v-else-if="detail">
          <eq-window title="Quick Actions" class="mb-3">
            <p class="zc-copy">
              Browse links open the JSON here. Live buttons talk to World telnet
              and queue the same zone-controller commands the GM hail menu uses.
            </p>
            <div class="zc-live-note mb-2">
              <template v-if="live && live.worldOk">
                World telnet is up.
                <template v-if="thisLive && thisLive.popped">
                  {{ zoneShort(zoneId) || zoneId }} is popped
                  <template v-if="thisLive.players">· {{ thisLive.players }} player(s)</template>.
                </template>
                <template v-else>
                  {{ zoneShort(zoneId) || zoneId }} is not popped. Queue waits until boot, or use Boot zone.
                </template>
              </template>
              <template v-else>
                {{ (live && live.worldNote) || "World telnet is down. Commands can still be queued for the next pop." }}
              </template>
            </div>
            <div class="zc-actions">
              <button
                v-for="a in browseActions"
                :key="a.id"
                type="button"
                class="btn btn-sm"
                :class="view === a.id ? 'btn-primary' : 'btn-dark'"
                @click="setView(a.id)"
              >
                {{ a.label }}
              </button>
            </div>
            <div class="zc-actions zc-actions-live">
              <button
                v-for="a in worldActions"
                :key="a.action"
                type="button"
                class="btn btn-sm btn-dark"
                :disabled="liveBusy"
                @click="runWorld(a.action)"
              >
                {{ a.label }}
              </button>
              <button
                v-for="a in liveActions"
                :key="a.command"
                type="button"
                class="btn btn-sm btn-outline-secondary"
                :disabled="liveBusy"
                @click="runLive(a.command)"
              >
                {{ a.label }}
              </button>
            </div>
            <div v-if="liveResult" class="zc-live-note">
              <div v-if="liveResult.worldNote">{{ liveResult.worldNote }}</div>
              <div v-if="(liveResult.reloaded || []).length">Reloaded: {{ liveResult.reloaded.join(", ") }}</div>
              <div v-if="(liveResult.rebooted || []).length">World: {{ liveResult.rebooted.join(", ") }}</div>
              <div v-if="(liveResult.queued || []).length">
                Queued: {{ liveResult.queued.map((q) => (q.short || q.zoneId) + " " + (q.commands || []).join(", ")).join("; ") }}
              </div>
              <div v-for="w in (liveResult.warnings || [])" :key="w" class="text-warning">{{ w }}</div>
            </div>
          </eq-window>

          <eq-window v-if="view === 'info'" title="Zone info">
            <div class="ui-field-grid">
              <div class="ui-field"><label>JSON name</label><div>{{ prettyInfo(detail.name) }}</div></div>
              <div class="ui-field"><label>PEQ name</label><div>{{ zoneLabel(detail.zoneId) }}</div></div>
              <div class="ui-field"><label>Respawn</label><div>{{ prettyInfo(detail.respawn) }}</div></div>
              <div class="ui-field"><label>Objective</label><div>{{ prettyInfo(detail.objective) }}</div></div>
              <div class="ui-field" style="grid-column: 1 / -1"><label>Tip</label><div>{{ prettyInfo(detail.tip) }}</div></div>
            </div>
            <table class="eq-table mt-3" style="width: 100%">
              <thead>
              <tr>
                <th>Type</th>
                <th v-for="stat in baseStats" :key="stat">{{ stat }}</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="kind in ['trash', 'boss', 'raid']" :key="kind">
                <td><b>{{ kind }}</b></td>
                <td v-for="stat in baseStats" :key="kind + stat" class="tabular">
                  {{ baseStat(kind, stat) }}
                </td>
              </tr>
              </tbody>
            </table>
          </eq-window>

          <eq-window v-else-if="view === 'mobs'" title="Custom mobs">
            <table class="eq-table eq-highlight-rows" style="width: 100%">
              <thead>
              <tr>
                <th>Name</th>
                <th>Type</th>
                <th>NPC ID</th>
                <th>Static</th>
                <th>Mods</th>
                <th>Loot tables</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="m in (detail.custom || [])" :key="m.name">
                <td><b>{{ m.name }}</b></td>
                <td>{{ m.type || "—" }}</td>
                <td>
                  <router-link v-if="usableId(m.mobId)" :to="'/npc/' + m.mobId">{{ m.mobId }}</router-link>
                  <span v-else>—</span>
                </td>
                <td>{{ m.static === "1" ? "yes" : "no" }}</td>
                <td class="text-muted">{{ formatMods(m.mods) }}</td>
                <td>
                  <button
                    v-for="l in (m.loot || [])"
                    :key="l.id"
                    type="button"
                    class="zc-pill ui-plain"
                    @click="setView('loot')"
                  >
                    {{ l.id }} <span class="text-muted">{{ l.chance }}</span>
                  </button>
                </td>
              </tr>
              <tr v-if="!(detail.custom || []).length">
                <td colspan="6" class="text-muted">No named custom mobs in this zone JSON.</td>
              </tr>
              </tbody>
            </table>
          </eq-window>

          <eq-window v-else-if="view === 'loot'" title="Loot tables">
            <table class="eq-table" style="width: 100%">
              <thead>
              <tr>
                <th style="width: 180px">Table</th>
                <th style="width: 70px" class="text-right">Count</th>
                <th>Items</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="t in (detail.loot || [])" :key="t.id">
                <td><code>{{ t.id }}</code></td>
                <td class="text-right tabular">{{ t.count }}</td>
                <td>{{ (t.items || []).join(", ") }}</td>
              </tr>
              <tr v-if="!(detail.loot || []).length">
                <td colspan="3" class="text-muted">Loot JSON is empty for this zone.</td>
              </tr>
              </tbody>
            </table>
          </eq-window>

          <eq-window v-else-if="view === 'items'" title="Items">
            <div class="ui-toolbar">
              <input
                v-model="itemSearch"
                type="search"
                class="form-control form-control-sm"
                style="width: 260px"
                placeholder="Item name or ID"
              >
            </div>
            <table class="eq-table eq-highlight-rows" style="width: 100%">
              <thead>
              <tr>
                <th>Name</th>
                <th style="width: 100px">Item ID</th>
                <th style="width: 100px">Spell</th>
                <th>Global key</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="it in filteredItems" :key="it.name + it.itemId">
                <td>
                  <router-link v-if="usableId(it.itemId)" :to="itemHref(it.itemId)">{{ it.name }}</router-link>
                  <span v-else>{{ it.name }}</span>
                </td>
                <td class="tabular">{{ blankDash(it.itemId) }}</td>
                <td>
                  <router-link v-if="usableId(it.spellId)" :to="spellHref(it.spellId)">{{ it.spellId }}</router-link>
                  <span v-else>{{ blankDash(it.spellId) }}</span>
                </td>
                <td class="text-muted">{{ blankDash(it.globalKey) }}</td>
              </tr>
              </tbody>
            </table>
          </eq-window>

          <eq-window v-else-if="view === 'ignore'" title="Ignored mobs">
            <p class="zc-copy">These names are skipped when the controller buffs new spawns.</p>
            <table class="eq-table" style="width: 100%">
              <tbody>
              <tr v-for="name in (detail.ignore || [])" :key="name">
                <td>{{ name }}</td>
              </tr>
              <tr v-if="!(detail.ignore || []).length">
                <td class="text-muted">Ignore list is empty.</td>
              </tr>
              </tbody>
            </table>
          </eq-window>

          <eq-window v-else-if="view === 'depop'" title="Depop mobs">
            <p class="zc-copy">These names are depopped when the controller applies spawn rules.</p>
            <table class="eq-table" style="width: 100%">
              <tbody>
              <tr v-for="name in (detail.depop || [])" :key="name">
                <td>{{ name }}</td>
              </tr>
              <tr v-if="!(detail.depop || []).length">
                <td class="text-muted">Depop list is empty.</td>
              </tr>
              </tbody>
            </table>
          </eq-window>

          <eq-window v-else-if="view === 'diag'" title="ZC files">
            <p class="zc-copy">
              Live hail <code>viewzcdiag</code> shows in-zone counters (signals, reconcile, apply).
              This page shows the JSON the controller loads from disk.
            </p>
            <table class="eq-table" style="width: 100%">
              <thead>
              <tr>
                <th>File</th>
                <th>Present</th>
                <th class="text-right">Bytes</th>
                <th></th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="f in fileRows" :key="f.kind">
                <td><code>{{ f.relPath }}</code></td>
                <td>{{ f.exists ? "yes" : "no" }}</td>
                <td class="text-right tabular">{{ f.size || 0 }}</td>
                <td>
                  <a v-if="f.exists" :href="fileHref(f.relPath)">Open in Server Files</a>
                </td>
              </tr>
              </tbody>
            </table>
          </eq-window>

          <eq-window v-else-if="view === 'commands'" title="Remote commands">
            <p class="zc-copy">
              Fallback say commands if World telnet is down. Live buttons on this page
              queue the same tokens for a popped controller.
            </p>
            <table class="eq-table" style="width: 100%">
              <thead>
              <tr>
                <th style="width: 360px">Command</th>
                <th>Info</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="c in remoteCommands" :key="c.cmd">
                <td><code>{{ c.cmd }}</code></td>
                <td>{{ c.info }}</td>
              </tr>
              </tbody>
            </table>
          </eq-window>
        </div>
      </div>
    </eq-window>
  </content-area>
</template>

<script>
import EqWindow from "../../components/eq-ui/EQWindow"
import ContentArea from "../../components/layout/ContentArea"
import {ROUTE} from "@/routes"
import {Zones} from "../../app/zones"
import {
  ZONE_CONTROLLER_BROWSE,
  ZONE_CONTROLLER_LIVE,
  ZONE_CONTROLLER_REMOTE,
  ZONE_CONTROLLER_WORLD,
  ZoneControllerApi,
} from "../../app/zone-controller"

const PLACEHOLDER = [
  "Zone name",
  "Helpful zone info.",
  "This is info to complete the zone.",
  "Amount of time it takes for trash and bosses to repop.",
]

export default {
  name: "ZoneController",
  components: {ContentArea, EqWindow},
  data() {
    return {
      ROUTE,
      loaded: false,
      error: "",
      index: null,
      detail: null,
      systems: {talents: [], unlocks: []},
      peqZones: {},
      zoneSearch: "",
      systemSearch: "",
      itemSearch: "",
      showLive: "",
      liveBusy: false,
      live: null,
      liveResult: null,
      browseActions: ZONE_CONTROLLER_BROWSE,
      liveActions: ZONE_CONTROLLER_LIVE,
      worldActions: ZONE_CONTROLLER_WORLD,
      remoteCommands: ZONE_CONTROLLER_REMOTE,
    }
  },
  computed: {
    zoneId() {
      const raw = this.$route.params.zoneId
      const n = parseInt(raw, 10)
      return n > 0 ? n : 0
    },
    isSystems() {
      return this.$route.path === ROUTE.ZONE_CONTROLLER_SYSTEMS
    },
    view() {
      const q = this.$route.query.view
      if (ZONE_CONTROLLER_BROWSE.some((a) => a.id === q)) {
        return q
      }
      return "mobs"
    },
    systemKind() {
      return this.$route.query.kind === "unlocks" ? "unlocks" : "talents"
    },
    pageTitle() {
      if (this.isSystems) {
        return "Zone Controller systems"
      }
      if (this.zoneId) {
        return "Zone Controller — " + this.zoneLabel(this.zoneId)
      }
      return "Zone Controller"
    },
    peqZone() {
      return this.peqZones[this.zoneId] || {}
    },
    scriptHref() {
      const rel = (this.index && this.index.scriptRel) || "global/zone_controller.pl"
      return "/admin/server-files?path=" + encodeURIComponent(rel)
    },
    filteredIndex() {
      const rows = (this.index && this.index.zones) || []
      const s = this.zoneSearch.trim().toLowerCase()
      if (!s) {
        return rows
      }
      return rows.filter((z) => {
        const peq = this.peqZones[z.zoneId] || {}
        return String(z.zoneId) === s
          || (peq.short_name || "").toLowerCase().includes(s)
          || (peq.long_name || "").toLowerCase().includes(s)
          || (z.name || "").toLowerCase().includes(s)
      })
    },
    filteredTalents() {
      const s = this.systemSearch.trim().toLowerCase()
      const rows = this.systems.talents || []
      if (!s) {
        return rows
      }
      return rows.filter((r) => {
        return [r.key, r.classFamily, r.bucketSuffix, r.legacyGlobalKey, r.valueType]
          .join(" ").toLowerCase().includes(s)
      })
    },
    filteredUnlocks() {
      const s = this.systemSearch.trim().toLowerCase()
      const rows = this.systems.unlocks || []
      if (!s) {
        return rows
      }
      return rows.filter((r) => {
        return [r.key, r.spellName, r.spellId, r.bucketSuffix, r.legacyGlobalKey]
          .join(" ").toLowerCase().includes(s)
      })
    },
    filteredItems() {
      const s = this.itemSearch.trim().toLowerCase()
      const rows = (this.detail && this.detail.items) || []
      if (!s) {
        return rows
      }
      return rows.filter((it) => {
        return (it.name || "").toLowerCase().includes(s) || String(it.itemId) === s
      })
    },
    liveById() {
      const out = {}
      ;((this.live && this.live.zones) || []).forEach((z) => {
        if (z && z.zoneId) {
          out[z.zoneId] = z
        }
      })
      return out
    },
    thisLive() {
      return this.liveById[this.zoneId] || null
    },
    fileRows() {
      const files = (this.detail && this.detail.files) || {}
      return ["mob", "loot", "item"].map((k) => files[k] || {kind: k, relPath: "", exists: false, size: 0})
    },
    baseStats() {
      return ["level", "max_hp", "min_hit", "max_hit", "atk", "ac", "attack_speed", "cash", "loot"]
    },
    talentPolicy() {
      const meta = this.systems.talentMeta || {}
      const storage = meta.policy && meta.policy.storage
      return storage
        ? "Storage: " + storage + ". Bucket suffix format _TalentRank_<legacy_global_key>."
        : "Bucket-backed talent ranks generated from quests and plugins."
    },
    unlockPolicy() {
      const meta = this.systems.unlockMeta || {}
      const storage = meta.policy && meta.policy.storage
      return storage
        ? "Storage: " + storage + ". Spell turn-in / unlock flags."
        : "Bucket-backed unlock flags from the spell-state inventory."
    },
  },
  watch: {
    $route() {
      this.load()
    },
  },
  async mounted() {
    await this.load()
  },
  methods: {
    async load() {
      this.loaded = false
      this.error = ""
      this.showLive = ""
      try {
        const zones = await Zones.getZones()
        const byId = {}
        ;(zones || []).forEach((z) => {
          if (z && z.zoneidnumber && !byId[z.zoneidnumber]) {
            byId[z.zoneidnumber] = z
          }
        })
        this.peqZones = byId

        this.index = await ZoneControllerApi.list()
        if (this.index && this.index.error && !this.index.ok && !this.zoneId && !this.isSystems) {
          this.error = this.index.error
        }
        const liveIds = this.zoneId
          ? [this.zoneId]
          : ((this.index && this.index.zones) || []).map((z) => z.zoneId)
        try {
          this.live = await ZoneControllerApi.live(liveIds)
        } catch (e) {
          this.live = {ok: false, worldOk: false, zones: [], worldNote: "Could not read World zone list."}
        }

        if (this.isSystems) {
          this.systems = await ZoneControllerApi.systems()
          this.detail = null
        } else if (this.zoneId) {
          this.detail = await ZoneControllerApi.zone(this.zoneId)
        } else {
          this.detail = null
        }
      } catch (e) {
        this.error = (e.response && e.response.data && e.response.data.error) || String(e)
      }
      this.loaded = true
    },
    zoneLabel(id) {
      const z = this.peqZones[id]
      if (z && z.long_name) {
        return z.long_name
      }
      return "Zone " + id
    },
    zoneShort(id) {
      const z = this.peqZones[id]
      return z && z.short_name ? z.short_name : ""
    },
    openZone(id) {
      this.$router.push("/zones/controller/" + id).catch(() => {
      })
    },
    setView(id) {
      this.showLive = ""
      this.$router.replace({path: this.$route.path, query: {view: id}}).catch(() => {
      })
    },
    setSystemKind(kind) {
      this.$router.replace({path: this.$route.path, query: {kind}}).catch(() => {
      })
    },
    prettyInfo(v) {
      if (!v || PLACEHOLDER.indexOf(v) !== -1) {
        return "—"
      }
      return v
    },
    baseStat(kind, stat) {
      const row = this.detail && this.detail.basedata && this.detail.basedata[kind]
      if (!row || row[stat] == null || row[stat] === "" || row[stat] === "-1") {
        return "—"
      }
      return row[stat]
    },
    formatMods(mods) {
      const keys = Object.keys(mods || {})
      if (!keys.length) {
        return "—"
      }
      return keys.map((k) => k.replace(/_/g, " ") + " " + mods[k]).join(", ")
    },
    usableId(v) {
      const n = parseInt(v, 10)
      return n > 0
    },
    blankDash(v) {
      if (v == null || v === "" || v === "-1") {
        return "—"
      }
      return v
    },
    itemHref(id) {
      return "/item/" + id
    },
    spellHref(id) {
      return "/spell/" + id
    },
    fileHref(rel) {
      return "/admin/server-files?path=" + encodeURIComponent(rel || "")
    },
    liveLabel(command) {
      const row = ZONE_CONTROLLER_LIVE.find((a) => a.command === command)
      return row ? row.label : command
    },
    async runLive(command) {
      if (!this.zoneId) {
        return
      }
      this.liveBusy = true
      this.liveResult = null
      try {
        this.liveResult = await ZoneControllerApi.apply({
          zoneIds: [this.zoneId],
          queueCommands: [command],
          rebootEmpty: true,
        })
        this.live = await ZoneControllerApi.live([this.zoneId])
      } catch (e) {
        this.error = (e.response && e.response.data && e.response.data.error) || String(e)
      }
      this.liveBusy = false
    },
    async runWorld(action) {
      if (!this.zoneId) {
        return
      }
      this.liveBusy = true
      this.liveResult = null
      try {
        this.liveResult = await ZoneControllerApi.apply({
          zoneIds: [this.zoneId],
          reloadQuests: action === "reload",
          rebootZones: action === "reboot",
          rebootEmpty: action === "reboot",
          bootIfDown: action === "boot" || action === "reboot",
          queueCommands: action === "reboot" ? ["refreshzonedata"] : [],
        })
        this.live = await ZoneControllerApi.live([this.zoneId])
      } catch (e) {
        this.error = (e.response && e.response.data && e.response.data.error) || String(e)
      }
      this.liveBusy = false
    },
  },
}
</script>

<style scoped>
.zc-dot {
  margin: 0 6px;
  color: var(--text-subtle);
}

.zc-copy {
  color: var(--text-muted, #b7c0cc);
  margin-bottom: 12px;
}

.zc-system-links,
.zc-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.zc-actions-live {
  margin-top: 10px;
}

.zc-live-note {
  margin-top: 12px;
  padding: 10px 12px;
  border: 1px solid var(--border, #3a4350);
  background: var(--surface-2, #12161d);
  border-radius: 6px;
}

.zc-row {
  cursor: pointer;
}

.zc-pill {
  display: inline-block;
  margin: 0 4px 4px 0;
  padding: 1px 7px;
  border: 1px solid var(--border, #3a4350);
  border-radius: 999px;
  background: var(--surface-2, #12161d);
  color: inherit;
  font-size: 12px;
}

.ml-auto {
  margin-left: auto;
}
</style>

<template>
  <div class="char-inv">
    <eq-window :title="windowTitle" class="p-3">
      <div class="char-inv-top">
        <div class="char-inv-toolbar">
          <input
            class="form-control form-control-sm"
            v-model="charSearch"
            @keyup.enter="searchCharacters"
            placeholder="Character name or id"
          >
          <button class="btn btn-sm btn-dark" @click="searchCharacters">Find</button>
          <router-link class="btn btn-sm btn-dark" to="/editors/players">Players</router-link>
          <router-link class="btn btn-sm btn-dark" to="/editors/keys">Keyring rows</router-link>
          <router-link class="btn btn-sm btn-dark" to="/editors/qglobals">QGlobals</router-link>
          <router-link class="btn btn-sm btn-dark" to="/editors/databuckets">Buckets</router-link>
          <router-link class="btn btn-sm btn-dark" to="/editors/ldon">LDoN</router-link>
          <span class="char-inv-status" v-if="status">{{ status }}</span>
        </div>
        <div class="char-inv-hero" v-if="character">
          <div class="char-inv-hero-text">
            <div class="char-inv-name">{{ displayName(character) }}</div>
            <div class="char-inv-meta">
              L{{ character.level }} {{ className(character) }}
              · #{{ character.id }}
              <template v-if="character.account_id"> · Acc {{ character.account_id }}</template>
            </div>
            <div class="char-inv-meta">
              {{ ownedKeys.length }} keys
              · {{ flags.length }} flags
            </div>
          </div>
        </div>
      </div>

      <div class="char-inv-hits" v-if="charHits.length > 1">
        <div
          v-for="hit in charHits"
          :key="hit.id"
          class="char-inv-hit"
          role="button"
          tabindex="0"
          :class="{ 'is-on': character && character.id === hit.id }"
          @click="selectCharacter(hit.id)"
          @keydown.enter.prevent="selectCharacter(hit.id)"
        >
          <div class="char-inv-hit-text">
            <b>{{ displayName(hit) }}</b>
            <span>L{{ hit.level }} {{ className(hit) }} · #{{ hit.id }}</span>
          </div>
        </div>
      </div>

      <div v-if="character">
        <div class="char-inv-tabs">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            type="button"
            class="char-inv-tab ui-plain"
            :class="{ 'is-on': activeTab === tab.id }"
            @click="setTab(tab.id)"
          >{{ tab.label }}</button>
        </div>

        <div v-if="loading" class="char-inv-empty">Loading…</div>

        <div v-else-if="activeTab === 'keys'" class="access-split">
          <div class="inv-panel">
            <div class="inv-panel-head">Owned ({{ ownedKeys.length }})</div>
            <table class="eq-table eq-highlight-rows" style="width: 100%">
              <thead><tr><th></th><th>Item</th><th class="text-right">ID</th><th></th></tr></thead>
              <tbody>
              <tr v-for="row in ownedKeys" :key="'own-' + row.item_id">
                <td><span v-if="row.icon" :class="'item-' + row.icon + '-sm'"></span></td>
                <td><b>{{ row.name }}</b></td>
                <td class="text-right tabular">{{ row.item_id }}</td>
                <td class="text-right">
                  <button class="btn btn-sm btn-outline-danger" :disabled="busy" @click="revokeKey(row)">Remove</button>
                </td>
              </tr>
              <tr v-if="!ownedKeys.length"><td colspan="4" class="char-inv-empty">No keyring rows.</td></tr>
              </tbody>
            </table>
            <div class="ui-toolbar mt-2">
              <input v-model="itemSearch" class="form-control form-control-sm" style="width: 220px" placeholder="Item name or id" @keyup.enter="searchItems">
              <button class="btn btn-sm btn-dark" @click="searchItems">Search</button>
            </div>
            <div class="access-hits" v-if="itemHits.length">
              <button
                v-for="item in itemHits"
                :key="'hit-' + item.id"
                type="button"
                class="zc-pill ui-plain"
                :disabled="busy || hasKey(item.id)"
                @click="grantKey(item.id)"
              >
                {{ item.name }} #{{ item.id }}
              </button>
            </div>
          </div>

          <div class="inv-panel">
            <div class="inv-panel-head">Known ({{ filteredKnown.length }})</div>
            <input v-model="knownSearch" type="search" class="form-control form-control-sm mb-2" placeholder="Filter known keys">
            <table class="eq-table eq-highlight-rows" style="width: 100%">
              <thead><tr><th></th><th>Item</th><th>Source</th><th></th></tr></thead>
              <tbody>
              <tr v-for="row in filteredKnown" :key="'know-' + row.item_id">
                <td><span v-if="row.icon" :class="'item-' + row.icon + '-sm'"></span></td>
                <td>
                  <b>{{ row.name }}</b>
                  <div class="text-muted">#{{ row.item_id }}</div>
                </td>
                <td class="text-muted">{{ row.source }}</td>
                <td class="text-right">
                  <span v-if="hasKey(row.item_id)" class="text-muted">Owned</span>
                  <button v-else class="btn btn-sm btn-dark" :disabled="busy" @click="grantKey(row.item_id)">Grant</button>
                </td>
              </tr>
              <tr v-if="!filteredKnown.length"><td colspan="4" class="char-inv-empty">No door keys or keyring items found.</td></tr>
              </tbody>
            </table>
          </div>
        </div>

        <div v-else-if="activeTab === 'flags'">
          <div class="ui-toolbar mb-2">
            <button
              v-for="src in flagSources"
              :key="src.id"
              type="button"
              class="btn btn-sm"
              :class="flagSource === src.id ? 'btn-primary' : 'btn-dark'"
              @click="flagSource = src.id"
            >{{ src.label }}</button>
          </div>
          <div class="access-split">
            <div class="inv-panel">
              <div class="inv-panel-head">{{ flagSourceLabel }} ({{ visibleFlags.length }})</div>
              <table class="eq-table eq-highlight-rows" style="width: 100%">
                <thead><tr><th>Key</th><th>Value</th><th>Scope</th><th></th></tr></thead>
                <tbody>
                <tr v-for="row in visibleFlags" :key="row.uid">
                  <td><b>{{ row.key }}</b></td>
                  <td class="tabular">{{ row.value }}</td>
                  <td class="text-muted">{{ row.scope }}</td>
                  <td class="text-right">
                    <button class="btn btn-sm btn-outline-danger" :disabled="busy" @click="deleteFlag(row)">Remove</button>
                  </td>
                </tr>
                <tr v-if="!visibleFlags.length"><td colspan="4" class="char-inv-empty">None for this source.</td></tr>
                </tbody>
              </table>
            </div>
            <div class="inv-panel">
              <div class="inv-panel-head">Add {{ flagSourceLabel.toLowerCase() }}</div>
              <div class="ui-field" v-if="flagSource !== 'zone'">
                <label>Name / key</label>
                <input v-model="flagDraft.key" class="form-control form-control-sm">
              </div>
              <div class="ui-field" v-if="flagSource === 'zone'">
                <label>Zone ID</label>
                <input v-model.number="flagDraft.zoneId" type="number" class="form-control form-control-sm">
                <div class="zc-hint" v-if="flagDraft.zoneId">{{ zoneName(flagDraft.zoneId) }}</div>
              </div>
              <div class="ui-field" v-if="flagSource !== 'zone'">
                <label>Value</label>
                <input v-model="flagDraft.value" class="form-control form-control-sm">
              </div>
              <div class="ui-field-grid" v-if="flagSource === 'quest'">
                <div class="ui-field"><label>NPC ID</label><input v-model.number="flagDraft.npcId" type="number" class="form-control form-control-sm"></div>
                <div class="ui-field"><label>Zone ID</label><input v-model.number="flagDraft.zoneId" type="number" class="form-control form-control-sm"></div>
              </div>
              <div class="ui-toolbar">
                <button class="btn btn-sm btn-primary" :disabled="busy" @click="addFlag">Add</button>
              </div>
            </div>
          </div>
        </div>

      </div>

      <div v-else class="char-inv-empty mt-3">
        Find a character to edit keys and flags.
      </div>
    </eq-window>
  </div>
</template>

<script>
import axios from "axios"
import EqWindow from "../../components/eq-ui/EQWindow"
import {SpireApi} from "../../app/api/spire-api"
import {SpireQueryBuilder} from "../../app/api/spire-query-builder"
import {CharacterDatumApi} from "../../app/api/api/character-datum-api"
import {ItemApi} from "../../app/api/api/item-api"
import {KeyringApi} from "../../app/api/api/keyring-api"
import {DoorApi} from "../../app/api/api/door-api"
import {QuestGlobalApi} from "../../app/api/api/quest-global-api"
import {DataBucketApi} from "../../app/api/api/data-bucket-api"
import {ZoneFlagApi} from "../../app/api/api/zone-flag-api"
import {AccountFlagApi} from "../../app/api/api/account-flag-api"
import {DB_PLAYER_CLASSES_ALL} from "../../app/constants/eq-classes-constants"
import {Zones} from "../../app/zones"
import {ACCESS_TABS, CHAR_ACCESS_FIELDS, FLAG_SOURCES} from "../../app/character-access"

export default {
  name: "CharacterAccess",
  components: {EqWindow},
  data() {
    return {
      charSearch: "",
      charHits: [],
      character: null,
      loading: false,
      busy: false,
      status: "",
      activeTab: "keys",
      tabs: ACCESS_TABS,
      flagSources: FLAG_SOURCES,
      flagSource: "quest",
      ownedKeys: [],
      knownKeys: [],
      knownSearch: "",
      itemSearch: "",
      itemHits: [],
      flags: [],
      flagDraft: {key: "", value: "1", npcId: 0, zoneId: 0},
    }
  },
  computed: {
    windowTitle() {
      return this.character ? (this.character.name + " — Keys / Flags") : "Keys / Flags"
    },
    filteredKnown() {
      const q = (this.knownSearch || "").toLowerCase()
      return this.knownKeys.filter((row) => {
        if (!q) {
          return true
        }
        return String(row.item_id).indexOf(q) !== -1 || (row.name || "").toLowerCase().indexOf(q) !== -1
      })
    },
    visibleFlags() {
      return this.flags.filter((row) => row.source === this.flagSource)
    },
    flagSourceLabel() {
      const src = FLAG_SOURCES.find((s) => s.id === this.flagSource)
      return src ? src.label : "Flags"
    },
  },
  watch: {
    "$route.query.c"() {
      this.bootFromRoute()
    },
    "$route.query.tab"() {
      this.bootTab()
    },
  },
  mounted() {
    if (String(this.$route.query.tab || "") === "ldon") {
      const query = {}
      if (this.$route.query.c) {
        query.c = this.$route.query.c
      }
      this.$router.replace({path: "/editors/ldon", query}).catch(() => {})
      return
    }
    Zones.getZones().catch(() => {})
    this.bootTab()
    this.bootFromRoute()
  },
  methods: {
    client() {
      return axios.create(SpireApi.getAxiosConfig())
    },
    charApi() {
      return new CharacterDatumApi(...SpireApi.cfg())
    },
    itemApi() {
      return new ItemApi(...SpireApi.cfg())
    },
    className(row) {
      const rec = row && DB_PLAYER_CLASSES_ALL[row.class != null ? row.class : row._class]
      return rec ? rec.class : ""
    },
    displayName(row) {
      if (!row) {
        return ""
      }
      return [row.name, row.last_name].filter(Boolean).join(" ")
    },
    zoneName(id) {
      const zone = Zones.zonesById[id]
      return zone ? (zone.long_name || zone.short_name) : ""
    },
    errorText(err) {
      if (err && err.response && err.response.data && err.response.data.error) {
        return err.response.data.error
      }
      return err && err.message ? err.message : "Request failed"
    },
    hasKey(itemId) {
      return this.ownedKeys.some((row) => Number(row.item_id) === Number(itemId))
    },
    setTab(id) {
      this.activeTab = id
      const query = Object.assign({}, this.$route.query, {tab: id})
      this.$router.replace({path: this.$route.path, query}).catch(() => {})
    },
    bootTab() {
      const tab = String(this.$route.query.tab || "")
      if (ACCESS_TABS.some((t) => t.id === tab)) {
        this.activeTab = tab
      }
    },
    bootFromRoute() {
      const id = parseInt(this.$route.query.c, 10) || 0
      if (id) {
        this.selectCharacter(id)
      }
    },
    pushQuery(id) {
      const query = Object.assign({}, this.$route.query)
      if (id) {
        query.c = String(id)
      }
      this.$router.replace({path: this.$route.path, query}).catch(() => {})
    },
    async searchCharacters() {
      this.status = ""
      const q = (this.charSearch || "").trim()
      if (!q) {
        this.charHits = []
        return
      }
      try {
        const builder = new SpireQueryBuilder()
        builder.limit(20)
        builder.orderBy(["name"])
        builder.select(CHAR_ACCESS_FIELDS)
        if (/^\d+$/.test(q)) {
          builder.where("id", "=", q)
        } else {
          builder.where("name", "like", q)
        }
        const r = await this.charApi().listCharacterData(builder.get())
        this.charHits = (r && r.data) ? r.data : []
        if (this.charHits.length === 1) {
          this.selectCharacter(this.charHits[0].id)
        }
      } catch (err) {
        this.status = this.errorText(err)
      }
    },
    async selectCharacter(id) {
      this.status = ""
      this.loading = true
      this.pushQuery(id)
      try {
        const cr = await this.charApi().getCharacterDatum({
          id,
          select: CHAR_ACCESS_FIELDS.join("."),
        })
        this.character = cr && cr.data ? cr.data : null
        if (this.character && !this.charHits.find((h) => h.id === this.character.id)) {
          this.charHits = [this.character].concat(this.charHits)
        }
        await Promise.all([this.loadKeys(), this.loadFlags()])
      } catch (err) {
        this.status = this.errorText(err)
        this.character = null
      } finally {
        this.loading = false
      }
    },
    async hydrateItems(ids) {
      const uniq = Array.from(new Set(ids.map((n) => Number(n)).filter((n) => n > 0)))
      const map = {}
      if (!uniq.length) {
        return map
      }
      const ir = await this.itemApi().getItemsBulk({body: {ids: uniq}})
      ;((ir && ir.data) || []).forEach((item) => {
        map[item.id] = item
      })
      return map
    },
    async loadKeys() {
      if (!this.character) {
        this.ownedKeys = []
        return
      }
      const keyApi = new KeyringApi(...SpireApi.cfg())
      const builder = new SpireQueryBuilder()
      builder.where("char_id", "=", this.character.id)
      builder.limit(1000)
      builder.orderBy(["item_id"])
      const r = await keyApi.listKeyrings(builder.get())
      const rows = (r && r.data) ? r.data : []
      const doorApi = new DoorApi(...SpireApi.cfg())
      const doorsQ = new SpireQueryBuilder()
      doorsQ.where("keyitem", ">", "0")
      doorsQ.select(["id", "zone", "name", "keyitem"])
      doorsQ.limit(4000)
      let doors = []
      try {
        const dr = await doorApi.listDoors(doorsQ.get())
        doors = (dr && dr.data) ? dr.data : []
      } catch (err) {
        doors = []
      }
      const doorCount = {}
      doors.forEach((door) => {
        const id = Number(door.keyitem)
        if (id > 0) {
          doorCount[id] = (doorCount[id] || 0) + 1
        }
      })
      const ids = rows.map((row) => row.item_id).concat(Object.keys(doorCount))
      const items = await this.hydrateItems(ids)
      this.ownedKeys = rows.map((row) => {
        const item = items[row.item_id] || {}
        return {
          id: row.id,
          item_id: row.item_id,
          name: item.name || ("Item " + row.item_id),
          icon: item.icon || 0,
        }
      })
      const knownIds = Object.keys(doorCount).map((n) => Number(n))
      rows.forEach((row) => {
        if (knownIds.indexOf(Number(row.item_id)) === -1) {
          knownIds.push(Number(row.item_id))
        }
      })
      this.knownKeys = knownIds.sort((a, b) => a - b).map((id) => {
        const item = items[id] || {}
        const doorsFor = doorCount[id] || 0
        return {
          item_id: id,
          name: item.name || ("Item " + id),
          icon: item.icon || 0,
          source: doorsFor ? (doorsFor + " door" + (doorsFor === 1 ? "" : "s")) : "keyring",
        }
      })
    },
    async loadFlags() {
      if (!this.character) {
        this.flags = []
        return
      }
      const id = this.character.id
      const acc = this.character.account_id
      const out = []
      const qg = new QuestGlobalApi(...SpireApi.cfg())
      const qb = new SpireQueryBuilder()
      qb.where("charid", "=", id)
      qb.limit(2000)
      try {
        const r = await qg.listQuestGlobals(qb.get())
        ;((r && r.data) || []).forEach((row) => {
          out.push({
            uid: "q-" + row.charid + "-" + row.npcid + "-" + row.zoneid + "-" + row.name,
            source: "quest",
            key: row.name,
            value: row.value,
            scope: "npc " + row.npcid + " · zone " + row.zoneid,
            raw: row,
          })
        })
      } catch (err) {}
      const db = new DataBucketApi(...SpireApi.cfg())
      const bb = new SpireQueryBuilder()
      bb.where("character_id", "=", id)
      bb.limit(2000)
      try {
        const r = await db.listDataBuckets(bb.get())
        ;((r && r.data) || []).forEach((row) => {
          out.push({
            uid: "b-" + row.id,
            source: "bucket",
            key: row.key,
            value: row.value,
            scope: row.zone_id ? ("zone " + row.zone_id) : "character",
            raw: row,
          })
        })
      } catch (err) {}
      const zf = new ZoneFlagApi(...SpireApi.cfg())
      const zb = new SpireQueryBuilder()
      zb.where("charID", "=", id)
      zb.limit(2000)
      try {
        const r = await zf.listZoneFlags(zb.get())
        ;((r && r.data) || []).forEach((row) => {
          const zid = row.zone_id
          out.push({
            uid: "z-" + row.char_id + "-" + zid,
            source: "zone",
            key: this.zoneName(zid) || ("Zone " + zid),
            value: "1",
            scope: "zone " + zid,
            raw: row,
          })
        })
      } catch (err) {}
      if (acc) {
        const af = new AccountFlagApi(...SpireApi.cfg())
        const ab = new SpireQueryBuilder()
        ab.where("p_accid", "=", acc)
        ab.limit(500)
        try {
          const r = await af.listAccountFlags(ab.get())
          ;((r && r.data) || []).forEach((row) => {
            out.push({
              uid: "a-" + row.p_accid + "-" + row.p_flag,
              source: "account",
              key: row.p_flag,
              value: row.p_value,
              scope: "account " + row.p_accid,
              raw: row,
            })
          })
        } catch (err) {}
      }
      this.flags = out
    },
    async searchItems() {
      const q = (this.itemSearch || "").trim()
      if (!q) {
        this.itemHits = []
        return
      }
      const builder = new SpireQueryBuilder()
      builder.limit(20)
      builder.orderBy(["id"])
      if (/^\d+$/.test(q)) {
        builder.where("id", "=", q)
      } else {
        builder.where("name", "like", q)
      }
      const r = await this.itemApi().listItems(builder.get())
      this.itemHits = (r && r.data) ? r.data : []
    },
    async grantKey(itemId) {
      if (!this.character || this.hasKey(itemId)) {
        return
      }
      this.busy = true
      this.status = ""
      try {
        const api = new KeyringApi(...SpireApi.cfg())
        await api.createKeyring({keyring: {char_id: this.character.id, item_id: Number(itemId)}})
        await this.loadKeys()
      } catch (err) {
        this.status = this.errorText(err)
      }
      this.busy = false
    },
    async revokeKey(row) {
      this.busy = true
      this.status = ""
      try {
        const api = new KeyringApi(...SpireApi.cfg())
        await api.deleteKeyring({id: row.id})
        await this.loadKeys()
      } catch (err) {
        this.status = this.errorText(err)
      }
      this.busy = false
    },
    async addFlag() {
      if (!this.character) {
        return
      }
      this.busy = true
      this.status = ""
      try {
        if (this.flagSource === "quest") {
          const api = new QuestGlobalApi(...SpireApi.cfg())
          await api.createQuestGlobal({
            questGlobal: {
              charid: this.character.id,
              npcid: Number(this.flagDraft.npcId || 0),
              zoneid: Number(this.flagDraft.zoneId || 0),
              name: this.flagDraft.key,
              value: this.flagDraft.value || "1",
            },
          })
        } else if (this.flagSource === "bucket") {
          const api = new DataBucketApi(...SpireApi.cfg())
          await api.createDataBucket({
            dataBucket: {
              character_id: this.character.id,
              key: this.flagDraft.key,
              value: this.flagDraft.value || "1",
              npc_id: 0,
              bot_id: 0,
            },
          })
        } else if (this.flagSource === "zone") {
          const api = new ZoneFlagApi(...SpireApi.cfg())
          await api.createZoneFlag({
            zoneFlag: {char_id: this.character.id, zone_id: Number(this.flagDraft.zoneId || 0)},
          })
        } else if (this.flagSource === "account" && this.character.account_id) {
          const api = new AccountFlagApi(...SpireApi.cfg())
          await api.createAccountFlag({
            accountFlag: {
              p_accid: this.character.account_id,
              p_flag: this.flagDraft.key,
              p_value: this.flagDraft.value || "1",
            },
          })
        }
        this.flagDraft = {key: "", value: "1", npcId: 0, zoneId: 0}
        await this.loadFlags()
      } catch (err) {
        this.status = this.errorText(err)
      }
      this.busy = false
    },
    async deleteFlag(row) {
      this.busy = true
      this.status = ""
      try {
        const http = this.client()
        if (row.source === "quest") {
          await http.delete("/quest_global/" + row.raw.charid, {
            params: {npcid: row.raw.npcid, zoneid: row.raw.zoneid, name: row.raw.name},
          })
        } else if (row.source === "bucket") {
          const api = new DataBucketApi(...SpireApi.cfg())
          await api.deleteDataBucket({id: row.raw.id})
        } else if (row.source === "zone") {
          await http.delete("/zone_flag/" + row.raw.char_id, {params: {zoneID: row.raw.zone_id}})
        } else if (row.source === "account") {
          await http.delete("/account_flag/" + row.raw.p_accid, {params: {p_flag: row.raw.p_flag}})
        }
        await this.loadFlags()
      } catch (err) {
        this.status = this.errorText(err)
      }
      this.busy = false
    },
  },
}
</script>

<style scoped>
.char-inv-top {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 12px;
}
.char-inv-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}
.char-inv-hero-text { text-align: right; }
.char-inv-name { font-size: 16px; font-weight: 650; }
.char-inv-meta, .char-inv-empty, .char-inv-status { font-size: 12px; color: var(--text-muted, #8b93a0); }
.char-inv-hits { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 12px; }
.char-inv-hit {
  border: 1px solid var(--border, #444);
  border-radius: 6px;
  padding: 6px 10px;
  cursor: pointer;
}
.char-inv-hit.is-on { outline: 1px solid #c9a227; }
.char-inv-hit-text { display: flex; flex-direction: column; font-size: 12px; }
.char-inv-tabs { display: flex; flex-wrap: wrap; gap: 4px; margin-bottom: 12px; }
.char-inv-tab {
  border: 1px solid var(--border, #444);
  background: var(--surface-2, rgba(255,255,255,0.04));
  color: inherit;
  border-radius: 6px;
  padding: 4px 10px;
  cursor: pointer;
  font-size: 12px;
}
.char-inv-tab.is-on { background: #2b3340; }
.access-split { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; align-items: start; }
.inv-panel {
  border: 1px solid var(--border-strong, #37404d);
  border-radius: 8px;
  background: var(--surface-3, #252b35);
  padding: 10px 12px 12px;
}
.inv-panel-head {
  font-size: 11px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--text-muted, #8b93a0);
  margin-bottom: 8px;
}
.access-hits { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 8px; }
.ui-field { margin-bottom: 8px; }
.ui-field label { display: block; font-size: 11px; color: var(--text-muted, #8b93a0); }
.ui-field-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.ui-toolbar { display: flex; flex-wrap: wrap; gap: 6px; align-items: center; }
.tabular { font-variant-numeric: tabular-nums; }
@media (max-width: 900px) {
  .access-split { grid-template-columns: 1fr; }
}
</style>

<template>
  <div class="char-inv">
    <eq-window :title="windowTitle" class="p-3">
      <div class="char-inv-top">
        <div class="char-inv-toolbar">
          <input
            class="form-control form-control-sm"
            v-model="search"
            @keyup.enter="runSearch"
            :placeholder="searchPlaceholder"
          >
          <button class="btn btn-sm btn-dark" @click="runSearch">Find</button>
          <router-link class="btn btn-sm btn-dark" to="/editors/access">Keys / Flags</router-link>
          <router-link class="btn btn-sm btn-dark" to="/editors/adventures">Raw templates</router-link>
          <span class="char-inv-status" v-if="status">{{ status }}</span>
        </div>
        <div class="char-inv-hero" v-if="character">
          <div class="char-inv-hero-text">
            <div class="char-inv-name">{{ displayName(character) }}</div>
            <div class="char-inv-meta">
              L{{ character.level }} {{ className(character) }}
              · #{{ character.id }}
            </div>
            <div class="char-inv-meta">{{ ldonAvailable }} available pts</div>
          </div>
        </div>
      </div>

      <div class="ui-toolbar mb-2">
        <button
          v-for="cat in categories"
          :key="cat.id"
          type="button"
          class="btn btn-sm"
          :class="category === cat.id ? 'btn-primary' : 'btn-dark'"
          @click="setCategory(cat.id)"
        >{{ cat.label }}</button>
      </div>
      <div class="ui-toolbar mb-3">
        <button
          type="button"
          class="btn btn-sm"
          :class="themeKey === 'all' ? 'btn-primary' : 'btn-dark'"
          @click="setTheme('all')"
        >All themes</button>
        <button
          v-for="theme in themes"
          :key="theme.key"
          type="button"
          class="btn btn-sm"
          :class="themeKey === theme.key ? 'btn-primary' : 'btn-dark'"
          @click="setTheme(theme.key)"
        >{{ theme.label }}</button>
      </div>

      <div v-if="loading" class="char-inv-empty">Loading…</div>

      <div v-else-if="category === 'characters'" class="char-inv-hits" v-show="charHits.length">
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

      <div v-else-if="category === 'zones'">
        <div v-if="!visibleZones.length" class="char-inv-empty">No LDoN adventure zones match that search.</div>
        <div v-for="group in zoneGroups" :key="group.key" class="inv-panel mb-3">
          <div class="inv-panel-head">{{ group.label }} ({{ group.zones.length }})</div>
          <div class="char-inv-hits">
            <div
              v-for="zone in group.zones"
              :key="zone.short_name"
              class="char-inv-hit"
              role="button"
              tabindex="0"
              :class="{ 'is-on': selectedZone && selectedZone.short_name === zone.short_name }"
              @click="selectZone(zone)"
              @keydown.enter.prevent="selectZone(zone)"
            >
              <div class="char-inv-hit-text">
                <b>{{ zone.long_name }}</b>
                <span>{{ zone.short_name }} · #{{ zone.zoneidnumber }}</span>
                <span>{{ adventureCount(zone.short_name) }} adventures</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else-if="category === 'adventures'">
        <div class="inv-panel">
          <div class="inv-panel-head">Adventures ({{ visibleAdventures.length }})</div>
          <table class="eq-table eq-highlight-rows" style="width: 100%">
            <thead>
            <tr>
              <th>Theme</th>
              <th>Zone</th>
              <th>Type</th>
              <th>Level</th>
              <th>Points</th>
              <th></th>
            </tr>
            </thead>
            <tbody>
            <tr v-for="row in visibleAdventures" :key="row.id">
              <td>{{ themeLabel(row) }}</td>
              <td>
                <b>{{ zoneLong(row.zone) }}</b>
                <div class="text-muted">{{ row.zone }}</div>
              </td>
              <td>{{ adventureType(row.type) }}<span v-if="row.is_hard"> · hard</span><span v-if="row.is_raid"> · raid</span></td>
              <td class="tabular">{{ row.min_level }}–{{ row.max_level }}</td>
              <td class="tabular">{{ row.win_points }} / {{ row.lose_points }}</td>
              <td class="text-right">
                <button class="btn btn-sm btn-dark" @click="openAdventureZone(row)">Zone</button>
              </td>
            </tr>
            <tr v-if="!visibleAdventures.length">
              <td colspan="6" class="char-inv-empty">No LDoN adventures match that search.</td>
            </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div v-if="character" class="mt-3">
        <div class="inv-panel mb-3">
          <div class="inv-panel-head">Available points</div>
          <input v-model.number="ldonDraft.available" type="number" class="form-control form-control-sm" style="max-width: 180px">
        </div>
        <div class="access-split">
          <div v-for="theme in themes" :key="'pts-' + theme.key" class="inv-panel">
            <div class="inv-panel-head">{{ theme.label }}</div>
            <div class="ui-field">
              <label>Points</label>
              <input v-model.number="ldonDraft.points[theme.key]" type="number" class="form-control form-control-sm">
            </div>
            <div class="ui-field-grid">
              <div class="ui-field">
                <label>Wins</label>
                <input v-model.number="ldonDraft.wins[theme.key]" type="number" class="form-control form-control-sm">
              </div>
              <div class="ui-field">
                <label>Losses</label>
                <input v-model.number="ldonDraft.losses[theme.key]" type="number" class="form-control form-control-sm">
              </div>
            </div>
          </div>
        </div>
        <div class="ui-toolbar mt-3">
          <button class="btn btn-sm btn-primary" :disabled="busy" @click="saveLdon">Save LDoN</button>
        </div>
      </div>

      <div v-if="selectedZone && category !== 'adventures'" class="inv-panel mt-3">
        <div class="inv-panel-head">{{ selectedZone.long_name }}</div>
        <div class="char-inv-meta mb-2">{{ selectedZone.short_name }} · #{{ selectedZone.zoneidnumber }} · {{ themeLabel({zone: selectedZone.short_name, theme: 0}) }}</div>
        <table class="eq-table eq-highlight-rows" style="width: 100%">
          <thead><tr><th>Type</th><th>Level</th><th>Points</th><th>Text</th></tr></thead>
          <tbody>
          <tr v-for="row in zoneAdventures" :key="'zadv-' + row.id">
            <td>{{ adventureType(row.type) }}<span v-if="row.is_hard"> · hard</span></td>
            <td class="tabular">{{ row.min_level }}–{{ row.max_level }}</td>
            <td class="tabular">{{ row.win_points }} / {{ row.lose_points }}</td>
            <td>{{ row.text }}</td>
          </tr>
          <tr v-if="!zoneAdventures.length">
            <td colspan="4" class="char-inv-empty">No adventure templates use this instance.</td>
          </tr>
          </tbody>
        </table>
        <div class="ui-toolbar mt-2">
          <router-link class="btn btn-sm btn-dark" :to="'/zone/' + selectedZone.short_name + '/atlas'">Atlas</router-link>
        </div>
      </div>

      <div v-if="!character && category === 'characters'" class="char-inv-empty mt-3">
        Find a character to edit LDoN points and theme wins.
      </div>
    </eq-window>
  </div>
</template>

<script>
import EqWindow from "../../components/eq-ui/EQWindow"
import {SpireApi} from "../../app/api/spire-api"
import {SpireQueryBuilder} from "../../app/api/spire-query-builder"
import {CharacterDatumApi} from "../../app/api/api/character-datum-api"
import {AdventureStatApi} from "../../app/api/api/adventure-stat-api"
import {AdventureTemplateApi} from "../../app/api/api/adventure-template-api"
import {DB_PLAYER_CLASSES_ALL} from "../../app/constants/eq-classes-constants"
import {Zones} from "../../app/zones"
import {
  CHAR_LDON_FIELDS,
  filterLdonAdventures,
  filterLdonZones,
  isLdonAdventureZone,
  LDON_SEARCH_CATEGORIES,
  LDON_THEMES,
  ldonAdventureType,
  ldonThemeFromId,
  ldonThemeFromShortName,
} from "../../app/ldon"

export default {
  name: "CharacterLdon",
  components: {EqWindow},
  data() {
    return {
      search: "",
      category: "zones",
      themeKey: "all",
      categories: LDON_SEARCH_CATEGORIES,
      themes: LDON_THEMES,
      charHits: [],
      character: null,
      selectedZone: null,
      ldonZones: [],
      adventures: [],
      loading: false,
      busy: false,
      status: "",
      ldonDraft: {available: 0, points: {}, wins: {}, losses: {}},
    }
  },
  computed: {
    windowTitle() {
      return this.character ? (this.character.name + " — LDoN") : "LDoN"
    },
    searchPlaceholder() {
      if (this.category === "characters") {
        return "Character name or id"
      }
      if (this.category === "adventures") {
        return "Adventure text, type, or LDoN zone"
      }
      return "LDoN zone name or short name"
    },
    ldonAvailable() {
      return this.character && this.character.ldon_points_available != null ? this.character.ldon_points_available : 0
    },
    visibleZones() {
      return filterLdonZones(this.ldonZones, this.themeKey, this.category === "zones" ? this.search : "")
    },
    zoneGroups() {
      const groups = []
      LDON_THEMES.forEach((theme) => {
        if (this.themeKey !== "all" && this.themeKey !== theme.key) {
          return
        }
        const zones = this.visibleZones.filter((zone) => {
          const found = ldonThemeFromShortName(zone)
          return found && found.key === theme.key
        })
        if (zones.length) {
          groups.push({key: theme.key, label: theme.label, zones})
        }
      })
      return groups
    },
    visibleAdventures() {
      return filterLdonAdventures(this.adventures, this.themeKey, this.category === "adventures" ? this.search : "")
    },
    zoneAdventures() {
      if (!this.selectedZone) {
        return []
      }
      const short = String(this.selectedZone.short_name || "").toLowerCase()
      return this.adventures.filter((row) => String(row.zone || "").toLowerCase() === short)
    },
  },
  watch: {
    "$route.query.c"() {
      this.bootFromRoute()
    },
    "$route.query.z"() {
      this.bootZone()
    },
    "$route.query.cat"() {
      this.bootCategory()
    },
    "$route.query.theme"() {
      this.bootTheme()
    },
  },
  mounted() {
    this.bootCategory()
    this.bootTheme()
    this.boot()
  },
  methods: {
    charApi() {
      return new CharacterDatumApi(...SpireApi.cfg())
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
    errorText(err) {
      if (err && err.response && err.response.data && err.response.data.error) {
        return err.response.data.error
      }
      return err && err.message ? err.message : "Request failed"
    },
    adventureType(type) {
      return ldonAdventureType(type)
    },
    themeLabel(row) {
      const theme = ldonThemeFromId(row && row.theme) || ldonThemeFromShortName(row && row.zone)
      return theme ? theme.label : ""
    },
    zoneLong(short) {
      const zone = Zones.zonesByShortName[String(short || "").toLowerCase()]
      return zone ? zone.long_name : short
    },
    adventureCount(short) {
      const name = String(short || "").toLowerCase()
      return this.adventures.filter((row) => String(row.zone || "").toLowerCase() === name).length
    },
    setCategory(id) {
      this.category = id
      this.pushQuery()
    },
    setTheme(key) {
      this.themeKey = key
      this.pushQuery()
    },
    bootCategory() {
      const cat = String(this.$route.query.cat || "")
      if (LDON_SEARCH_CATEGORIES.some((row) => row.id === cat)) {
        this.category = cat
      }
    },
    bootTheme() {
      const key = String(this.$route.query.theme || "")
      if (key === "all" || LDON_THEMES.some((theme) => theme.key === key)) {
        this.themeKey = key
      }
    },
    bootFromRoute() {
      const id = parseInt(this.$route.query.c, 10) || 0
      if (id) {
        this.selectCharacter(id)
      }
    },
    bootZone() {
      const short = String(this.$route.query.z || "").toLowerCase()
      if (!short || !isLdonAdventureZone(short)) {
        return
      }
      const zone = Zones.zonesByShortName[short]
      if (zone) {
        this.selectedZone = zone
      }
    },
    async boot() {
      this.loading = true
      try {
        const zones = await Zones.getZones()
        this.ldonZones = (Array.isArray(zones) ? zones : []).filter((zone) => isLdonAdventureZone(zone))
        const api = new AdventureTemplateApi(...SpireApi.cfg())
        const builder = new SpireQueryBuilder()
        builder.limit(4000)
        builder.orderBy(["theme", "zone", "id"])
        const r = await api.listAdventureTemplates(builder.get())
        this.adventures = ((r && r.data) || []).filter((row) => isLdonAdventureZone(row.zone))
        this.bootZone()
        this.bootFromRoute()
      } catch (err) {
        this.status = this.errorText(err)
      } finally {
        this.loading = false
      }
    },
    pushQuery() {
      const query = Object.assign({}, this.$route.query, {
        cat: this.category,
        theme: this.themeKey === "all" ? undefined : this.themeKey,
      })
      if (this.character) {
        query.c = String(this.character.id)
      }
      if (this.selectedZone) {
        query.z = this.selectedZone.short_name
      }
      Object.keys(query).forEach((key) => {
        if (query[key] == null || query[key] === "") {
          delete query[key]
        }
      })
      this.$router.replace({path: this.$route.path, query}).catch(() => {})
    },
    runSearch() {
      if (this.category === "characters") {
        this.searchCharacters()
        return
      }
      this.status = this.category === "zones"
        ? (this.visibleZones.length + " LDoN zones")
        : (this.visibleAdventures.length + " adventures")
    },
    async searchCharacters() {
      this.status = ""
      const q = (this.search || "").trim()
      if (!q) {
        this.charHits = []
        return
      }
      try {
        const builder = new SpireQueryBuilder()
        builder.limit(20)
        builder.orderBy(["name"])
        builder.select(CHAR_LDON_FIELDS)
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
    selectZone(zone) {
      if (!isLdonAdventureZone(zone)) {
        return
      }
      this.selectedZone = zone
      this.pushQuery()
    },
    openAdventureZone(row) {
      const zone = Zones.zonesByShortName[String(row.zone || "").toLowerCase()]
      if (zone) {
        this.selectedZone = zone
        this.category = "zones"
        this.pushQuery()
      }
    },
    emptyStats() {
      const wins = {}
      const losses = {}
      LDON_THEMES.forEach((theme) => {
        wins[theme.key] = 0
        losses[theme.key] = 0
      })
      return {wins, losses}
    },
    async selectCharacter(id) {
      this.status = ""
      this.loading = true
      this.character = this.character && this.character.id === id ? this.character : {id}
      try {
        const cr = await this.charApi().getCharacterDatum({
          id,
          select: CHAR_LDON_FIELDS.join("."),
        })
        this.character = cr && cr.data ? cr.data : null
        if (this.character && !this.charHits.find((h) => h.id === this.character.id)) {
          this.charHits = [this.character].concat(this.charHits)
        }
        await this.loadLdon()
        this.pushQuery()
      } catch (err) {
        this.status = this.errorText(err)
        this.character = null
      } finally {
        this.loading = false
      }
    },
    async loadLdon() {
      const points = {}
      LDON_THEMES.forEach((theme) => {
        points[theme.key] = this.character ? Number(this.character[theme.points] || 0) : 0
      })
      const stats = this.emptyStats()
      if (this.character) {
        try {
          const api = new AdventureStatApi(...SpireApi.cfg())
          const qb = new SpireQueryBuilder()
          qb.where("player_id", "=", this.character.id)
          qb.limit(1)
          const r = await api.listAdventureStats(qb.get())
          const row = r && r.data && r.data.length ? r.data[0] : null
          if (row) {
            LDON_THEMES.forEach((theme) => {
              stats.wins[theme.key] = Number(row[theme.wins] || 0)
              stats.losses[theme.key] = Number(row[theme.losses] || 0)
            })
          }
        } catch (err) {}
      }
      this.ldonDraft = {
        available: this.character ? Number(this.character.ldon_points_available || 0) : 0,
        points,
        wins: stats.wins,
        losses: stats.losses,
      }
    },
    async saveLdon() {
      if (!this.character) {
        return
      }
      this.busy = true
      this.status = ""
      try {
        const body = {ldon_points_available: Number(this.ldonDraft.available || 0)}
        LDON_THEMES.forEach((theme) => {
          body[theme.points] = Number(this.ldonDraft.points[theme.key] || 0)
        })
        await this.charApi().updateCharacterDatum({id: this.character.id, characterDatum: body})
        const stats = {player_id: this.character.id}
        LDON_THEMES.forEach((theme) => {
          stats[theme.wins] = Number(this.ldonDraft.wins[theme.key] || 0)
          stats[theme.losses] = Number(this.ldonDraft.losses[theme.key] || 0)
        })
        const api = new AdventureStatApi(...SpireApi.cfg())
        try {
          await api.updateAdventureStat({id: this.character.id, adventureStat: stats})
        } catch (err) {
          await api.createAdventureStat({adventureStat: stats})
        }
        const cr = await this.charApi().getCharacterDatum({
          id: this.character.id,
          select: CHAR_LDON_FIELDS.join("."),
        })
        this.character = cr && cr.data ? cr.data : this.character
        this.status = "Saved LDoN"
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
.access-split { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 12px; align-items: start; }
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
.ui-field { margin-bottom: 8px; }
.ui-field label { display: block; font-size: 11px; color: var(--text-muted, #8b93a0); }
.ui-field-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.ui-toolbar { display: flex; flex-wrap: wrap; gap: 6px; align-items: center; }
.tabular { font-variant-numeric: tabular-nums; }
.mb-2 { margin-bottom: 8px; }
.mb-3 { margin-bottom: 12px; }
.mt-2 { margin-top: 8px; }
.mt-3 { margin-top: 12px; }
</style>

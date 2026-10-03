<template>
  <content-area>
    <eq-window title="Tier Factory">
      <b-alert show variant="danger" v-if="error">
        <i class="fa fa-warning"></i> {{ error }}
      </b-alert>
      <b-alert show variant="info" v-if="plan && plan.dryRun && plan.ok">
        Preview only. {{ (plan.writes || []).length }} file action(s), {{ (plan.dbWrites || []).length }} database write(s). Nothing written yet.
      </b-alert>
      <b-alert show variant="success" v-if="plan && !plan.dryRun && plan.ok">
        Wrote the recipe onto {{ (plan.steps || []).length }} zone(s). Use Apply to reload quests, queue controller commands, or reboot the zones.
      </b-alert>

      <div class="ui-toolbar">
        <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER">All ZC zones</router-link>
        <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER_GUIDE">Guide</router-link>
        <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER_BUILDER">Builder</router-link>
        <router-link class="btn btn-sm btn-dark" :to="ROUTE.CONTENT_FACTORY">Content Factory</router-link>
        <button type="button" class="btn btn-sm" :class="tab === 'recipe' ? 'btn-primary' : 'btn-dark'" @click="tab = 'recipe'">Recipe</button>
        <button type="button" class="btn btn-sm" :class="tab === 'run' ? 'btn-primary' : 'btn-dark'" @click="tab = 'run'">Ladder / stamp</button>
        <button type="button" class="btn btn-sm" :class="tab === 'coverage' ? 'btn-primary' : 'btn-dark'" @click="tab = 'coverage'">Coverage</button>
        <button type="button" class="btn btn-sm" :class="tab === 'validate' ? 'btn-primary' : 'btn-dark'" @click="tab = 'validate'">Validate</button>
        <button type="button" class="btn btn-sm" :class="tab === 'apply' ? 'btn-primary' : 'btn-dark'" @click="tab = 'apply'">Apply</button>
      </div>

      <p class="zc-copy">
        A recipe is the tier: trash/boss/raid numbers plus a gear kit. Stamp copies that kit onto many zones.
        Ladder steps each zone up by the multiplier and mints a new item ID range so T2 is not wearing T1 loot.
        Mint the kit itself in Content Factory (spell clones, 18×16 grid, loot tables), then stamp it here.
      </p>

      <eq-window v-if="tab === 'recipe'" title="Named recipe">
        <div class="ui-toolbar">
          <select v-model="recipeId" class="form-control form-control-sm" style="width: 260px" @change="loadSelectedRecipe">
            <option value="">New recipe</option>
            <option v-for="r in recipes" :key="r.id" :value="r.id">{{ r.name }} ({{ r.id }})</option>
          </select>
          <select v-model.number="importZone" class="form-control form-control-sm" style="width: 260px">
            <option :value="0">Import kit from a finished zone</option>
            <option v-for="z in configured" :key="z.zoneId" :value="z.zoneId">
              {{ z.zoneId }} — {{ zoneLabel(z.zoneId) }}
            </option>
          </select>
          <button type="button" class="btn btn-sm btn-dark" :disabled="!importZone || busy" @click="importFromZone">Load zone kit</button>
          <button type="button" class="btn btn-sm btn-primary" :disabled="busy" @click="saveRecipe">Save recipe</button>
          <button type="button" class="btn btn-sm btn-outline-danger" :disabled="busy || !recipe.id" @click="deleteRecipe">Delete</button>
        </div>

        <div class="ui-field-grid">
          <div class="ui-field"><label>Recipe id</label><input v-model="recipe.id" class="form-control form-control-sm" placeholder="classic-t1"></div>
          <div class="ui-field"><label>Display name</label><input v-model="recipe.name" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>Zone group tag</label><input v-model="recipe.group" class="form-control form-control-sm" placeholder="classic-t1"></div>
          <div class="ui-field"><label>Default step</label><input v-model.number="recipe.step" type="number" step="0.05" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>Item prefix</label><input v-model="recipe.prefix" class="form-control form-control-sm" placeholder="T1"></div>
        </div>

        <table class="eq-table mt-3" style="width: 100%">
          <thead>
          <tr>
            <th></th>
            <th>Level</th>
            <th>Max HP</th>
            <th>Min hit</th>
            <th>Max hit</th>
            <th>ATK</th>
            <th>AC</th>
            <th>Cash</th>
            <th>Loot tables</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="kind in kinds" :key="kind">
            <td><b>{{ kind }}</b></td>
            <td><input v-model="recipe.basedata[kind].level" class="form-control form-control-sm"></td>
            <td><input v-model="recipe.basedata[kind].max_hp" class="form-control form-control-sm"></td>
            <td><input v-model="recipe.basedata[kind].min_hit" class="form-control form-control-sm"></td>
            <td><input v-model="recipe.basedata[kind].max_hit" class="form-control form-control-sm"></td>
            <td><input v-model="recipe.basedata[kind].atk" class="form-control form-control-sm"></td>
            <td><input v-model="recipe.basedata[kind].ac" class="form-control form-control-sm"></td>
            <td><input v-model="recipe.basedata[kind].cash" class="form-control form-control-sm"></td>
            <td><input v-model="recipe.basedata[kind].loot" class="form-control form-control-sm" placeholder="trash@3"></td>
          </tr>
          </tbody>
        </table>

        <div class="zc-hint mt-3">
          Kit: <b>{{ (recipe.kit.items || []).length }}</b> items
          · <b>{{ (recipe.kit.tables || []).length }}</b> loot tables
          <span v-if="recipe.kit.sourceZone"> · source zone {{ recipe.kit.sourceZone }}</span>
        </div>
      </eq-window>

      <eq-window v-else-if="tab === 'run'" title="Stamp or ladder">
        <div class="ui-field-grid">
          <div class="ui-field">
            <label>Mode</label>
            <select v-model="run.mode" class="form-control form-control-sm">
              <option value="stamp">Stamp — same tier on every zone</option>
              <option value="ladder">Ladder — each zone is the next step</option>
            </select>
          </div>
          <div class="ui-field">
            <label>{{ run.mode === 'ladder' ? 'Step multiplier' : 'Scale (1 = recipe as-is)' }}</label>
            <input v-model.number="run.step" type="number" step="0.05" class="form-control form-control-sm">
          </div>
          <div class="ui-field">
            <label>Item prefix</label>
            <input v-model="run.prefix" class="form-control form-control-sm" :placeholder="recipe.prefix || 'T1'">
          </div>
          <div class="ui-field">
            <label>First new item ID (0 = next free / 800000+)</label>
            <input v-model.number="run.itemStartId" type="number" class="form-control form-control-sm">
          </div>
          <div class="ui-field">
            <label>Evolving kill count</label>
            <input v-model.number="run.evoKills" type="number" class="form-control form-control-sm">
          </div>
        </div>

        <div class="ui-field mt-3">
          <label>Target zone IDs</label>
          <textarea v-model="run.idsText" class="form-control form-control-sm" rows="3" placeholder="17, 31, 39"></textarea>
          <div class="zc-hint">{{ parseIds(run.idsText).length }} zone(s)</div>
        </div>

        <div class="zc-missing" v-if="missing.length">
          <div class="zc-hint">PEQ zones with no ultimatedata yet (click to add):</div>
          <button v-for="z in missing.slice(0, 40)" :key="z.zoneidnumber" type="button" class="zc-pill ui-plain" @click="addRunId(z.zoneidnumber)">
            {{ z.zoneidnumber }} {{ z.short_name }}
          </button>
        </div>

        <div class="zc-checks">
          <label><input type="checkbox" v-model="run.createZones"> Create missing zone JSON</label>
          <label><input type="checkbox" v-model="run.createItems"> Clone / scale PEQ items</label>
          <label><input type="checkbox" v-model="run.copyGear"> Write loot + item JSON</label>
          <label><input type="checkbox" v-model="run.classify"> Classify spawns from PEQ</label>
          <label><input type="checkbox" v-model="run.merchants"> Add kit to a zone merchant</label>
          <label><input type="checkbox" v-model="run.traits"> Trait vendor rows (click/proc)</label>
          <label><input type="checkbox" v-model="run.evolving"> Evolving T1→Tn (ladder only)</label>
          <label><input type="checkbox" v-model="run.overwrite"> Overwrite existing custom names</label>
        </div>

        <div class="ui-toolbar">
          <button class="btn btn-sm btn-dark" :disabled="busy" @click="runFactory(true)">Preview</button>
          <button class="btn btn-sm btn-primary" :disabled="busy" @click="runFactory(false)">Write JSON + items</button>
        </div>
      </eq-window>

      <eq-window v-else-if="tab === 'coverage'" title="Class / slot coverage">
        <p class="zc-copy">
          Empty cells are missing gear for that class and slot. Preview a run or import a zone kit first.
          {{ coverageMissing }} empty of {{ coverageCells.length }}.
        </p>
        <div class="cov-wrap" v-if="coverageSlots.length">
          <table class="eq-table cov-table">
            <thead>
            <tr>
              <th></th>
              <th v-for="cls in coverageClasses" :key="cls">{{ cls }}</th>
            </tr>
            </thead>
            <tbody>
            <tr v-for="slot in coverageSlots" :key="slot">
              <td><b>{{ slot }}</b></td>
              <td v-for="cls in coverageClasses" :key="slot + cls" :class="cellClass(slot, cls)">
                {{ cellCount(slot, cls) || "" }}
              </td>
            </tr>
            </tbody>
          </table>
        </div>
        <div v-else class="text-muted">No kit items with slot/class bits yet. Import a zone that already has items.</div>
      </eq-window>

      <eq-window v-else-if="tab === 'validate'" title="Validate JSON + PEQ">
        <p class="zc-copy">
          Checks loot names against item JSON, item IDs against the items table, custom names against spawn2,
          and whether a zone controller NPC exists.
        </p>
        <div class="ui-field">
          <label>Zone IDs (blank = all configured)</label>
          <input v-model="validateIds" class="form-control form-control-sm" placeholder="17, 31">
        </div>
        <div class="ui-toolbar">
          <button class="btn btn-sm btn-primary" :disabled="busy" @click="runValidate">Validate</button>
        </div>
        <div v-if="report" class="zc-hint mt-2">
          {{ report.zones }} zone(s) · <b>{{ report.errors }}</b> errors · {{ report.warnings }} warnings
        </div>
        <table v-if="report" class="eq-table mt-3" style="width: 100%">
          <thead>
          <tr><th>Zone</th><th>Level</th><th>Kind</th><th>Message</th></tr>
          </thead>
          <tbody>
          <tr v-for="(issue, i) in (report.issues || [])" :key="i">
            <td class="tabular">{{ issue.zoneId }}</td>
            <td>{{ issue.severity }}</td>
            <td>{{ issue.kind }}</td>
            <td>{{ issue.message }}</td>
          </tr>
          <tr v-if="!(report.issues || []).length">
            <td colspan="4" class="text-muted">No issues on the checked zones.</td>
          </tr>
          </tbody>
        </table>
      </eq-window>

      <eq-window v-else title="Apply live">
        <p class="zc-copy">
          Spire talks to World over telnet and drops a command file the zone controller
          picks up on its next tick. Occupied zones are not kicked unless you check reboot.
        </p>
        <div class="ui-field">
          <label>Zone IDs</label>
          <input v-model="run.idsText" class="form-control form-control-sm">
        </div>
        <div class="zc-live-note" v-if="live">
          <template v-if="live.worldOk">World telnet is up. {{ livePopped }} popped of {{ applyIds.length || (live.zones || []).length }} listed.</template>
          <template v-else>{{ live.worldNote || "World telnet is down." }}</template>
        </div>
        <div class="zc-checks">
          <label><input type="checkbox" v-model="apply.reloadQuests"> Reload quest scripts</label>
          <label><input type="checkbox" v-model="apply.reloadWorld"> World <code>api reload quests</code></label>
          <label><input type="checkbox" v-model="apply.refresh"> Queue refresh JSON</label>
          <label><input type="checkbox" v-model="apply.rebuff"> Queue rebuff</label>
          <label><input type="checkbox" v-model="apply.repopBosses"> Queue boss repop</label>
          <label><input type="checkbox" v-model="apply.rebootEmpty"> Reboot empty popped zones</label>
          <label><input type="checkbox" v-model="apply.bootIfDown"> Boot zones that are down</label>
          <label><input type="checkbox" v-model="apply.rebootZones"> Reboot occupied zones (kicks players)</label>
        </div>
        <div class="ui-toolbar">
          <button class="btn btn-sm btn-dark" :disabled="busy" @click="loadLive">Refresh live</button>
          <button class="btn btn-sm btn-dark" :disabled="busy" @click="runApply(false)">Preview only</button>
          <button class="btn btn-sm btn-primary" :disabled="busy" @click="runApply(true)">Apply to World</button>
        </div>
        <div v-if="applyResult">
          <div v-if="applyResult.worldNote" class="zc-hint mt-2">{{ applyResult.worldNote }}</div>
          <div v-if="(applyResult.reloaded || []).length" class="zc-hint">Reloaded: {{ applyResult.reloaded.join(", ") }}</div>
          <div v-if="(applyResult.rebooted || []).length" class="zc-hint">World: {{ applyResult.rebooted.join(", ") }}</div>
          <div v-if="(applyResult.queued || []).length" class="zc-hint">
            Queued: {{ applyResult.queued.map((q) => (q.short || q.zoneId) + " " + (q.commands || []).join(", ")).join("; ") }}
          </div>
          <div v-for="w in (applyResult.warnings || [])" :key="w" class="text-warning">{{ w }}</div>
          <table v-if="(applyResult.live || []).length" class="eq-table mt-3" style="width: 100%">
            <thead>
            <tr><th>Zone</th><th>State</th><th>Players</th></tr>
            </thead>
            <tbody>
            <tr v-for="z in applyResult.live" :key="z.zoneId">
              <td class="tabular">{{ z.zoneId }} {{ z.long || z.short }}</td>
              <td>{{ z.popped ? "popped" : "down" }}</td>
              <td class="tabular">{{ z.players || 0 }}</td>
            </tr>
            </tbody>
          </table>
          <div class="zc-hint mt-2">Fallback if a zone never ticks: <span v-for="cmd in (applyResult.commands || [])" :key="cmd"><code>{{ cmd }}</code> </span></div>
        </div>
      </eq-window>

      <eq-window v-if="plan && plan.steps && plan.steps.length" title="Preview" class="mt-3">
        <table class="eq-table" style="width: 100%">
          <thead>
          <tr>
            <th>Zone</th>
            <th>Prefix</th>
            <th class="text-right">Scale</th>
            <th>Trash</th>
            <th>Boss</th>
            <th>Raid</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="step in plan.steps" :key="step.zoneId + '-' + step.index">
            <td class="tabular"><b>{{ step.zoneId }}</b> {{ zoneLabel(step.zoneId) }}</td>
            <td>{{ step.prefix }}</td>
            <td class="text-right tabular">{{ Number(step.multiplier).toFixed(2) }}×</td>
            <td class="tabular">{{ baseLine(step, 'trash') }}</td>
            <td class="tabular">{{ baseLine(step, 'boss') }}</td>
            <td class="tabular">{{ baseLine(step, 'raid') }}</td>
          </tr>
          </tbody>
        </table>
        <div v-if="(plan.itemClones || []).length" class="zc-hint mt-2">
          {{ plan.itemClones.length }} item clone(s), first id {{ plan.itemClones[0].newId }}
        </div>
      </eq-window>

      <eq-window v-if="plan" title="Plan" class="mt-3">
        <div v-if="plan.error" class="text-danger mb-2">{{ plan.error }}</div>
        <div v-if="(plan.warnings || []).length" class="mb-2">
          <div v-for="w in plan.warnings" :key="w" class="text-warning">{{ w }}</div>
        </div>
        <table class="eq-table" style="width: 100%">
          <thead>
          <tr><th>Zone</th><th>File</th><th>Action</th></tr>
          </thead>
          <tbody>
          <tr v-for="w in (plan.writes || [])" :key="(w.relPath || '') + w.action + w.zoneId">
            <td class="tabular">{{ w.zoneId || "—" }}</td>
            <td><code>{{ w.relPath }}</code></td>
            <td>{{ w.action }}</td>
          </tr>
          <tr v-if="!(plan.writes || []).length">
            <td colspan="3" class="text-muted">No file writes in this plan.</td>
          </tr>
          </tbody>
        </table>
        <table v-if="(plan.dbWrites || []).length" class="eq-table mt-3" style="width: 100%">
          <thead>
          <tr><th>Table</th><th>Action</th><th>ID</th><th>Note</th></tr>
          </thead>
          <tbody>
          <tr v-for="(w, i) in plan.dbWrites" :key="i">
            <td>{{ w.table }}</td>
            <td>{{ w.action }}</td>
            <td class="tabular">{{ w.id }}</td>
            <td>{{ w.note }}</td>
          </tr>
          </tbody>
        </table>
      </eq-window>
    </eq-window>
  </content-area>
</template>

<script>
import EqWindow from "../../components/eq-ui/EQWindow"
import ContentArea from "../../components/layout/ContentArea"
import {ROUTE} from "@/routes"
import {Zones} from "../../app/zones"
import {ZoneControllerApi} from "../../app/zone-controller"

function emptyBasedata() {
  const row = () => ({level: "", max_hp: "", min_hit: "", max_hit: "", atk: "", ac: "", cash: "", loot: ""})
  return {trash: row(), boss: row(), raid: row()}
}

function emptyRecipe() {
  return {
    id: "",
    name: "",
    group: "",
    step: 1.35,
    prefix: "T1",
    basedata: emptyBasedata(),
    kit: {sourceZone: 0, items: [], tables: []},
  }
}

export default {
  name: "ZoneTierFactory",
  components: {ContentArea, EqWindow},
  data() {
    return {
      ROUTE,
      kinds: ["trash", "boss", "raid"],
      error: "",
      busy: false,
      tab: "recipe",
      plan: null,
      report: null,
      applyResult: null,
      live: null,
      recipes: [],
      recipeId: "",
      recipe: emptyRecipe(),
      importZone: 0,
      configured: [],
      peqZones: {},
      peqList: [],
      validateIds: "",
      run: {
        mode: "stamp",
        step: 1,
        prefix: "",
        itemStartId: 0,
        evoKills: 100,
        idsText: "",
        createZones: true,
        createItems: true,
        copyGear: true,
        classify: true,
        merchants: false,
        traits: false,
        evolving: false,
        overwrite: false,
      },
      apply: {
        reloadQuests: true,
        reloadWorld: false,
        refresh: true,
        rebuff: true,
        repopBosses: false,
        rebootEmpty: true,
        bootIfDown: false,
        rebootZones: false,
      },
    }
  },
  computed: {
    missing() {
      const have = {}
      this.configured.forEach((z) => {
        have[z.zoneId] = true
      })
      return (this.peqList || []).filter((z) => z.version === 0 && z.zoneidnumber && !have[z.zoneidnumber])
        .sort((a, b) => a.zoneidnumber - b.zoneidnumber)
    },
    coverageCells() {
      if (this.plan && this.plan.coverage && this.plan.coverage.length) {
        return this.plan.coverage
      }
      return this.coverageFromKit(this.recipe.kit && this.recipe.kit.items)
    },
    coverageSlots() {
      const seen = []
      this.coverageCells.forEach((c) => {
        if (seen.indexOf(c.slot) === -1) {
          seen.push(c.slot)
        }
      })
      return seen
    },
    coverageClasses() {
      const seen = []
      this.coverageCells.forEach((c) => {
        if (seen.indexOf(c.class) === -1) {
          seen.push(c.class)
        }
      })
      return seen
    },
    coverageMissing() {
      return this.coverageCells.filter((c) => !c.filled).length
    },
    applyIds() {
      return this.parseIds(this.run.idsText)
    },
    livePopped() {
      return ((this.live && this.live.zones) || []).filter((z) => z.popped).length
    },
  },
  async mounted() {
    try {
      const zones = await Zones.getZones()
      this.peqList = zones || []
      const byId = {}
      this.peqList.forEach((z) => {
        if (z && z.zoneidnumber && !byId[z.zoneidnumber]) {
          byId[z.zoneidnumber] = z
        }
      })
      this.peqZones = byId
      const index = await ZoneControllerApi.list()
      this.configured = index.zones || []
      await this.reloadRecipes()
      await this.loadLive()
    } catch (e) {
      this.error = this.errText(e)
    }
  },
  watch: {
    tab(next) {
      if (next === "apply") {
        this.loadLive()
      }
    },
  },
  methods: {
    parseIds(text) {
      return String(text || "").split(/[,\s]+/).map((s) => parseInt(s, 10)).filter((n) => n > 0)
    },
    zoneLabel(id) {
      const z = this.peqZones[id]
      return z && z.long_name ? z.long_name : ("Zone " + id)
    },
    addRunId(id) {
      const cur = this.parseIds(this.run.idsText)
      if (cur.indexOf(id) === -1) {
        cur.push(id)
      }
      this.run.idsText = cur.join(", ")
    },
    mergeBasedata(src) {
      const out = emptyBasedata()
      this.kinds.forEach((kind) => {
        out[kind] = Object.assign(out[kind], (src && src[kind]) || {})
      })
      return out
    },
    applyRecipe(row) {
      const next = emptyRecipe()
      Object.assign(next, row || {})
      next.basedata = this.mergeBasedata(row && row.basedata)
      next.kit = Object.assign({sourceZone: 0, items: [], tables: []}, (row && row.kit) || {})
      this.recipe = next
    },
    loadSelectedRecipe() {
      const row = this.recipes.find((r) => r.id === this.recipeId)
      if (row) {
        this.applyRecipe(row)
      } else {
        this.recipe = emptyRecipe()
      }
    },
    async reloadRecipes() {
      const listed = await ZoneControllerApi.recipes()
      this.recipes = listed.recipes || []
    },
    async importFromZone() {
      this.busy = true
      this.error = ""
      try {
        const row = await ZoneControllerApi.recipeFromZone(this.importZone)
        this.applyRecipe(row)
        this.recipeId = row.id || ""
      } catch (e) {
        this.error = this.errText(e)
      }
      this.busy = false
    },
    async saveRecipe() {
      this.busy = true
      this.error = ""
      try {
        const plan = await ZoneControllerApi.saveRecipe(this.recipe)
        if (plan && plan.error) {
          this.error = plan.error
        } else {
          this.recipeId = this.recipe.id
          await this.reloadRecipes()
        }
      } catch (e) {
        this.error = this.errText(e)
      }
      this.busy = false
    },
    async deleteRecipe() {
      if (!this.recipe.id) {
        return
      }
      this.busy = true
      this.error = ""
      try {
        await ZoneControllerApi.deleteRecipe(this.recipe.id)
        this.recipe = emptyRecipe()
        this.recipeId = ""
        await this.reloadRecipes()
      } catch (e) {
        this.error = this.errText(e)
      }
      this.busy = false
    },
    factoryBody(dryRun) {
      return {
        recipeId: this.recipe.id,
        recipe: this.recipe,
        zoneIds: this.parseIds(this.run.idsText),
        mode: this.run.mode,
        step: Number(this.run.step) || 0,
        prefix: this.run.prefix,
        itemStartId: Number(this.run.itemStartId) || 0,
        createZones: this.run.createZones,
        createItems: this.run.createItems,
        copyGear: this.run.copyGear,
        classify: this.run.classify,
        merchants: this.run.merchants,
        traits: this.run.traits,
        evolving: this.run.evolving,
        overwrite: this.run.overwrite,
        evoKills: Number(this.run.evoKills) || 100,
        dryRun,
      }
    },
    async runFactory(dryRun) {
      this.busy = true
      this.error = ""
      this.plan = null
      try {
        this.plan = await ZoneControllerApi.factory(this.factoryBody(dryRun))
        if (this.plan && this.plan.error) {
          this.error = this.plan.error
        } else if (this.plan && !this.plan.dryRun) {
          const index = await ZoneControllerApi.list()
          this.configured = index.zones || []
        }
        if (this.plan && this.plan.coverage) {
          this.tab = dryRun ? "coverage" : this.tab
        }
      } catch (e) {
        this.error = this.errText(e)
        if (e.response && e.response.data && e.response.data.steps) {
          this.plan = e.response.data
        }
      }
      this.busy = false
    },
    async runValidate() {
      this.busy = true
      this.error = ""
      try {
        this.report = await ZoneControllerApi.validate(this.parseIds(this.validateIds || this.run.idsText))
      } catch (e) {
        this.error = this.errText(e)
      }
      this.busy = false
    },
    applyQueue() {
      const cmds = []
      if (this.apply.refresh) {
        cmds.push("refreshzonedata")
      }
      if (this.apply.rebuff) {
        cmds.push("rebuffzone")
      }
      if (this.apply.repopBosses) {
        cmds.push("repopallstaticbosses")
      }
      return cmds
    },
    async loadLive() {
      try {
        this.live = await ZoneControllerApi.live(this.parseIds(this.run.idsText))
      } catch (e) {
        this.live = {ok: false, worldOk: false, zones: [], worldNote: this.errText(e)}
      }
    },
    async runApply(live) {
      this.busy = true
      this.error = ""
      try {
        this.applyResult = await ZoneControllerApi.apply({
          zoneIds: this.parseIds(this.run.idsText),
          reloadQuests: live && this.apply.reloadQuests,
          reloadWorld: live && this.apply.reloadWorld,
          rebootZones: live && this.apply.rebootZones,
          rebootEmpty: live && this.apply.rebootEmpty,
          bootIfDown: live && this.apply.bootIfDown,
          queueCommands: live ? this.applyQueue() : [],
        })
        if (live) {
          await this.loadLive()
        }
      } catch (e) {
        this.error = this.errText(e)
      }
      this.busy = false
    },
    cellCount(slot, cls) {
      const cell = this.coverageCells.find((c) => c.slot === slot && c.class === cls)
      return cell ? cell.count : 0
    },
    coverageFromKit(items) {
      const slots = [
        ["Charm", [1]], ["Ear", [2, 16]], ["Head", [4]], ["Face", [8]], ["Neck", [32]],
        ["Shoulders", [64]], ["Arms", [128]], ["Back", [256]], ["Wrist", [512, 1024]],
        ["Range", [2048]], ["Hands", [4096]], ["Primary", [8192]], ["Secondary", [16384]],
        ["Ring", [32768, 65536]], ["Chest", [131072]], ["Legs", [262144]], ["Feet", [524288]], ["Waist", [1048576]],
      ]
      const classes = [
        ["WAR", 1], ["CLR", 2], ["PAL", 4], ["RNG", 8], ["SK", 16], ["DRU", 32],
        ["MNK", 64], ["BRD", 128], ["ROG", 256], ["SHM", 512], ["NEC", 1024],
        ["WIZ", 2048], ["MAG", 4096], ["ENC", 8192], ["BST", 16384], ["BER", 32768],
      ]
      const rows = items || []
      const out = []
      slots.forEach((slot) => {
        classes.forEach((cls) => {
          let n = 0
          rows.forEach((it) => {
            const okSlot = !it.slots || slot[1].some((bit) => (it.slots & bit) !== 0)
            const okClass = !it.classes || (it.classes & cls[1]) !== 0
            if (it.slots || it.classes) {
              if (okSlot && okClass) {
                n++
              }
            }
          })
          out.push({slot: slot[0], class: cls[0], count: n, filled: n > 0})
        })
      })
      return rows.some((it) => it.slots || it.classes) ? out : []
    },
    cellClass(slot, cls) {
      return this.cellCount(slot, cls) ? "cov-yes" : "cov-no"
    },
    baseLine(step, kind) {
      const row = (step.basedata && step.basedata[kind]) || {}
      const parts = [row.level, row.max_hp, row.min_hit, row.max_hit].filter(Boolean)
      return parts.join(" / ")
    },
    errText(e) {
      const status = e.response && e.response.status
      if (status === 404) {
        return "Factory routes are not on the running backend yet. Restart Spire with start_spire_dev.bat."
      }
      return (e.response && e.response.data && e.response.data.error) || String(e)
    },
  },
}
</script>

<style scoped>
.zc-copy, .zc-hint {
  color: var(--text-muted, #b7c0cc);
  margin-bottom: 12px;
}

.zc-checks {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 18px;
  margin: 12px 0;
}

.zc-checks label {
  margin: 0;
}

.zc-live-note {
  margin: 10px 0 12px;
  padding: 10px 12px;
  border: 1px solid var(--border, #3a4350);
  background: var(--surface-2, #12161d);
  border-radius: 6px;
}

.zc-missing {
  margin: 10px 0 14px;
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

.cov-wrap {
  overflow: auto;
}

.cov-table td, .cov-table th {
  text-align: center;
  min-width: 42px;
}

.cov-table td:first-child, .cov-table th:first-child {
  text-align: left;
}

.cov-yes {
  background: rgba(46, 160, 67, 0.18);
}

.cov-no {
  background: rgba(248, 81, 73, 0.12);
}

.mt-3 {
  margin-top: 1rem;
}
</style>

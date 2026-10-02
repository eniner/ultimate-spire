<template>
  <content-area>
    <eq-window title="Ultimate systems">
      <b-alert show variant="danger" v-if="error">
        <i class="fa fa-warning"></i> {{ error }}
      </b-alert>
      <b-alert show variant="info" v-if="plan && plan.dryRun && plan.ok">
        Preview only. {{ plan.action }} <code>{{ plan.relPath }}</code> ({{ plan.key }}).
      </b-alert>
      <b-alert show variant="success" v-if="plan && !plan.dryRun && plan.ok">
        Wrote <code>{{ plan.relPath }}</code>. In-game still needs a quest reload / hail / login to pick this up.
      </b-alert>

      <app-loader :is-loading="!loaded" class="mt-3 mb-3"/>

      <div v-if="loaded">
        <div class="ui-toolbar">
          <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER">Zone Controller</router-link>
          <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER_BUILDER">ZC builder</router-link>
          <button v-for="t in tabs" :key="t.id" type="button" class="btn btn-sm" :class="tab === t.id ? 'btn-primary' : 'btn-dark'" @click="tab = t.id">
            {{ t.label }}
          </button>
          <div class="ui-stat-line ml-auto">
            <b>{{ status.talentCount || 0 }}</b> talents
            <span class="zc-dot">·</span>
            <b>{{ status.rankCount || 0 }}</b> ranks
            <span class="zc-dot">·</span>
            <b>{{ status.traitCount || 0 }}</b> traits
            <span class="zc-dot">·</span>
            <b>{{ status.runewordCombos || 0 }}</b> runewords
          </div>
        </div>

        <p class="zc-copy">
          Reading <code class="privacy-hide">{{ status.questsDir }}</code>.
          Talents live in <code>talentdata/talent_catalog.json</code>,
          ranks/unlocks next to zone JSON,
          traits in <code>vendordata</code>,
          runewords as a Spire catalog (runtime still uses the custom_runeword_* tables when those exist).
        </p>

        <eq-window v-if="tab === 'talents'" title="Talent catalog">
          <div class="ui-toolbar">
            <select v-model.number="classFilter" class="form-control form-control-sm" style="width: 180px">
              <option :value="0">All classes</option>
              <option v-for="n in 16" :key="n" :value="n">{{ className(n) }}</option>
            </select>
            <input v-model="search" type="search" class="form-control form-control-sm" style="width: 260px" placeholder="Name or talent_id">
            <button type="button" class="btn btn-sm btn-dark" @click="newTalent">New talent</button>
          </div>
          <div class="sys-split">
            <table class="eq-table eq-highlight-rows" style="width: 100%">
              <thead>
              <tr>
                <th>Talent</th>
                <th>Class</th>
                <th>Tree</th>
                <th class="text-right">Ranks</th>
                <th class="text-right">Cost</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="row in filteredTalents" :key="row.talentId" class="zc-row" @click="editTalent(row)">
                <td>
                  <b>{{ row.displayName }}</b>
                  <div class="text-muted">{{ row.talentId }}</div>
                </td>
                <td>{{ classList(row.classIds) }}</td>
                <td>{{ row.treeId }}</td>
                <td class="text-right tabular">{{ row.rankCap }}</td>
                <td class="text-right tabular">{{ row.costAmount || 1 }}</td>
              </tr>
              </tbody>
            </table>
            <div class="sys-form" v-if="talentForm.open">
              <div class="ui-field"><label>talent_id</label><input v-model="talentForm.talent_id" class="form-control form-control-sm" :disabled="!!talentForm._existing"></div>
              <div class="ui-field"><label>Display name</label><input v-model="talentForm.display_name" class="form-control form-control-sm"></div>
              <div class="ui-field"><label>Description</label><textarea v-model="talentForm.description" class="form-control form-control-sm" rows="3"></textarea></div>
              <div class="ui-field-grid">
                <div class="ui-field">
                  <label>Class</label>
                  <select v-model.number="talentForm.class_id" class="form-control form-control-sm">
                    <option v-for="n in 16" :key="n" :value="n">{{ className(n) }}</option>
                  </select>
                </div>
                <div class="ui-field">
                  <label>Tree</label>
                  <select v-model="talentForm.tree_id" class="form-control form-control-sm">
                    <option v-for="tree in treesForClass(talentForm.class_id)" :key="tree.treeId" :value="tree.treeId">{{ tree.label }}</option>
                  </select>
                </div>
                <div class="ui-field">
                  <label>Point pool</label>
                  <select v-model="talentForm.point_pool_id" class="form-control form-control-sm">
                    <option v-for="p in (catalog.pools || [])" :key="p.poolId" :value="p.poolId">{{ p.displayName }}</option>
                  </select>
                </div>
                <div class="ui-field"><label>Rank cap</label><input v-model.number="talentForm.rank_cap" type="number" class="form-control form-control-sm"></div>
                <div class="ui-field"><label>Cost</label><input v-model.number="talentForm.cost_amount" type="number" class="form-control form-control-sm"></div>
                <div class="ui-field">
                  <label>Phase</label>
                  <select v-model="talentForm.progression_phase" class="form-control form-control-sm">
                    <option value="demigod">demigod</option>
                    <option value="god">god</option>
                    <option value="specialization">specialization</option>
                  </select>
                </div>
              </div>
              <div class="ui-field"><label>Required talents (comma ids)</label><input v-model="talentForm.required_talents" class="form-control form-control-sm"></div>
              <div class="zc-checks">
                <label><input type="checkbox" v-model="talentForm.hidden"> Hidden</label>
                <label><input type="checkbox" v-model="talentForm.create_rank"> Also create rank registry row</label>
              </div>
              <div class="ui-toolbar">
                <button class="btn btn-sm btn-dark" :disabled="busy" @click="saveTalent(true)">Preview</button>
                <button class="btn btn-sm btn-primary" :disabled="busy" @click="saveTalent(false)">Write JSON</button>
                <button v-if="talentForm._existing" class="btn btn-sm btn-outline-danger" :disabled="busy" @click="deleteKind('talent', {talent_id: talentForm.talent_id})">Delete</button>
              </div>
            </div>
          </div>
        </eq-window>

        <eq-window v-else-if="tab === 'ranks'" title="Talent rank registry">
          <p class="zc-copy">Bucket keys the server uses for spent ranks. Creating a catalog talent does not automatically add this unless you check that box.</p>
          <div class="ui-toolbar">
            <input v-model="search" type="search" class="form-control form-control-sm" style="width: 260px" placeholder="Key or class family">
            <button type="button" class="btn btn-sm btn-dark" @click="newRank">New rank</button>
          </div>
          <div class="sys-split">
            <table class="eq-table eq-highlight-rows" style="width: 100%">
              <thead><tr><th>Key</th><th>Family</th><th>Bucket</th></tr></thead>
              <tbody>
              <tr v-for="row in filteredRanks" :key="row.key" class="zc-row" @click="editRank(row)">
                <td><b>{{ row.key }}</b></td>
                <td>{{ row.classFamily }}</td>
                <td class="text-muted">{{ row.bucketSuffix }}</td>
              </tr>
              </tbody>
            </table>
            <div class="sys-form" v-if="rankForm.open">
              <div class="ui-field"><label>Key</label><input v-model="rankForm.key" class="form-control form-control-sm" :disabled="!!rankForm._existing"></div>
              <div class="ui-field"><label>Class family</label><input v-model="rankForm.class_family" class="form-control form-control-sm" placeholder="war"></div>
              <div class="ui-field"><label>Bucket suffix</label><input v-model="rankForm.bucket_suffix" class="form-control form-control-sm"></div>
              <div class="ui-toolbar">
                <button class="btn btn-sm btn-dark" :disabled="busy" @click="saveRank(true)">Preview</button>
                <button class="btn btn-sm btn-primary" :disabled="busy" @click="saveRank(false)">Write JSON</button>
                <button v-if="rankForm._existing" class="btn btn-sm btn-outline-danger" :disabled="busy" @click="deleteKind('rank', {key: rankForm.key})">Delete</button>
              </div>
            </div>
          </div>
        </eq-window>

        <eq-window v-else-if="tab === 'unlocks'" title="Unlock states">
          <div class="ui-toolbar">
            <input v-model="search" type="search" class="form-control form-control-sm" style="width: 260px" placeholder="Key or spell">
            <button type="button" class="btn btn-sm btn-dark" @click="newUnlock">New unlock</button>
          </div>
          <div class="sys-split">
            <table class="eq-table eq-highlight-rows" style="width: 100%">
              <thead><tr><th>Key</th><th>Spell</th><th class="text-right">ID</th></tr></thead>
              <tbody>
              <tr v-for="row in filteredUnlocks" :key="row.key" class="zc-row" @click="editUnlock(row)">
                <td><b>{{ row.key }}</b></td>
                <td>{{ row.spellName }}</td>
                <td class="text-right tabular">{{ row.spellId }}</td>
              </tr>
              </tbody>
            </table>
            <div class="sys-form" v-if="unlockForm.open">
              <div class="ui-field"><label>Key</label><input v-model="unlockForm.key" class="form-control form-control-sm" :disabled="!!unlockForm._existing"></div>
              <div class="ui-field"><label>Spell name</label><input v-model="unlockForm.spell_name" class="form-control form-control-sm"></div>
              <div class="ui-field"><label>Spell ID</label><input v-model.number="unlockForm.spell_id" type="number" class="form-control form-control-sm"></div>
              <div class="ui-field"><label>Expected value</label><input v-model="unlockForm.expected_value" class="form-control form-control-sm"></div>
              <div class="ui-toolbar">
                <button class="btn btn-sm btn-dark" :disabled="busy" @click="saveUnlock(true)">Preview</button>
                <button class="btn btn-sm btn-primary" :disabled="busy" @click="saveUnlock(false)">Write JSON</button>
                <button v-if="unlockForm._existing" class="btn btn-sm btn-outline-danger" :disabled="busy" @click="deleteKind('unlock', {key: unlockForm.key})">Delete</button>
              </div>
            </div>
          </div>
        </eq-window>

        <eq-window v-else-if="tab === 'specs'" title="Specializations">
          <div class="ui-toolbar">
            <select v-model.number="classFilter" class="form-control form-control-sm" style="width: 180px">
              <option :value="0">All classes</option>
              <option v-for="n in 16" :key="n" :value="n">{{ className(n) }}</option>
            </select>
            <input v-model="search" type="search" class="form-control form-control-sm" style="width: 220px" placeholder="Name">
            <button type="button" class="btn btn-sm btn-dark" @click="newSpec">New specialization</button>
          </div>
          <div class="sys-split">
            <table class="eq-table eq-highlight-rows" style="width: 100%">
              <thead><tr><th>Name</th><th>Class</th><th class="text-right">Ranks</th><th class="text-right">Cost</th></tr></thead>
              <tbody>
              <tr v-for="row in filteredSpecs" :key="row.classId + row.name" class="zc-row" @click="editSpec(row)">
                <td><b>{{ row.name }}</b></td>
                <td>{{ className(row.classId) }}</td>
                <td class="text-right tabular">{{ row.rankCount }}</td>
                <td class="text-right tabular">{{ row.cost }}</td>
              </tr>
              </tbody>
            </table>
            <div class="sys-form" v-if="specForm.open">
              <div class="ui-field"><label>Name</label><input v-model="specForm.name" class="form-control form-control-sm" :disabled="!!specForm._existing"></div>
              <div class="ui-field">
                <label>Class</label>
                <select v-model.number="specForm.class_id" class="form-control form-control-sm">
                  <option v-for="n in 16" :key="n" :value="n">{{ className(n) }}</option>
                </select>
              </div>
              <div class="ui-field"><label>Description</label><textarea v-model="specForm.description" class="form-control form-control-sm" rows="3"></textarea></div>
              <div class="ui-field-grid">
                <div class="ui-field"><label>Cost</label><input v-model.number="specForm.cost" type="number" class="form-control form-control-sm"></div>
                <div class="ui-field"><label>Ranks</label><input v-model.number="specForm.rankcount" type="number" class="form-control form-control-sm"></div>
              </div>
              <div class="ui-toolbar">
                <button class="btn btn-sm btn-dark" :disabled="busy" @click="saveSpec(true)">Preview</button>
                <button class="btn btn-sm btn-primary" :disabled="busy" @click="saveSpec(false)">Write JSON</button>
                <button v-if="specForm._existing" class="btn btn-sm btn-outline-danger" :disabled="busy" @click="deleteKind('spec', {name: specForm.name, class_id: specForm.class_id})">Delete</button>
              </div>
            </div>
          </div>
        </eq-window>

        <eq-window v-else-if="tab === 'traits'" title="Trait items">
          <p class="zc-copy">Vendor catalog of non-god trait drops. This is the list the server uses for trait loot notes, not the item row itself.</p>
          <div class="ui-toolbar">
            <input v-model="search" type="search" class="form-control form-control-sm" style="width: 260px" placeholder="Item or zone">
            <button type="button" class="btn btn-sm btn-dark" @click="newTrait">New trait</button>
          </div>
          <div class="sys-split">
            <table class="eq-table eq-highlight-rows" style="width: 100%">
              <thead><tr><th>Item</th><th>Zone</th><th>Classes</th><th class="text-right">Proc</th></tr></thead>
              <tbody>
              <tr v-for="row in filteredTraits" :key="row.itemId" class="zc-row" @click="editTrait(row)">
                <td><b>{{ row.itemName }}</b><div class="text-muted">{{ row.itemId }}</div></td>
                <td>{{ row.zoneShortName || row.zoneId }}</td>
                <td>{{ row.classes }}</td>
                <td class="text-right tabular">{{ row.procSpell || "" }}</td>
              </tr>
              </tbody>
            </table>
            <div class="sys-form" v-if="traitForm.open">
              <div class="ui-field"><label>Item ID</label><input v-model.number="traitForm.item_id" type="number" class="form-control form-control-sm" :disabled="!!traitForm._existing"></div>
              <div class="ui-field"><label>Name</label><input v-model="traitForm.item_name" class="form-control form-control-sm"></div>
              <div class="ui-field-grid">
                <div class="ui-field"><label>Zone ID</label><input v-model.number="traitForm.zone_id" type="number" class="form-control form-control-sm"></div>
                <div class="ui-field"><label>Short name</label><input v-model="traitForm.zone_short_name" class="form-control form-control-sm"></div>
                <div class="ui-field"><label>Tier</label><input v-model="traitForm.tier" class="form-control form-control-sm"></div>
                <div class="ui-field"><label>Classes</label><input v-model="traitForm.classes" class="form-control form-control-sm" placeholder="WAR, PAL"></div>
                <div class="ui-field"><label>Proc spell</label><input v-model.number="traitForm.proc_spell" type="number" class="form-control form-control-sm"></div>
                <div class="ui-field"><label>Focus spell</label><input v-model.number="traitForm.focus_spell" type="number" class="form-control form-control-sm"></div>
                <div class="ui-field"><label>Click spell</label><input v-model.number="traitForm.click_spell" type="number" class="form-control form-control-sm"></div>
                <div class="ui-field"><label>Worn spell</label><input v-model.number="traitForm.worn_spell" type="number" class="form-control form-control-sm"></div>
              </div>
              <div class="ui-toolbar">
                <button class="btn btn-sm btn-dark" :disabled="busy" @click="saveTrait(true)">Preview</button>
                <button class="btn btn-sm btn-primary" :disabled="busy" @click="saveTrait(false)">Write JSON</button>
                <button v-if="traitForm._existing" class="btn btn-sm btn-outline-danger" :disabled="busy" @click="deleteKind('trait', {item_id: traitForm.item_id})">Delete</button>
              </div>
            </div>
          </div>
        </eq-window>

        <eq-window v-else title="Runewords">
          <p class="zc-copy">
            Runtime craft is <code>custom_runeword_combo_map</code> (4 runes + ethereal base → output).
            Local peq
            <template v-if="status.runewordTables && status.runewordTables.available">has those tables.</template>
            <template v-else>does not have those tables yet, so Spire edits <code>ultimatedata/runeword_catalog.json</code>.</template>
            Output IDs of 0 mean “recipe known, live item not filled in”.
          </p>
          <div class="ui-toolbar">
            <button type="button" class="btn btn-sm" :class="rwPane === 'combos' ? 'btn-primary' : 'btn-dark'" @click="rwPane = 'combos'">Combos</button>
            <button type="button" class="btn btn-sm" :class="rwPane === 'items' ? 'btn-primary' : 'btn-dark'" @click="rwPane = 'items'">Runes / bases</button>
            <input v-model="search" type="search" class="form-control form-control-sm" style="width: 220px" placeholder="Family, name, id">
            <button type="button" class="btn btn-sm btn-dark" @click="newCombo">New combo</button>
            <button type="button" class="btn btn-sm btn-dark" @click="seedRunewords(true)">Preview seed</button>
            <button type="button" class="btn btn-sm btn-primary" @click="seedRunewords(false)">Write starter catalog</button>
          </div>
          <div v-if="rwPane === 'combos'" class="sys-split">
            <table class="eq-table eq-highlight-rows" style="width: 100%">
              <thead><tr><th>Family</th><th>Tier</th><th>Output</th><th>Base</th></tr></thead>
              <tbody>
              <tr v-for="row in filteredCombos" :key="row.comboKey" class="zc-row" @click="editCombo(row)">
                <td><b>{{ row.familyKey }}</b></td>
                <td>{{ row.qualityTier }}</td>
                <td>{{ row.outputName || row.outputItemId || "—" }}</td>
                <td class="tabular">{{ row.baseItemId }}</td>
              </tr>
              </tbody>
            </table>
            <div class="sys-form" v-if="comboForm.open">
              <div class="ui-field"><label>Combo key</label><input v-model="comboForm.combo_key" class="form-control form-control-sm" placeholder="annihilus.normal"></div>
              <div class="ui-field-grid">
                <div class="ui-field"><label>Family</label><input v-model="comboForm.family_key" class="form-control form-control-sm"></div>
                <div class="ui-field">
                  <label>Tier</label>
                  <select v-model="comboForm.quality_tier" class="form-control form-control-sm">
                    <option value="rusty">rusty</option>
                    <option value="normal">normal</option>
                    <option value="ultimate">ultimate</option>
                  </select>
                </div>
                <div class="ui-field"><label>Output item ID</label><input v-model.number="comboForm.output_item_id" type="number" class="form-control form-control-sm"></div>
                <div class="ui-field"><label>Base item ID</label><input v-model.number="comboForm.base_item_id" type="number" class="form-control form-control-sm"></div>
              </div>
              <div class="ui-field"><label>Output name</label><input v-model="comboForm.output_name" class="form-control form-control-sm"></div>
              <div class="ui-field-grid">
                <div class="ui-field"><label>Rune 1</label><input v-model.number="comboForm.rune1" type="number" class="form-control form-control-sm"></div>
                <div class="ui-field"><label>Rune 2</label><input v-model.number="comboForm.rune2" type="number" class="form-control form-control-sm"></div>
                <div class="ui-field"><label>Rune 3</label><input v-model.number="comboForm.rune3" type="number" class="form-control form-control-sm"></div>
                <div class="ui-field"><label>Rune 4</label><input v-model.number="comboForm.rune4" type="number" class="form-control form-control-sm"></div>
              </div>
              <div class="ui-toolbar">
                <button class="btn btn-sm btn-dark" :disabled="busy" @click="saveCombo(true)">Preview</button>
                <button class="btn btn-sm btn-primary" :disabled="busy" @click="saveCombo(false)">Write JSON</button>
                <button v-if="comboForm._existing" class="btn btn-sm btn-outline-danger" :disabled="busy" @click="deleteKind('runeword_combo', {combo_key: comboForm.combo_key})">Delete</button>
              </div>
            </div>
          </div>
          <table v-else class="eq-table" style="width: 100%">
            <thead><tr><th>Kind</th><th>Key</th><th>Name</th><th class="text-right">Item ID</th></tr></thead>
            <tbody>
            <tr v-for="row in filteredRwItems" :key="row.sourceKey">
              <td>{{ row.itemKind }}</td>
              <td><code>{{ row.sourceKey }}</code></td>
              <td>{{ row.name }}</td>
              <td class="text-right tabular">{{ row.targetItemId }}</td>
            </tr>
            </tbody>
          </table>
        </eq-window>
      </div>
    </eq-window>
  </content-area>
</template>

<script>
import EqWindow from "../../components/eq-ui/EQWindow"
import ContentArea from "../../components/layout/ContentArea"
import {ROUTE} from "@/routes"
import {UltimateSystemsApi, classLabel} from "../../app/ultimate-systems"

export default {
  name: "UltimateSystems",
  components: {ContentArea, EqWindow},
  data() {
    return {
      ROUTE,
      loaded: false,
      busy: false,
      error: "",
      plan: null,
      tab: "talents",
      rwPane: "combos",
      search: "",
      classFilter: 0,
      status: {},
      catalog: {talents: [], trees: [], pools: [], ranks: [], unlocks: [], specs: []},
      traits: {items: []},
      runewords: {items: [], combos: [], families: [], exists: false, seeded: true},
      talentForm: {},
      rankForm: {},
      unlockForm: {},
      specForm: {},
      traitForm: {},
      comboForm: {},
      tabs: [
        {id: "talents", label: "Talents"},
        {id: "ranks", label: "Ranks"},
        {id: "unlocks", label: "Unlocks"},
        {id: "specs", label: "Specializations"},
        {id: "traits", label: "Traits"},
        {id: "runewords", label: "Runewords"},
      ],
    }
  },
  computed: {
    filteredTalents() {
      return (this.catalog.talents || []).filter((row) => {
        if (this.classFilter && (row.classIds || []).indexOf(this.classFilter) < 0) return false
        return this.matches(row.displayName + " " + row.talentId + " " + row.treeId)
      })
    },
    filteredRanks() {
      return (this.catalog.ranks || []).filter((row) => this.matches(row.key + " " + row.classFamily + " " + row.bucketSuffix))
    },
    filteredUnlocks() {
      return (this.catalog.unlocks || []).filter((row) => this.matches(row.key + " " + row.spellName + " " + row.spellId))
    },
    filteredSpecs() {
      return (this.catalog.specs || []).filter((row) => {
        if (this.classFilter && row.classId !== this.classFilter) return false
        return this.matches(row.name + " " + row.description)
      })
    },
    filteredTraits() {
      return (this.traits.items || []).filter((row) => this.matches(row.itemName + " " + row.itemId + " " + row.zoneShortName + " " + row.classes))
    },
    filteredCombos() {
      return (this.runewords.combos || []).filter((row) => this.matches(row.familyKey + " " + row.outputName + " " + row.comboKey + " " + row.outputItemId))
    },
    filteredRwItems() {
      return (this.runewords.items || []).filter((row) => this.matches(row.name + " " + row.sourceKey + " " + row.targetItemId))
    },
  },
  async created() {
    await this.reload()
  },
  methods: {
    className: classLabel,
    classList(ids) {
      return (ids || []).map((id) => classLabel(id)).join(", ")
    },
    matches(text) {
      const q = String(this.search || "").toLowerCase().trim()
      if (!q) return true
      return String(text || "").toLowerCase().indexOf(q) >= 0
    },
    treesForClass(classId) {
      return (this.catalog.trees || []).filter((t) => !classId || t.classId === classId)
    },
    async reload() {
      this.error = ""
      try {
        this.status = await UltimateSystemsApi.status()
        this.catalog = await UltimateSystemsApi.catalog()
        try { this.traits = await UltimateSystemsApi.traits() } catch (e) { this.traits = {items: []} }
        this.runewords = await UltimateSystemsApi.runewords()
        this.loaded = true
      } catch (e) {
        const status = e.response && e.response.status
        this.loaded = true
        if (status === 404) {
          this.error = "Ultimate systems routes are not on the running backend yet. Restart Spire with start_spire_dev.bat."
        } else {
          this.error = (e.response && e.response.data && e.response.data.error) || String(e)
        }
      }
    },
    newTalent() {
      this.talentForm = {
        open: true,
        talent_id: "",
        display_name: "",
        description: "",
        class_id: this.classFilter || 1,
        tree_id: "",
        point_pool_id: "demigod_talent",
        rank_cap: 1,
        cost_amount: 1,
        progression_phase: "demigod",
        required_talents: "",
        hidden: false,
        create_rank: true,
      }
    },
    editTalent(row) {
      this.talentForm = {
        open: true,
        _existing: true,
        talent_id: row.talentId,
        display_name: row.displayName,
        description: row.description,
        class_id: (row.classIds && row.classIds[0]) || 1,
        tree_id: row.treeId,
        point_pool_id: row.pointPoolId,
        rank_cap: row.rankCap,
        cost_amount: row.costAmount || 1,
        progression_phase: row.progressionPhase || "demigod",
        required_talents: (row.requiredTalents || []).join(", "),
        hidden: !!row.hidden,
        create_rank: false,
      }
    },
    newRank() {
      this.rankForm = {open: true, key: "", class_family: "", bucket_suffix: ""}
    },
    editRank(row) {
      this.rankForm = {
        open: true,
        _existing: true,
        key: row.key,
        class_family: row.classFamily,
        bucket_suffix: row.bucketSuffix,
      }
    },
    newUnlock() {
      this.unlockForm = {open: true, key: "Has_Turned_In_", spell_name: "", spell_id: 0, expected_value: "1"}
    },
    editUnlock(row) {
      this.unlockForm = {
        open: true,
        _existing: true,
        key: row.key,
        spell_name: row.spellName,
        spell_id: row.spellId,
        expected_value: row.expectedValue || "1",
      }
    },
    newSpec() {
      this.specForm = {open: true, name: "", class_id: this.classFilter || 1, description: "", cost: 1, rankcount: 1}
    },
    editSpec(row) {
      this.specForm = {
        open: true,
        _existing: true,
        name: row.name,
        class_id: row.classId,
        description: row.description,
        cost: row.cost,
        rankcount: row.rankCount,
      }
    },
    newTrait() {
      this.traitForm = {open: true, item_id: 0, item_name: "", zone_id: 0, zone_short_name: "", tier: "Demigod", classes: "", proc_spell: 0, focus_spell: 0, click_spell: 0, worn_spell: 0}
    },
    editTrait(row) {
      this.traitForm = {
        open: true,
        _existing: true,
        item_id: row.itemId,
        item_name: row.itemName,
        zone_id: row.zoneId,
        zone_short_name: row.zoneShortName,
        tier: row.tier,
        classes: row.classes,
        proc_spell: row.procSpell,
        focus_spell: row.focusSpell,
        click_spell: row.clickSpell,
        worn_spell: row.wornSpell,
      }
    },
    newCombo() {
      this.comboForm = {open: true, combo_key: "", family_key: "", quality_tier: "normal", output_item_id: 0, output_name: "", base_item_id: 0, rune1: 899300, rune2: 899301, rune3: 899304, rune4: 899305}
    },
    editCombo(row) {
      this.comboForm = {
        open: true,
        _existing: true,
        combo_key: row.comboKey,
        family_key: row.familyKey,
        quality_tier: row.qualityTier,
        output_item_id: row.outputItemId,
        output_name: row.outputName,
        base_item_id: row.baseItemId,
        rune1: row.rune1,
        rune2: row.rune2,
        rune3: row.rune3,
        rune4: row.rune4,
      }
    },
    async saveTalent(dryRun) {
      const entry = {
        talent_id: this.talentForm.talent_id,
        display_name: this.talentForm.display_name,
        description: this.talentForm.description,
        class_ids: [this.talentForm.class_id],
        tree_id: this.talentForm.tree_id,
        point_pool_id: this.talentForm.point_pool_id,
        rank_cap: this.talentForm.rank_cap,
        cost_amount: this.talentForm.cost_amount,
        progression_phase: this.talentForm.progression_phase,
        required_talents: this.talentForm.required_talents,
        hidden: this.talentForm.hidden,
      }
      await this.runWrite("talent", "upsert", entry, dryRun)
      if (!dryRun && this.plan && this.plan.ok && this.talentForm.create_rank) {
        await this.runWrite("rank", "upsert", {key: this.talentForm.talent_id, class_family: String(this.talentForm.talent_id || "").split("_")[0]}, false)
      }
    },
    saveRank(dryRun) {
      return this.runWrite("rank", "upsert", {
        key: this.rankForm.key,
        class_family: this.rankForm.class_family,
        bucket_suffix: this.rankForm.bucket_suffix,
      }, dryRun)
    },
    saveUnlock(dryRun) {
      return this.runWrite("unlock", "upsert", this.unlockForm, dryRun)
    },
    saveSpec(dryRun) {
      return this.runWrite("spec", "upsert", this.specForm, dryRun)
    },
    saveTrait(dryRun) {
      return this.runWrite("trait", "upsert", this.traitForm, dryRun)
    },
    saveCombo(dryRun) {
      return this.runWrite("runeword_combo", "upsert", this.comboForm, dryRun)
    },
    seedRunewords(dryRun) {
      return this.runWrite("runeword_seed", "seed", {overwrite: false}, dryRun)
    },
    deleteKind(kind, entry) {
      if (!window.confirm("Delete this " + kind + "?")) return
      return this.runWrite(kind, "delete", entry, false)
    },
    async runWrite(kind, action, entry, dryRun) {
      this.busy = true
      this.error = ""
      this.plan = null
      try {
        this.plan = await UltimateSystemsApi.write({kind, action, entry, dryRun})
        if (this.plan && this.plan.error) {
          this.error = this.plan.error
        } else if (this.plan && !this.plan.dryRun) {
          await this.reload()
        }
      } catch (e) {
        const status = e.response && e.response.status
        if (status === 404) {
          this.error = "Ultimate systems write route is not on the running backend yet. Restart Spire with start_spire_dev.bat."
        } else {
          this.error = (e.response && e.response.data && e.response.data.error) || String(e)
        }
        if (e.response && e.response.data && e.response.data.relPath) {
          this.plan = e.response.data
        }
      }
      this.busy = false
    },
  },
}
</script>

<style scoped>
.zc-copy {
  color: var(--text-muted, #b7c0cc);
  margin-bottom: 12px;
}
.zc-dot {
  margin: 0 6px;
  color: var(--text-subtle);
}
.zc-row { cursor: pointer; }
.ml-auto { margin-left: auto; }
.sys-split {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(280px, 0.9fr);
  gap: 16px;
  align-items: start;
}
.sys-form {
  border: 1px solid var(--border, #3a4350);
  background: var(--surface-2, #12161d);
  border-radius: 8px;
  padding: 12px;
}
.zc-checks {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 18px;
  margin: 12px 0;
}
@media (max-width: 980px) {
  .sys-split { grid-template-columns: 1fr; }
}
</style>

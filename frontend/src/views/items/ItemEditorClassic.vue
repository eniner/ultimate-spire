<template>
  <content-area>
    <eq-window>
      <div class="d-flex align-items-center mb-3" v-if="item">
        <h4 class="mb-0 mr-3">
          {{ editing ? 'Edit' : 'View' }} Item - {{ item.id }} {{ item.name }}
          (<a :href="'https://lucy.allakhazam.com/item.html?id=' + item.id" target="_blank">Lucy</a>)
        </h4>
        <div class="ml-auto">
          <b-button size="sm" :variant="editing ? 'outline-warning' : 'warning'" @click="editing = false">
            <i class="fa fa-eye"></i> View
          </b-button>
          <b-button size="sm" :variant="editing ? 'warning' : 'outline-warning'" class="ml-1" @click="editing = true">
            <i class="fa fa-edit"></i> Edit
          </b-button>
          <router-link :to="spireEditorRoute" class="btn btn-sm btn-outline-light ml-3">
            Spire Editor
          </router-link>
        </div>
      </div>

      <div v-if="notification" class="text-center eq-header fade-in mb-2" @click="notification = ''">
        {{ notification }}
      </div>
      <b-alert show dismissable variant="danger" v-if="error">
        <i class="fa fa-warning"></i> {{ error }}
      </b-alert>

      <app-loader :is-loading="!item" class="mt-3 mb-3"/>

      <div v-if="item" class="classic-item-form">
        <fieldset v-for="section in sections" :key="section.title" class="classic-fieldset">
          <legend>{{ section.title }}</legend>

          <div v-for="(group, gi) in (section.groups || [section])" :key="gi">
            <div v-if="section.groups && group.title" class="classic-subtitle">{{ group.title }}</div>

            <div class="row">
              <div
                v-for="field in group.fields"
                :key="field.f"
                :class="field.wide ? 'col-lg-4 col-md-6' : 'col-lg-2 col-md-3 col-sm-4'"
                class="mb-2"
              >
                <label :for="'classic-' + field.f" class="classic-label">{{ field.l }}:</label>

                <div v-if="!editing" class="classic-value">{{ displayValue(field) }}</div>

                <select
                  v-else-if="field.o"
                  :id="'classic-' + field.f"
                  class="form-control form-control-sm"
                  v-model.number="item[field.f]"
                >
                  <option v-for="opt in optionsFor(field)" :key="opt.value" :value="opt.value">
                    {{ opt.text }}
                  </option>
                </select>

                <input
                  v-else-if="field.text"
                  :id="'classic-' + field.f"
                  type="text"
                  class="form-control form-control-sm"
                  :readonly="field.readonly"
                  v-model="item[field.f]"
                >

                <input
                  v-else
                  :id="'classic-' + field.f"
                  type="number"
                  class="form-control form-control-sm"
                  v-model.number="item[field.f]"
                >
              </div>
            </div>

            <div v-if="section.evolving" class="mt-2">
              <div v-if="!item.evoid" class="text-muted">Not part of an evolving chain (Evolution ID is 0).</div>
              <div v-else-if="!evolvingRows.length" class="text-danger">
                No rows in items_evolving_details for Evolution ID {{ item.evoid }}.
              </div>
              <table v-else class="eq-table bordered" style="width: 100%">
                <thead>
                <tr>
                  <th>Evolution ID</th>
                  <th>Level</th>
                  <th>Item</th>
                  <th>Type</th>
                  <th>Subtype</th>
                  <th>Required Amt</th>
                </tr>
                </thead>
                <tbody>
                <tr v-for="d in evolvingRows" :key="d.id" :class="d.item_id === item.id ? 'font-weight-bold text-warning' : ''">
                  <td>{{ d.item_evo_id }}</td>
                  <td>{{ d.item_evolve_level }}<span v-if="d.item_id === item.id"> (this item)</span></td>
                  <td>
                    <router-link :to="classicRoute(d.item_id)">{{ d.item_id }} - {{ evolvingNames[d.item_id] || '?' }}</router-link>
                  </td>
                  <td>{{ evolvingTypes[d.type] || 'UNK' }} ({{ d.type }})</td>
                  <td>{{ describeSubType(d) }} ({{ d.sub_type }})</td>
                  <td>{{ d.required_amount }}</td>
                </tr>
                </tbody>
              </table>
              <div v-if="evolvingProblems.length" class="mt-2">
                <div v-for="p in evolvingProblems" :key="p" class="text-danger">
                  <i class="fa fa-exclamation-triangle"></i> {{ p }}
                </div>
              </div>
              <router-link v-if="item.evoid" :to="chainRoute(item.evoid)" class="btn btn-sm btn-outline-warning mt-2">
                <i class="ra ra-cycle"></i> Open chain editor
              </router-link>
            </div>

            <div v-for="mask in (group.masks || [])" :key="mask.f" class="mt-2 mb-3">
              <div class="classic-label">{{ mask.l }} ({{ item[mask.f] }}):</div>

              <div v-if="!editing" class="classic-value">
                <template v-if="allBitsSet(mask)">All</template>
                <template v-else-if="!item[mask.f]">{{ mask.zero || "None" }}</template>
                <span
                  v-else
                  v-for="[bit, name] in selectedMaskBits(mask)"
                  :key="bit"
                  class="classic-mask-view"
                >
                  <span v-if="maskIcon(mask.f, bit)" :class="'item-' + maskIcon(mask.f, bit) + '-sm'"></span>
                  {{ name }}
                </span>
              </div>

              <div v-else class="classic-mask">
                <label v-for="[bit, name] in mask.bits" :key="bit" class="classic-mask-option">
                  <input type="checkbox" :checked="(item[mask.f] & bit) !== 0" @change="toggleBit(mask.f, bit)">
                  <span v-if="maskIcon(mask.f, bit)" :class="'item-' + maskIcon(mask.f, bit) + '-sm'"></span>
                  {{ name }}
                </label>
                <label class="classic-mask-option font-weight-bold">
                  <input type="checkbox" :checked="allBitsSet(mask)" @change="toggleAll(mask)"> All/None
                </label>
              </div>
            </div>
          </div>
        </fieldset>

        <div class="text-center mt-3" v-if="editing">
          <b-button variant="warning" @click="save()">
            <i class="fa fa-save"></i> Submit Changes
          </b-button>
        </div>
      </div>
    </eq-window>
  </content-area>
</template>

<script>
import * as util                                    from "util";
import EqWindow                                     from "../../components/eq-ui/EQWindow";
import ContentArea                                  from "../../components/layout/ContentArea";
import {SpireApi}                                   from "../../app/api/spire-api";
import {SpireQueryBuilder}                          from "../../app/api/spire-query-builder";
import {FactionListApi, ItemApi}                    from "../../app/api";
import {Items}                                      from "../../app/items";
import {ROUTE}                                      from "../../routes";
import {
  DB_BAG_TYPES,
  DB_ITEM_AUG_RESTRICT,
  DB_ITEM_CLASS,
  DB_ITEM_MATERIAL,
  DB_ITEM_TYPES
}                                                   from "../../app/constants/eq-item-constants";
import {DB_RACE_NAMES}                              from "../../app/constants/eq-races-constants";
import {BODYTYPES}                                  from "../../app/constants/eq-bodytype-constants";
import {DB_SKILLS}                                  from "../../app/constants/eq-skill-constants";
import {DB_ITEM_BARD_TYPE}                          from "../../app/constants/eq-bard-types";
import {EVOLVING_TYPES, EvolvingItems}              from "../../app/evolving-items";
import {maskFieldIcon}                              from "../../app/eq-chip-icons";
import {
  CLASSIC_AUG_TYPE_BITS,
  CLASSIC_CLASS_BITS,
  CLASSIC_DEITY_BITS,
  CLASSIC_INVERTED_NO_YES,
  CLASSIC_ITEM_BAG_SIZE,
  CLASSIC_ITEM_CLICK_TYPE,
  CLASSIC_ITEM_FOCUS_TYPE,
  CLASSIC_ITEM_LDON_THEME,
  CLASSIC_ITEM_POINT_TYPE,
  CLASSIC_ITEM_PROC_TYPE,
  CLASSIC_ITEM_SCROLL_TYPE,
  CLASSIC_ITEM_SIZE,
  CLASSIC_ITEM_WORN_TYPE,
  CLASSIC_NO_YES,
  CLASSIC_RACE_BITS,
  CLASSIC_SLOT_BITS
}                                                   from "../../app/constants/eq-item-classic-constants";

const n  = (f, l, extra = {}) => ({f, l, ...extra})
const t  = (f, l, extra = {}) => ({f, l, text: true, ...extra})
const s  = (f, l, o, extra = {}) => ({f, l, o, ...extra})
const yn = (f, l) => s(f, l, CLASSIC_NO_YES)

const effectRow = (prefix, label, typeOptions, extra = []) => [
  s(prefix + "type", label + " Type", typeOptions, {wide: true}),
  n(prefix + "effect", label + " Effect"),
  n(prefix + "level", label + " Level"),
  n(prefix + "level_2", label + " Level 2"),
  ...extra,
  t(prefix + "name", label + " Name", {wide: true}),
]

export default {
  name: "ItemEditorClassic",
  components: {ContentArea, EqWindow},
  data() {
    return {
      item: null,
      editing: this.$route.query.mode === "edit",
      factions: {0: "None"},
      evolvingTypes: EVOLVING_TYPES,
      evolvingRows: [],
      evolvingNames: {},
      evolvingProblems: [],
      error: "",
      notification: "",
    }
  },
  computed: {
    spireEditorRoute() {
      return util.format(ROUTE.ITEM_EDIT, this.$route.params.id)
    },
    sections() {
      return [
        {
          title: "General",
          fields: [
            t("name", "Name", {wide: true}),
            s("itemtype", "Type", DB_ITEM_TYPES, {wide: true}),
            s("itemclass", "Class", DB_ITEM_CLASS),
            t("lore", "Lore Name", {wide: true}),
            yn("stackable", "Stackable"),
            n("stacksize", "Stacksize"),
            n("maxcharges", "Charges"),
            t("filename", "File Name"),
            s("book", "Book", {0: "No", 1: "Yes", 2: "Message"}),
            n("booktype", "Booktype"),
            t("charmfile", "Charmfile"),
            t("charmfileid", "Charmfile ID"),
            n("scriptfileid", "Script File ID"),
            n("powersourcecapacity", "Power Source Capacity"),
            n("potionbeltslots", "Potion Belt Slots"),
            n("subtype", "Subtype"),
            s("bagsize", "Bag Size", CLASSIC_ITEM_BAG_SIZE),
            n("bagslots", "Bag Slots"),
            n("bagwr", "Bag Weight Reduction"),
            s("bagtype", "Bag Type", DB_BAG_TYPES, {wide: true}),
            yn("heirloom", "Heirloom"),
            yn("placeable", "Placeable"),
            yn("epicitem", "Epic Item"),
          ],
        },
        {
          title: "Restrictions",
          fields: [
            s("nodrop", "No Drop", CLASSIC_INVERTED_NO_YES),
            s("norent", "No Rent", {...CLASSIC_INVERTED_NO_YES, 255: "255"}),
            yn("magic", "Magic"),
            yn("tradeskills", "Tradeskill"),
            yn("artifactflag", "Artifact"),
            yn("questitemflag", "Quest"),
            yn("attuneable", "Attuneable"),
            yn("nopet", "No Pet"),
            yn("fvnodrop", "FV No Drop"),
            yn("notransfer", "No Transfer"),
            yn("potionbelt", "Potion Belt"),
            s("benefitflag", "Benefit", {0: "No", 1: "Yes", 3: "3"}),
            yn("expendablearrow", "Expendable Arrow"),
            n("loregroup", "Lore"),
            n("reqlevel", "Req Level"),
            n("reclevel", "Rec Level"),
            n("recskill", "Rec Skill"),
          ],
          masks: [
            {f: "slots", l: "Slots", bits: CLASSIC_SLOT_BITS},
            {f: "races", l: "Races", bits: CLASSIC_RACE_BITS},
            {f: "classes", l: "Classes", bits: CLASSIC_CLASS_BITS},
            {f: "deity", l: "Deities", bits: CLASSIC_DEITY_BITS, zero: "None (usable by any deity)"},
          ],
        },
        {
          title: "Stats",
          groups: [
            {
              title: "Damage",
              fields: [
                n("damage", "Damage"),
                n("delay", "Delay"),
                n("range", "Range"),
                n("banedmgamt", "Bane Dmg"),
                n("banedmgraceamt", "Bane Race Dmg"),
                s("banedmgrace", "Bane Race", DB_RACE_NAMES, {wide: true}),
                s("banedmgbody", "Bane Bodytype", BODYTYPES, {wide: true}),
                s("extradmgskill", "Extra Dmg Skill", DB_SKILLS, {wide: true}),
                n("extradmgamt", "Extra Dmg Amt"),
                n("elemdmgtype", "Elem Dmg Type"),
                n("elemdmgamt", "Elem Dmg Amt"),
                n("spelldmg", "Spell Dmg"),
                n("backstabdmg", "Backstab Dmg"),
              ],
            },
            {
              title: "Base Stats",
              fields: [
                n("hp", "HP"), n("mana", "Mana"), n("endur", "Endurance"), n("ac", "AC"),
                n("accuracy", "Accuracy"), n("attack", "Attack"), n("regen", "HP Regen"),
                n("manaregen", "Mana Regen"), n("enduranceregen", "End Regen"), n("haste", "Haste"),
                n("light", "Light"),
              ],
            },
            {
              title: "Stats",
              fields: [
                n("aagi", "AGI"), n("acha", "CHA"), n("adex", "DEX"), n("aint", "INT"),
                n("asta", "STA"), n("astr", "STR"), n("awis", "WIS"),
                n("cr", "Cold"), n("dr", "Disease"), n("fr", "Fire"), n("mr", "Magic"),
                n("pr", "Poison"), n("svcorruption", "Corruption"), n("stunresist", "Stun"),
              ],
            },
            {
              title: "Heroic Stats",
              fields: [
                n("heroic_agi", "Heroic AGI"), n("heroic_cha", "Heroic CHA"), n("heroic_dex", "Heroic DEX"),
                n("heroic_int", "Heroic INT"), n("heroic_sta", "Heroic STA"), n("heroic_str", "Heroic STR"),
                n("heroic_wis", "Heroic WIS"), n("heroic_cr", "Heroic Cold"), n("heroic_dr", "Heroic Disease"),
                n("heroic_fr", "Heroic Fire"), n("heroic_mr", "Heroic Magic"), n("heroic_pr", "Heroic Poison"),
                n("heroic_svcorrup", "Heroic Corruption"),
              ],
            },
            {
              title: "Spell/Ability Stats",
              fields: [
                n("damageshield", "DMG Shield"), n("dotshielding", "DOT Shielding"), n("shielding", "Shielding"),
                n("spellshield", "Spell Shield"), n("strikethrough", "Strikethrough"),
                n("combateffects", "Combat Effects"), n("avoidance", "Avoidance"),
                n("dsmitigation", "Dmg Shield Mit"), n("healamt", "Heal Amt"),
                n("clairvoyance", "Clairvoyance"), n("purity", "Purity"),
              ],
            },
            {
              title: "Skill Stats",
              fields: [
                s("skillmodtype", "Skill Mod", DB_SKILLS, {wide: true}),
                n("skillmodvalue", "Skill Mod Value"),
                n("skillmodmax", "Skill Mod Max"),
              ],
            },
          ],
        },
        {
          title: "Evolving Item Details",
          evolving: true,
          fields: [
            yn("evoitem", "Evolving"),
            n("evoid", "Evolution ID"),
            n("evolvinglevel", "Level"),
            n("evomax", "Max Level"),
          ],
        },
        {
          title: "Costs",
          fields: [
            n("price", "Price"), n("sellrate", "Sellrate"), n("favor", "Favor"), n("guildfavor", "Guild Favor"),
            n("ldonprice", "LDoN Price"), n("ldonsellbackrate", "LDoN Sellback"),
            yn("ldonsold", "LDoN Sold"),
            s("ldontheme", "LDoN Theme", CLASSIC_ITEM_LDON_THEME),
            s("pointtype", "Point Type", CLASSIC_ITEM_POINT_TYPE),
          ],
        },
        {
          title: "Appearance",
          fields: [
            n("icon", "Icon"), t("idfile", "IDFile"), n("weight", "Weight"), n("color", "Color"),
            s("size", "Size", CLASSIC_ITEM_SIZE),
            s("material", "Material", DB_ITEM_MATERIAL, {wide: true}),
            n("elitematerial", "Elite Material"),
          ],
        },
        {
          title: "Spells",
          groups: [
            {
              fields: [
                n("casttime", "Casttime"), n("casttime_", "Casttime_"),
                n("recastdelay", "Recast Delay"), n("recasttype", "Recast Type"),
              ],
            },
            {fields: effectRow("click", "Click", CLASSIC_ITEM_CLICK_TYPE)},
            {fields: effectRow("proc", "Proc", CLASSIC_ITEM_PROC_TYPE, [n("procrate", "Proc Rate")])},
            {fields: effectRow("worn", "Worn", CLASSIC_ITEM_WORN_TYPE)},
            {fields: effectRow("focus", "Focus", CLASSIC_ITEM_FOCUS_TYPE)},
            {fields: effectRow("scroll", "Scroll", CLASSIC_ITEM_SCROLL_TYPE)},
            {fields: effectRow("bard", "Bard", DB_ITEM_BARD_TYPE, [n("bardvalue", "Bard Value")])},
          ],
        },
        {
          title: "Augment",
          fields: [
            ...[1, 2, 3, 4, 5].flatMap((i) => [
              n("augslot_" + i + "_type", "Slot " + i + " Type"),
              yn("augslot_" + i + "_visible", "Slot " + i + " Visible"),
            ]),
            s("augrestrict", "Augment Restrictions", DB_ITEM_AUG_RESTRICT, {wide: true}),
            n("augdistiller", "Augment Distiller"),
          ],
          masks: [
            {f: "augtype", l: "Type", bits: CLASSIC_AUG_TYPE_BITS},
          ],
        },
        {
          title: "Faction",
          fields: [1, 2, 3, 4].flatMap((i) => [
            s("factionmod_" + i, "Faction Mod " + i, this.factions, {wide: true}),
            n("factionamt_" + i, "Amt " + i),
          ]),
        },
        {
          title: "Verification",
          fields: [
            t("created", "Created", {readonly: true}),
            t("verified", "Verified", {readonly: true}),
            t("updated", "Updated", {readonly: true}),
            t("source", "Source"),
            t("comment", "Comment", {wide: true}),
          ],
        },
      ]
    },
  },
  watch: {
    "$route.params.id"() {
      this.load()
    },
  },
  mounted() {
    this.load()
    this.loadFactions()
  },
  methods: {
    async load() {
      this.item  = null
      this.error = ""
      try {
        const r   = await (new ItemApi(...SpireApi.cfg())).getItem({id: this.$route.params.id})
        this.item = r.data
      } catch (e) {
        this.error = (e.response && e.response.data && e.response.data.error) || "Failed to load item"
        return
      }
      this.loadEvolving()
    },

    async loadEvolving() {
      this.evolvingRows     = []
      this.evolvingProblems = []
      if (!this.item.evoid) {
        return
      }
      try {
        const all   = await EvolvingItems.listDetails()
        const rows  = all
          .filter((d) => d.item_evo_id === this.item.evoid)
          .sort((a, b) => a.item_evolve_level - b.item_evolve_level)
        const items = await EvolvingItems.loadItemsFor(rows)

        const names = {}
        Object.values(items).forEach((i) => names[i.id] = i.name)
        this.evolvingNames = names
        this.evolvingRows  = rows
        if (rows.length) {
          this.evolvingProblems = EvolvingItems.analyzeChain(rows, items, all).problems
        }
      } catch (e) {
        console.log("[ItemEditorClassic] failed to load evolving details", e)
      }
    },

    describeSubType(d) {
      return EvolvingItems.describeSubType(d.type, d.sub_type)
    },

    classicRoute(id) {
      return util.format(ROUTE.ITEM_EDIT_CLASSIC, id)
    },

    chainRoute(evoId) {
      return util.format(ROUTE.ITEM_EVOLVING_CHAIN, evoId)
    },

    async loadFactions() {
      const request = (new SpireQueryBuilder()).select(["id", "name"]).limit(100000).get()
      try {
        const r        = await (new FactionListApi(...SpireApi.cfg())).listFactionLists(request)
        const factions = {0: "None"}
        r.data.forEach((f) => {
          factions[f.id] = f.name
        })
        this.factions = factions
      } catch (e) {
        console.log("[ItemEditorClassic] failed to load factions", e)
      }
    },

    optionsFor(field) {
      const opts = Object.entries(field.o).map(([k, v]) => ({value: Number(k), text: k + ": " + v}))
      const cur  = this.item[field.f]
      if (!opts.some((o) => o.value === cur)) {
        opts.unshift({value: cur, text: cur + ": (unknown)"})
      }
      return opts
    },

    displayValue(field) {
      const v = this.item[field.f]
      if (field.o) {
        return v in field.o ? v + ": " + field.o[v] : v + ": (unknown)"
      }
      return v === null || v === undefined || v === "" ? "-" : v
    },

    maskIcon(field, bit) {
      return maskFieldIcon(field, bit)
    },
    selectedMaskBits(mask) {
      const v = this.item[mask.f]
      if (!v || this.allBitsSet(mask)) {
        return []
      }
      return mask.bits.filter(([bit]) => (v & bit) !== 0)
    },
    maskSummary(mask) {
      const v = this.item[mask.f]
      if (!v) {
        return mask.zero || "None"
      }
      if (this.allBitsSet(mask)) {
        return "All"
      }
      return mask.bits.filter(([bit]) => (v & bit) !== 0).map(([, name]) => name).join(", ")
    },

    allBitsSet(mask) {
      const all = mask.bits.reduce((sum, [bit]) => sum + bit, 0)
      return (this.item[mask.f] & all) === all
    },

    toggleBit(field, bit) {
      this.item[field] = this.item[field] ^ bit
    },

    toggleAll(mask) {
      this.item[mask.f] = this.allBitsSet(mask) ? 0 : mask.bits.reduce((sum, [bit]) => sum + bit, 0)
    },

    async save() {
      this.error        = ""
      this.notification = ""
      try {
        const r = await (new ItemApi(...SpireApi.cfg())).updateItem({id: this.item.id, item: this.item})
        if (r.status === 200) {
          Items.setItem(this.item.id, undefined)
          this.notification = "Item " + this.item.id + " saved"
          setTimeout(() => this.notification = "", 5000)
        }
      } catch (e) {
        this.error = (e.response && e.response.data && e.response.data.error) || "Failed to save item"
      }
    },
  },
}
</script>

<style scoped>
.classic-fieldset {
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 4px;
  padding: 10px 15px;
  margin-bottom: 15px;
}

.classic-fieldset legend {
  width: auto;
  padding: 0 8px;
  font-size: 1.2rem;
  font-weight: bold;
}

.classic-subtitle {
  font-weight: bold;
  margin: 8px 0 4px;
  opacity: 0.85;
}

.classic-label {
  margin-bottom: 2px;
  font-size: 0.85rem;
  opacity: 0.8;
}

.classic-value {
  font-weight: bold;
}

.classic-mask {
  display: flex;
  flex-wrap: wrap;
}

.classic-mask-option {
  width: 150px;
  margin-bottom: 2px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.classic-mask-view {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-right: 10px;
}

.classic-mask-option [class*="item-"],
.classic-mask-view [class*="item-"] {
  display: block !important;
  margin: 0 !important;
}
</style>

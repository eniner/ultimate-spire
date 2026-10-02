<template>
  <content-area>
    <eq-window :title="'Evolving Chain ' + evoId">
      <div class="d-flex align-items-center mb-3">
        <router-link :to="listRoute" class="btn btn-sm btn-outline-light mr-3">
          <i class="fa fa-arrow-left"></i> All Evolving Chains
        </router-link>
        <div class="mr-auto">
          Two tables, in order: define the chain, then match the item columns.
        </div>
        <b-button size="sm" variant="outline-warning" class="mr-2" @click="load()">
          <i class="fa fa-refresh"></i> Reload
        </b-button>
      </div>

      <ol class="evo-steps">
        <li><b>items_evolving_details</b> — level, item, type, subtype, required amount. Type 4 is zone-kill: zone IDs in subtype, kill count in required amount.</li>
        <li><b>items</b> — set <code>evoitem=1</code>, <code>evoid</code>, <code>evolvinglevel</code>, and <code>evomax</code> to match step 1.</li>
        <li>Restart zones. They load this table at boot.</li>
      </ol>

      <div v-if="notification" class="text-center eq-header fade-in mb-2" @click="notification = ''">
        {{ notification }}
      </div>
      <b-alert show dismissable variant="danger" v-if="error">
        <i class="fa fa-warning"></i> {{ error }}
      </b-alert>

      <app-loader :is-loading="!loaded" class="mt-3 mb-3"/>

      <div v-if="loaded">
        <div v-if="analysis.problems.length" class="mb-3">
          <div v-for="p in analysis.problems" :key="p" class="ui-problem">
            {{ split(p).title }}
            <small v-if="split(p).detail">{{ split(p).detail }}</small>
          </div>
        </div>
        <div v-else-if="rows.length" class="text-success mb-3">
          <i class="fa fa-check"></i> This chain is consistent: levels 1-{{ analysis.maxLevel }} are all present and every item matches.
        </div>

        <h5 class="evo-table-title">1. items_evolving_details</h5>
        <div class="text-muted mb-2">
          Source of truth for the chain. Type 4 subtype is zone IDs (comma or period separated). Required amount is the kill count.
        </div>
        <div class="evo-table-wrap">
        <table class="eq-table bordered" style="width: 100%">
          <thead>
          <tr>
            <th>id</th>
            <th>level</th>
            <th>item_id</th>
            <th>item</th>
            <th>type</th>
            <th>sub_type</th>
            <th>required_amount</th>
            <th></th>
          </tr>
          </thead>
          <tbody>
          <template v-for="r in rows">
            <tr :key="r.id" :class="analysis.rowIssues[r.id] ? 'evo-row-bad' : ''">
              <td class="tabular">{{ r.id }}</td>
              <td><input type="number" class="form-control form-control-sm" v-model.number="r.item_evolve_level"></td>
              <td><input type="number" class="form-control form-control-sm evo-id" v-model.number="r.item_id"></td>
              <td class="evo-item-cell">
                <span v-if="itemIcon(r.item_id)" :class="'item-' + itemIcon(r.item_id) + ' evo-item-icon'"></span>
                <router-link v-if="items[r.item_id]" :to="classicRoute(r.item_id)">{{ items[r.item_id].name }}</router-link>
                <span v-else class="text-danger">(item {{ r.item_id }} not found)</span>
              </td>
              <td>
                <select class="form-control form-control-sm evo-type" v-model.number="r.type">
                  <option v-for="(name, id) in types" :key="id" :value="Number(id)">{{ id }}: {{ name }}</option>
                </select>
              </td>
              <td>
                <input
                  type="text"
                  class="form-control form-control-sm evo-sub"
                  v-model="r.sub_type"
                  :placeholder="Number(r.type) === 4 ? 'zone ids, e.g. 36,186,103' : ''"
                >
                <div v-if="Number(r.type) !== 4" class="evo-sub-help">{{ describeSubType(r) }}</div>
                <div v-if="Number(r.type) === 4" class="evo-zones">
                  <span
                    v-for="z in zoneList(r.sub_type)"
                    :key="z.id"
                    class="evo-zone-pill"
                    :class="{ 'is-missing': !z.name }"
                  >{{ z.id }}{{ z.name ? ' ' + z.name : '' }}</span>
                </div>
              </td>
              <td>
                <input type="number" class="form-control form-control-sm" v-model.number="r.required_amount">
                <div class="evo-sub-help">{{ requiredLabel(r.type) }}</div>
              </td>
              <td class="text-nowrap">
                <b-button size="sm" variant="outline-success" title="Save row" :disabled="busy" @click="saveRow(r)">
                  <i class="fa fa-save"></i>
                </b-button>
                <b-button size="sm" variant="outline-danger" class="ml-1" title="Delete row" :disabled="busy" @click="deleteRow(r)">
                  <i class="fa fa-trash"></i>
                </b-button>
              </td>
            </tr>
            <tr v-if="analysis.rowIssues[r.id]" :key="'issues-' + r.id" class="evo-row-bad">
              <td colspan="8" class="text-danger pt-0">
                <span v-for="issue in analysis.rowIssues[r.id]" :key="issue" class="mr-3">
                  <i class="fa fa-exclamation-circle"></i> {{ issue }}
                </span>
              </td>
            </tr>
          </template>
          </tbody>
        </table>
        </div>

        <div v-if="gapSuggestions.length" class="mt-4">
          <h5 class="text-warning">Missing levels</h5>
          <div class="mb-2 text-muted">
            Suggested items come from items already flagged with evoid {{ evoId }} at that level, otherwise from the
            item ID that follows the chain's numbering (level 1 item ID + level - 1). Check the name before adding.
          </div>
          <table class="eq-table bordered" style="width: 100%">
            <thead>
            <tr>
              <th style="width: 70px">Level</th>
              <th style="width: 120px">Item ID</th>
              <th>Suggested item</th>
              <th style="width: 120px">Required Amt</th>
              <th style="width: 200px"></th>
            </tr>
            </thead>
            <tbody>
            <tr v-for="g in gapSuggestions" :key="g.level">
              <td><b>{{ g.level }}</b></td>
              <td><input type="number" class="form-control form-control-sm evo-id" v-model.number="g.item_id"></td>
              <td class="evo-item-cell">
                <span v-if="itemIcon(g.item_id)" :class="'item-' + itemIcon(g.item_id) + ' evo-item-icon'"></span>
                <span v-if="items[g.item_id]">
                  {{ items[g.item_id].name }}
                  <small class="text-muted">({{ g.reason }})</small>
                </span>
                <span v-else class="text-muted">No item found - enter an item ID</span>
              </td>
              <td><input type="number" class="form-control form-control-sm" v-model.number="g.required_amount"></td>
              <td>
                <b-button size="sm" variant="warning" :disabled="busy || !g.item_id" @click="addGap(g)">
                  <i class="fa fa-plus"></i> Add level {{ g.level }}
                </b-button>
              </td>
            </tr>
            </tbody>
          </table>
        </div>

        <div class="mt-4">
          <h5>Add a level</h5>
          <div class="d-flex align-items-end flex-wrap">
            <div class="mr-2 mb-2">
              Level
              <input type="number" class="form-control form-control-sm" style="width: 80px" v-model.number="newRow.item_evolve_level">
            </div>
            <div class="mr-2 mb-2">
              Item ID
              <input type="number" class="form-control form-control-sm" style="width: 120px" v-model.number="newRow.item_id">
            </div>
            <div class="mr-2 mb-2">
              Type
              <select class="form-control form-control-sm" v-model.number="newRow.type">
                <option v-for="(name, id) in types" :key="id" :value="Number(id)">{{ id }}: {{ name }}</option>
              </select>
            </div>
            <div class="mr-2 mb-2">
              {{ Number(newRow.type) === 4 ? "Zone IDs" : "Sub Type" }}
              <input
                type="text"
                class="form-control form-control-sm"
                style="width: 220px"
                v-model="newRow.sub_type"
                :placeholder="Number(newRow.type) === 4 ? '36,186,103' : ''"
              >
            </div>
            <div class="mr-2 mb-2">
              {{ requiredLabel(newRow.type) }}
              <input type="number" class="form-control form-control-sm" style="width: 130px" v-model.number="newRow.required_amount">
            </div>
            <b-button class="mb-2" size="sm" variant="warning" :disabled="busy || !newRow.item_id || !newRow.item_evolve_level" @click="addNewRow()">
              <i class="fa fa-plus"></i> Add
            </b-button>
          </div>
          <div v-if="Number(newRow.type) === 4" class="evo-zones mt-1">
            <span
              v-for="z in zoneList(newRow.sub_type)"
              :key="'new-' + z.id"
              class="evo-zone-pill"
              :class="{ 'is-missing': !z.name }"
            >{{ z.id }}{{ z.name ? ' ' + z.name : '' }}</span>
          </div>
        </div>

        <h5 class="evo-table-title mt-4">2. items evolving columns</h5>
        <div class="d-flex align-items-center mb-2">
          <div class="text-muted mr-auto">
            Match <code>evoitem</code>, <code>evoid</code>, <code>evolvinglevel</code>, and <code>evomax</code> to the table above. The server uses these to find the next item.
          </div>
          <b-button size="sm" variant="warning" :disabled="!rows.length || busy || !itemMismatchCount" @click="syncAllItems()">
            <i class="fa fa-magic"></i> Sync {{ itemMismatchCount || "all" }} item{{ itemMismatchCount === 1 ? "" : "s" }}
          </b-button>
        </div>
        <div class="evo-table-wrap">
        <table class="eq-table bordered" style="width: 100%">
          <thead>
          <tr>
            <th>item_id</th>
            <th>name</th>
            <th>evoitem</th>
            <th>evoid</th>
            <th>evolvinglevel</th>
            <th>evomax</th>
            <th>status</th>
            <th></th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="r in rows" :key="'item-' + r.id" :class="itemRowMismatched(r) ? 'evo-row-bad' : ''">
            <td class="tabular">{{ r.item_id }}</td>
            <td class="evo-item-cell">
              <span v-if="itemIcon(r.item_id)" :class="'item-' + itemIcon(r.item_id) + ' evo-item-icon'"></span>
              <router-link v-if="items[r.item_id]" :to="classicRoute(r.item_id)">{{ items[r.item_id].name }}</router-link>
              <span v-else class="text-danger">(missing)</span>
            </td>
            <td :class="cellClass(r, 'evoitem', 1)">
              {{ itemField(r, 'evoitem') }}
              <span v-if="itemField(r, 'evoitem') !== 1" class="text-muted">→ 1</span>
            </td>
            <td :class="cellClass(r, 'evoid', evoId)">
              {{ itemField(r, 'evoid') }}
              <span v-if="itemField(r, 'evoid') !== evoId" class="text-muted">→ {{ evoId }}</span>
            </td>
            <td :class="cellClass(r, 'evolvinglevel', r.item_evolve_level)">
              {{ itemField(r, 'evolvinglevel') }}
              <span v-if="itemField(r, 'evolvinglevel') !== r.item_evolve_level" class="text-muted">→ {{ r.item_evolve_level }}</span>
            </td>
            <td :class="cellClass(r, 'evomax', analysis.maxLevel)">
              {{ itemField(r, 'evomax') }}
              <span v-if="itemField(r, 'evomax') !== analysis.maxLevel" class="text-muted">→ {{ analysis.maxLevel }}</span>
            </td>
            <td>
              <span v-if="!items[r.item_id]" class="badge badge-danger">Missing item</span>
              <span v-else-if="itemRowMismatched(r)" class="badge badge-warning">Needs sync</span>
              <span v-else class="badge badge-success">Matched</span>
            </td>
            <td>
              <b-button
                size="sm"
                variant="outline-warning"
                :disabled="busy || !items[r.item_id] || !itemRowMismatched(r)"
                @click="syncItem(r).then(load)"
              >Sync</b-button>
            </td>
          </tr>
          </tbody>
        </table>
        </div>

        <div class="mt-4">
          <h5>Characters holding items from this chain (character_evolving_items)</h5>
          <div v-if="!characterRows.length" class="text-muted">No characters have items from this chain.</div>
          <table v-else class="eq-table bordered" style="width: 100%">
            <thead>
            <tr>
              <th>ID</th>
              <th>Character</th>
              <th>Item</th>
              <th>Current Amount</th>
              <th>Progression</th>
              <th>Equipped</th>
              <th>Activated</th>
              <th>Final Item ID</th>
              <th>Deleted At</th>
            </tr>
            </thead>
            <tbody>
            <tr v-for="c in characterRows" :key="c.id">
              <td>{{ c.id }}</td>
              <td>{{ characterNames[c.character_id] || c.character_id }}</td>
              <td class="evo-item-cell">
                <span v-if="itemIcon(c.item_id)" :class="'item-' + itemIcon(c.item_id) + ' evo-item-icon'"></span>
                {{ c.item_id }} - {{ items[c.item_id] ? items[c.item_id].name : '?' }}
              </td>
              <td>{{ c.current_amount }}</td>
              <td>{{ Number(c.progression || 0).toFixed(2) }}%</td>
              <td>{{ c.equipped }}</td>
              <td>{{ c.activated }}</td>
              <td :class="finalItemId && c.final_item_id !== finalItemId ? 'text-danger font-weight-bold' : ''">
                {{ c.final_item_id }}
                <span v-if="finalItemId && c.final_item_id !== finalItemId">(chain final is {{ finalItemId }})</span>
              </td>
              <td>{{ c.deleted_at || '' }}</td>
            </tr>
            </tbody>
          </table>
        </div>

        <div class="mt-3 text-muted">
          Zone servers load items_evolving_details when they boot, so restart zones after changing chains.
        </div>
      </div>
    </eq-window>
  </content-area>
</template>

<script>
import * as util         from "util";
import EqWindow          from "../../components/eq-ui/EQWindow";
import ContentArea       from "../../components/layout/ContentArea";
import {SpireApi}        from "../../app/api/spire-api";
import {ItemApi}           from "../../app/api";
import {CharacterDatumApi} from "../../app/api/api/character-datum-api";
import {Items}           from "../../app/items";
import {EVOLVING_TYPES, EvolvingItems} from "../../app/evolving-items";
import {Zones}           from "../../app/zones";
import {ROUTE}           from "../../routes";

export default {
  name: "EvolvingChainEditor",
  components: {ContentArea, EqWindow},
  data() {
    return {
      loaded: false,
      busy: false,
      error: "",
      notification: "",
      types: EVOLVING_TYPES,
      rows: [],
      allDetails: [],
      items: {},
      analysis: {problems: [], rowIssues: {}, missing: [], maxLevel: 0},
      gapSuggestions: [],
      characterRows: [],
      characterNames: {},
      zoneNames: {},
      newRow: this.blankRow(),
    }
  },
  computed: {
    evoId() {
      return Number(this.$route.params.evoId)
    },
    listRoute() {
      return ROUTE.ITEMS_EVOLVING
    },
    finalItemId() {
      const last = this.rows.find((r) => r.item_evolve_level === this.analysis.maxLevel)
      return last ? last.item_id : 0
    },
    itemMismatchCount() {
      return this.rows.filter((r) => this.items[r.item_id] && this.itemRowMismatched(r)).length
    },
  },
  watch: {
    "$route.params.evoId"() {
      this.load()
    },
  },
  mounted() {
    this.load()
  },
  methods: {
    blankRow() {
      return {item_evolve_level: null, item_id: null, type: 1, sub_type: "0", required_amount: 0}
    },

    async load() {
      this.loaded = false
      this.error  = ""
      try {
        this.rows = (await EvolvingItems.listDetails(this.evoId))
          .sort((a, b) => a.item_evolve_level - b.item_evolve_level || a.id - b.id)
        const itemIds = this.rows.map((r) => r.item_id)
        this.allDetails = await EvolvingItems.listDetailsForItems(itemIds)

        const items = Object.assign(
          {},
          await EvolvingItems.listFlaggedForChain(this.evoId),
          await EvolvingItems.loadItemsFor(this.rows.concat(this.allDetails)),
        )
        if (itemIds.length) {
          Object.assign(items, await EvolvingItems.loadItemRange(Math.min(...itemIds), Math.max(...itemIds) + 30))
        }
        this.items    = items
        this.analysis = this.rows.length
          ? EvolvingItems.analyzeChain(this.rows, this.items, this.allDetails)
          : {problems: ["No rows in items_evolving_details for chain " + this.evoId], rowIssues: {}, missing: [], maxLevel: 0}

        this.buildGapSuggestions()
        this.newRow = {...this.blankRow(), item_evolve_level: this.analysis.maxLevel + 1, ...this.templateFrom(this.rows[this.rows.length - 1])}
        await this.loadZones()
        await this.loadCharacters()
      } catch (e) {
        this.error = this.errorText(e)
      }
      this.loaded = true
    },

    templateFrom(row) {
      return row ? {type: row.type, sub_type: row.sub_type, required_amount: row.required_amount} : {}
    },

    buildGapSuggestions() {
      const first = this.rows.find((r) => r.item_evolve_level === 1)
      this.gapSuggestions = this.analysis.missing.map((level) => {
        const flagged = Object.values(this.items).find((i) => i.evoid === this.evoId && i.evolvinglevel === level)
        const byId    = first ? first.item_id + level - 1 : null
        const below   = [...this.rows].reverse().find((r) => r.item_evolve_level < level)

        let item_id = null
        let reason  = ""
        if (flagged) {
          item_id = flagged.id
          reason  = "already flagged evoid " + this.evoId + " level " + level
        } else if (byId && this.items[byId]) {
          item_id = byId
          reason  = "next ID in sequence"
        }

        return {level, item_id, reason, ...this.templateFrom(below)}
      })
    },

    async loadCharacters() {
      const ids          = this.rows.map((r) => r.item_id)
      this.characterRows = await EvolvingItems.listCharacterItems(ids)

      const api = new CharacterDatumApi(...SpireApi.cfg())
      for (const cid of new Set(this.characterRows.map((c) => c.character_id))) {
        if (this.characterNames[cid]) {
          continue
        }
        try {
          const r = await api.getCharacterDatum({id: cid})
          this.$set(this.characterNames, cid, r.data.name)
        } catch (e) {
          this.$set(this.characterNames, cid, "(deleted character " + cid + ")")
        }
      }
    },

    async loadZones() {
      const zones = await Zones.getZones()
      const names = {}
      ;(zones || []).forEach((z) => {
        if (z && z.zoneidnumber && !names[z.zoneidnumber]) {
          names[z.zoneidnumber] = z.long_name || z.short_name
        }
      })
      this.zoneNames = names
    },

    zoneList(subType) {
      return EvolvingItems.parseIdList(subType).map((id) => ({
        id,
        name: this.zoneNames[id] || "",
      }))
    },

    requiredLabel(type) {
      return EvolvingItems.requiredLabel(type)
    },

    describeSubType(r) {
      return EvolvingItems.describeSubType(r.type, r.sub_type, this.zoneNames)
    },

    itemRowMismatched(r) {
      const item = this.items[r.item_id]
      if (!item) {
        return true
      }
      return item.evoitem !== 1
        || item.evoid !== this.evoId
        || item.evolvinglevel !== r.item_evolve_level
        || item.evomax !== this.analysis.maxLevel
    },

    split(p) {
      return EvolvingItems.splitProblem(p)
    },

    itemField(r, field) {
      return this.items[r.item_id] ? this.items[r.item_id][field] : "-"
    },

    cellClass(r, field, expected) {
      const item = this.items[r.item_id]
      return item && item[field] !== expected ? "text-danger font-weight-bold" : ""
    },

    classicRoute(id) {
      return util.format(ROUTE.ITEM_EDIT_CLASSIC, id)
    },
    itemIcon(id) {
      return this.items[id] && this.items[id].icon ? this.items[id].icon : 0
    },

    toDetail(r) {
      return {
        id: r.id,
        item_evo_id: this.evoId,
        item_evolve_level: Number(r.item_evolve_level),
        item_id: Number(r.item_id),
        type: Number(r.type),
        sub_type: String(r.sub_type),
        required_amount: Number(r.required_amount),
      }
    },

    async run(action, message) {
      this.busy  = true
      this.error = ""
      try {
        await action()
        this.notify(message)
        await this.load()
      } catch (e) {
        this.error = this.errorText(e)
      }
      this.busy = false
    },

    saveRow(r) {
      return this.run(() => EvolvingItems.updateDetail(this.toDetail(r)), "Saved row " + r.id)
    },

    deleteRow(r) {
      if (!confirm("Delete level " + r.item_evolve_level + " (item " + r.item_id + ") from chain " + this.evoId + "?")) {
        return
      }
      return this.run(() => EvolvingItems.deleteDetail(r.id), "Deleted row " + r.id)
    },

    addGap(g) {
      const detail = this.toDetail({...g, item_evolve_level: g.level})
      delete detail.id
      return this.run(() => EvolvingItems.createDetail(detail), "Added level " + g.level)
    },

    addNewRow() {
      const detail = this.toDetail(this.newRow)
      delete detail.id
      return this.run(() => EvolvingItems.createDetail(detail), "Added level " + detail.item_evolve_level)
    },

    // items.evoid / evolvinglevel / evomax / evoitem are what the server reads to find the next item
    async syncItem(r) {
      const maxLevel = Math.max(...this.rows.map((x) => x.item_evolve_level))
      const api      = new ItemApi(...SpireApi.cfg())
      await api.updateItem({
        id: r.item_id,
        item: {
          evoitem: 1,
          evoid: this.evoId,
          evolvinglevel: r.item_evolve_level,
          evomax: maxLevel,
        },
      })
      Items.setItem(r.item_id, undefined)
    },

    syncAllItems() {
      return this.run(
        () => EvolvingItems.synchronize(this.evoId),
        "Synced item columns to chain " + this.evoId,
      )
    },

    notify(message) {
      this.notification = message
      setTimeout(() => this.notification = "", 5000)
    },

    errorText(e) {
      return (e.response && e.response.data && e.response.data.error) || String(e)
    },
  },
}
</script>

<style scoped>
.evo-row-bad td:first-child {
  box-shadow: inset 2px 0 0 var(--danger);
}

.evo-table-wrap {
  overflow-x: auto;
}

input[type=number],
input.evo-sub {
  min-width: 60px;
  padding-left: 6px !important;
  padding-right: 6px !important;
  -moz-appearance: textfield;
}

input.evo-id {
  min-width: 95px !important;
}

input.evo-sub {
  min-width: 180px !important;
}

.evo-type {
  min-width: 120px !important;
}

input[type=number]::-webkit-inner-spin-button,
input[type=number]::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.evo-item-icon {
  display: inline-block;
  flex-shrink: 0;
  vertical-align: middle;
}

.evo-item-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.evo-steps {
  margin: 0 0 16px;
  padding-left: 20px;
  color: var(--text-muted);
  font-size: 13px;
}

.evo-steps li + li {
  margin-top: 4px;
}

.evo-table-title {
  margin: 8px 0 6px;
}

.evo-sub-help {
  font-size: 11px;
  color: var(--text-subtle);
  margin-top: 3px;
}

.evo-zones {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 4px;
  max-width: 420px;
}

.evo-zone-pill {
  display: inline-block;
  padding: 1px 7px;
  border: 1px solid var(--border);
  border-radius: 999px;
  background: var(--surface-2);
  color: var(--text-muted);
  font-size: 11px;
  font-variant-numeric: tabular-nums;
}

.evo-zone-pill.is-missing {
  border-color: var(--danger);
  color: var(--danger);
}

.tabular {
  font-variant-numeric: tabular-nums;
}
</style>

<template>
  <div>
    <div class="row" v-for="field in itemFields" :key="field.field">
      <div class="col-4 text-right mt-2 p-0 mr-3">{{ field.description }}</div>
      <div class="col-2 p-0" :style="(item[field.field] === 0 ? 'opacity: .5' : '')">
        <b-form-input v-model.number="item[field.field]" :id="field.field"/>
      </div>
      <div class="col-5 mt-2 text-muted" style="font-size: 12px">{{ field.help }}</div>
    </div>
    <div class="text-muted text-center mt-1 mb-3" style="font-size: 12px">
      These four item columns save with "Save Item". The chain rows below save with their own buttons.
    </div>

    <div class="d-flex align-items-center mb-2">
      <h6 class="eq-header m-0 mr-auto">
        items_evolving_details
        <span v-if="item.evoid"> - chain {{ item.evoid }}</span>
      </h6>
      <b-button size="sm" variant="outline-warning" class="mr-2" @click="load()">
        <i class="fa fa-refresh"></i>
      </b-button>
      <router-link v-if="item.evoid" :to="chainRoute" class="btn btn-sm btn-outline-light">
        <i class="fa fa-external-link"></i> Full chain editor
      </router-link>
      <router-link v-else to="/items/evolving" class="btn btn-sm btn-outline-light">
        <i class="fa fa-list"></i> All evolving chains
      </router-link>
    </div>

    <app-loader :is-loading="!loaded" class="mt-3 mb-3"/>

    <div v-if="loaded">
      <b-alert show variant="danger" v-if="error">{{ error }}</b-alert>
      <div v-if="notification" class="text-center text-success mb-2">{{ notification }}</div>

      <div v-if="!item.evoid && !ownRows.length" class="text-muted text-center">
        This item is not part of an evolving chain (evoid is 0 and no items_evolving_details row uses item {{ item.id }}).
      </div>

      <div v-for="r in ownRows.filter((r) => r.item_evo_id !== item.evoid)" :key="'other-' + r.id" class="text-danger mb-2">
        <i class="fa fa-exclamation-triangle"></i>
        items_evolving_details row {{ r.id }} puts this item in chain {{ r.item_evo_id }} level {{ r.item_evolve_level }},
        but items.evoid is {{ item.evoid }}
      </div>

      <div v-for="p in analysis.problems" :key="p" class="ui-problem">
        {{ split(p).title }}
        <small v-if="split(p).detail">{{ split(p).detail }}</small>
      </div>

      <table v-if="rows.length" class="eq-table bordered mt-2" style="width: 100%; font-size: 12px">
        <thead>
        <tr>
          <th>Lvl</th>
          <th>Item</th>
          <th style="width: 110px">Type</th>
          <th style="width: 80px">Sub Type</th>
          <th style="width: 90px">Required</th>
          <th></th>
        </tr>
        </thead>
        <tbody>
        <tr v-for="r in rows" :key="r.id" :class="analysis.rowIssues[r.id] ? 'evo-row-bad' : ''">
          <td>
            {{ r.item_evolve_level }}
            <span v-if="r.item_id === item.id" class="text-warning">(this)</span>
          </td>
          <td>
            <router-link :to="itemRoute(r.item_id)">
              {{ items[r.item_id] ? items[r.item_id].name : '(missing)' }} ({{ r.item_id }})
            </router-link>
            <div v-for="issue in (analysis.rowIssues[r.id] || [])" :key="issue" class="text-danger">
              {{ issue }}
            </div>
          </td>
          <td>
            <select class="form-control form-control-sm evo-type" v-model.number="r.type">
              <option v-for="(name, id) in types" :key="id" :value="Number(id)">{{ id }}: {{ name }}</option>
            </select>
          </td>
          <td>
            <input type="text" class="form-control form-control-sm" v-model="r.sub_type">
            <small class="text-muted">{{ describeSubType(r) }}</small>
          </td>
          <td><input type="number" class="form-control form-control-sm" v-model.number="r.required_amount"></td>
          <td>
            <b-button size="sm" variant="outline-success" title="Save row" :disabled="busy" @click="saveRow(r)">
              <i class="fa fa-save"></i>
            </b-button>
          </td>
        </tr>
        </tbody>
      </table>

      <div v-if="analysis.missing.length" class="text-warning mt-2" style="font-size: 12px">
        Levels {{ analysis.missing.join(", ") }} have no row. Use the full chain editor to add them.
      </div>
    </div>
  </div>
</template>

<script>
import * as util                         from "util";
import {ROUTE}                           from "../../../routes";
import {EVOLVING_TYPES, EvolvingItems}   from "../../../app/evolving-items";

export default {
  name: "ItemEvolvingTab",
  props: {
    item: { type: Object, required: true },
  },
  data() {
    return {
      loaded: false,
      busy: false,
      error: "",
      notification: "",
      types: EVOLVING_TYPES,
      rows: [],
      ownRows: [],
      items: {},
      analysis: { problems: [], rowIssues: {}, missing: [], maxLevel: 0 },
      itemFields: [
        { field: "evoitem", description: "Evolving Item (evoitem)", help: "1 = this item evolves" },
        { field: "evoid", description: "Evolution ID (evoid)", help: "items_evolving_details.item_evo_id of the chain" },
        { field: "evolvinglevel", description: "Evolving Level (evolvinglevel)", help: "The server looks up level + 1 for the next item" },
        { field: "evomax", description: "Max Level (evomax)", help: "Highest level in the chain" },
      ],
    }
  },
  computed: {
    chainRoute() {
      return util.format(ROUTE.ITEM_EVOLVING_CHAIN, this.item.evoid)
    },
  },
  watch: {
    "item.id"() {
      this.load()
    },
  },
  created() {
    this.load()
  },
  methods: {
    itemRoute(id) {
      return util.format(ROUTE.ITEM_EDIT, id)
    },
    describeSubType(r) {
      return EvolvingItems.describeSubType(r.type, r.sub_type)
    },
    split(p) {
      return EvolvingItems.splitProblem(p)
    },
    async load() {
      this.loaded       = false
      this.error        = ""
      this.notification = ""
      try {
        const all     = await EvolvingItems.listDetails()
        this.ownRows  = all.filter((d) => d.item_id === this.item.id)
        this.rows     = EvolvingItems.groupChains(all)[this.item.evoid] || []
        this.items    = this.rows.length ? await EvolvingItems.loadItemsFor(this.rows) : {}
        this.analysis = this.rows.length
          ? EvolvingItems.analyzeChain(this.rows, this.items, all)
          : { problems: [], rowIssues: {}, missing: [], maxLevel: 0 }
      } catch (e) {
        this.error = (e.response && e.response.data && e.response.data.error) || e.message
      }
      this.loaded = true
    },
    async saveRow(r) {
      this.busy = true
      try {
        await EvolvingItems.updateDetail(r)
        this.notification = "Saved level " + r.item_evolve_level + ". Restart zones for the server to use it."
        this.analysis     = EvolvingItems.analyzeChain(this.rows, this.items, this.rows)
      } catch (e) {
        this.error = (e.response && e.response.data && e.response.data.error) || e.message
      }
      this.busy = false
    },
  },
}
</script>

<style scoped>
.evo-row-bad td:first-child {
  box-shadow: inset 2px 0 0 var(--danger);
}

.evo-type {
  min-width: 110px !important;
}

input[type=number] {
  min-width: 60px;
  padding-left: 6px !important;
  padding-right: 6px !important;
  -moz-appearance: textfield;
}

input[type=number]::-webkit-inner-spin-button,
input[type=number]::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
</style>

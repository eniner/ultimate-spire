<template>
  <content-area>
    <eq-window title="Evolving Items">
      <b-alert show dismissable variant="danger" v-if="error">
        <i class="fa fa-warning"></i> {{ error }}
      </b-alert>

      <app-loader :is-loading="!loaded" class="mt-3 mb-3"/>

      <div v-if="loaded">
        <div class="evo-toolbar">
          <div class="ui-stat-line mr-auto">
            <b>{{ chainList.length }}</b> chains
            <span class="evo-dot">·</span>
            <b>{{ details.length }}</b> rows
            <span class="evo-dot">·</span>
            <b :class="brokenCount ? 'text-danger' : 'text-success'">{{ brokenCount }}</b> with problems
          </div>

          <input
            v-model="search"
            type="search"
            class="form-control form-control-sm"
            style="width: 260px"
            placeholder="Chain ID, item ID, or item name"
            aria-label="Search chains"
          >

          <select v-model="typeFilter" class="form-control form-control-sm evo-type-filter" aria-label="Filter by type">
            <option value="">All types</option>
            <option v-for="(name, id) in types" :key="id" :value="String(id)">{{ id }}: {{ name }}</option>
          </select>

          <b-form-checkbox v-model="onlyProblems" switch class="evo-switch">Only with problems</b-form-checkbox>

          <b-button size="sm" variant="outline-secondary" @click="load()">
            <i class="fa fa-refresh"></i> Reload
          </b-button>
        </div>

        <table class="eq-table" style="width: 100%">
          <thead>
          <tr>
            <th style="width: 90px">Chain</th>
            <th>Levels</th>
            <th style="width: 140px">Type</th>
            <th style="width: 60px" class="text-right">Max</th>
            <th style="width: 150px">Status</th>
          </tr>
          </thead>
          <tbody>
          <template v-for="c in filteredChains">
            <tr :key="c.evoId" class="evo-row" @click="openChain(c.evoId)">
              <td class="tabular"><b>{{ c.evoId }}</b></td>
              <td>
                <span
                  v-for="r in c.rows"
                  :key="r.id"
                  class="evo-pill"
                  :title="'Level ' + r.item_evolve_level + ': ' + itemName(r.item_id)"
                >
                  <span v-if="itemIcon(r.item_id)" :class="'item-' + itemIcon(r.item_id) + ' evo-item-icon'"></span>
                  <span class="evo-pill-level">{{ r.item_evolve_level }}</span>
                  {{ r.item_id }}
                </span>
              </td>
              <td class="text-muted">{{ chainType(c) }}</td>
              <td class="text-right tabular">{{ c.analysis.maxLevel }}</td>
              <td>
                <span v-if="!c.analysis.problems.length" class="badge badge-success">No problems</span>
                <button
                  v-else
                  type="button"
                  class="evo-problem-badge ui-plain"
                  :aria-expanded="expanded[c.evoId] ? 'true' : 'false'"
                  @click.stop="toggle(c.evoId)"
                >
                  {{ c.analysis.problems.length }} {{ c.analysis.problems.length === 1 ? 'problem' : 'problems' }}
                  <i :class="'fa fa-chevron-' + (expanded[c.evoId] ? 'up' : 'down')"></i>
                </button>
              </td>
            </tr>
            <tr v-if="expanded[c.evoId]" :key="'p-' + c.evoId" class="evo-detail-row">
              <td></td>
              <td colspan="4">
                <div v-for="p in c.analysis.problems" :key="p" class="ui-problem">
                  {{ split(p).title }}
                  <small v-if="split(p).detail">{{ split(p).detail }}</small>
                </div>
              </td>
            </tr>
          </template>
          </tbody>
        </table>

        <div v-if="!filteredChains.length" class="ui-empty mt-3">
          <div class="ui-empty-title">No chains match</div>
          Try a different chain ID, item ID or name, or clear the type / problems filters.
        </div>

        <div v-if="orphanItems.length" class="mt-4">
          <h5>Items with evoid set but no chain row</h5>
          <table class="eq-table" style="width: 100%">
            <thead>
            <tr>
              <th>Item</th>
              <th>evoitem</th>
              <th>evoid</th>
              <th>evolvinglevel</th>
              <th>evomax</th>
            </tr>
            </thead>
            <tbody>
            <tr v-for="i in orphanItems" :key="i.id">
              <td class="evo-item-cell">
                <span v-if="i.icon" :class="'item-' + i.icon + ' evo-item-icon'"></span>
                <router-link :to="classicRoute(i.id)">{{ i.id }} - {{ i.name }}</router-link>
              </td>
              <td>{{ i.evoitem }}</td>
              <td>
                <router-link :to="chainRoute(i.evoid)">{{ i.evoid }}</router-link>
              </td>
              <td>{{ i.evolvinglevel }}</td>
              <td>{{ i.evomax }}</td>
            </tr>
            </tbody>
          </table>
        </div>
      </div>
    </eq-window>
  </content-area>
</template>

<script>
import * as util       from "util";
import EqWindow        from "../../components/eq-ui/EQWindow";
import ContentArea     from "../../components/layout/ContentArea";
import {EVOLVING_TYPES, EvolvingItems} from "../../app/evolving-items";
import {ROUTE}         from "../../routes";

export default {
  name: "EvolvingItems",
  components: {ContentArea, EqWindow},
  data() {
    return {
      loaded: false,
      error: "",
      types: EVOLVING_TYPES,
      details: [],
      items: {},
      chainList: [],
      search: this.$route.query.q || "",
      typeFilter: this.$route.query.type || "",
      onlyProblems: this.$route.query.problems === "1",
      expanded: {},
    }
  },
  computed: {
    brokenCount() {
      return this.chainList.filter((c) => c.analysis.problems.length).length
    },
    filteredChains() {
      const s = this.search.trim().toLowerCase()
      const type = this.typeFilter === "" ? null : Number(this.typeFilter)
      return this.chainList.filter((c) => {
        if (this.onlyProblems && !c.analysis.problems.length) {
          return false
        }
        if (type !== null && !(c.rows || []).some((r) => Number(r.type) === type)) {
          return false
        }
        if (!s) {
          return true
        }
        if (String(c.evoId) === s) {
          return true
        }
        return c.rows.some((r) => String(r.item_id) === s || this.itemName(r.item_id).toLowerCase().includes(s))
      })
    },
    orphanItems() {
      const inDetails = new Set(this.details.map((d) => d.item_id))
      return Object.values(this.items).filter((i) => i.evoid > 0 && !inDetails.has(i.id))
    },
  },
  watch: {
    search() {
      this.updateQuery()
    },
    typeFilter() {
      this.updateQuery()
    },
    onlyProblems() {
      this.updateQuery()
    },
  },
  mounted() {
    this.load()
  },
  methods: {
    async load() {
      this.loaded = false
      this.error  = ""
      try {
        this.details = await EvolvingItems.listDetails()
        this.items   = await EvolvingItems.loadItemsFor(this.details)

        const chains   = EvolvingItems.groupChains(this.details)
        this.chainList = Object.keys(chains).map((evoId) => ({
          evoId: Number(evoId),
          rows: chains[evoId],
          analysis: EvolvingItems.analyzeChain(chains[evoId], this.items, this.details),
        })).sort((a, b) => a.evoId - b.evoId)
      } catch (e) {
        this.error = (e.response && e.response.data && e.response.data.error) || String(e)
      }
      this.loaded = true
    },
    chainType(c) {
      const types = [...new Set((c.rows || []).map((r) => Number(r.type)))]
      return types.map((t) => EVOLVING_TYPES[t] || ("Type " + t)).join(" / ")
    },
    itemName(id) {
      return this.items[id] ? this.items[id].name : "(missing item)"
    },
    itemIcon(id) {
      return this.items[id] && this.items[id].icon ? this.items[id].icon : 0
    },
    split(p) {
      return EvolvingItems.splitProblem(p)
    },
    toggle(evoId) {
      this.$set(this.expanded, evoId, !this.expanded[evoId])
    },
    openChain(evoId) {
      this.$router.push(this.chainRoute(evoId))
    },
    chainRoute(evoId) {
      return util.format(ROUTE.ITEM_EVOLVING_CHAIN, evoId)
    },
    classicRoute(id) {
      return util.format(ROUTE.ITEM_EDIT_CLASSIC, id)
    },
    updateQuery() {
      const q = {}
      if (this.search) {
        q.q = this.search
      }
      if (this.typeFilter) {
        q.type = this.typeFilter
      }
      if (this.onlyProblems) {
        q.problems = "1"
      }
      this.$router.replace({path: this.$route.path, query: q}).catch(() => {
      })
    },
  },
}
</script>

<style scoped>
.evo-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.evo-dot {
  margin: 0 6px;
  color: var(--text-subtle);
}

.evo-switch {
  white-space: nowrap;
}

.evo-type-filter {
  width: 170px;
}

.evo-row {
  cursor: pointer;
}

.evo-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin: 2px 6px 2px 0;
  padding: 3px 10px 3px 4px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--surface-2);
  color: var(--text-muted);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
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

.evo-pill-level {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 999px;
  background: var(--accent-soft);
  color: var(--accent);
  font-weight: 600;
}

.evo-problem-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 2px 9px !important;
  border: 1px solid transparent !important;
  border-radius: 999px !important;
  background: var(--danger-soft) !important;
  color: var(--danger) !important;
  font-size: 11.5px !important;
  font-weight: 600;
  line-height: 1.5 !important;
  cursor: pointer;
  box-shadow: none !important;
}

.evo-problem-badge:hover {
  border-color: var(--danger) !important;
}

.evo-problem-badge .fa {
  font-size: 9px;
}

.evo-detail-row,
.evo-detail-row:hover {
  background: var(--surface-2) !important;
}
</style>

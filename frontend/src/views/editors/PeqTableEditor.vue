<template>
  <div>
    <div class="row">
      <div :class="selected ? 'col-7' : 'col-12'">
        <eq-window :title="title" class="p-3">
          <div class="ui-toolbar">
            <div class="ui-field" style="flex: 1 1 220px; margin-bottom: 0">
              <label>Search</label>
              <input
                class="form-control form-control-sm"
                v-model="search"
                @keyup.enter="page = 1; load()"
                placeholder="Name or id"
              >
            </div>
            <div class="ui-field" style="flex: 1 1 220px; margin-bottom: 0" v-if="hasZoneFilter">
              <label>Zone</label>
              <select class="form-control form-control-sm" v-model="zoneValue" @change="page = 1; load()">
                <option value="">All zones</option>
                <option v-for="z in zones" :key="z.short_name + '-' + z.version" :value="zoneOptionValue(z)">
                  {{ z.short_name }} ({{ z.version }}) {{ z.long_name }}
                </option>
              </select>
            </div>
            <button class="btn btn-sm btn-dark" @click="page = 1; load()">Search</button>
            <button class="btn btn-sm btn-dark" @click="reset()">Reset</button>
            <button class="btn btn-sm btn-dark" v-if="canCreate" @click="createRow()">New</button>
            <router-link class="btn btn-sm btn-dark" to="/editors">All editors</router-link>
            <div class="ui-stat-line" v-if="status">{{ status }}</div>
          </div>
        </eq-window>

        <eq-window class="p-0">
          <div v-if="loading" class="p-3">Loading…</div>
          <div v-else-if="!rows.length" class="p-3">No rows.</div>
          <div v-else style="overflow: auto; max-height: 78vh">
            <table class="eq-table eq-highlight-rows bordered" style="font-size: 13px">
              <thead class="eq-table-floating-header">
                <tr>
                  <th v-if="rowActions.length"></th>
                  <th v-for="col in columns" :key="col" class="text-center">{{ col }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="row in rows"
                  :key="rowKey(row)"
                  @click="selectRow(row)"
                  :class="{ 'peq-row-selected': selected && rowKey(selected) === rowKey(row) }"
                >
                  <td v-if="rowActions.length" class="text-center" @click.stop>
                    <router-link
                      v-for="action in rowActions"
                      :key="action.label"
                      class="btn btn-sm btn-dark"
                      :to="actionTo(action, row)"
                    >{{ action.label }}</router-link>
                  </td>
                  <td v-for="col in columns" :key="col" class="text-center">{{ formatCell(row[col]) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="p-2 text-center" v-if="totalRows > pageSize">
            <b-pagination
              class="mb-0"
              v-model="page"
              :total-rows="totalRows"
              :per-page="pageSize"
              :hide-ellipsis="true"
              @change="onPage"
            />
          </div>
        </eq-window>
      </div>

      <div class="col-5" v-if="selected">
        <eq-window title="Edit row" class="p-3">
          <div class="ui-field-grid">
            <div class="form-group" v-for="col in editColumns" :key="col">
              <label>{{ col }}</label>
              <input
                class="form-control form-control-sm"
                :value="selected[col]"
                :disabled="isIdentityField(col)"
                @input="onField(col, $event)"
              >
            </div>
          </div>
          <div class="ui-toolbar" style="margin-top: 8px; margin-bottom: 0">
            <button class="btn btn-sm btn-dark" :disabled="saving" @click="saveRow">
              {{ saving ? "Saving…" : "Save" }}
            </button>
            <button class="btn btn-sm btn-outline-danger" v-if="canDelete" :disabled="saving" @click="deleteRow">Delete</button>
            <router-link
              v-for="action in rowActions"
              :key="'side-' + action.label"
              class="btn btn-sm btn-dark"
              :to="actionTo(action, selected)"
            >{{ action.label }}</router-link>
          </div>
        </eq-window>
      </div>
    </div>
  </div>
</template>

<script>
import EqWindow from "../../components/eq-ui/EQWindow"
import { SpireApi } from "../../app/api/spire-api"
import { SpireQueryBuilder } from "../../app/api/spire-query-builder"
import { Zones } from "../../app/zones"
import { getPeqEditor } from "../../app/peq-editors"
import { getPeqTableConfig } from "../../app/peq-editor-tables"

export default {
  name: "PeqTableEditor",
  components: { EqWindow },
  data() {
    return {
      editor: null,
      tableCfg: null,
      rows: [],
      selected: null,
      columns: [],
      search: "",
      zoneValue: "",
      zones: [],
      page: 1,
      pageSize: 100,
      totalRows: 0,
      loading: false,
      saving: false,
      status: "",
      original: null,
    }
  },
  computed: {
    title() {
      return this.editor ? this.editor.title : "Editor"
    },
    hasZoneFilter() {
      return !!(this.tableCfg && this.tableCfg.zoneField)
    },
    canCreate() {
      return !!(this.tableCfg && this.tableCfg.allowCreate !== false)
    },
    canDelete() {
      return !!(this.tableCfg && this.tableCfg.allowDelete !== false)
    },
    rowActions() {
      return (this.tableCfg && this.tableCfg.rowActions) || []
    },
    idField() {
      return (this.tableCfg && this.tableCfg.idField) || "id"
    },
    editColumns() {
      if (!this.selected) {
        return []
      }
      const hide = (this.table() && this.table().hideColumns) || []
      return Object.keys(this.selected).filter((key) => {
        return this.isScalar(this.selected[key]) && hide.indexOf(key) === -1
      })
    },
    identityFields() {
      const fields = []
      const add = (field) => {
        if (field && fields.indexOf(field) === -1) {
          fields.push(field)
        }
      }
      add(this.idField)
      const table = this.table()
      if (table && table.rowKeyFields) {
        table.rowKeyFields.forEach(add)
      }
      if (table && table.idQuery) {
        table.idQuery.forEach((item) => add(item.field))
      }
      return fields
    },
  },
  watch: {
    "$route.params.id"() {
      this.boot()
    },
  },
  mounted() {
    this.boot()
  },
  methods: {
    table() {
      return this.tableCfg
    },
    client() {
      const table = this.table()
      return table ? new table.Api(...SpireApi.cfg()) : null
    },
    rowKey(row) {
      if (!row) {
        return ""
      }
      const table = this.table()
      const fields = (table && table.rowKeyFields && table.rowKeyFields.length)
        ? table.rowKeyFields
        : [this.idField]
      return fields.map((field) => String(row[field] != null ? row[field] : "")).join("|")
    },
    rowId(row) {
      return row ? row[this.idField] : undefined
    },
    actionTo(action, row) {
      if (!action) {
        return "/"
      }
      return typeof action.to === "function" ? action.to(row) : action.to
    },
    extraOptions(row) {
      const table = this.table()
      const source = this.original || row
      if (!table || !table.idQuery || !table.idQuery.length || !source) {
        return undefined
      }
      const query = {}
      table.idQuery.forEach((item) => {
        const value = source[item.field]
        if (value !== undefined && value !== null && value !== "") {
          query[item.query] = value
        }
      })
      return Object.keys(query).length ? { query } : undefined
    },
    isIdentityField(col) {
      return this.identityFields.indexOf(col) !== -1
    },
    dirtyBody(row) {
      const body = {}
      Object.keys(row || {}).forEach((key) => {
        if (!this.isScalar(row[key]) || this.isIdentityField(key)) {
          return
        }
        if (this.original && String(this.original[key]) === String(row[key])) {
          return
        }
        body[key] = row[key]
      })
      return body
    },
    isScalar(value) {
      return value === null || value === undefined || typeof value !== "object"
    },
    formatCell(value) {
      if (value === null || value === undefined) {
        return ""
      }
      if (typeof value === "object") {
        return ""
      }
      return value
    },
    zoneOptionValue(zone) {
      const table = this.table()
      if (table && table.zoneFieldIsId) {
        return String(zone.zoneidnumber)
      }
      return zone.short_name
    },
    visibleColumns(row) {
      const hide = (this.table() && this.table().hideColumns) || []
      return Object.keys(row).filter((key) => {
        return this.isScalar(row[key]) && hide.indexOf(key) === -1
      })
    },
    async boot() {
      this.editor = getPeqEditor(this.$route.params.id)
      this.tableCfg = getPeqTableConfig(this.$route.params.id)
      this.rows = []
      this.selected = null
      this.original = null
      this.search = this.$route.query.search || ""
      this.zoneValue = this.$route.query.zone || ""
      this.page = parseInt(this.$route.query.page, 10) || 1
      this.status = ""
      if (!this.editor || !this.tableCfg) {
        this.status = "This editor is not a table editor."
        return
      }
      if (this.hasZoneFilter && !this.zones.length) {
        this.zones = await Zones.getZones()
      }
      await this.load()
    },
    reset() {
      this.search = ""
      this.zoneValue = ""
      this.page = 1
      this.selected = null
      this.original = null
      this.load()
    },
    onPage(next) {
      this.page = next
      this.load()
    },
    selectRow(row) {
      this.selected = Object.assign({}, row)
      this.original = Object.assign({}, row)
    },
    onField(col, event) {
      if (!this.selected) {
        return
      }
      const raw = event.target.value
      const current = this.selected[col]
      this.$set(this.selected, col, typeof current === "number" ? Number(raw) : raw)
    },
    pushQuery() {
      const query = {}
      if (this.search) {
        query.search = this.search
      }
      if (this.zoneValue) {
        query.zone = this.zoneValue
      }
      if (this.page > 1) {
        query.page = String(this.page)
      }
      this.$router.replace({ path: this.$route.path, query }).catch(() => {})
    },
    async load() {
      const table = this.table()
      const api = this.client()
      if (!table || !api) {
        return
      }
      this.loading = true
      this.status = ""
      this.pushQuery()
      try {
        const builder = new SpireQueryBuilder()
        builder.limit(this.pageSize)
        builder.page(this.page)
        if (table.orderBy && table.orderBy.length) {
          builder.orderBy(table.orderBy)
        }
        if (table.orderDirection) {
          builder.orderDirection(table.orderDirection)
        }
        if (table.select && table.select.length) {
          builder.select(table.select)
        }
        if (this.zoneValue && table.zoneField) {
          builder.where(table.zoneField, "=", this.zoneValue)
        }
        if (this.search) {
          const numeric = /^\d+$/.test(this.search)
          const idField = this.idField
          table.searchFields.forEach((field) => {
            if ((field === "id" || field === idField || field === "msgid" || field === "charid") && numeric) {
              builder.whereOr(field, "=", this.search)
            } else if (field !== "id" && field !== idField) {
              builder.whereOr(field, "like", this.search)
            }
          })
        }
        const r = await api[table.list](builder.get())
        this.rows = (r && r.data) ? r.data : []
        this.columns = this.rows.length ? this.visibleColumns(this.rows[0]) : []
        if (table.count && api[table.count]) {
          const cr = await api[table.count](builder.get())
          const countBody = cr && cr.data
          this.totalRows = typeof countBody === "number"
            ? countBody
            : (countBody && (countBody.count || countBody.data || this.rows.length))
        } else {
          this.totalRows = this.rows.length
        }
        if (this.selected) {
          const key = this.original ? this.rowKey(this.original) : this.rowKey(this.selected)
          const still = this.rows.find((row) => this.rowKey(row) === key)
          if (still) {
            this.selected = Object.assign({}, still)
            this.original = Object.assign({}, still)
          }
        }
      } catch (err) {
        this.status = this.errorText(err)
        this.rows = []
      } finally {
        this.loading = false
      }
    },
    async createRow() {
      const table = this.table()
      const api = this.client()
      if (!table || !api) {
        return
      }
      this.saving = true
      this.status = ""
      try {
        const body = Object.assign({}, table.newRow)
        if (this.zoneValue && table.zoneField) {
          body[table.zoneField] = table.zoneFieldIsId ? Number(this.zoneValue) : this.zoneValue
        }
        const payload = {}
        payload[table.bodyKey] = body
        const r = await api[table.create](payload)
        this.status = "Created"
        this.page = 1
        await this.load()
        if (r && r.data) {
          const createdKey = this.rowKey(r.data)
          const created = this.rows.find((row) => this.rowKey(row) === createdKey)
          this.selected = created ? Object.assign({}, created) : Object.assign({}, r.data)
          this.original = Object.assign({}, this.selected)
        }
      } catch (err) {
        this.status = this.errorText(err)
      } finally {
        this.saving = false
      }
    },
    async saveRow() {
      const table = this.table()
      const api = this.client()
      const identity = this.original || this.selected
      if (!table || !api || !this.selected || this.rowId(identity) === undefined || this.rowId(identity) === null) {
        return
      }
      const body = this.dirtyBody(this.selected)
      if (!Object.keys(body).length) {
        this.status = "No changes"
        return
      }
      this.saving = true
      this.status = ""
      try {
        const payload = { id: this.rowId(identity) }
        payload[table.bodyKey] = body
        await api[table.update](payload, this.extraOptions(identity))
        this.status = "Saved"
        await this.load()
      } catch (err) {
        this.status = this.errorText(err)
      } finally {
        this.saving = false
      }
    },
    async deleteRow() {
      const table = this.table()
      const api = this.client()
      const identity = this.original || this.selected
      if (!table || !api || !identity || this.rowId(identity) === undefined || this.rowId(identity) === null) {
        return
      }
      if (!confirm("Delete " + this.title + " #" + this.rowId(identity) + "?")) {
        return
      }
      this.saving = true
      this.status = ""
      try {
        await api[table.delete]({ id: this.rowId(identity) }, this.extraOptions(identity))
        this.status = "Deleted"
        this.selected = null
        this.original = null
        await this.load()
      } catch (err) {
        this.status = this.errorText(err)
      } finally {
        this.saving = false
      }
    },
    errorText(err) {
      if (err && err.response && err.response.data && err.response.data.error) {
        return err.response.data.error
      }
      return err && err.message ? err.message : "Request failed"
    },
  },
}
</script>

<style scoped>
.peq-row-selected {
  outline: 1px solid #c9a24a;
}
</style>

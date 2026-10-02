<template>
  <div>
    <div class="row">
      <div :class="selected ? 'col-7' : 'col-12'">
        <eq-window title="Mercs" class="p-3">
          <div class="row align-items-end">
            <div class="col-md-4">
              <label class="small mb-0">Table</label>
              <select class="form-control form-control-sm" v-model="table" @change="page = 1; load()">
                <option value="">Select a merc table</option>
                <option v-for="name in tables" :key="name" :value="name">{{ name }}</option>
              </select>
            </div>
            <div class="col-md-4">
              <label class="small mb-0">Search</label>
              <input
                class="form-control form-control-sm"
                v-model="search"
                @keyup.enter="page = 1; load()"
                placeholder="Name or id"
              >
            </div>
            <div class="col-md-4">
              <button class="btn btn-sm btn-dark mr-1" @click="page = 1; load()">Search</button>
              <button class="btn btn-sm btn-dark mr-1" :disabled="!table" @click="createRow">New</button>
              <router-link class="btn btn-sm btn-dark" to="/editors">All editors</router-link>
            </div>
          </div>
          <div class="small mt-2" v-if="status">{{ status }}</div>
        </eq-window>

        <eq-window class="p-0">
          <div v-if="loading" class="p-3">Loading…</div>
          <div v-else-if="!table" class="p-3">Pick a merc table. These are the PHP Mercs tabs.</div>
          <div v-else-if="!rows.length" class="p-3">No rows.</div>
          <div v-else style="overflow: auto; max-height: 78vh">
            <table class="eq-table eq-highlight-rows bordered" style="font-size: 13px">
              <thead class="eq-table-floating-header">
                <tr>
                  <th v-for="col in columns" :key="col" class="text-center">{{ col }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="row in rows"
                  :key="rowKey(row)"
                  @click="selectRow(row)"
                  :class="{ 'peq-row-selected': selected && rowKey(original || selected) === rowKey(row) }"
                >
                  <td v-for="col in columns" :key="col" class="text-center">{{ row[col] }}</td>
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
          <div class="form-group" v-for="col in columns" :key="col">
            <label class="small mb-0">{{ col }}</label>
            <input class="form-control form-control-sm" v-model="selected[col]" :disabled="isPrimaryKey(col)">
          </div>
          <div>
            <button class="btn btn-sm btn-dark mr-1" :disabled="saving" @click="saveRow">
              {{ saving ? "Saving…" : "Save" }}
            </button>
            <button class="btn btn-sm btn-outline-danger" :disabled="saving" @click="deleteRow">Delete</button>
          </div>
        </eq-window>
      </div>
    </div>
  </div>
</template>

<script>
import EqWindow from "../../components/eq-ui/EQWindow"
import { PeqRawApi } from "../../app/peq-raw"

export default {
  name: "PeqRawEditor",
  components: { EqWindow },
  data() {
    return {
      tables: [],
      table: "",
      rows: [],
      columns: [],
      primaryKey: [],
      selected: null,
      search: "",
      page: 1,
      pageSize: 100,
      totalRows: 0,
      loading: false,
      saving: false,
      status: "",
      original: null,
    }
  },
  async mounted() {
    try {
      this.tables = await PeqRawApi.tables()
    } catch (err) {
      this.status = this.errorText(err)
    }
    if (this.$route.query.table) {
      this.table = this.$route.query.table
    } else if (this.tables.indexOf("merc_templates") !== -1) {
      this.table = "merc_templates"
    } else if (this.tables.length) {
      this.table = this.tables[0]
    }
    if (this.table) {
      await this.load()
    }
  },
  methods: {
    rowKey(row) {
      if (!row) {
        return ""
      }
      const keys = this.primaryKey.length ? this.primaryKey : Object.keys(row).slice(0, 3)
      return keys.map((key) => String(row[key] != null ? row[key] : "")).join("|")
    },
    pk(row) {
      const source = row || {}
      const out = {}
      this.primaryKey.forEach((key) => {
        out[key] = source[key]
      })
      return out
    },
    isPrimaryKey(col) {
      return this.primaryKey.indexOf(col) !== -1
    },
    selectRow(row) {
      this.selected = Object.assign({}, row)
      this.original = Object.assign({}, row)
    },
    dirtyRow(row) {
      const out = {}
      this.columns.forEach((col) => {
        if (this.isPrimaryKey(col)) {
          return
        }
        if (this.original && String(this.original[col]) === String(row[col])) {
          return
        }
        out[col] = row[col]
      })
      return out
    },
    onPage(next) {
      this.page = next
      this.load()
    },
    async load() {
      if (!this.table) {
        return
      }
      this.loading = true
      this.status = ""
      this.$router.replace({ path: this.$route.path, query: { table: this.table, search: this.search || undefined } }).catch(() => {})
      try {
        const [list, count] = await Promise.all([
          PeqRawApi.list(this.table, this.search, this.page, this.pageSize),
          PeqRawApi.count(this.table, this.search),
        ])
        this.rows = list.rows || []
        this.columns = list.columns || (this.rows[0] ? Object.keys(this.rows[0]) : [])
        this.primaryKey = list.primaryKey || []
        this.totalRows = count
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
      if (!this.table || !this.columns.length) {
        return
      }
      const body = {}
      this.columns.forEach((col) => {
        body[col] = this.selected && this.selected[col] != null ? this.selected[col] : 0
      })
      this.saving = true
      try {
        await PeqRawApi.create(this.table, body)
        this.status = "Created"
        await this.load()
      } catch (err) {
        this.status = this.errorText(err)
      } finally {
        this.saving = false
      }
    },
    async saveRow() {
      if (!this.table || !this.selected) {
        return
      }
      this.saving = true
      try {
        const body = this.dirtyRow(this.selected)
        if (!Object.keys(body).length) {
          this.status = "No changes"
          return
        }
        await PeqRawApi.update(this.table, body, this.pk(this.original || this.selected))
        this.status = "Saved"
        await this.load()
      } catch (err) {
        this.status = this.errorText(err)
      } finally {
        this.saving = false
      }
    },
    async deleteRow() {
      if (!this.table || !this.selected) {
        return
      }
      if (!confirm("Delete this " + this.table + " row?")) {
        return
      }
      this.saving = true
      try {
        await PeqRawApi.remove(this.table, this.pk(this.original || this.selected))
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

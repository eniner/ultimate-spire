<template>
  <div class="server-files">
    <div class="row justify-content-center" style="position: absolute; top: 10%; z-index: 40; width: 100%">
      <div class="col-6">
        <info-error-banner
          style="width: 100%"
          :slim="true"
          :notification="notification"
          :error="error"
          @dismiss-error="error = ''"
          @dismiss-notification="notification = ''"
          class="mt-3"
        />
      </div>
    </div>

    <component :is="embedded ? 'div' : 'eq-window'" v-bind="embedded ? {} : { title: 'Server Files' }">
      <p class="text-muted mb-3">
        Point Spire at the folder that holds your server files (usually <span class="mono">quests</span>,
        or the EQEmu server folder that contains it). Use a path on this PC, a mounted AWS/EFS/SMB share,
        or an HTTP(S) directory listing if the files live on another host.
      </p>

      <div class="form-row">
        <div class="form-group col-md-3">
          <label class="form-label">Source</label>
          <select class="form-control" v-model="form.source" @change="onSourceChange">
            <option value="local">Local or mounted folder</option>
            <option value="http">HTTP(S) host</option>
          </select>
        </div>
        <div class="form-group col-md-7">
          <label class="form-label">{{ form.source === 'http' ? 'Base URL' : 'Folder path' }}</label>
          <input
            type="text"
            class="form-control mono privacy-hide"
            v-model="form.root"
            :placeholder="form.source === 'http' ? 'https://files.example.com/quests' : 'C:\\eqemu\\quests'"
            @keyup.enter="connect(true)"
          >
        </div>
        <div class="form-group col-md-2 d-flex align-items-end">
          <button class="btn btn-dark btn-block" type="button" :disabled="busy" @click="connect(true)">
            {{ busy ? 'Connecting…' : 'Connect' }}
          </button>
        </div>
      </div>

      <div class="small text-muted mb-2" v-if="status.root">
        <span v-if="status.ok" class="text-success font-weight-bold">Connected</span>
        <span v-else class="text-warning font-weight-bold">Not reachable</span>
        · {{ status.source || 'local' }}
        · <span class="mono privacy-hide">{{ status.root }}</span>
        · {{ status.writable ? 'editable' : 'read-only' }}
      </div>
      <div class="small text-danger" v-if="status.error">{{ status.error }}</div>

      <div class="mt-3" v-if="suggestions.length">
        <div class="small font-weight-bold mb-2">Detected on this host</div>
        <button
          v-for="path in suggestions"
          :key="path"
          type="button"
          class="btn btn-sm btn-dark mr-2 mb-2 privacy-hide"
          @click="useSuggestion(path)"
        >
          {{ path }}
        </button>
      </div>
    </component>

    <div class="row mt-4" v-if="status.ok">
      <div :class="openFile ? 'col-lg-5' : 'col-12'">
        <eq-window :title="listTitle">
          <div class="d-flex flex-wrap align-items-center mb-3">
            <nav class="crumbs mr-auto">
              <a href="javascript:void(0)" @click="openFolder('')">root</a>
              <span v-for="(crumb, i) in crumbs" :key="crumb.path">
                <span class="text-muted"> / </span>
                <a href="javascript:void(0)" @click="openFolder(crumb.path)">{{ crumb.name }}</a>
              </span>
            </nav>
            <input
              class="form-control form-control-sm search-box"
              v-model="search"
              placeholder="Search this folder…"
              @keyup.enter="runSearch"
            >
            <button class="btn btn-sm btn-dark ml-2" type="button" @click="runSearch">Search</button>
          </div>

          <div class="d-flex mb-3" v-if="status.writable">
            <input class="form-control form-control-sm mono mr-2" v-model="newName" placeholder="new_script.pl">
            <button class="btn btn-sm btn-dark" type="button" :disabled="!newName" @click="createFile">New file</button>
          </div>

          <div v-if="!entries.length" class="text-muted p-4 text-center">No quest or text files in this folder.</div>

          <table class="eq-table eq-highlight-rows bordered" v-if="entries.length">
            <thead>
            <tr>
              <th style="width: 70px"></th>
              <th>Name</th>
              <th class="text-right" style="width: 110px">Size</th>
            </tr>
            </thead>
            <tbody>
            <tr v-for="entry in entries" :key="entry.path" @dblclick="activate(entry)">
              <td class="text-center">
                <button class="btn btn-sm btn-dark" type="button" @click="activate(entry)">
                  {{ entry.isDir ? 'Open' : 'Edit' }}
                </button>
              </td>
              <td>
                <i :class="entry.isDir ? 'fa fa-folder mr-2' : 'fa fa-file-text-o mr-2'"></i>
                {{ entry.path }}
              </td>
              <td class="text-right">{{ entry.isDir ? '' : formatBytes(entry.size) }}</td>
            </tr>
            </tbody>
          </table>
        </eq-window>
      </div>

      <div class="col-lg-7" v-if="openFile">
        <eq-window :title="openFile">
          <div class="mb-2">
            <button class="btn btn-sm btn-dark mr-2" type="button" :disabled="!canSave || saving" @click="saveFile">
              {{ saving ? 'Saving…' : 'Save' }}
            </button>
            <button class="btn btn-sm btn-dark mr-2" type="button" @click="reloadFile">Reload</button>
            <button class="btn btn-sm btn-dark" type="button" @click="closeFile">Close</button>
            <span class="small text-muted ml-2" v-if="!status.writable">Remote HTTP sources are read-only.</span>
            <span class="small text-muted ml-2" v-if="dirty">Unsaved changes</span>
            <span class="small text-muted ml-2" v-if="editorLangLabel">{{ editorLangLabel }}</span>
          </div>
          <editor
            :key="openFile + '-' + aceTheme"
            v-model="fileContent"
            @init="editorInit"
            :lang="editorLang"
            :theme="aceTheme"
            width="100%"
            height="68vh"
            class="file-ace"
          />
        </eq-window>
      </div>
    </div>
  </div>
</template>

<script>
import EqWindow from "@/components/eq-ui/EQWindow.vue"
import InfoErrorBanner from "@/components/InfoErrorBanner.vue"
import {ServerFilesApi} from "@/app/server-files"
import {Theme} from "@/app/theme"
import {EventBus} from "@/app/event-bus/event-bus"

export default {
  name: "ServerFiles",
  components: {EqWindow, InfoErrorBanner, editor: require("vue2-ace-editor")},
  props: {
    embedded: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      busy: false,
      saving: false,
      error: "",
      notification: "",
      form: {source: "local", root: ""},
      status: {ok: false, source: "local", root: "", writable: false, detected: {}},
      folder: "",
      entries: [],
      search: "",
      newName: "",
      openFile: "",
      fileContent: "",
      savedContent: "",
      fileHash: "",
      themeTick: 0,
    }
  },
  computed: {
    suggestions() {
      return (this.status.detected && this.status.detected.suggestions) || []
    },
    crumbs() {
      const parts = String(this.folder || "").split("/").filter(Boolean)
      const out = []
      let path = ""
      for (const name of parts) {
        path = path ? path + "/" + name : name
        out.push({name, path})
      }
      return out
    },
    listTitle() {
      return this.folder ? `Files / ${this.folder}` : "Files"
    },
    dirty() {
      return this.openFile && this.fileContent !== this.savedContent
    },
    canSave() {
      return this.status.writable && this.openFile && this.dirty
    },
    aceTheme() {
      this.themeTick
      return Theme.resolved() === "dark" ? "monokai" : "chrome"
    },
    editorLang() {
      const n = String(this.openFile || "").toLowerCase()
      if (n.endsWith(".lua")) {
        return "lua"
      }
      if (n.endsWith(".py")) {
        return "python"
      }
      if (n.endsWith(".json")) {
        return "json"
      }
      if (n.endsWith(".sql")) {
        return "sql"
      }
      if (n.endsWith(".xml")) {
        return "xml"
      }
      if (n.endsWith(".md")) {
        return "markdown"
      }
      if (n.endsWith(".pl") || n.endsWith(".pm") || n.endsWith(".perl") || n.endsWith(".inc")) {
        return "perl"
      }
      return "text"
    },
    editorLangLabel() {
      return {lua: "Lua", perl: "Perl", python: "Python", json: "JSON", sql: "SQL", xml: "XML", markdown: "Markdown"}[this.editorLang] || ""
    },
  },
  async created() {
    this.onThemeChange = () => {
      this.themeTick += 1
    }
    EventBus.$on("SPIRE_THEME", this.onThemeChange)
    await this.loadStatus()
    const start = this.$route.query.file || this.$route.query.path || ""
    if (this.status.ok && start) {
      if (this.$route.query.file) {
        const file = String(this.$route.query.file)
        const slash = file.lastIndexOf("/")
        this.folder = slash >= 0 ? file.slice(0, slash) : ""
        await this.loadList()
        await this.openEntry({path: file, isDir: false})
        return
      }
      this.folder = String(start)
    }
    if (this.status.ok) {
      await this.loadList()
    }
  },
  beforeDestroy() {
    EventBus.$off("SPIRE_THEME", this.onThemeChange)
  },
  methods: {
    editorInit(editor) {
      require("brace/ext/language_tools")
      require("brace/theme/monokai")
      require("brace/theme/chrome")
      require("brace/mode/lua")
      require("brace/mode/perl")
      require("brace/mode/python")
      require("brace/mode/json")
      require("brace/mode/sql")
      require("brace/mode/xml")
      require("brace/mode/markdown")
      require("brace/mode/text")
      if (editor && editor.setShowPrintMargin) {
        editor.setShowPrintMargin(false)
        editor.setHighlightActiveLine(false)
      }
    },
    onSourceChange() {
      if (this.form.source === "http" && this.form.root && !/^https?:/i.test(this.form.root)) {
        this.form.root = ""
      }
    },
    async loadStatus() {
      try {
        this.status = await ServerFilesApi.status()
        if (!this.form.root && this.status.root) {
          this.form.root = this.status.root
          this.form.source = this.status.source || "local"
        }
      } catch (e) {
        this.error = this.readError(e, "Could not load server files status.")
      }
    },
    async connect(save) {
      this.busy = true
      this.error = ""
      try {
        const source = /^https?:/i.test(this.form.root) ? "http" : this.form.source
        this.status = Object.assign({}, this.status, await ServerFilesApi.connect(source, this.form.root, save))
        this.form.source = this.status.source || source
        this.form.root = this.status.root || this.form.root
        this.notification = save ? "Server files folder saved." : "Folder is reachable."
        this.folder = ""
        this.openFile = ""
        await this.loadList()
      } catch (e) {
        this.error = this.readError(e, "Could not connect to that folder or URL.")
      } finally {
        this.busy = false
      }
    },
    async useSuggestion(path) {
      this.form.root = path
      this.form.source = /^https?:/i.test(path) ? "http" : "local"
      await this.connect(true)
    },
    async loadList() {
      try {
        const r = await ServerFilesApi.list(this.folder)
        this.entries = r.entries || []
        this.status.writable = !!r.writable
      } catch (e) {
        this.error = this.readError(e, "Could not list files.")
        this.entries = []
      }
    },
    async openFolder(path) {
      this.folder = path || ""
      this.search = ""
      await this.loadList()
    },
    async activate(entry) {
      if (entry.isDir) {
        await this.openFolder(entry.path)
        return
      }
      await this.openEntry(entry)
    },
    async openEntry(entry) {
      try {
        const r = await ServerFilesApi.read(entry.path)
        this.openFile = r.path
        this.fileContent = r.content || ""
        this.savedContent = this.fileContent
        this.fileHash = r.hash || ""
      } catch (e) {
        this.error = this.readError(e, "Could not open file.")
      }
    },
    async saveFile() {
      if (!this.canSave) {
        return
      }
      this.saving = true
      try {
        const r = await ServerFilesApi.save(this.openFile, this.fileContent, this.fileHash)
        this.savedContent = this.fileContent
        this.fileHash = (r && r.hash) || this.fileHash
        this.notification = `Saved ${this.openFile}`
      } catch (e) {
        this.error = this.readError(e, "Could not save file.")
      } finally {
        this.saving = false
      }
    },
    async reloadFile() {
      if (!this.openFile) {
        return
      }
      await this.openEntry({path: this.openFile, isDir: false})
    },
    closeFile() {
      this.openFile = ""
      this.fileContent = ""
      this.savedContent = ""
      this.fileHash = ""
    },
    async createFile() {
      const name = String(this.newName || "").replace(/\\/g, "/").split("/").pop()
      if (!name) {
        return
      }
      const path = this.folder ? `${this.folder}/${name}` : name
      try {
        await ServerFilesApi.create(path, "")
        this.newName = ""
        this.notification = `Created ${path}`
        await this.loadList()
        await this.openEntry({path, isDir: false})
      } catch (e) {
        this.error = this.readError(e, "Could not create file.")
      }
    },
    async runSearch() {
      const q = String(this.search || "").trim()
      if (q.length < 2) {
        await this.loadList()
        return
      }
      if ((this.status.source || "local") !== "local") {
        const needle = q.toLowerCase()
        const r = await ServerFilesApi.list(this.folder)
        this.entries = (r.entries || []).filter((e) => String(e.name).toLowerCase().includes(needle))
        return
      }
      try {
        this.entries = await ServerFilesApi.search(q, this.folder)
      } catch (e) {
        this.error = this.readError(e, "Search failed.")
      }
    },
    formatBytes(n) {
      const size = Number(n) || 0
      if (size < 1024) {
        return size + " B"
      }
      if (size < 1024 * 1024) {
        return (size / 1024).toFixed(1) + " KB"
      }
      return (size / (1024 * 1024)).toFixed(1) + " MB"
    },
    readError(e, fallback) {
      return e && e.response && e.response.data && e.response.data.error ? e.response.data.error : fallback
    },
  },
}
</script>

<style scoped>
.server-files {
  position: relative;
}
.mono {
  font-family: ui-monospace, Consolas, monospace;
}
.crumbs a {
  color: inherit;
}
.search-box {
  max-width: 240px;
}
.file-ace {
  width: 100%;
  border: 1px solid #2c3542;
  border-radius: 4px;
  overflow: hidden;
}
.file-ace >>> .ace_editor {
  font-family: ui-monospace, Consolas, monospace;
  font-size: 13px;
  line-height: 1.45;
}
.file-ace >>> .ace_print-margin {
  display: none;
}
</style>

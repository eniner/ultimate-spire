<template>
  <div>
    <div class="zone-editor-bar">
      <label class="mb-0 mr-2">
        <input type="checkbox" v-model="editMode">
        Edit placements
      </label>
      <button class="btn btn-sm btn-dark mr-1" :disabled="!canUndo" @click="undo">Undo</button>
      <button class="btn btn-sm btn-dark mr-1" :disabled="!canRedo" @click="redo">Redo</button>
      <button class="btn btn-sm btn-dark mr-1" :disabled="!dirtyCount || saving" @click="save">
        {{ saving ? 'Saving...' : 'Save to local DB' }}
      </button>
      <button class="btn btn-sm btn-dark mr-1" :disabled="!dirtyCount || saving" @click="discard">Discard</button>
      <router-link class="btn btn-sm btn-dark mr-1" :to="atlasHref">3D Atlas</router-link>
      <router-link class="btn btn-sm btn-dark mr-1" :to="questFilesHref">Quest files</router-link>
      <router-link class="btn btn-sm btn-dark mr-1" :to="zoneControllerHref">Zone Controller</router-link>
      <span class="zone-editor-status">
        {{ dirtyCount ? dirtyCount + ' unsaved spawn2 move(s)' : 'No unsaved moves' }}
        <span v-if="saveMessage"> — {{ saveMessage }}</span>
      </span>
    </div>

    <div class="row">
      <div class="col-7">
        <eq-zone-map
          v-if="zone"
          ref="zoneMap"
          :zone="zone"
          :version="version"
          :edit-mode="editMode"
          @npc-marker-hover="processNpcMarkerHover"
          @npc-marker-select="processNpcMarkerSelect"
          @spell-marker-hover="processSpellMarkerHover"
          @spawn2-moved="onSpawn2Moved"
          @map-loaded="onMapLoaded"
        />
      </div>
      <div class="col-5">

        <eq-window
          v-if="editMode && selectedSpawn"
          class="mb-2 p-2"
          title="Spawn2 placement"
        >
          <div class="small mb-2">
            spawn2 #{{ selectedSpawn.id }}
            <span v-if="selectedNpcName"> — {{ selectedNpcName }}</span>
          </div>
          <div class="form-row">
            <div class="col">
              <label class="small mb-0">X</label>
              <input class="form-control form-control-sm" type="number" step="0.1" :value="selectedSpawn.x" @change="onCoordField('x', $event)">
            </div>
            <div class="col">
              <label class="small mb-0">Y</label>
              <input class="form-control form-control-sm" type="number" step="0.1" :value="selectedSpawn.y" @change="onCoordField('y', $event)">
            </div>
            <div class="col">
              <label class="small mb-0">Z</label>
              <input class="form-control form-control-sm" type="number" step="0.1" :value="selectedSpawn.z" @change="onCoordField('z', $event)">
            </div>
            <div class="col">
              <label class="small mb-0">Heading</label>
              <input class="form-control form-control-sm" type="number" step="1" min="0" max="511" :value="selectedSpawn.heading" @change="onCoordField('heading', $event)">
            </div>
          </div>
          <div class="mt-2">
            <button class="btn btn-sm btn-dark mr-1" @click="nudgeHeading(-45)">Heading -45</button>
            <button class="btn btn-sm btn-dark" @click="nudgeHeading(45)">Heading +45</button>
          </div>
        </eq-window>

        <!-- Zone Card -->
        <eq-zone-card-preview
          style="height: 96vh"
          v-show="selectorActive['zone-preview'] && zoneData"
          :zone="zoneData"
        />

        <eq-window
          v-if="!isZoneCardActive()"
          class="text-center"
        >
          <b-button
            class="btn-dark btn-sm btn-dark"
            @click="setSelectorActive('zone-preview', true)"
          >
            <i class="fa fa-chevron-up"></i> Return to Zone
          </b-button>
        </eq-window>

        <!-- NPC -->
        <eq-window
          class="fade-in"
          id="preview-pane"
          :style="'max-height: ' + (isZoneCardActive() ? '95' : '87') + 'vh; overflow-y: scroll; overflow-x: hidden'"
          v-if="selectorActive['npc-hover'] && npc"
        >
          <eq-npc-card-preview
            :npc="npc"
          />
        </eq-window>

        <!-- Spell -->
        <eq-window
          class="fade-in"
          id="preview-pane"
          :style="'max-height: ' + (isZoneCardActive() ? '95' : '87') + 'vh; overflow-y: scroll; overflow-x: hidden'"
          v-if="selectorActive['spell-hover'] && spell"
        >
          <eq-spell-preview
            :spell-data="spell"
          />
        </eq-window>

      </div>
    </div>
  </div>
</template>

<script>
import ContentArea       from "../../components/layout/ContentArea";
import EqWindow          from "../../components/eq-ui/EQWindow";
import {Navbar}          from "../../app/navbar";
import EqZoneMap         from "../../components/EqZoneMap";
import EqNpcCardPreview  from "../../components/preview/EQNpcCardPreview";
import EqSpellPreview    from "../../components/preview/EQSpellCardPreview";
import {Zones}           from "../../app/zones";
import EqZoneCardPreview from "../../components/preview/EQZoneCardPreview";
import {EventBus}        from "../../app/event-bus/event-bus";
import {Npcs}            from "../../app/npcs";
import {ZoneEditorApi, coordsEqual} from "../../app/zone-editor";
import {roundCoord}      from "../../app/eq-coords";

const MILLISECONDS_BEFORE_WINDOW_RESET = 5000;

export default {
  name: "Zone",
  components: { EqZoneCardPreview, EqSpellPreview, EqNpcCardPreview, EqZoneMap, EqWindow, ContentArea },
  data() {
    return {
      zone: "",
      version: "",

      zoneData: {},

      selectorActive: {},
      editMode: false,
      selectedSpawn: null,
      undoStack: [],
      redoStack: [],
      dirtyOriginals: {},
      dirtyCurrent: {},
      saving: false,
      saveMessage: "",
    }
  },
  computed: {
    dirtyCount() {
      return Object.keys(this.dirtyCurrent).filter((id) => {
        return this.dirtyOriginals[id] && !coordsEqual(this.dirtyOriginals[id], this.dirtyCurrent[id])
      }).length
    },
    canUndo() {
      return this.undoStack.length > 0
    },
    canRedo() {
      return this.redoStack.length > 0
    },
    selectedNpcName() {
      if (!this.npc || !this.npc.name) {
        return ""
      }
      return Npcs.getCleanName(this.npc.name)
    },
    atlasHref() {
      const v = this.version === "" || this.version == null ? 0 : this.version
      return "/zone/" + this.zone + "/atlas?v=" + v
    },
    questFilesHref() {
      return "/admin/server-files?path=" + encodeURIComponent(this.zone || "")
    },
    zoneControllerHref() {
      const id = this.zoneData && this.zoneData.zoneidnumber
      if (id) {
        return "/zones/controller/" + id
      }
      return "/zones/controller"
    },
  },
  beforeDestroy() {
    Navbar.expand()
    EventBus.$off("NPC_SHOW_CARD", this.handleNpcShowCardEvent);
    window.removeEventListener("keydown", this.onEditorKeydown)
  },
  created() {
    this.npc           = {}
    this.lastResetTime = Date.now()

    EventBus.$on("NPC_SHOW_CARD", this.handleNpcShowCardEvent);
  },
  watch: {
    '$route'() {
      console.log("route trigger")
      this.init()
    },
  },

  mounted() {
    this.init()
    window.addEventListener("keydown", this.onEditorKeydown)
  },

  methods: {

    isZoneCardActive() {
      return Object.keys(this.selectorActive).length > 0 && this.selectorActive['zone-preview']
    },

    handleNpcShowCardEvent(e) {
      this.processNpcMarkerHover(e)
    },

    previewZone() {
      console.log("[Zone] previewZone trigger")
      this.setSelectorActive('zone-preview')
    },

    async init() {
      this.npc   = {}
      this.spell = {}
      this.resetSelectors()
      this.resetEditorState()

      Navbar.collapse()

      this.zone    = this.$route.params.zone
      this.version = this.$route.query.v != null && this.$route.query.v !== "" ? this.$route.query.v : "0"

      this.zoneData = (await Zones.getZoneByShortName(this.zone))

      this.setSelectorActive('zone-preview', true)
    },

    resetEditorState() {
      this.selectedSpawn = null
      this.undoStack = []
      this.redoStack = []
      this.dirtyOriginals = {}
      this.dirtyCurrent = {}
      this.saveMessage = ""
    },

    shouldReset() {
      if (this.editMode && this.selectedSpawn) {
        return false
      }
      return (Date.now() - this.lastResetTime) > MILLISECONDS_BEFORE_WINDOW_RESET
    },

    resetSelectors() {
      for (const [k, v] of Object.entries(this.selectorActive)) {
        this.selectorActive[k] = false
      }
    },

    setSelectorActive(selector, force = false) {
      if (this.selectorActive[selector] && !force) {
        return
      }

      if (this.shouldReset() || force) {
        this.lastResetTime = Date.now()
        this.resetSelectors()
        this.selectorActive[selector] = true
        this.$forceUpdate()
        return
      }

      console.log(
        "[Zone] Tried to set selector [%s] but reset time was not met (%s) ms remaining",
        selector,
        MILLISECONDS_BEFORE_WINDOW_RESET - (Date.now() - this.lastResetTime)
      )
    },

    processSpellMarkerHover(s) {
      this.spell = {}
      this.setSelectorActive("spell-hover", true)
      this.spell = s

      const t = document.getElementById("preview-pane");
      if (t) {
        t.scrollTop = 0;
      }
    },

    processNpcMarkerHover(n) {
      this.npc = {}
      this.setSelectorActive("npc-hover", true)
      this.npc = n

      const t = document.getElementById("preview-pane");
      if (t) {
        t.scrollTop = 0;
      }
    },

    processNpcMarkerSelect(payload) {
      this.processNpcMarkerHover(payload.npc)
      this.selectedSpawn = Object.assign({}, payload.spawn2)
    },

    onMapLoaded() {
      if (!this.dirtyCount) {
        this.resetEditorState()
      }
    },

    copyCoords(c) {
      return {
        x: roundCoord(Number(c.x) || 0),
        y: roundCoord(Number(c.y) || 0),
        z: roundCoord(Number(c.z) || 0),
        heading: roundCoord(Number(c.heading) || 0),
      }
    },

    applyToMap(id, coords) {
      if (this.$refs.zoneMap && this.$refs.zoneMap.applySpawn2Placement) {
        this.$refs.zoneMap.applySpawn2Placement(id, coords)
      }
      if (this.selectedSpawn && this.selectedSpawn.id === id) {
        this.selectedSpawn = Object.assign({}, this.selectedSpawn, coords)
      }
    },

    rememberDirty(id, from, to) {
      if (!this.dirtyOriginals[id]) {
        this.$set(this.dirtyOriginals, id, this.copyCoords(from))
      }
      this.$set(this.dirtyCurrent, id, this.copyCoords(to))
    },

    pushUndo(entry) {
      this.undoStack.push(entry)
      this.redoStack = []
      this.rememberDirty(entry.id, entry.from, entry.to)
      this.saveMessage = ""
    },

    onSpawn2Moved(e) {
      this.pushUndo({
        id: e.id,
        from: this.copyCoords(e.from),
        to: this.copyCoords(e.to)
      })
      this.selectedSpawn = Object.assign({}, e.spawn2, e.to)
      if (e.npc) {
        this.processNpcMarkerHover(e.npc)
      }
    },

    onCoordField(field, event) {
      if (!this.selectedSpawn) {
        return
      }
      const next = Number(event.target.value)
      if (!isFinite(next)) {
        return
      }
      const from = this.copyCoords(this.selectedSpawn)
      const to = this.copyCoords(this.selectedSpawn)
      to[field] = roundCoord(next)
      if (field === "heading") {
        to.heading = ((to.heading % 512) + 512) % 512
      }
      this.applyToMap(this.selectedSpawn.id, to)
      this.pushUndo({id: this.selectedSpawn.id, from, to})
    },

    nudgeHeading(delta) {
      if (!this.selectedSpawn) {
        return
      }
      const from = this.copyCoords(this.selectedSpawn)
      const to = this.copyCoords(this.selectedSpawn)
      to.heading = ((from.heading + delta) % 512 + 512) % 512
      this.applyToMap(this.selectedSpawn.id, to)
      this.pushUndo({id: this.selectedSpawn.id, from, to})
    },

    undo() {
      const entry = this.undoStack.pop()
      if (!entry) {
        return
      }
      this.redoStack.push(entry)
      this.applyToMap(entry.id, entry.from)
      this.$set(this.dirtyCurrent, entry.id, this.copyCoords(entry.from))
    },

    redo() {
      const entry = this.redoStack.pop()
      if (!entry) {
        return
      }
      this.undoStack.push(entry)
      this.applyToMap(entry.id, entry.to)
      this.$set(this.dirtyCurrent, entry.id, this.copyCoords(entry.to))
    },

    async save() {
      const changes = []
      for (const id of Object.keys(this.dirtyCurrent)) {
        const original = this.dirtyOriginals[id]
        const current = this.dirtyCurrent[id]
        if (!original || !current || coordsEqual(original, current)) {
          continue
        }
        changes.push({
          table: "spawn2",
          id: Number(id),
          x: current.x,
          y: current.y,
          z: current.z,
          heading: current.heading
        })
      }
      if (!changes.length) {
        this.saveMessage = "Nothing to save"
        return
      }

      this.saving = true
      this.saveMessage = ""
      try {
        const r = await ZoneEditorApi.savePlacements({
          zone: this.zone,
          version: parseInt(this.version, 10) || 0,
          changes
        })
        const updated = r && r.data ? r.data.updated : changes.length
        this.dirtyOriginals = {}
        this.dirtyCurrent = {}
        this.undoStack = []
        this.redoStack = []
        this.saveMessage = "Saved " + updated + " spawn2 row(s) to local peq"
      } catch (err) {
        const msg = err && err.response && err.response.data && err.response.data.error
          ? err.response.data.error
          : (err && err.message ? err.message : "Save failed")
        this.saveMessage = msg
      } finally {
        this.saving = false
      }
    },

    discard() {
      this.resetEditorState()
      if (this.$refs.zoneMap && this.$refs.zoneMap.loadMap) {
        this.$refs.zoneMap.loadMap()
      }
    },

    onEditorKeydown(e) {
      if (!this.editMode) {
        return
      }
      const key = (e.key || "").toLowerCase()
      if ((e.ctrlKey || e.metaKey) && key === "z") {
        e.preventDefault()
        if (e.shiftKey) {
          this.redo()
        } else {
          this.undo()
        }
        return
      }
      if ((e.ctrlKey || e.metaKey) && key === "y") {
        e.preventDefault()
        this.redo()
        return
      }
      if ((e.ctrlKey || e.metaKey) && key === "s") {
        e.preventDefault()
        this.save()
      }
    }
  }
}
</script>

<style>
.zone-editor-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  margin-bottom: 8px;
  background: rgba(18, 22, 27, 0.92);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 6px;
  color: #e7eaef;
  font-size: 13px;
}

.zone-editor-status {
  opacity: 0.9;
}
</style>

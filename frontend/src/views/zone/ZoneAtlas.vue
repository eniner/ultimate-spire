<template>
  <div class="atlas-page" :class="{ 'atlas-bright': brightLighting, 'atlas-walk': cameraMode === 'walk' }">
    <div class="atlas-bar">
      <div class="atlas-brand">3D Atlas</div>
      <div class="atlas-zone mono">{{ zone || "—" }}</div>
      <div class="atlas-modes">
        <button class="btn btn-sm" :class="cameraMode === 'orbit' ? 'btn-warning' : 'btn-dark'" type="button" @click="setCameraMode('orbit')">Orbit</button>
        <button class="btn btn-sm" :class="cameraMode === 'walk' ? 'btn-warning' : 'btn-dark'" type="button" @click="setCameraMode('walk')">Walk</button>
        <button class="btn btn-sm" :class="editMode ? 'btn-warning' : 'btn-dark'" type="button" @click="editMode = !editMode">Edit</button>
      </div>
      <label class="atlas-check">
        <input type="checkbox" v-model="showScenery" @change="onToggleScenery">
        Client scenery
      </label>
      <label class="atlas-check">
        <input type="checkbox" v-model="showNpcs" @change="onToggleNpcs">
        NPCs
      </label>
      <label class="atlas-check">
        <input type="checkbox" v-model="showObjects" @change="onToggleObjects">
        Server objects
      </label>
      <label class="atlas-check">
        <input type="checkbox" v-model="showPaths" @change="redrawPaths">
        Path grids
      </label>
      <label class="atlas-check">
        <input type="checkbox" v-model="animatePathing">
        Move on path
      </label>
      <label class="atlas-check">
        <input type="checkbox" v-model="brightLighting" @change="onToggleBrightLighting">
        Bright lighting
      </label>
      <label class="atlas-check">
        <input type="checkbox" v-model="wireframe" @change="onToggleWireframe">
        Wireframe
      </label>
      <button class="btn btn-sm btn-dark" type="button" @click="resetCamera">Reset camera</button>
      <router-link class="btn btn-sm btn-dark" :to="mapHref">2D Map</router-link>
      <router-link class="btn btn-sm btn-dark" :to="sageHref">Sage</router-link>
      <button class="btn btn-sm btn-dark" type="button" @click="openSageWindow">Sage window</button>
      <button class="btn btn-sm btn-dark" type="button" :disabled="!canUndo" @click="undo">Undo</button>
      <button class="btn btn-sm btn-warning" type="button" :disabled="!dirtyCount || saving" @click="saveEdits">{{ saving ? "Saving…" : ("Save " + dirtyCount) }}</button>
      <span class="atlas-status" :class="{ error: !!error }">{{ status }}</span>
    </div>
    <div class="atlas-body">
      <aside v-if="editMode" class="atlas-catalog">
        <div class="atlas-catalog-title">Catalog</div>
        <p class="atlas-help">Walk or orbit, click the ground to place. Drag / G move, R / Q / E rotate. Save writes peq and, if you edited client scenery, Lantern <code>object_instances.txt</code>.</p>
        <label class="atlas-field">
          <span>Find NPC</span>
          <input v-model="npcQuery" placeholder="name or id" @input="onNpcQuery">
        </label>
        <div class="atlas-list">
          <button
            v-for="npc in npcHits"
            :key="'npc-' + npc.id"
            type="button"
            class="atlas-item"
            :class="{ active: placeKind === 'npc' && placeItem && placeItem.id === npc.id }"
            @click="selectPlaceNpc(npc)"
          >
            {{ npc.name }} <span class="mono">#{{ npc.id }}</span>
          </button>
        </div>
        <label class="atlas-field">
          <span>Ground items (icons)</span>
          <input v-model="itemQuery" placeholder="item name" @input="onItemQuery">
        </label>
        <div class="atlas-list">
          <button
            v-for="item in itemHits"
            :key="'item-' + item.id"
            type="button"
            class="atlas-item atlas-item-icon"
            :class="{ active: placeKind === 'item' && placeItem && placeItem.id === item.id }"
            @click="selectPlaceItem(item)"
          >
            <span v-if="item.icon" :class="'item-' + item.icon + '-sm'"></span>
            {{ item.name }} <span class="mono">#{{ item.id }}</span>
          </button>
        </div>
        <label class="atlas-field">
          <span>Object models</span>
          <input v-model="objectQuery" placeholder="filter lantern / object name">
        </label>
        <div class="atlas-place-as">
          <label><input type="radio" value="scenery" v-model="placeAs"> Client scenery</label>
          <label><input type="radio" value="object" v-model="placeAs"> Server object</label>
        </div>
        <div class="atlas-list">
          <button
            v-for="model in filteredModels"
            :key="'mdl-' + model.modelName"
            type="button"
            class="atlas-item"
            :class="{ active: (placeKind === 'object' || placeKind === 'scenery') && placeItem && placeItem.modelName === model.modelName }"
            @click="selectPlaceObject(model)"
          >
            {{ model.modelName }}
          </button>
        </div>
        <div v-if="selected && selected.table === 'object'" class="atlas-icon-edit">
          <span>Object icon</span>
          <span v-if="selected.icon" :class="'item-' + selected.icon + '-sm'"></span>
          <input v-model.number="selected.icon" type="number" min="0" @change="onSelectedIcon">
        </div>
        <button v-if="placeKind" class="btn btn-sm btn-dark btn-block" type="button" @click="clearPlace">Cancel place</button>
      </aside>
      <div class="atlas-stage">
        <div ref="viewport" class="atlas-viewport" :class="{ placing: !!placeKind }"></div>
        <div v-if="cameraMode === 'walk'" class="atlas-walk-hint">
          Click the view to look around. WASD move · Shift sprint · Space up · Esc unlock. Click ground to place when a catalog item is selected.
        </div>
        <div v-if="selected" class="atlas-tip">
          <strong>{{ selected.name }}</strong>
          <span class="mono"> {{ selected.table }} #{{ selected.id }}  X {{ fmt(selected.x) }} Y {{ fmt(selected.y) }} Z {{ fmt(selected.z) }} H {{ fmt(selected.heading) }}</span>
          <span v-if="selected.icon" :class="'item-' + selected.icon + '-sm atlas-tip-icon'"></span>
          <span v-if="selected.pathLabel"> · {{ selected.pathLabel }}</span>
        </div>
        <div v-if="saveMessage" class="atlas-save">{{ saveMessage }}</div>
      </div>
    </div>
  </div>
</template>

<script>
import * as THREE from "three"
import {OrbitControls} from "three/examples/jsm/controls/OrbitControls"
import {PointerLockControls} from "three/examples/jsm/controls/PointerLockControls"
import {TransformControls} from "three/examples/jsm/controls/TransformControls"
import {GLTFLoader} from "three/examples/jsm/loaders/GLTFLoader"
import {Navbar} from "../../app/navbar"
import {LanternApi} from "../../app/lantern"
import {NPC_COORD_TRANSFORMS, WORLD_SCALE, degToRad, eqHeadingToYaw, eqToWorld, lanternInstanceToWorld, radToDeg, worldToEq, worldToLantern, yawToEqHeading} from "../../app/lantern-coords"
import {Spawn} from "../../app/spawn"
import {Npcs} from "../../app/npcs"
import {Zones} from "../../app/zones"
import {ZoneEditorApi, coordsEqual} from "../../app/zone-editor"
import {currentPathPose, eqWalkSpeed, gridMoves, gridTypeLabel, loadZonePathing, makePathState, stepPathState} from "../../app/zone-pathing"
import {createNpcSpawn, createZoneObject, listZoneObjects, searchItems, searchNpcs, updateZoneObjectIcon} from "../../app/zone-atlas-place"
import {createAtlasSageBridge, sageHref} from "../../app/atlas-sage-bridge"

const NPC_SPRITE_LOCAL = "/eq-asset-preview-master/assets/npc_models/"
const NPC_SPRITE_REMOTE = "https://raw.githubusercontent.com/EQEmuTools/eq-asset-preview/master/assets/npc_models/"
const ITEM_ICON_LOCAL = "/eq-asset-preview-master/assets/item_icons/"
const ITEM_ICON_REMOTE = "https://raw.githubusercontent.com/EQEmuTools/eq-asset-preview/master/assets/item_icons/"

export default {
  name: "ZoneAtlas",
  data() {
    return {
      zone: "",
      version: 0,
      zoneId: 0,
      status: "Starting…",
      error: "",
      showScenery: true,
      showNpcs: true,
      showObjects: true,
      showPaths: true,
      animatePathing: true,
      brightLighting: true,
      wireframe: false,
      cameraMode: "orbit",
      editMode: true,
      selected: null,
      npcQuery: "",
      objectQuery: "",
      itemQuery: "",
      npcHits: [],
      itemHits: [],
      lanternModels: [],
      placeKind: "",
      placeAs: "scenery",
      placeItem: null,
      sceneryDirty: false,
      dirtyCount: 0,
      canUndo: false,
      saving: false,
      saveMessage: "",
    }
  },
  computed: {
    mapHref() {
      const v = this.version === "" || this.version == null ? 0 : this.version
      return `/zone/${this.zone}?v=${v}`
    },
    sageHref() {
      return sageHref(this.zone, this.selected)
    },
    filteredModels() {
      const q = String(this.objectQuery || "").toLowerCase()
      return (this.lanternModels || []).filter((m) => !q || m.modelName.indexOf(q) !== -1).slice(0, 80)
    },
  },
  watch: {
    "$route.fullPath"() {
      const zone = String(this.$route.params.zone || "").toLowerCase()
      const version = this.$route.query.v == null || this.$route.query.v === "" ? 0 : Number(this.$route.query.v)
      if (zone !== this.zone || Number(version) !== Number(this.version)) {
        this.zone = zone
        this.version = version
        this.boot()
        return
      }
      this.applyPendingFocus()
    },
  },
  mounted() {
    Navbar.collapse()
    this.zone = String(this.$route.params.zone || "").toLowerCase()
    this.version = this.$route.query.v == null || this.$route.query.v === "" ? 0 : Number(this.$route.query.v)
    this.objectUrls = []
    this.templateCache = new Map()
    this.surfaceTargets = []
    this.npcMarkers = []
    this.objectMarkers = []
    this.sceneryMarkers = []
    this.sceneryInstances = []
    this.pathGrids = new Map()
    this.dirty = {}
    this.undoStack = []
    this.keys = {}
    this.lastTick = performance.now()
    this.loadToken = 0
    this.sageBridge = createAtlasSageBridge("atlas", (msg) => this.onSageBridge(msg))
    this.initRenderer()
    this.bindEditorInput()
    this.boot()
  },
  beforeDestroy() {
    this.loadToken += 1
    if (this.sageBridge) {
      this.sageBridge.dispose()
      this.sageBridge = null
    }
    window.removeEventListener("resize", this.onResize)
    window.removeEventListener("keydown", this.onKeyDown)
    window.removeEventListener("keyup", this.onKeyUp)
    if (this.rendererEl && this.rendererEl.domElement) {
      this.rendererEl.domElement.removeEventListener("click", this.onCanvasClick)
      this.rendererEl.domElement.removeEventListener("pointerdown", this.onPointerDown)
      this.rendererEl.domElement.removeEventListener("pointermove", this.onPointerMove)
      this.rendererEl.domElement.removeEventListener("pointerup", this.onPointerUp)
    }
    this.disposeScene()
    this.revokeUrls()
    if (this.rendererEl) {
      this.rendererEl.dispose()
      if (this.rendererEl.domElement && this.rendererEl.domElement.parentNode) {
        this.rendererEl.domElement.parentNode.removeChild(this.rendererEl.domElement)
      }
    }
    Navbar.expand()
  },
  methods: {
    fmt(n) {
      return Number(n || 0).toFixed(1)
    },
    setStatus(text, isError = false) {
      this.status = text
      this.error = isError ? text : ""
    },
    initRenderer() {
      const el = this.$refs.viewport
      this.scene = new THREE.Scene()
      this.scene.background = new THREE.Color(0x2b3544)
      this.camera = new THREE.PerspectiveCamera(55, 1, 0.1, 250000)
      this.camera.position.set(400, 260, 400)
      this.rendererEl = new THREE.WebGLRenderer({antialias: true})
      this.rendererEl.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2))
      this.rendererEl.outputEncoding = THREE.sRGBEncoding
      this.rendererEl.toneMapping = THREE.ACESFilmicToneMapping
      this.rendererEl.toneMappingExposure = 1.55
      el.appendChild(this.rendererEl.domElement)
      this.controls = new OrbitControls(this.camera, this.rendererEl.domElement)
      this.controls.enableDamping = true
      this.controls.dampingFactor = 0.06
      this.walkControls = new PointerLockControls(this.camera, this.rendererEl.domElement)
      this.transform = new TransformControls(this.camera, this.rendererEl.domElement)
      this.transform.setSize(0.85)
      this.transform.addEventListener("dragging-changed", (e) => {
        if (this.controls) {
          this.controls.enabled = !e.value && this.cameraMode === "orbit"
        }
        if (!e.value) {
          this.commitTransform()
        }
      })
      this.transform.addEventListener("objectChange", () => this.syncSelectedFromObject())
      this.scene.add(this.transform)
      this.loader = new GLTFLoader()
      this.textureLoader = new THREE.TextureLoader()
      this.textureLoader.crossOrigin = "anonymous"
      this.npcTextureCache = new Map()
      this.itemIconCache = new Map()
      this.raycaster = new THREE.Raycaster()
      this.pointer = new THREE.Vector2()

      this.hemi = new THREE.HemisphereLight(0xf2f6ff, 0x8d7b62, 1.35)
      this.ambient = new THREE.AmbientLight(0xffffff, 1.15)
      this.sun = new THREE.DirectionalLight(0xfff6e8, 1.7)
      this.sun.position.set(140, 240, 90)
      this.fill = new THREE.DirectionalLight(0xd4e4ff, 0.9)
      this.fill.position.set(-160, 140, -110)
      this.headlamp = new THREE.PointLight(0xffffff, 0.7, 0, 2)
      this.scene.add(this.hemi)
      this.scene.add(this.ambient)
      this.scene.add(this.sun)
      this.scene.add(this.fill)
      this.camera.add(this.headlamp)
      this.scene.add(this.camera)
      this.applyLighting()

      this.zoneGroup = new THREE.Group()
      this.sceneryGroup = new THREE.Group()
      this.npcGroup = new THREE.Group()
      this.objectGroup = new THREE.Group()
      this.pathGroup = new THREE.Group()
      this.ghost = null
      this.scene.add(this.zoneGroup)
      this.scene.add(this.sceneryGroup)
      this.scene.add(this.npcGroup)
      this.scene.add(this.objectGroup)
      this.scene.add(this.pathGroup)

      this.onResize = this.onResize.bind(this)
      this.onCanvasClick = this.onCanvasClick.bind(this)
      this.onPointerDown = this.onPointerDown.bind(this)
      this.onPointerMove = this.onPointerMove.bind(this)
      this.onPointerUp = this.onPointerUp.bind(this)
      this.onKeyDown = this.onKeyDown.bind(this)
      this.onKeyUp = this.onKeyUp.bind(this)
      window.addEventListener("resize", this.onResize)
      this.rendererEl.domElement.addEventListener("click", this.onCanvasClick)
      this.rendererEl.domElement.addEventListener("pointerdown", this.onPointerDown)
      this.rendererEl.domElement.addEventListener("pointermove", this.onPointerMove)
      this.rendererEl.domElement.addEventListener("pointerup", this.onPointerUp)
      this.onResize()
      this.tick = this.tick.bind(this)
      this.tick()
    },
    bindEditorInput() {
      window.addEventListener("keydown", this.onKeyDown)
      window.addEventListener("keyup", this.onKeyUp)
    },
    onResize() {
      const el = this.$refs.viewport
      if (!el || !this.rendererEl) {
        return
      }
      const w = Math.max(320, el.clientWidth)
      const h = Math.max(240, el.clientHeight)
      this.camera.aspect = w / h
      this.camera.updateProjectionMatrix()
      this.rendererEl.setSize(w, h, false)
    },
    tick() {
      this.raf = requestAnimationFrame(this.tick)
      const now = performance.now()
      const dt = Math.min(0.05, (now - this.lastTick) / 1000)
      this.lastTick = now
      if (this.cameraMode === "walk") {
        this.stepWalk(dt)
      } else if (this.controls) {
        this.controls.update()
      }
      this.stepPathing(dt)
      this.updateNpcSpriteScales()
      if (this.rendererEl && this.scene && this.camera) {
        this.rendererEl.render(this.scene, this.camera)
      }
    },
    async boot() {
      const token = ++this.loadToken
      try {
        const zoneRow = await Zones.getZoneByShortName(this.zone)
        this.zoneId = Number(zoneRow && zoneRow.zoneidnumber) || 0
        this.setStatus("Checking Lantern exports…")
        const status = await LanternApi.status()
        if (token !== this.loadToken) {
          return
        }
        if (!status.ok) {
          this.setStatus(status.error || "Lantern export root not found.", true)
          return
        }
        this.setStatus(`Lantern root ready (${status.zonesWithModels} zones). Loading ${this.zone}…`)
        await this.loadZone(token)
      } catch (err) {
        this.setStatus(err && err.message ? err.message : "Failed to start Atlas.", true)
      }
    },
    async loadZone(token) {
      const info = await LanternApi.zone(this.zone)
      if (token !== this.loadToken) {
        return
      }
      if (!info || !info.hasMesh) {
        this.setStatus(`No Lantern zone mesh for ${this.zone}. Expected Zone/${this.zone}.glb`, true)
        return
      }
      this.setStatus(`Loading ${this.zone} terrain…`)
      const meshUrl = await LanternApi.loadFileUrl(info.modelRelPath)
      this.objectUrls.push(meshUrl)
      const gltf = await this.loader.loadAsync(meshUrl)
      if (token !== this.loadToken) {
        return
      }
      this.clearGroup(this.zoneGroup)
      this.surfaceTargets = []
      this.zoneGroup.add(gltf.scene)
      gltf.scene.traverse((node) => {
        if (node.isMesh && node.geometry) {
          const mats = Array.isArray(node.material) ? node.material : [node.material]
          mats.forEach((m) => this.prepareMaterial(m))
          this.surfaceTargets.push(node)
        }
      })
      this.frameObject(gltf.scene)
      if (this.zoneId) {
        this.pathGrids = await loadZonePathing(this.zoneId)
      }
      this.lanternModels = await LanternApi.models(this.zone).catch(() => [])
      if (this.showNpcs) {
        await this.loadNpcs(token)
      }
      if (this.showObjects) {
        await this.loadObjects(token)
      }
      if (this.showScenery) {
        await this.loadScenery(token, info.instanceCount || 0)
      } else {
        this.setStatus(this.finalStatus())
      }
      this.redrawPaths()
      this.applyPendingFocus()
      this.broadcastZone()
    },
    async loadNpcs(token) {
      const rows = await Spawn.getByZone(this.zone, this.version, false)
      if (token !== this.loadToken) {
        return
      }
      const spawns = []
      for (const spawn2 of rows || []) {
        let name = `spawn2 ${spawn2.id}`
        let npc = {race: 1, gender: 2, texture: 0, helmtexture: 0, size: 6, runspeed: 1.25}
        if (spawn2.spawnentries) {
          for (const entry of spawn2.spawnentries) {
            if (entry.npc_type) {
              const n = entry.npc_type
              if (n.name) {
                name = Npcs.getCleanName(n.name)
              }
              npc = {
                race: Number(n.race) || 1,
                gender: Number.isFinite(Number(n.gender)) ? Number(n.gender) : 2,
                texture: Number(n.texture) || 0,
                helmtexture: Number(n.helmtexture) || 0,
                size: Number(n.size) || 6,
                runspeed: Number(n.runspeed) || 1.25,
                walkspeed: Number(n.walkspeed) || 0,
              }
              break
            }
          }
        }
        spawns.push({
          id: spawn2.id,
          name,
          x: Number(spawn2.x) || 0,
          y: Number(spawn2.y) || 0,
          z: Number(spawn2.z) || 0,
          heading: Number(spawn2.heading) || 0,
          pathgrid: Number(spawn2.pathgrid) || 0,
          pathWhenIdle: Number(spawn2.path_when_zone_idle) || 0,
          ...npc,
        })
      }
      this.coordMap = this.detectCoordMap(spawns)
      this.clearGroup(this.npcGroup)
      this.npcMarkers = []
      for (const spawn of spawns) {
        this.npcMarkers.push(this.makeNpcMarker(spawn, token))
      }
      this.npcGroup.visible = this.showNpcs
      this.setStatus(this.finalStatus())
    },
    makeNpcMarker(spawn, token) {
      const pos = eqToWorld(spawn.x, spawn.y, spawn.z, this.coordMap || NPC_COORD_TRANSFORMS[1])
      const grounded = this.snapY(pos.x, pos.y, pos.z)
      const dims = this.npcSpriteSize(spawn)
      const marker = new THREE.Sprite(new THREE.SpriteMaterial({
        map: this.placeholderNpcTexture(),
        transparent: true,
        depthWrite: false,
      }))
      marker.scale.set(dims.w, dims.h, 1)
      marker.position.set(grounded.x, grounded.y + dims.h * 0.5, grounded.z)
      marker.userData.kind = "npc"
      marker.userData.spawn = spawn
      marker.userData.groundY = grounded.y
      marker.userData.path = this.makePathState(spawn)
      this.npcGroup.add(marker)
      this.getNpcTexture(spawn).then((tex) => {
        if (!tex || (token && token !== this.loadToken) || !marker.parent) {
          return
        }
        marker.material.map = tex
        marker.material.needsUpdate = true
      })
      return marker
    },
    makePathState(spawn) {
      const grid = this.pathGrids && this.pathGrids.get(Number(spawn.pathgrid))
      return makePathState(grid, spawn)
    },
    async loadObjects(token) {
      if (!this.zoneId) {
        return
      }
      const rows = await listZoneObjects(this.zoneId, this.version)
      if (token !== this.loadToken) {
        return
      }
      this.clearGroup(this.objectGroup)
      this.objectMarkers = []
      for (const row of rows || []) {
        this.objectMarkers.push(await this.makeObjectMarker(row))
      }
      this.objectGroup.visible = this.showObjects
    },
    async makeObjectMarker(row) {
      const name = String((row.objectname && row.objectname.String) || row.objectname || "object")
      const icon = Number(row.icon) || 0
      const itemid = Number(row.itemid) || 0
      const model = (this.lanternModels || []).find((m) => m.modelName === name.toLowerCase())
      const pos = eqToWorld(Number(row.xpos) || 0, Number(row.ypos) || 0, Number(row.zpos) || 0, this.coordMap || NPC_COORD_TRANSFORMS[1])
      const grounded = this.snapY(pos.x, pos.y, pos.z)
      let obj
      if (model) {
        const tpl = await this.getTemplate(model.modelRelPath)
        obj = tpl ? tpl.clone(true) : this.fallbackObjectMesh()
        if (tpl) {
          obj.scale.multiplyScalar(WORLD_SCALE)
        }
      } else {
        obj = this.fallbackObjectMesh()
      }
      obj.position.set(grounded.x, grounded.y, grounded.z)
      obj.rotation.y = eqHeadingToYaw(row.heading)
      obj.userData.kind = "object"
      obj.userData.row = {
        id: row.id,
        name,
        x: Number(row.xpos) || 0,
        y: Number(row.ypos) || 0,
        z: Number(row.zpos) || 0,
        heading: Number(row.heading) || 0,
        objectname: name,
        icon,
        itemid,
      }
      if (icon) {
        this.attachObjectIcon(obj, icon)
      }
      this.objectGroup.add(obj)
      return obj
    },
    attachObjectIcon(obj, icon) {
      if (!obj || !icon) {
        return
      }
      if (obj.userData.iconSprite) {
        obj.remove(obj.userData.iconSprite)
        obj.userData.iconSprite = null
      }
      const sprite = new THREE.Sprite(new THREE.SpriteMaterial({
        map: this.placeholderNpcTexture(),
        transparent: true,
        depthWrite: false,
      }))
      sprite.scale.set(0.9, 0.9, 1)
      sprite.position.set(0, 1.4, 0)
      sprite.userData.kind = "object-icon"
      obj.add(sprite)
      obj.userData.iconSprite = sprite
      this.getItemIconTexture(icon).then((tex) => {
        if (!tex || !sprite.parent) {
          return
        }
        sprite.material.map = tex
        sprite.material.needsUpdate = true
      })
    },
    getItemIconTexture(icon) {
      const id = Number(icon) || 0
      if (!id) {
        return Promise.resolve(null)
      }
      if (this.itemIconCache.has(id)) {
        return this.itemIconCache.get(id)
      }
      const files = [`${id}.png`, `item_${id}.png`, `item-${id}.png`]
      const bases = [ITEM_ICON_LOCAL, ITEM_ICON_REMOTE]
      const load = (async () => {
        for (const file of files) {
          for (const base of bases) {
            try {
              return await this.loadTexture(base + file)
            } catch (_err) {
              // try next
            }
          }
        }
        return null
      })()
      this.itemIconCache.set(id, load)
      return load
    },
    fallbackObjectMesh() {
      const mesh = new THREE.Mesh(
        new THREE.BoxGeometry(1.2, 1.2, 1.2),
        new THREE.MeshStandardMaterial({color: 0xc9a24a, transparent: true, opacity: 0.85})
      )
      return mesh
    },
    npcSpriteFiles(spawn) {
      const race = Math.max(1, Number(spawn.race) || 1)
      const gender = [0, 1, 2].includes(Number(spawn.gender)) ? Number(spawn.gender) : 2
      const texture = Math.max(0, Number(spawn.texture) || 0)
      const helm = Math.max(0, Number(spawn.helmtexture) || 0)
      return [...new Set([
        `CTN_${race}_${gender}_${texture}_${helm}.png`,
        `CTN_${race}_${gender}_${texture}_0.png`,
        `CTN_${race}_${gender}_0_0.png`,
        `CTN_${race}_2_${texture}_${helm}.png`,
        `CTN_${race}_2_0_0.png`,
        "CTN_1_0_0_0.png",
      ])]
    },
    npcSpriteSize(spawn) {
      const eqSize = Math.max(2, Number(spawn.size) || 6)
      const h = Math.min(2.4, Math.max(1.05, eqSize * WORLD_SCALE * 2.1))
      return {w: h * 0.58, h}
    },
    updateNpcSpriteScales() {
      if (!this.camera || !this.npcMarkers || !this.npcMarkers.length) {
        return
      }
      const cam = this.camera.position
      for (const marker of this.npcMarkers) {
        if (this.animatePathing && marker.userData.path) {
          continue
        }
        const base = this.npcSpriteSize(marker.userData.spawn || {})
        const dist = cam.distanceTo(marker.position)
        const h = Math.min(4.2, Math.max(base.h, dist * 0.032))
        marker.scale.set(h * 0.58, h, 1)
        const groundY = marker.userData.groundY
        if (typeof groundY === "number") {
          marker.position.y = groundY + h * 0.5
        }
      }
    },
    placeholderNpcTexture() {
      if (this._npcPlaceholder) {
        return this._npcPlaceholder
      }
      const canvas = document.createElement("canvas")
      canvas.width = 64
      canvas.height = 96
      const ctx = canvas.getContext("2d")
      ctx.fillStyle = "rgba(0,0,0,0)"
      ctx.fillRect(0, 0, 64, 96)
      ctx.fillStyle = "#d7c4a0"
      ctx.beginPath()
      ctx.arc(32, 18, 10, 0, Math.PI * 2)
      ctx.fill()
      ctx.fillStyle = "#8a7a5c"
      ctx.fillRect(22, 28, 20, 34)
      ctx.fillRect(16, 32, 8, 22)
      ctx.fillRect(40, 32, 8, 22)
      ctx.fillRect(22, 60, 8, 28)
      ctx.fillRect(34, 60, 8, 28)
      const tex = new THREE.CanvasTexture(canvas)
      tex.minFilter = THREE.LinearFilter
      this._npcPlaceholder = tex
      return tex
    },
    loadTexture(url) {
      return new Promise((resolve, reject) => {
        this.textureLoader.load(url, (tex) => {
          if (THREE.sRGBEncoding != null) {
            tex.encoding = THREE.sRGBEncoding
          }
          tex.minFilter = THREE.LinearFilter
          resolve(tex)
        }, undefined, reject)
      })
    },
    getNpcTexture(spawn) {
      const files = this.npcSpriteFiles(spawn)
      const cacheKey = files.join("|")
      if (this.npcTextureCache.has(cacheKey)) {
        return this.npcTextureCache.get(cacheKey)
      }
      const load = (async () => {
        const bases = [NPC_SPRITE_LOCAL, NPC_SPRITE_REMOTE]
        for (const file of files) {
          for (const base of bases) {
            try {
              return await this.loadTexture(base + file)
            } catch (_err) {
              // try next source
            }
          }
        }
        return this.placeholderNpcTexture()
      })()
      this.npcTextureCache.set(cacheKey, load)
      return load
    },
    detectCoordMap(spawns) {
      if (!spawns.length || !this.surfaceTargets.length) {
        return NPC_COORD_TRANSFORMS[1]
      }
      const sample = spawns.slice(0, Math.min(160, spawns.length))
      let best = NPC_COORD_TRANSFORMS[1]
      let bestScore = -1
      for (const transform of NPC_COORD_TRANSFORMS) {
        let hits = 0
        for (const spawn of sample) {
          const pos = eqToWorld(spawn.x, spawn.y, spawn.z, transform)
          if (this.terrainHeight(pos.x, pos.z) != null) {
            hits += 1
          }
        }
        const score = hits / sample.length
        if (score > bestScore) {
          bestScore = score
          best = transform
        }
      }
      return best
    },
    terrainHeight(x, z) {
      if (!this.surfaceTargets.length) {
        return null
      }
      this.raycaster.set(new THREE.Vector3(x, 80000, z), new THREE.Vector3(0, -1, 0))
      const hits = this.raycaster.intersectObjects(this.surfaceTargets, false)
      return hits.length ? hits[0].point.y : null
    },
    snapY(x, y, z) {
      const h = this.terrainHeight(x, z)
      return {x, y: h == null ? y : h, z}
    },
    groundFromEvent(event) {
      const rect = this.rendererEl.domElement.getBoundingClientRect()
      this.pointer.x = ((event.clientX - rect.left) / rect.width) * 2 - 1
      this.pointer.y = -((event.clientY - rect.top) / rect.height) * 2 + 1
      this.raycaster.setFromCamera(this.pointer, this.camera)
      const hits = this.raycaster.intersectObjects(this.surfaceTargets, false)
      return hits.length ? hits[0].point : null
    },
    async loadScenery(token, expected) {
      this.setStatus(`Loading client scenery${expected ? ` (0/${expected})` : ""}…`)
      const instances = await LanternApi.instances(this.zone)
      if (token !== this.loadToken) {
        return
      }
      this.clearGroup(this.sceneryGroup)
      this.sceneryMarkers = []
      this.sceneryInstances = instances.slice()
      let placed = 0
      for (let i = 0; i < instances.length; i++) {
        if (token !== this.loadToken) {
          return
        }
        const obj = await this.makeSceneryMarker(instances[i], i)
        if (!obj) {
          continue
        }
        placed += 1
        if (placed % 80 === 0) {
          this.setStatus(`Loading client scenery ${placed}/${instances.length}…`)
          await new Promise((r) => setTimeout(r, 0))
        }
      }
      this.sceneryGroup.visible = this.showScenery
      this.frameLoaded()
      this.setStatus(this.finalStatus(placed))
    },
    async makeSceneryMarker(inst, index) {
      const tpl = await this.getTemplate(inst.modelRelPath)
      if (!tpl) {
        return null
      }
      const obj = tpl.clone(true)
      const pos = lanternInstanceToWorld(inst.pos || [])
      obj.position.set(pos.x, pos.y, pos.z)
      obj.rotation.set(
        degToRad((inst.rot && inst.rot[0]) || 0),
        degToRad((inst.rot && inst.rot[1]) || 0),
        degToRad((inst.rot && inst.rot[2]) || 0),
        "YXZ"
      )
      const sx = ((inst.scale && inst.scale[0]) || 1) * WORLD_SCALE
      const sy = ((inst.scale && inst.scale[1]) || 1) * WORLD_SCALE
      const sz = ((inst.scale && inst.scale[2]) || 1) * WORLD_SCALE
      obj.scale.set(sx, sy, sz)
      obj.userData.kind = "scenery"
      obj.userData.index = index
      obj.userData.instance = inst
      this.sceneryGroup.add(obj)
      this.sceneryMarkers.push(obj)
      return obj
    },
    async getTemplate(rel) {
      const key = String(rel || "").toLowerCase()
      if (!key) {
        return null
      }
      if (!this.templateCache.has(key)) {
        this.templateCache.set(key, (async () => {
          const url = await LanternApi.loadFileUrl(rel)
          this.objectUrls.push(url)
          const gltf = await this.loader.loadAsync(url)
          gltf.scene.traverse((node) => {
            if (!node.isMesh || !node.material) {
              return
            }
            const mats = Array.isArray(node.material) ? node.material : [node.material]
            mats.forEach((m) => this.prepareMaterial(m))
          })
          return gltf.scene
        })().catch(() => null))
      }
      return this.templateCache.get(key)
    },
    finalStatus(sceneryCount) {
      const bits = [`${this.zone} terrain`]
      if (this.showNpcs) {
        bits.push(`${this.npcMarkers.length} NPCs`)
      }
      if (this.showObjects) {
        bits.push(`${this.objectMarkers.length} objects`)
      }
      if (this.showPaths) {
        bits.push(`${this.pathGrids.size} grids`)
      }
      if (this.showScenery && sceneryCount != null) {
        bits.push(`${sceneryCount} client objects`)
      }
      return bits.join(" · ")
    },
    redrawPaths() {
      this.clearGroup(this.pathGroup)
      if (!this.showPaths || !this.pathGrids) {
        return
      }
      const selectedGrid = this.selected && this.selected.table === "spawn2" ? Number(this.selected.pathgrid) : 0
      for (const grid of this.pathGrids.values()) {
        if (grid.points.length < 2) {
          continue
        }
        const highlight = selectedGrid && grid.id === selectedGrid
        const pts = grid.points.map((p) => {
          const w = eqToWorld(p.x, p.y, p.z, this.coordMap || NPC_COORD_TRANSFORMS[1])
          const g = this.snapY(w.x, w.y, w.z)
          return new THREE.Vector3(g.x, g.y + 0.3, g.z)
        })
        if (grid.type === 0) {
          pts.push(pts[0].clone())
        }
        if (grid.type === 7 && pts.length > 1) {
          const star = []
          for (let i = 1; i < pts.length; i++) {
            star.push(pts[0], pts[i])
          }
          const geo = new THREE.BufferGeometry().setFromPoints(star)
          const line = new THREE.LineSegments(geo, new THREE.LineBasicMaterial({
            color: highlight ? 0xffc14d : 0x4d7ea8,
            transparent: true,
            opacity: highlight ? 0.95 : 0.35,
          }))
          this.pathGroup.add(line)
          continue
        }
        const geo = new THREE.BufferGeometry().setFromPoints(pts)
        const line = new THREE.Line(geo, new THREE.LineBasicMaterial({
          color: highlight ? 0xffc14d : 0x4d7ea8,
          transparent: true,
          opacity: highlight ? 0.95 : 0.35,
        }))
        this.pathGroup.add(line)
      }
    },
    stepPathing(dt) {
      if (!this.animatePathing || !this.npcMarkers) {
        return
      }
      for (const marker of this.npcMarkers) {
        const spawn = marker.userData.spawn
        if (!spawn || this.dirtyKey("spawn2", spawn.id) || (this.selected && this.selected.table === "spawn2" && this.selected.id === spawn.id)) {
          continue
        }
        const grid = this.pathGrids.get(Number(spawn.pathgrid))
        let state = marker.userData.path
        if (!grid || !gridMoves(grid.type, grid.points.length)) {
          continue
        }
        if (!state) {
          state = this.makePathState(spawn)
          marker.userData.path = state
        }
        stepPathState(grid, state, eqWalkSpeed(spawn.walkspeed, spawn.runspeed), dt)
        marker.visible = !state.hidden
        if (state.hidden) {
          continue
        }
        const here = currentPathPose(grid, state)
        const w = eqToWorld(here.x, here.y, here.z, this.coordMap || NPC_COORD_TRANSFORMS[1])
        const g = this.snapY(w.x, w.y, w.z)
        const dims = this.npcSpriteSize(spawn)
        marker.userData.groundY = g.y
        marker.position.set(g.x, g.y + dims.h * 0.5, g.z)
        spawn.heading = here.heading
      }
    },
    setCameraMode(mode) {
      this.cameraMode = mode
      if (this.controls) {
        this.controls.enabled = mode === "orbit"
      }
      if (mode !== "walk" && this.walkControls && this.walkControls.isLocked) {
        this.walkControls.unlock()
      }
    },
    stepWalk(dt) {
      if (!this.walkControls) {
        return
      }
      const sprint = this.keys.ShiftLeft || this.keys.ShiftRight
      const speed = (sprint ? 42 : 16) * dt
      if (this.keys.KeyW) {
        this.walkControls.moveForward(speed)
      }
      if (this.keys.KeyS) {
        this.walkControls.moveForward(-speed)
      }
      if (this.keys.KeyA) {
        this.walkControls.moveRight(-speed)
      }
      if (this.keys.KeyD) {
        this.walkControls.moveRight(speed)
      }
      const pos = this.camera.position
      if (this.keys.Space) {
        pos.y += speed
      }
      if (this.keys.ControlLeft || this.keys.KeyC) {
        pos.y -= speed
      }
      const floor = this.terrainHeight(pos.x, pos.z)
      const eye = 1.7
      if (floor != null && !this.keys.Space && pos.y < floor + eye) {
        pos.y = floor + eye
      }
    },
    onNpcQuery() {
      clearTimeout(this.npcTimer)
      this.npcTimer = setTimeout(async () => {
        this.npcHits = await searchNpcs(this.npcQuery)
      }, 250)
    },
    selectPlaceNpc(npc) {
      this.placeKind = "npc"
      this.placeItem = npc
      this.saveMessage = `Click ground to place ${npc.name}`
    },
    selectPlaceObject(model) {
      this.placeKind = this.placeAs === "object" ? "object" : "scenery"
      this.placeItem = model
      this.saveMessage = `Click ground to place ${model.modelName} as ${this.placeKind === "scenery" ? "client scenery" : "server object"}`
    },
    selectPlaceItem(item) {
      this.placeKind = "item"
      this.placeItem = item
      this.saveMessage = `Click ground to place ${item.name} (icon ${item.icon || 0})`
    },
    onItemQuery() {
      clearTimeout(this.itemTimer)
      this.itemTimer = setTimeout(async () => {
        this.itemHits = await searchItems(this.itemQuery)
      }, 250)
    },
    clearPlace() {
      this.placeKind = ""
      this.placeItem = null
    },
    async onCanvasClick(event) {
      if (this.transform && this.transform.dragging) {
        return
      }
      if (this.cameraMode === "walk" && this.walkControls && !this.walkControls.isLocked) {
        this.walkControls.lock()
      }
      if (this.placeKind && this.editMode) {
        const hit = this.groundFromEvent(event)
        if (hit) {
          await this.placeAtWorld(hit)
        }
        return
      }
      const hit = this.pickSelectable(event)
      if (!hit) {
        this.clearSelection()
        return
      }
      this.selectObject3D(hit.object)
    },
    pickSelectable(event) {
      const rect = this.rendererEl.domElement.getBoundingClientRect()
      this.pointer.x = ((event.clientX - rect.left) / rect.width) * 2 - 1
      this.pointer.y = -((event.clientY - rect.top) / rect.height) * 2 + 1
      this.raycaster.setFromCamera(this.pointer, this.camera)
      const targets = []
      if (this.showNpcs) {
        targets.push(...this.npcGroup.children)
      }
      if (this.showObjects) {
        targets.push(...this.objectGroup.children)
      }
      if (this.showScenery && this.editMode) {
        targets.push(...this.sceneryGroup.children)
      }
      const hits = this.raycaster.intersectObjects(targets, true)
      if (!hits.length) {
        return null
      }
      let obj = hits[0].object
      while (obj && !obj.userData.kind && obj.parent) {
        obj = obj.parent
      }
      return obj && obj.userData.kind ? {object: obj} : null
    },
    selectObject3D(obj) {
      if (!obj || !obj.userData.kind) {
        return
      }
      if (obj.userData.kind === "npc") {
        const spawn = obj.userData.spawn
        const grid = this.pathGrids.get(Number(spawn.pathgrid))
        this.selected = {
          table: "spawn2",
          id: spawn.id,
          name: spawn.name,
          x: spawn.x,
          y: spawn.y,
          z: spawn.z,
          heading: spawn.heading,
          pathgrid: spawn.pathgrid,
          pathLabel: grid ? `grid ${grid.id} ${gridTypeLabel(grid.type)}` : (spawn.pathgrid ? `grid ${spawn.pathgrid}` : "no path"),
        }
        this.selectedObj = obj
        if (this.transform) {
          this.transform.detach()
        }
        this.broadcastFocus()
      } else if (obj.userData.kind === "scenery") {
        this.syncSceneryFromObject(obj)
        const inst = obj.userData.instance
        const eq = worldToEq(obj.position.x, obj.position.y, obj.position.z, this.coordMap || NPC_COORD_TRANSFORMS[1])
        this.selected = {
          table: "scenery",
          id: obj.userData.index,
          name: inst.modelName,
          x: eq.x,
          y: eq.y,
          z: eq.z,
          heading: yawToEqHeading(obj.rotation.y),
        }
        this.selectedObj = obj
        if (this.editMode && this.transform && this.cameraMode === "orbit") {
          this.transform.attach(obj)
          this.transform.setMode("translate")
        }
        this.broadcastFocus()
      } else {
        const row = obj.userData.row
        this.selected = {
          table: "object",
          id: row.id,
          name: row.name,
          x: row.x,
          y: row.y,
          z: row.z,
          heading: row.heading,
          icon: row.icon || 0,
        }
        this.selectedObj = obj
        if (this.editMode && this.transform && this.cameraMode === "orbit") {
          this.transform.attach(obj)
          this.transform.setMode("translate")
        }
        this.broadcastFocus()
      }
      if (this.controls && this.cameraMode === "orbit") {
        this.controls.target.copy(obj.position)
      }
      this.redrawPaths()
    },
    clearSelection() {
      this.selected = null
      this.selectedObj = null
      if (this.transform) {
        this.transform.detach()
      }
      this.redrawPaths()
    },
    onPointerDown(event) {
      if (!this.editMode || this.placeKind || this.cameraMode === "walk") {
        return
      }
      const hit = this.pickSelectable(event)
      if (hit && hit.object.userData.kind === "npc") {
        this.selectObject3D(hit.object)
        this.draggingNpc = hit.object
        this.dragFrom = this.copySelected()
      }
    },
    onPointerMove(event) {
      if (!this.draggingNpc) {
        return
      }
      const hit = this.groundFromEvent(event)
      if (!hit) {
        return
      }
      this.applyWorldToNpc(this.draggingNpc, hit)
    },
    onPointerUp() {
      if (this.draggingNpc && this.dragFrom) {
        this.pushDirty("spawn2", this.draggingNpc.userData.spawn.id, this.dragFrom, this.coordsFromNpc(this.draggingNpc))
      }
      this.draggingNpc = null
      this.dragFrom = null
    },
    applyWorldToNpc(marker, world) {
      const eq = worldToEq(world.x, world.y, world.z, this.coordMap || NPC_COORD_TRANSFORMS[1])
      const spawn = marker.userData.spawn
      spawn.x = eq.x
      spawn.y = eq.y
      spawn.z = eq.z
      const dims = this.npcSpriteSize(spawn)
      marker.userData.groundY = world.y
      marker.position.set(world.x, world.y + dims.h * 0.5, world.z)
      if (this.selected && this.selected.table === "spawn2" && this.selected.id === spawn.id) {
        this.selected.x = eq.x
        this.selected.y = eq.y
        this.selected.z = eq.z
      }
    },
    coordsFromNpc(marker) {
      const s = marker.userData.spawn
      return {x: s.x, y: s.y, z: s.z, heading: s.heading}
    },
    copySelected() {
      if (!this.selected) {
        return null
      }
      return {x: this.selected.x, y: this.selected.y, z: this.selected.z, heading: this.selected.heading}
    },
    syncSelectedFromObject() {
      if (!this.selectedObj) {
        return
      }
      if (this.selectedObj.userData.kind === "scenery") {
        this.syncSceneryFromObject(this.selectedObj)
        const eq = worldToEq(this.selectedObj.position.x, this.selectedObj.position.y, this.selectedObj.position.z, this.coordMap || NPC_COORD_TRANSFORMS[1])
        if (this.selected) {
          this.selected.x = eq.x
          this.selected.y = eq.y
          this.selected.z = eq.z
          this.selected.heading = yawToEqHeading(this.selectedObj.rotation.y)
        }
        return
      }
      if (this.selectedObj.userData.kind !== "object") {
        return
      }
      const eq = worldToEq(this.selectedObj.position.x, this.selectedObj.position.y, this.selectedObj.position.z, this.coordMap || NPC_COORD_TRANSFORMS[1])
      const heading = yawToEqHeading(this.selectedObj.rotation.y)
      const row = this.selectedObj.userData.row
      row.x = eq.x
      row.y = eq.y
      row.z = eq.z
      row.heading = heading
      if (this.selected) {
        this.selected.x = eq.x
        this.selected.y = eq.y
        this.selected.z = eq.z
        this.selected.heading = heading
      }
    },
    syncSceneryFromObject(obj) {
      if (!obj || obj.userData.kind !== "scenery") {
        return
      }
      const lantern = worldToLantern(obj.position.x, obj.position.y, obj.position.z)
      const inst = obj.userData.instance || {}
      inst.pos = lantern.pos
      inst.rot = [radToDeg(obj.rotation.x), radToDeg(obj.rotation.y), radToDeg(obj.rotation.z)]
      obj.userData.instance = inst
      const idx = obj.userData.index
      if (this.sceneryInstances && this.sceneryInstances[idx]) {
        this.sceneryInstances[idx] = inst
      }
    },
    commitTransform() {
      if (!this.selected || (this.selected.table !== "object" && this.selected.table !== "scenery")) {
        return
      }
      this.syncSelectedFromObject()
      if (this.selected.table === "scenery") {
        this.markSceneryDirty()
        return
      }
      this.pushDirty("object", this.selected.id, this.transformFrom || this.copySelected(), this.copySelected())
      this.transformFrom = this.copySelected()
    },
    markSceneryDirty() {
      this.sceneryDirty = true
      this.dirtyCount = Object.keys(this.dirty).length + 1
      this.saveMessage = ""
    },
    refreshDirtyCount() {
      this.dirtyCount = Object.keys(this.dirty).length + (this.sceneryDirty ? 1 : 0)
      this.canUndo = this.undoStack.length > 0
    },
    pushDirty(table, id, from, to) {
      if (!from || !to || coordsEqual(from, to)) {
        return
      }
      const key = table + ":" + id
      if (!this.dirty[key]) {
        this.$set(this.dirty, key, {table, id, from: {...from}, to: {...to}})
      } else {
        this.dirty[key].to = {...to}
      }
      this.undoStack.push({table, id, from: {...from}, to: {...to}})
      this.refreshDirtyCount()
      this.saveMessage = ""
    },
    dirtyKey(table, id) {
      return !!this.dirty[table + ":" + id]
    },
    undo() {
      const entry = this.undoStack.pop()
      if (!entry) {
        return
      }
      this.applyCoords(entry.table, entry.id, entry.from)
      delete this.dirty[entry.table + ":" + entry.id]
      this.refreshDirtyCount()
    },
    applyCoords(table, id, coords) {
      if (table === "spawn2") {
        const marker = this.npcMarkers.find((m) => m.userData.spawn && m.userData.spawn.id === id)
        if (!marker) {
          return
        }
        const spawn = marker.userData.spawn
        spawn.x = coords.x
        spawn.y = coords.y
        spawn.z = coords.z
        spawn.heading = coords.heading
        const w = eqToWorld(coords.x, coords.y, coords.z, this.coordMap || NPC_COORD_TRANSFORMS[1])
        const g = this.snapY(w.x, w.y, w.z)
        const dims = this.npcSpriteSize(spawn)
        marker.userData.groundY = g.y
        marker.position.set(g.x, g.y + dims.h * 0.5, g.z)
      } else {
        const obj = this.objectMarkers.find((m) => m.userData.row && m.userData.row.id === id)
        if (!obj) {
          return
        }
        obj.userData.row.x = coords.x
        obj.userData.row.y = coords.y
        obj.userData.row.z = coords.z
        obj.userData.row.heading = coords.heading
        const w = eqToWorld(coords.x, coords.y, coords.z, this.coordMap || NPC_COORD_TRANSFORMS[1])
        const g = this.snapY(w.x, w.y, w.z)
        obj.position.set(g.x, g.y, g.z)
        obj.rotation.y = eqHeadingToYaw(coords.heading)
      }
      if (this.selected && this.selected.table === table && this.selected.id === id) {
        Object.assign(this.selected, coords)
      }
    },
    applySceneryWorld(obj, world, heading) {
      obj.position.set(world.x, world.y, world.z)
      if (heading != null) {
        obj.rotation.y = eqHeadingToYaw(heading)
      }
      this.syncSceneryFromObject(obj)
      if (this.selected && this.selected.table === "scenery" && this.selected.id === obj.userData.index) {
        const eq = worldToEq(world.x, world.y, world.z, this.coordMap || NPC_COORD_TRANSFORMS[1])
        this.selected.x = eq.x
        this.selected.y = eq.y
        this.selected.z = eq.z
        if (heading != null) {
          this.selected.heading = heading
        }
      }
    },
    async placeAtWorld(world) {
      const eq = worldToEq(world.x, world.y, world.z, this.coordMap || NPC_COORD_TRANSFORMS[1])
      try {
        if (this.placeKind === "npc") {
          const spawn = await createNpcSpawn({
            zone: this.zone,
            version: this.version,
            npc: this.placeItem,
            x: eq.x,
            y: eq.y,
            z: eq.z,
            heading: 0,
          })
          const marker = this.makeNpcMarker({
            id: spawn.id,
            name: this.placeItem.name,
            x: eq.x,
            y: eq.y,
            z: eq.z,
            heading: 0,
            pathgrid: 0,
            ...this.placeItem,
          })
          this.npcMarkers.push(marker)
          this.selectObject3D(marker)
          this.saveMessage = `Placed ${this.placeItem.name} as spawn2 #${spawn.id}`
        } else if (this.placeKind === "object") {
          const row = await createZoneObject({
            zoneId: this.zoneId,
            version: this.version,
            objectname: this.placeItem.modelName,
            x: eq.x,
            y: eq.y,
            z: eq.z,
            heading: 0,
          })
          const obj = await this.makeObjectMarker({
            id: row.id,
            objectname: this.placeItem.modelName,
            xpos: eq.x,
            ypos: eq.y,
            zpos: eq.z,
            heading: 0,
          })
          this.objectMarkers.push(obj)
          this.selectObject3D(obj)
          this.saveMessage = `Placed ${this.placeItem.modelName} as object #${row.id}`
        } else if (this.placeKind === "item") {
          const objectname = String(this.placeItem.idfile || this.placeItem.name || "IT63").replace(/\s+/g, "_")
          const row = await createZoneObject({
            zoneId: this.zoneId,
            version: this.version,
            objectname,
            x: eq.x,
            y: eq.y,
            z: eq.z,
            heading: 0,
            icon: this.placeItem.icon || 0,
            itemid: this.placeItem.id,
            type: 1,
          })
          const obj = await this.makeObjectMarker({
            id: row.id,
            objectname,
            xpos: eq.x,
            ypos: eq.y,
            zpos: eq.z,
            heading: 0,
            icon: this.placeItem.icon || 0,
            itemid: this.placeItem.id,
          })
          this.objectMarkers.push(obj)
          this.selectObject3D(obj)
          this.saveMessage = `Placed ${this.placeItem.name} as object #${row.id}`
        } else if (this.placeKind === "scenery") {
          const inst = {
            modelName: this.placeItem.modelName,
            modelRelPath: this.placeItem.modelRelPath,
            pos: worldToLantern(world.x, world.y, world.z).pos,
            rot: [0, 0, 0],
            scale: [1, 1, 1],
            colorIndex: -1,
          }
          const obj = await this.makeSceneryMarker(inst, this.sceneryInstances.length)
          if (!obj) {
            throw new Error("Could not load that Lantern model")
          }
          this.sceneryInstances.push(inst)
          obj.position.set(world.x, world.y, world.z)
          this.syncSceneryFromObject(obj)
          this.markSceneryDirty()
          this.selectObject3D(obj)
          this.saveMessage = `Placed ${inst.modelName} as client scenery (Save writes object_instances.txt)`
        }
      } catch (err) {
        this.saveMessage = err && err.message ? err.message : "Place failed"
      }
    },
    async saveEdits() {
      const changes = Object.values(this.dirty).map((row) => ({
        table: row.table,
        id: row.id,
        x: row.to.x,
        y: row.to.y,
        z: row.to.z,
        heading: row.to.heading,
      }))
      if (!changes.length && !this.sceneryDirty) {
        this.saveMessage = "Nothing to save"
        return
      }
      this.saving = true
      const parts = []
      try {
        if (changes.length) {
          const r = await ZoneEditorApi.savePlacements({
            zone: this.zone,
            version: this.version,
            changes,
          })
          const updated = r && r.data ? r.data.updated : changes.length
          parts.push(updated + " peq row(s)")
        }
        if (this.sceneryDirty) {
          this.sceneryMarkers.forEach((obj) => this.syncSceneryFromObject(obj))
          const written = await LanternApi.saveInstances(this.zone, this.sceneryInstances)
          this.sceneryInstances = written
          this.sceneryDirty = false
          parts.push(written.length + " client scenery")
        }
        this.dirty = {}
        this.undoStack = []
        this.refreshDirtyCount()
        this.saveMessage = "Saved " + parts.join(" + ")
      } catch (err) {
        this.saveMessage = err && err.response && err.response.data && err.response.data.error
          ? err.response.data.error
          : (err && err.message ? err.message : "Save failed")
      } finally {
        this.saving = false
      }
    },
    onKeyDown(e) {
      this.keys[e.code] = true
      if (e.code === "Escape") {
        if (this.walkControls && this.walkControls.isLocked) {
          this.walkControls.unlock()
        }
        this.clearPlace()
        return
      }
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "s") {
        e.preventDefault()
        this.saveEdits()
        return
      }
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "z") {
        e.preventDefault()
        this.undo()
        return
      }
      if (!this.selected || !this.editMode) {
        return
      }
      if (e.key.toLowerCase() === "q" || e.key.toLowerCase() === "e") {
        const from = this.copySelected()
        const delta = e.key.toLowerCase() === "q" ? -16 : 16
        const heading = (((this.selected.heading || 0) + delta) % 512 + 512) % 512
        if (this.selected.table === "scenery" && this.selectedObj) {
          this.selectedObj.rotation.y = eqHeadingToYaw(heading)
          this.syncSelectedFromObject()
          this.markSceneryDirty()
        } else {
          this.applyCoords(this.selected.table, this.selected.id, {...from, heading})
          this.pushDirty(this.selected.table, this.selected.id, from, this.copySelected())
        }
      }
      if ((this.selected.table === "object" || this.selected.table === "scenery") && this.transform) {
        if (e.key.toLowerCase() === "g") {
          this.transform.setMode("translate")
        }
        if (e.key.toLowerCase() === "r") {
          this.transform.setMode("rotate")
        }
      }
    },
    onKeyUp(e) {
      this.keys[e.code] = false
    },
    frameObject(object3D) {
      const box = new THREE.Box3().setFromObject(object3D)
      this.applyFrameBox(box)
    },
    frameLoaded() {
      if (this.zoneGroup && this.zoneGroup.children.length) {
        this.frameObject(this.zoneGroup)
        return
      }
      const box = new THREE.Box3()
      if (this.sceneryGroup) {
        box.expandByObject(this.sceneryGroup)
      }
      this.applyFrameBox(box)
    },
    applyFrameBox(box) {
      if (!box || box.isEmpty()) {
        return
      }
      const size = box.getSize(new THREE.Vector3())
      const center = box.getCenter(new THREE.Vector3())
      const span = Math.max(size.x, size.z, 80)
      this.camera.position.set(
        center.x + span * 0.55,
        Math.max(box.max.y, center.y) + span * 0.25,
        center.z + span * 0.55
      )
      this.controls.target.set(center.x, box.min.y + Math.min(size.y * 0.2, 40), center.z)
      this.controls.update()
      const reach = Math.max(span, size.y, 80)
      if (this.sun) {
        this.sun.position.set(center.x + reach * 0.7, center.y + reach * 1.1, center.z + reach * 0.45)
      }
      if (this.fill) {
        this.fill.position.set(center.x - reach * 0.8, center.y + reach * 0.65, center.z - reach * 0.55)
      }
    },
    prepareMaterial(m) {
      if (!m) {
        return
      }
      m.side = THREE.DoubleSide
      if ("metalness" in m) {
        m.metalness = 0
      }
      if (!m.userData.atlasLit) {
        m.userData.atlasLit = true
        m.userData.origColor = m.color ? m.color.clone() : null
        m.userData.origEmissive = m.emissive ? m.emissive.clone() : null
        m.userData.origEmissiveIntensity = m.emissiveIntensity
        m.userData.origRoughness = "roughness" in m ? m.roughness : null
      }
      this.applyMaterialLook(m)
    },
    applyMaterialLook(m) {
      if (!m || !m.userData.atlasLit) {
        return
      }
      if (this.brightLighting) {
        if ("roughness" in m) {
          m.roughness = Math.min(m.userData.origRoughness != null ? m.userData.origRoughness : 1, 0.82)
        }
        if (m.color && m.userData.origColor) {
          m.color.copy(m.userData.origColor)
          if ((m.color.r + m.color.g + m.color.b) < 0.45) {
            m.color.multiplyScalar(1.4)
          }
        }
        if ("emissive" in m) {
          m.emissive.setHex(0x2a241c)
          m.emissiveIntensity = 0.2
        }
        return
      }
      if (m.color && m.userData.origColor) {
        m.color.copy(m.userData.origColor)
      }
      if (m.emissive && m.userData.origEmissive) {
        m.emissive.copy(m.userData.origEmissive)
      }
      if (m.userData.origEmissiveIntensity != null) {
        m.emissiveIntensity = m.userData.origEmissiveIntensity
      }
      if ("roughness" in m && m.userData.origRoughness != null) {
        m.roughness = m.userData.origRoughness
      }
    },
    applyAllMaterialLooks() {
      if (!this.scene) {
        return
      }
      this.scene.traverse((node) => {
        if (!node.isMesh || !node.material) {
          return
        }
        const mats = Array.isArray(node.material) ? node.material : [node.material]
        mats.forEach((m) => this.applyMaterialLook(m))
      })
    },
    applyLighting() {
      const bright = this.brightLighting
      if (this.scene) {
        this.scene.background.setHex(bright ? 0x2b3544 : 0x0f1216)
      }
      if (this.hemi) {
        this.hemi.color.setHex(bright ? 0xf2f6ff : 0xc5d4e8)
        this.hemi.groundColor.setHex(bright ? 0x8d7b62 : 0x3a3228)
        this.hemi.intensity = bright ? 1.35 : 0.7
      }
      if (this.ambient) {
        this.ambient.intensity = bright ? 1.15 : 0.45
      }
      if (this.sun) {
        this.sun.intensity = bright ? 1.7 : 1.05
      }
      if (this.fill) {
        this.fill.intensity = bright ? 0.9 : 0
      }
      if (this.headlamp) {
        this.headlamp.intensity = bright ? 0.7 : 0.35
      }
      if (this.rendererEl) {
        this.rendererEl.toneMappingExposure = bright ? 1.55 : 1
      }
    },
    onToggleBrightLighting() {
      this.applyLighting()
      this.applyAllMaterialLooks()
    },
    resetCamera() {
      this.setCameraMode("orbit")
      this.frameLoaded()
    },
    openSageWindow() {
      const href = this.sageHref
      window.open(href, "spire-sage")
      this.broadcastFocus()
    },
    broadcastZone() {
      if (!this.sageBridge) {
        return
      }
      this.sageBridge.send({type: "zone", zone: this.zone})
    },
    broadcastFocus() {
      if (!this.sageBridge) {
        return
      }
      const sel = this.selected
      this.sageBridge.send({
        type: "focus",
        zone: this.zone,
        x: sel ? sel.x : undefined,
        y: sel ? sel.y : undefined,
        z: sel ? sel.z : undefined,
        heading: sel ? sel.heading : undefined,
        id: sel ? sel.id : undefined,
        table: sel ? sel.table : undefined,
        name: sel ? sel.name : undefined,
      })
    },
    applyPendingFocus() {
      const q = this.$route.query || {}
      const pending = this.sageBridge ? this.sageBridge.consumePending() : null
      const focus = {
        x: q.x != null ? Number(q.x) : (pending && pending.x),
        y: q.y != null ? Number(q.y) : (pending && pending.y),
        z: q.z != null ? Number(q.z) : (pending && pending.z),
        id: q.id != null ? Number(q.id) : (pending && pending.id),
      }
      if (focus.x == null && focus.id == null) {
        return
      }
      this.focusEq(focus)
    },
    focusEq(msg) {
      if (msg && msg.id && this.npcMarkers) {
        const marker = this.npcMarkers.find((m) => m.userData.spawn && m.userData.spawn.id === Number(msg.id))
        if (marker) {
          this.selectObject3D(marker)
          return
        }
      }
      if (msg && msg.x != null && msg.y != null) {
        const w = eqToWorld(Number(msg.x) || 0, Number(msg.y) || 0, Number(msg.z) || 0, this.coordMap || NPC_COORD_TRANSFORMS[1])
        const g = this.snapY(w.x, w.y, w.z)
        if (this.controls) {
          this.controls.target.set(g.x, g.y, g.z)
          this.camera.position.set(g.x + 18, g.y + 12, g.z + 18)
          this.controls.update()
        }
        if (this.npcMarkers && this.npcMarkers.length) {
          let best = null
          let bestD = 12
          for (const marker of this.npcMarkers) {
            const d = marker.position.distanceTo(new THREE.Vector3(g.x, g.y, g.z))
            if (d < bestD) {
              bestD = d
              best = marker
            }
          }
          if (best) {
            this.selectObject3D(best)
          }
        }
      }
    },
    onSageBridge(msg) {
      if (!msg) {
        return
      }
      const zone = String(msg.zone || "").toLowerCase()
      if (zone && zone !== this.zone && (msg.type === "click" || msg.type === "focus" || msg.type === "zone")) {
        this.$router.push({
          path: "/zone/" + zone + "/atlas",
          query: {
            v: this.version,
            x: msg.x,
            y: msg.y,
            z: msg.z,
            id: msg.id,
          },
        })
        return
      }
      if (msg.type === "click" || msg.type === "focus") {
        this.focusEq(msg)
        this.saveMessage = msg.name ? ("Sage: " + msg.name) : "Sage click"
      }
    },
    async onSelectedIcon() {
      if (!this.selected || this.selected.table !== "object" || !this.selectedObj) {
        return
      }
      const icon = Number(this.selected.icon) || 0
      this.selectedObj.userData.row.icon = icon
      this.attachObjectIcon(this.selectedObj, icon)
      try {
        await updateZoneObjectIcon(this.selected.id, icon, this.selectedObj.userData.row.itemid)
        this.saveMessage = "Saved object icon " + icon
      } catch (err) {
        this.saveMessage = err && err.message ? err.message : "Icon save failed"
      }
    },
    onToggleScenery() {
      if (!this.showScenery) {
        this.sceneryGroup.visible = false
        this.setStatus(this.finalStatus())
        return
      }
      if (this.sceneryGroup.children.length) {
        this.sceneryGroup.visible = true
        this.setStatus(this.finalStatus(this.sceneryGroup.children.length))
        return
      }
      this.loadScenery(this.loadToken, 0)
    },
    onToggleNpcs() {
      if (this.npcGroup) {
        this.npcGroup.visible = this.showNpcs
      }
    },
    onToggleObjects() {
      if (this.objectGroup) {
        this.objectGroup.visible = this.showObjects
      }
      if (this.showObjects && !this.objectMarkers.length) {
        this.loadObjects(this.loadToken)
      }
    },
    onToggleWireframe() {
      this.scene.traverse((node) => {
        if (!node.isMesh || !node.material) {
          return
        }
        const mats = Array.isArray(node.material) ? node.material : [node.material]
        mats.forEach((m) => {
          m.wireframe = this.wireframe
        })
      })
    },
    disposeNode(node) {
      if (!node) {
        return
      }
      if (node.geometry && node.geometry.dispose) {
        node.geometry.dispose()
      }
      const mats = node.material ? (Array.isArray(node.material) ? node.material : [node.material]) : []
      mats.forEach((mat) => {
        if (!mat) {
          return
        }
        Object.keys(mat).forEach((key) => {
          const val = mat[key]
          if (val && val.isTexture && val.dispose) {
            val.dispose()
          }
        })
        if (mat.dispose) {
          mat.dispose()
        }
      })
    },
    clearGroup(group) {
      if (!group) {
        return
      }
      while (group.children.length) {
        const child = group.children[0]
        group.remove(child)
        child.traverse((node) => this.disposeNode(node))
      }
    },
    disposeScene() {
      cancelAnimationFrame(this.raf)
      if (this.controls) {
        this.controls.dispose()
      }
      if (this.transform) {
        this.transform.detach()
        this.transform.dispose()
      }
      this.clearGroup(this.zoneGroup)
      this.clearGroup(this.sceneryGroup)
      this.clearGroup(this.npcGroup)
      this.clearGroup(this.objectGroup)
      this.clearGroup(this.pathGroup)
      if (this.scene) {
        this.scene.traverse((node) => this.disposeNode(node))
      }
    },
    revokeUrls() {
      for (const url of this.objectUrls || []) {
        URL.revokeObjectURL(url)
      }
      this.objectUrls = []
    },
  },
}
</script>

<style scoped>
.atlas-page {
  position: relative;
  height: calc(100vh - 24px);
  margin: -8px;
  background: #0f1216;
  color: #e7eaef;
}
.atlas-page.atlas-bright { background: #2b3544; }
.atlas-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  background: #161a20;
  border-bottom: 1px solid #272e38;
}
.atlas-brand { font-weight: 700; color: #c9a24a; }
.atlas-zone { font-family: ui-monospace, Consolas, monospace; }
.atlas-modes { display: flex; gap: 4px; }
.atlas-check { margin: 0; font-size: 13px; }
.atlas-status { margin-left: auto; font-size: 12px; color: #a3adbb; }
.atlas-status.error { color: #f0a8a0; }
.atlas-body { display: flex; height: calc(100% - 46px); }
.atlas-catalog {
  width: 280px;
  min-width: 240px;
  overflow: auto;
  padding: 12px;
  background: #161a20;
  border-right: 1px solid #272e38;
}
.atlas-catalog-title { font-weight: 700; margin-bottom: 6px; }
.atlas-help { color: #a3adbb; font-size: 12px; line-height: 1.45; }
.atlas-field { display: block; margin: 10px 0 6px; font-size: 12px; color: #a3adbb; }
.atlas-field input {
  display: block;
  width: 100%;
  margin-top: 4px;
  background: #0f1216;
  border: 1px solid #272e38;
  color: #e7eaef;
  border-radius: 6px;
  padding: 6px 8px;
}
.atlas-list { max-height: 220px; overflow: auto; }
.atlas-item {
  display: block;
  width: 100%;
  text-align: left;
  background: transparent;
  border: 0;
  color: #e7eaef;
  padding: 6px 4px;
  font-size: 13px;
}
.atlas-item.active, .atlas-item:hover { background: #1c2129; }
.atlas-item-icon { display: flex; align-items: center; gap: 6px; }
.atlas-place-as { display: flex; gap: 10px; font-size: 12px; color: #a3adbb; margin: 4px 0 8px; }
.atlas-place-as label { margin: 0; }
.atlas-icon-edit { display: flex; align-items: center; gap: 8px; margin-top: 10px; font-size: 12px; color: #a3adbb; }
.atlas-icon-edit input { width: 80px; background: #0f1216; border: 1px solid #272e38; color: #e7eaef; border-radius: 6px; padding: 4px 6px; }
.atlas-tip-icon { display: inline-block; vertical-align: middle; margin-right: 4px; }
.btn-block { width: 100%; margin-top: 8px; }
.atlas-stage { position: relative; flex: 1; min-width: 0; }
.atlas-viewport { height: 100%; width: 100%; }
.atlas-viewport.placing { cursor: crosshair; }
.atlas-walk-hint, .atlas-tip, .atlas-save {
  position: absolute;
  left: 12px;
  padding: 8px 10px;
  border-radius: 8px;
  background: rgba(16, 20, 26, 0.92);
  border: 1px solid #37404d;
  font-size: 13px;
}
.atlas-walk-hint { top: 12px; right: 12px; left: auto; max-width: 420px; }
.atlas-tip { bottom: 12px; }
.atlas-save { bottom: 52px; color: #c9a24a; }
.mono { font-family: ui-monospace, Consolas, monospace; }
</style>

<template>
  <div class="atlas-page" :class="{ 'atlas-bright': brightLighting }">
    <div class="atlas-bar">
      <div class="atlas-brand">3D Atlas</div>
      <div class="atlas-zone mono">{{ zone || "—" }}</div>
      <label class="atlas-check">
        <input type="checkbox" v-model="showScenery" @change="onToggleScenery">
        Client scenery
      </label>
      <label class="atlas-check">
        <input type="checkbox" v-model="showNpcs" @change="onToggleNpcs">
        NPCs
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
      <span class="atlas-status" :class="{ error: !!error }">{{ status }}</span>
    </div>
    <div ref="viewport" class="atlas-viewport"></div>
    <div v-if="selectedNpc" class="atlas-tip">
      {{ selectedNpc.name }}
      <span class="mono"> spawn2 #{{ selectedNpc.id }}  X {{ selectedNpc.x }} Y {{ selectedNpc.y }} Z {{ selectedNpc.z }}</span>
    </div>
  </div>
</template>

<script>
import * as THREE from "three"
import {OrbitControls} from "three/examples/jsm/controls/OrbitControls"
import {GLTFLoader} from "three/examples/jsm/loaders/GLTFLoader"
import {Navbar} from "../../app/navbar"
import {LanternApi} from "../../app/lantern"
import {NPC_COORD_TRANSFORMS, WORLD_SCALE, eqToWorld, lanternInstanceToWorld} from "../../app/lantern-coords"
import {Spawn} from "../../app/spawn"
import {Npcs} from "../../app/npcs"

const NPC_SPRITE_LOCAL = "/eq-asset-preview-master/assets/npc_models/"
const NPC_SPRITE_REMOTE = "https://raw.githubusercontent.com/EQEmuTools/eq-asset-preview/master/assets/npc_models/"

export default {
  name: "ZoneAtlas",
  data() {
    return {
      zone: "",
      version: 0,
      status: "Starting…",
      error: "",
      showScenery: true,
      showNpcs: true,
      brightLighting: true,
      wireframe: false,
      selectedNpc: null,
      rootLabel: "",
    }
  },
  computed: {
    mapHref() {
      const v = this.version === "" || this.version == null ? 0 : this.version
      return `/zone/${this.zone}?v=${v}`
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
    this.loadToken = 0
    this.initRenderer()
    this.boot()
  },
  beforeDestroy() {
    this.loadToken += 1
    window.removeEventListener("resize", this.onResize)
    if (this.rendererEl && this.rendererEl.domElement) {
      this.rendererEl.domElement.removeEventListener("click", this.onCanvasClick)
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
      this.loader = new GLTFLoader()
      this.textureLoader = new THREE.TextureLoader()
      this.textureLoader.crossOrigin = "anonymous"
      this.npcTextureCache = new Map()
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
      this.scene.add(this.zoneGroup)
      this.scene.add(this.sceneryGroup)
      this.scene.add(this.npcGroup)

      this.onResize = this.onResize.bind(this)
      this.onCanvasClick = this.onCanvasClick.bind(this)
      window.addEventListener("resize", this.onResize)
      this.rendererEl.domElement.addEventListener("click", this.onCanvasClick)
      this.onResize()
      this.tick = this.tick.bind(this)
      this.tick()
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
      if (this.controls) {
        this.controls.update()
      }
      this.updateNpcSpriteScales()
      if (this.rendererEl && this.scene && this.camera) {
        this.rendererEl.render(this.scene, this.camera)
      }
    },
    async boot() {
      const token = ++this.loadToken
      try {
        this.setStatus("Checking Lantern exports…")
        const status = await LanternApi.status()
        if (token !== this.loadToken) {
          return
        }
        if (!status.ok) {
          this.setStatus(status.error || "Lantern export root not found.", true)
          return
        }
        this.rootLabel = status.root
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
      this.setStatus(`${this.zone} terrain loaded. Overlaying NPCs…`)

      if (this.showNpcs) {
        await this.loadNpcs(token)
      }
      if (this.showScenery) {
        await this.loadScenery(token, info.instanceCount || 0)
      } else {
        this.setStatus(this.finalStatus())
      }
    },
    async loadNpcs(token) {
      const rows = await Spawn.getByZone(this.zone, this.version, false)
      if (token !== this.loadToken) {
        return
      }
      const spawns = []
      for (const spawn2 of rows || []) {
        let name = `spawn2 ${spawn2.id}`
        let npc = {race: 1, gender: 2, texture: 0, helmtexture: 0, size: 6}
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
          ...npc,
        })
      }
      this.coordMap = this.detectCoordMap(spawns)
      this.clearGroup(this.npcGroup)
      this.npcMarkers = []
      for (const spawn of spawns) {
        const pos = eqToWorld(spawn.x, spawn.y, spawn.z, this.coordMap)
        const grounded = this.snapY(pos.x, pos.y, pos.z)
        const dims = this.npcSpriteSize(spawn)
        const marker = new THREE.Sprite(new THREE.SpriteMaterial({
          map: this.placeholderNpcTexture(),
          transparent: true,
          depthWrite: false,
        }))
        marker.scale.set(dims.w, dims.h, 1)
        marker.position.set(grounded.x, grounded.y + dims.h * 0.5, grounded.z)
        marker.userData.spawn = spawn
        marker.userData.groundY = grounded.y
        this.npcGroup.add(marker)
        this.npcMarkers.push(marker)
        this.getNpcTexture(spawn).then((tex) => {
          if (!tex || token !== this.loadToken || !marker.parent) {
            return
          }
          marker.material.map = tex
          marker.material.needsUpdate = true
        })
      }
      this.npcGroup.visible = this.showNpcs
      this.setStatus(this.finalStatus())
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
    async loadScenery(token, expected) {
      this.setStatus(`Loading client scenery${expected ? ` (0/${expected})` : ""}…`)
      const instances = await LanternApi.instances(this.zone)
      if (token !== this.loadToken) {
        return
      }
      this.clearGroup(this.sceneryGroup)
      let placed = 0
      for (const inst of instances) {
        if (token !== this.loadToken) {
          return
        }
        const tpl = await this.getTemplate(inst.modelRelPath)
        if (!tpl) {
          continue
        }
        const obj = tpl.clone(true)
        const pos = lanternInstanceToWorld(inst.pos || [])
        obj.position.set(pos.x, pos.y, pos.z)
        obj.rotation.set(
          ((inst.rot && inst.rot[0]) || 0) * Math.PI / 180,
          ((inst.rot && inst.rot[1]) || 0) * Math.PI / 180,
          ((inst.rot && inst.rot[2]) || 0) * Math.PI / 180,
          "YXZ"
        )
        const sx = ((inst.scale && inst.scale[0]) || 1) * WORLD_SCALE
        const sy = ((inst.scale && inst.scale[1]) || 1) * WORLD_SCALE
        const sz = ((inst.scale && inst.scale[2]) || 1) * WORLD_SCALE
        obj.scale.set(sx, sy, sz)
        this.sceneryGroup.add(obj)
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
        bits.push(`${this.npcMarkers.length} NPCs${this.coordMap ? ` (${this.coordMap.id})` : ""}`)
      }
      if (this.showScenery && sceneryCount != null) {
        bits.push(`${sceneryCount} client objects`)
      }
      return bits.join(" · ")
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
        this.headlamp.intensity = bright ? 0.7 : 0
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
      this.frameLoaded()
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
    onCanvasClick(event) {
      if (!this.showNpcs || !this.npcGroup.children.length) {
        return
      }
      const rect = this.rendererEl.domElement.getBoundingClientRect()
      this.pointer.x = ((event.clientX - rect.left) / rect.width) * 2 - 1
      this.pointer.y = -((event.clientY - rect.top) / rect.height) * 2 + 1
      this.raycaster.setFromCamera(this.pointer, this.camera)
      const hits = this.raycaster.intersectObjects(this.npcGroup.children, true)
      if (!hits.length) {
        this.selectedNpc = null
        return
      }
      const spawn = hits[0].object.userData.spawn
      this.selectedNpc = spawn
      this.controls.target.copy(hits[0].object.position)
      this.controls.update()
    },
    clearGroup(group) {
      if (!group) {
        return
      }
      while (group.children.length) {
        group.remove(group.children[0])
      }
    },
    disposeScene() {
      cancelAnimationFrame(this.raf)
      if (this.controls) {
        this.controls.dispose()
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
.atlas-page.atlas-bright {
  background: #2b3544;
}
.atlas-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  background: #161a20;
  border-bottom: 1px solid #272e38;
}
.atlas-brand {
  font-weight: 700;
  color: #c9a24a;
}
.atlas-zone {
  font-family: ui-monospace, Consolas, monospace;
}
.atlas-check {
  margin: 0;
  font-size: 13px;
}
.atlas-status {
  margin-left: auto;
  font-size: 12px;
  color: #a3adbb;
}
.atlas-status.error {
  color: #f0a8a0;
}
.atlas-viewport {
  height: calc(100% - 46px);
  width: 100%;
}
.atlas-tip {
  position: absolute;
  left: 12px;
  bottom: 12px;
  padding: 8px 10px;
  border-radius: 8px;
  background: rgba(16, 20, 26, 0.92);
  border: 1px solid #37404d;
  font-size: 13px;
}
.mono {
  font-family: ui-monospace, Consolas, monospace;
}
</style>

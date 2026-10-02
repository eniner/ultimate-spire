<template>
  <div class="sage-container" ref="sage-root">
    <div class="sage-bar">
      <div class="sage-brand">EQ Sage</div>
      <button class="btn btn-sm btn-warning" type="button" :disabled="connecting" @click="connectFolder">
        {{ connecting ? "Connecting…" : (dirName ? "Reconnect EQ folder" : "Connect EQ folder") }}
      </button>
      <button class="btn btn-sm btn-dark" type="button" :disabled="!pendingZone" @click="openZoneDialog">
        Open {{ pendingZone || "zone" }}
      </button>
      <button class="btn btn-sm btn-dark" type="button" @click="openSageWindow">Sage window</button>
      <span class="sage-bar-status" :class="{ error: !!error }">{{ barText }}</span>
    </div>
    <div v-if="status === 'loading'" class="sage-status">
      <p>Loading EQ Sage…</p>
    </div>
    <iframe
      ref="sage-iframe"
      class="sage-iframe"
      src="/eqsage/"
      title="EQSage"
      allow="fullscreen; clipboard-read; clipboard-write"
      referrerpolicy="no-referrer"
      @load="onSageLoad"
    ></iframe>
  </div>
</template>

<script type="ts">
import * as SpireApiTypes from "@/app/api";
import { SpireApi } from "../../app/api/spire-api";
import { SpireQueryBuilder } from "@/app/api/spire-query-builder";
import {Navbar}          from "../../app/navbar";
import { Zones } from "../../app/zones";
import { Spawn } from "../../app/spawn";
import { Npcs } from "../../app/npcs";
import { Grid } from "../../app/grid";
import {createAtlasSageBridge, installSageIframeBridge, sageHref} from "../../app/atlas-sage-bridge";
import {connectEqDirectory, eqDirGranted, readEqDirHandle, zoneLooksConverted} from "../../app/sage-eqdir";
export default {
  components: {

  },
  data() {
    return {
      status: "loading",
      error: "",
      hint: "",
      dirName: "",
      connecting: false,
      loadRetries: 0,
      mountTimer: null,
      fileTimer: null,
      pendingFocus: null,
      appliedZone: false,
      openedChooser: false,
    }
  },
  computed: {
    pendingZone() {
      const pending = this.pendingFocus || {}
      return String(pending.zone || "").toLowerCase()
    },
    barText() {
      if (this.error) {
        return this.error
      }
      if (this.hint) {
        return this.hint
      }
      if (this.status === "loading") {
        return "Loading EQ Sage…"
      }
      return this.dirName ? ("Folder: " + this.dirName) : "Connect your EQ directory, then pick a zone."
    },
  },
  watch: {

  },

  destroyed() {
    if (this.mountTimer) {
      clearTimeout(this.mountTimer)
      this.mountTimer = null
    }
    if (this.fileTimer) {
      clearInterval(this.fileTimer)
      this.fileTimer = null
    }
    if (this.bridge) {
      this.bridge.dispose()
      this.bridge = null
    }
    Navbar.expand();
  },
  async mounted() {
    Navbar.collapse()
    const q = this.$route.query || {}
    this.pendingFocus = {
      zone: String(q.zone || "").toLowerCase(),
      x: q.x != null ? Number(q.x) : undefined,
      y: q.y != null ? Number(q.y) : undefined,
      z: q.z != null ? Number(q.z) : undefined,
    }
    this.bridge = createAtlasSageBridge("sage", (msg) => {
      this.forwardToIframe(msg)
    })
    await this.refreshStoredDir()
    if (this.pendingZone) {
      this.hint = "Atlas sent " + this.pendingZone + ". Connect the EQ folder in this bar (not only Sage’s button), then Open " + this.pendingZone + "."
    }
  },
  methods: {
    injectSpire(win) {
      if (!win) {
        return
      }
      win.Spire = {
        SpireApi,
        SpireApiTypes,
        SpireQueryBuilder,
        Grid,
        Zones,
        Spawn,
        Npcs,
        backendBaseUrl: SpireApi.getBasePath(),
      }
      installSageIframeBridge(win, window.location.origin)
    },
    forwardToIframe(msg) {
      const iframe = this.$refs["sage-iframe"]
      if (!iframe || !iframe.contentWindow) {
        return
      }
      iframe.contentWindow.postMessage(Object.assign({channel: "spire-atlas-sage"}, msg), window.location.origin)
    },
    async refreshStoredDir() {
      try {
        const handle = await readEqDirHandle()
        this.dirName = handle && handle.name ? handle.name : ""
      } catch (_err) {
        this.dirName = ""
      }
    },
    sageWin() {
      const iframe = this.$refs["sage-iframe"]
      return iframe && iframe.contentWindow
    },
    sageDoc() {
      const iframe = this.$refs["sage-iframe"]
      try {
        return iframe && iframe.contentDocument
      } catch (_err) {
        return null
      }
    },
    sageBodyText() {
      const doc = this.sageDoc()
      return (doc && doc.body && doc.body.innerText) ? doc.body.innerText : ""
    },
    async sageReady() {
      const win = this.sageWin()
      const gc = win && (win.gameController || win.gc)
      const handle = (gc && gc.rootFileSystemHandle) || await readEqDirHandle()
      if (handle && gc) {
        gc.rootFileSystemHandle = handle
      }
      if (await eqDirGranted(handle)) {
        this.dirName = handle.name || this.dirName
        return true
      }
      const body = this.sageBodyText()
      if (body.indexOf("Request Permissions") !== -1) {
        this.hint = "Sage has the folder. Click Request Permissions in the Sage window so it can write the eqsage cache."
      }
      return false
    },
    async connectFolder() {
      this.connecting = true
      this.error = ""
      try {
        const existing = await readEqDirHandle()
        const alreadyGranted = await eqDirGranted(existing)
        const result = await connectEqDirectory(alreadyGranted)
        if (result.error) {
          this.error = result.error
          return
        }
        if (!result.granted) {
          return
        }
        this.dirName = result.handle && result.handle.name ? result.handle.name : "EQ folder"
        const win = this.sageWin()
        if (win && win.gameController) {
          win.gameController.rootFileSystemHandle = result.handle
        }
        this.hint = "Folder connected. Sage will reload with access, then use Open " + (this.pendingZone || "zone") + "."
        this.appliedZone = false
        this.openedChooser = false
        this.reloadSage()
      } finally {
        this.connecting = false
      }
    },
    openSageWindow() {
      const href = this.pendingZone ? sageHref(this.pendingZone, this.pendingFocus) : "/sage"
      window.open(href, "spire-sage")
    },
    clickSageText(label) {
      const doc = this.sageDoc()
      if (!doc) {
        return false
      }
      const nodes = doc.querySelectorAll("button, a, p, span, div, h1, h2, h3, h4, h5, h6")
      for (let i = 0; i < nodes.length; i++) {
        const el = nodes[i]
        const text = (el.textContent || "").replace(/\s+/g, " ").trim()
        if (text === label || text.indexOf(label) === 0) {
          el.dispatchEvent(new MouseEvent("click", {bubbles: true, cancelable: true}))
          return true
        }
      }
      return false
    },
    openZoneDialog() {
      if (this.clickSageText("Select a Zone")) {
        this.hint = this.pendingZone
          ? ("Pick " + this.pendingZone + " in Sage. It converts the zone the first time, then you can walk it.")
          : "Pick a zone in Sage."
        return
      }
      this.hint = "Select a Zone is in Sage’s top bar after the folder is connected."
    },
    async applyPendingToSage(win) {
      const pending = this.pendingFocus || (this.bridge && this.bridge.consumePending())
      if (!pending || !win || this.appliedZone) {
        return
      }
      const zone = String(pending.zone || "").toLowerCase()
      if (!zone) {
        return
      }
      if (!(await this.sageReady())) {
        return
      }
      const gc = win.gameController || win.gc
      const handle = (gc && gc.rootFileSystemHandle) || await readEqDirHandle()
      const converted = await zoneLooksConverted(handle, zone)
      if (!converted) {
        this.hint = "Folder is connected. Use Open " + zone + " — Sage has to convert it before it can walk the zone."
        if (!this.openedChooser) {
          this.openedChooser = true
          this.openZoneDialog()
        }
        return
      }
      if (!gc || !gc.ZoneController || typeof gc.ZoneController.loadModel !== "function") {
        this.openZoneDialog()
        return
      }
      this.appliedZone = true
      this.hint = "Loading " + zone + "…"
      gc.ZoneController.loadModel(zone).then(() => {
        gc.ZoneController.zoneName = zone
        if (pending.x != null && gc.CameraController && gc.CameraController.camera) {
          gc.CameraController.camera.position.x = pending.y || 0
          gc.CameraController.camera.position.y = (pending.z || 0) + 5
          gc.CameraController.camera.position.z = pending.x || 0
        }
        this.hint = "Loaded " + zone + "."
      }).catch(() => {
        this.appliedZone = false
        this.hint = "Sage could not load " + zone + " yet. Use Open " + zone + " and let it convert."
        this.openZoneDialog()
      })
    },
    sageDidMount(doc) {
      const root = doc && doc.getElementById("root")
      return !!(root && root.childElementCount > 0)
    },
    sageStillFetching(win) {
      try {
        const entries = win && win.performance && win.performance.getEntriesByType("resource")
        if (!entries || !entries.length) {
          return true
        }
        return !entries.some((e) => String(e.name || "").indexOf("eqsage-main") !== -1)
      } catch (_err) {
        return true
      }
    },
    reloadSage() {
      const iframe = this.$refs["sage-iframe"]
      if (!iframe) {
        return
      }
      this.status = "loading"
      this.error = ""
      iframe.src = "/eqsage/?r=" + Date.now()
    },
    onSageLoad(e) {
      const iframe = e.target
      const win = iframe.contentWindow
      this.injectSpire(win)

      let doc = null
      try {
        doc = iframe.contentDocument
      } catch (err) {
        this.status = "error"
        this.error = "EQ Sage loaded cross-origin, so Spire cannot hook its APIs. Use the same-origin /eqsage proxy."
        return
      }

      const bodyText = (doc && doc.body && doc.body.innerText) ? doc.body.innerText : ""
      if (
        bodyText.indexOf("Cannot GET") !== -1 ||
        (doc && doc.title && doc.title.indexOf("Error") === 0 && bodyText.indexOf("Cannot GET") !== -1)
      ) {
        this.status = "error"
        this.error = "EQ Sage did not load (/eqsage 404). Restart the Vue frontend so the Sage proxy in vue.config.js is picked up."
        return
      }

      this.watchForMount(iframe)
    },
    watchForMount(iframe) {
      if (this.mountTimer) {
        clearTimeout(this.mountTimer)
        this.mountTimer = null
      }
      let attempts = 0
      const poll = async () => {
        let live = null
        try {
          live = iframe.contentDocument
        } catch (_err) {
          live = null
        }
        if (this.sageDidMount(live)) {
          this.status = "ready"
          this.loadRetries = 0
          await this.applyPendingToSage(iframe.contentWindow)
          this.startReadyWatch()
          return
        }
        attempts += 1
        const shellOk = !!(live && live.getElementById && live.getElementById("root") && live.title === "EQ Sage")
        const stillFetching = this.sageStillFetching(iframe.contentWindow)
        if (!shellOk && !stillFetching && attempts === 50 && this.loadRetries < 1) {
          this.loadRetries += 1
          this.reloadSage()
          return
        }
        if (attempts > 150 && !stillFetching) {
          this.status = "error"
          this.error = "EQ Sage stayed blank after loading. Refresh /sage — a Sage JS chunk failed through the local /static proxy."
          return
        }
        this.status = "loading"
        this.mountTimer = setTimeout(poll, 400)
      }
      poll()
    },
    startReadyWatch() {
      if (this.fileTimer) {
        return
      }
      this.fileTimer = setInterval(async () => {
        const win = this.sageWin()
        if (!win) {
          return
        }
        await this.applyPendingToSage(win)
      }, 1500)
    },
  },
}
</script>

<style>
.sage-container {
  position: relative;
}
.sage-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(20, 16, 12, 0.94);
  border-bottom: 1px solid #8a6a32;
  color: #e6dcc3;
  font-size: 13px;
}
.sage-brand {
  font-weight: 700;
  letter-spacing: 0.04em;
  margin-right: 6px;
}
.sage-bar-status {
  flex: 1;
  min-width: 0;
  opacity: 0.9;
}
.sage-bar-status.error {
  color: #e6b0a0;
}
.sage-status {
  position: absolute;
  z-index: 2;
  left: 24px;
  top: 56px;
  color: #e6dcc3;
  font-size: 14px;
}
.sage-iframe {
  height: calc(100vh - 52px);
  border: none;
  width: 100%;
}
.code-display {
  font-size: 14px !important;
  max-width: 100% !important;
  width: 100%;
}
</style>

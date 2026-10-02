export const ATLAS_SAGE_CHANNEL = "spire-atlas-sage"

export type AtlasSageSource = "atlas" | "sage"

export type AtlasSageMsg = {
  source: AtlasSageSource
  type: "hello" | "zone" | "focus" | "click"
  zone?: string
  x?: number
  y?: number
  z?: number
  heading?: number
  id?: number
  table?: string
  name?: string
  t?: number
}

function parseMsg(raw: any): AtlasSageMsg | null {
  if (!raw || typeof raw !== "object") {
    return null
  }
  if (raw.channel && raw.channel !== ATLAS_SAGE_CHANNEL) {
    return null
  }
  if (raw.source !== "atlas" && raw.source !== "sage") {
    return null
  }
  return raw as AtlasSageMsg
}

export function createAtlasSageBridge(source: AtlasSageSource, onMessage: (msg: AtlasSageMsg) => void) {
  const ch = typeof BroadcastChannel !== "undefined" ? new BroadcastChannel(ATLAS_SAGE_CHANNEL) : null
  const deliver = (raw: any) => {
    const msg = parseMsg(raw)
    if (!msg || msg.source === source) {
      return
    }
    onMessage(msg)
  }
  if (ch) {
    ch.onmessage = (ev) => deliver(ev.data)
  }
  const onStorage = (e: StorageEvent) => {
    if (e.key !== ATLAS_SAGE_CHANNEL || !e.newValue) {
      return
    }
    try {
      deliver(JSON.parse(e.newValue))
    } catch (_err) {
      // ignore
    }
  }
  const onWindow = (e: MessageEvent) => {
    deliver(e.data)
  }
  window.addEventListener("storage", onStorage)
  window.addEventListener("message", onWindow)

  function send(msg: Omit<AtlasSageMsg, "source" | "t">) {
    const payload: AtlasSageMsg = Object.assign({}, msg, {source, t: Date.now()})
    if (ch) {
      ch.postMessage(payload)
    }
    try {
      sessionStorage.setItem(ATLAS_SAGE_CHANNEL, JSON.stringify(payload))
    } catch (_err) {
      // ignore
    }
    try {
      localStorage.setItem(ATLAS_SAGE_CHANNEL, JSON.stringify(payload))
    } catch (_err) {
      // ignore
    }
    return payload
  }

  function consumePending(): AtlasSageMsg | null {
    try {
      const raw = sessionStorage.getItem(ATLAS_SAGE_CHANNEL)
      if (!raw) {
        return null
      }
      const msg = parseMsg(JSON.parse(raw))
      if (msg && msg.source !== source) {
        return msg
      }
    } catch (_err) {
      // ignore
    }
    return null
  }

  function dispose() {
    if (ch) {
      ch.close()
    }
    window.removeEventListener("storage", onStorage)
    window.removeEventListener("message", onWindow)
  }

  return {send, consumePending, dispose}
}

export function sageHref(zone: string, coords?: {x?: number; y?: number; z?: number}) {
  const q = ["zone=" + encodeURIComponent(String(zone || "").toLowerCase())]
  if (coords) {
    if (coords.x != null) {
      q.push("x=" + Number(coords.x).toFixed(2))
    }
    if (coords.y != null) {
      q.push("y=" + Number(coords.y).toFixed(2))
    }
    if (coords.z != null) {
      q.push("z=" + Number(coords.z).toFixed(2))
    }
  }
  return "/sage?" + q.join("&")
}

export function installSageIframeBridge(win: Window, parentOrigin: string) {
  if (!win || (win as any).__spireAtlasBridge) {
    return
  }
  ;(win as any).__spireAtlasBridge = true
  const Channel = (win as any).BroadcastChannel
  const ch = typeof Channel !== "undefined" ? new Channel(ATLAS_SAGE_CHANNEL) : null

  function currentZone() {
    const gc = (win as any).gameController || (win as any).gc
    const store = (win as any).GlobalStore
    const fromZone = gc && gc.ZoneController && gc.ZoneController.zoneName
    const fromStore = store && store.getState && store.getState().zoneInfo && store.getState().zoneInfo.shortName
    return String(fromZone || fromStore || "").toLowerCase()
  }

  function sageCanLoad(gc: any) {
    return !!(gc && gc.ZoneController && gc.ZoneController.zoneLoaded)
  }

  function applyFocus(msg: AtlasSageMsg) {
    const gc = (win as any).gameController || (win as any).gc
    if (!gc) {
      return false
    }
    const zc = gc.ZoneController
    const cam = gc.CameraController && gc.CameraController.camera
    const zone = String(msg.zone || "").toLowerCase()
    const here = currentZone()
    if (zone && here && here !== zone && sageCanLoad(gc) && zc && typeof zc.loadModel === "function") {
      zc.loadModel(zone).then(() => {
        zc.zoneName = zone
        applyFocus(msg)
      }).catch(() => {
        // Sage only loads zones it has already converted
      })
      return true
    }
    if (cam && msg.x != null && msg.y != null) {
      cam.position.x = Number(msg.y) || 0
      cam.position.y = (Number(msg.z) || 0) + 5
      cam.position.z = Number(msg.x) || 0
    }
    return true
  }

  function emit(type: AtlasSageMsg["type"], extra: Partial<AtlasSageMsg> = {}) {
    const payload: AtlasSageMsg = Object.assign({
      source: "sage" as AtlasSageSource,
      type,
      zone: currentZone(),
      t: Date.now(),
    }, extra)
    if (ch) {
      ch.postMessage(payload)
    }
    try {
      win.parent.postMessage(Object.assign({channel: ATLAS_SAGE_CHANNEL}, payload), parentOrigin)
    } catch (_err) {
      // ignore
    }
    try {
      win.sessionStorage.setItem(ATLAS_SAGE_CHANNEL, JSON.stringify(payload))
      win.localStorage.setItem(ATLAS_SAGE_CHANNEL, JSON.stringify(payload))
    } catch (_err) {
      // ignore
    }
  }

  function hook() {
    const gc = (win as any).gameController || (win as any).gc
    if (!gc || !gc.ZoneController) {
      return false
    }
    const zc = gc.ZoneController
    if (zc._spireAtlasHooked) {
      return true
    }
    zc._spireAtlasHooked = true
    const onClick = (spawn: any) => {
      emit("click", {
        x: spawn && spawn.x,
        y: spawn && spawn.y,
        z: spawn && spawn.z,
        heading: spawn && spawn.heading,
        id: spawn && (spawn.id || spawn.spawn_id),
        table: "spawn2",
        name: spawn && (spawn.name || spawn.displayed_name),
      })
    }
    if (typeof zc.addClickCallback === "function") {
      zc.addClickCallback(onClick)
    }
    if (gc.SpawnController && typeof gc.SpawnController.addClickCallback === "function") {
      gc.SpawnController.addClickCallback(onClick)
    }
    if (typeof zc.addLoadCallback === "function") {
      zc.addLoadCallback(() => emit("zone"))
    }
    emit("hello")
    return true
  }

  const timer = win.setInterval(() => {
    if (hook()) {
      win.clearInterval(timer)
    }
  }, 400)
  if (ch) {
    ch.onmessage = (ev: MessageEvent) => {
      const msg = parseMsg(ev.data)
      if (!msg || msg.source === "sage") {
        return
      }
      if (msg.type === "focus" || msg.type === "zone" || msg.type === "hello") {
        applyFocus(msg)
      }
    }
  }
  win.addEventListener("message", (ev: MessageEvent) => {
    const msg = parseMsg(ev.data)
    if (!msg || msg.source === "sage") {
      return
    }
    if (msg.type === "focus" || msg.type === "zone" || msg.type === "hello") {
      applyFocus(msg)
    }
  })
}

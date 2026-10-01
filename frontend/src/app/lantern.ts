import {SpireApi} from "@/app/api/spire-api"

export type LanternStatus = {
  ok: boolean
  root: string
  totalZones: number
  zonesWithModels: number
  error?: string
}

export type LanternZoneInfo = {
  zone: string
  modelRelPath: string
  hasMesh: boolean
  instanceCount?: number
}

export type LanternObjectInstance = {
  modelName: string
  modelRelPath: string
  pos: number[]
  rot: number[]
  scale: number[]
  colorIndex: number
}

export class LanternApi {
  static async status(): Promise<LanternStatus> {
    const r = await SpireApi.v1().get("/zone-editor/lantern/status")
    return r.data
  }

  static async zone(name: string): Promise<LanternZoneInfo & { ok?: boolean; error?: string }> {
    const r = await SpireApi.v1().get(`/zone-editor/lantern/zones/${encodeURIComponent(name)}`)
    return r.data
  }

  static async instances(name: string): Promise<LanternObjectInstance[]> {
    const r = await SpireApi.v1().get(`/zone-editor/lantern/zones/${encodeURIComponent(name)}/instances`)
    return r.data && r.data.instances ? r.data.instances : []
  }

  static async loadFileUrl(rel: string): Promise<string> {
    const r = await SpireApi.v1().get("/zone-editor/lantern/file", {
      params: {rel},
      responseType: "arraybuffer",
      timeout: 180000,
    })
    const blob = new Blob([r.data], {type: "model/gltf-binary"})
    return URL.createObjectURL(blob)
  }
}

import {SpireApi} from "@/app/api/spire-api"

export type PlacementCoords = {
  x: number
  y: number
  z: number
  heading: number
}

export type Spawn2PlacementChange = PlacementCoords & {
  table: "spawn2"
  id: number
}

export type SavePlacementsRequest = {
  zone: string
  version: number
  changes: Spawn2PlacementChange[]
}

export type SavePlacementsResponse = {
  data?: {
    updated: number
    ids: number[]
  }
  error?: string
}

export class ZoneEditorApi {
  static async savePlacements(body: SavePlacementsRequest): Promise<SavePlacementsResponse> {
    const r = await SpireApi.v1().post("/zone-editor/placements", body)
    return r.data
  }
}

export function coordsEqual(a: PlacementCoords, b: PlacementCoords): boolean {
  return a.x === b.x && a.y === b.y && a.z === b.z && a.heading === b.heading
}

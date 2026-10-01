import { SpireApi } from "@/app/api/spire-api"

export type PeqRawListResponse = {
  rows?: Record<string, any>[]
  primaryKey?: string[]
  columns?: string[]
  error?: string
}

export class PeqRawApi {
  static async tables(): Promise<string[]> {
    const r = await SpireApi.v1().get("/peq-raw/tables")
    return (r.data && r.data.tables) || []
  }

  static async list(table: string, search: string, page: number, limit: number): Promise<PeqRawListResponse> {
    const r = await SpireApi.v1().get("/peq-raw/" + encodeURIComponent(table), {
      params: { search, page, limit },
    })
    return r.data || {}
  }

  static async count(table: string): Promise<number> {
    const r = await SpireApi.v1().get("/peq-raw/" + encodeURIComponent(table) + "/count")
    const body = r.data
    if (typeof body === "number") {
      return body
    }
    return (body && (body.count || 0)) || 0
  }

  static async create(table: string, row: Record<string, any>) {
    return SpireApi.v1().put("/peq-raw/" + encodeURIComponent(table), row)
  }

  static async update(table: string, row: Record<string, any>, pk: Record<string, any>) {
    return SpireApi.v1().patch("/peq-raw/" + encodeURIComponent(table), row, { params: pk })
  }

  static async remove(table: string, pk: Record<string, any>) {
    return SpireApi.v1().delete("/peq-raw/" + encodeURIComponent(table), { params: pk })
  }
}

import {SpireApi} from "@/app/api/spire-api"

export type ServerFilesStatus = {
  ok: boolean
  connected?: boolean
  source: string
  root: string
  writable: boolean
  error?: string
  detected?: {
    serverPath?: string
    questsDir?: string
    configQuests?: string
    envRoot?: string
    suggestions?: string[]
  }
}

export type ServerFileEntry = {
  name: string
  path: string
  isDir: boolean
  size: number
  modified: number
  editable: boolean
}

export class ServerFilesApi {
  static async status(): Promise<ServerFilesStatus> {
    const r = await SpireApi.v1().get("/admin/server-files/status")
    return r.data
  }

  static async connect(source: string, root: string, save = true): Promise<ServerFilesStatus> {
    const r = await SpireApi.v1().post("/admin/server-files/connect", {source, root, save})
    return r.data
  }

  static async list(path = ""): Promise<{entries: ServerFileEntry[]; writable: boolean; root: string; source: string}> {
    const r = await SpireApi.v1().get("/admin/server-files/list", {params: {path}})
    return r.data
  }

  static async read(path: string): Promise<{content: string; writable: boolean; path: string}> {
    const r = await SpireApi.v1().get("/admin/server-files/file", {params: {path}})
    return r.data
  }

  static async save(path: string, content: string) {
    const r = await SpireApi.v1().put("/admin/server-files/file", {path, content})
    return r.data
  }

  static async create(path: string, content = "") {
    const r = await SpireApi.v1().post("/admin/server-files/file", {path, content})
    return r.data
  }

  static async search(q: string, path = ""): Promise<ServerFileEntry[]> {
    const r = await SpireApi.v1().get("/admin/server-files/search", {params: {q, path}})
    return r.data && r.data.entries ? r.data.entries : []
  }
}

import {SpireApi} from "@/app/api/spire-api"

export type WebsiteTable = {name: string; present: boolean; count: number}

export class WebsiteApi {
  static async status() {
    const r = await SpireApi.v1().get("/admin/website/status")
    return r.data
  }

  static async users(q = "") {
    const r = await SpireApi.v1().get("/admin/website/users", {params: {q}})
    return r.data
  }

  static async updateUserRole(id: number, role: string) {
    const r = await SpireApi.v1().patch("/admin/website/users/" + id + "/role", {role})
    return r.data
  }

  static async accountLinks(q = "") {
    const r = await SpireApi.v1().get("/admin/website/account-links", {params: {q}})
    return r.data
  }

  static async createAccountLink(web_user_id: number, eq_account_id: number) {
    const r = await SpireApi.v1().post("/admin/website/account-links", {web_user_id, eq_account_id})
    return r.data
  }

  static async deleteAccountLink(id: number) {
    const r = await SpireApi.v1().delete("/admin/website/account-links/" + id)
    return r.data
  }

  static async characterLinks(q = "") {
    const r = await SpireApi.v1().get("/admin/website/character-links", {params: {q}})
    return r.data
  }

  static async createCharacterLink(web_user_id: number, character_id: number) {
    const r = await SpireApi.v1().post("/admin/website/character-links", {web_user_id, character_id})
    return r.data
  }

  static async deleteCharacterLink(id: number) {
    const r = await SpireApi.v1().delete("/admin/website/character-links/" + id)
    return r.data
  }

  static async eqAccounts(q: string) {
    const r = await SpireApi.v1().get("/admin/website/eq-accounts", {params: {q}})
    return r.data
  }

  static async eqCharacters(q: string) {
    const r = await SpireApi.v1().get("/admin/website/eq-characters", {params: {q}})
    return r.data
  }
}

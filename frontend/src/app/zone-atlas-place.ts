import {ItemApi, ObjectApi, Spawn2Api, SpawnentryApi, SpawngroupApi} from "@/app/api"
import {SpireApi} from "@/app/api/spire-api"
import {SpireQueryBuilder} from "@/app/api/spire-query-builder"
import {Npcs} from "@/app/npcs"

function unwrapCreated(data: any) {
  if (!data) {
    return null
  }
  if (Array.isArray(data)) {
    return data[0] || null
  }
  if (data.data) {
    return unwrapCreated(data.data)
  }
  return data
}

export async function searchNpcs(query: string) {
  const q = String(query || "").trim()
  if (!q) {
    return []
  }
  const rows = await Npcs.listNpcsByName(`%${q}%`, [])
  return (rows || []).slice(0, 40).map((n: any) => ({
    id: Number(n.id),
    name: Npcs.getCleanName(n.name || `NPC ${n.id}`),
    race: Number(n.race) || 1,
    gender: Number.isFinite(Number(n.gender)) ? Number(n.gender) : 2,
    texture: Number(n.texture) || 0,
    helmtexture: Number(n.helmtexture) || 0,
    size: Number(n.size) || 6,
    runspeed: Number(n.runspeed) || 1.25,
    walkspeed: Number(n.walkspeed) || 0,
  }))
}

export async function searchItems(query: string) {
  const q = String(query || "").trim()
  if (!q) {
    return []
  }
  const api = new ItemApi(...SpireApi.cfg())
  const r = await api.listItems(
    (new SpireQueryBuilder())
      .where("name", "like", `%${q}%`)
      .limit(40)
      .get() as any
  )
  return (r.data || []).slice(0, 40).map((item: any) => ({
    id: Number(item.id),
    name: item.name || `Item ${item.id}`,
    icon: Number(item.icon) || 0,
    idfile: String(item.idfile || ""),
  }))
}

export async function listZoneObjects(zoneId: number, version: number) {
  const api = new ObjectApi(...SpireApi.cfg())
  const r = await api.listObjects(
    (new SpireQueryBuilder())
      .where("zoneid", "=", zoneId)
      .where("version", "=", version)
      .limit(10000)
      .get() as any
  )
  return r.data || []
}

export async function createNpcSpawn(opts: {
  zone: string
  version: number
  npc: any
  x: number
  y: number
  z: number
  heading: number
}) {
  const groupApi = new SpawngroupApi(...SpireApi.cfg())
  const spawnApi = new Spawn2Api(...SpireApi.cfg())
  const entryApi = new SpawnentryApi(...SpireApi.cfg())
  const stamp = Date.now() % 100000
  const groupRes = await groupApi.createSpawngroup({
    spawngroup: {
      name: `${opts.zone}_${opts.npc.id}_${stamp}`,
      spawn_limit: 0,
      dist: 0,
      max_x: 0,
      min_x: 0,
      max_y: 0,
      min_y: 0,
      delay: 0,
      mindelay: 15000,
    } as any,
  })
  const group = unwrapCreated(groupRes.data)
  const groupId = Number(group && group.id)
  if (!groupId) {
    throw new Error("Could not create spawn group")
  }
  await entryApi.createSpawnentry({
    spawnentry: {
      spawngroup_id: groupId,
      npc_id: Number(opts.npc.id),
      chance: 100,
    } as any,
  })
  const spawnRes = await spawnApi.createSpawn2({
    spawn2: {
      spawngroup_id: groupId,
      zone: opts.zone,
      version: opts.version,
      x: opts.x,
      y: opts.y,
      z: opts.z,
      heading: opts.heading,
      respawntime: 60,
      variance: 0,
      pathgrid: 0,
      path_when_zone_idle: 0,
    } as any,
  })
  const spawn = unwrapCreated(spawnRes.data)
  if (!spawn || !spawn.id) {
    throw new Error("Could not create spawn2")
  }
  return spawn
}

export async function createZoneObject(opts: {
  zoneId: number
  version: number
  objectname: string
  x: number
  y: number
  z: number
  heading: number
  icon?: number
  itemid?: number
  type?: number
}) {
  const api = new ObjectApi(...SpireApi.cfg())
  const r = await api.createObject({
    object: {
      zoneid: opts.zoneId,
      version: opts.version,
      xpos: opts.x,
      ypos: opts.y,
      zpos: opts.z,
      heading: opts.heading,
      objectname: opts.objectname,
      itemid: Number(opts.itemid) || 0,
      charges: 0,
      type: Number(opts.type) || 0,
      icon: Number(opts.icon) || 0,
      size: 100,
    } as any,
  })
  const row = unwrapCreated(r.data)
  if (!row || !row.id) {
    throw new Error("Could not create object")
  }
  return row
}

export async function updateZoneObjectIcon(id: number, icon: number, itemid = 0) {
  const api = new ObjectApi(...SpireApi.cfg())
  await api.updateObject({
    id,
    object: {
      id,
      icon: Number(icon) || 0,
      itemid: Number(itemid) || 0,
    } as any,
  })
}

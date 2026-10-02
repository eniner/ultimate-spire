import {GridApi, GridEntryApi} from "@/app/api"
import {SpireApi} from "@/app/api/spire-api"
import {SpireQueryBuilder} from "@/app/api/spire-query-builder"

export type PathWaypoint = {
  number: number
  x: number
  y: number
  z: number
  heading: number
  pause: number
  centerpoint: number
}

export type PathGrid = {
  id: number
  type: number
  type2: number
  points: PathWaypoint[]
}

export type PathState = {
  from: PathWaypoint
  toIndex: number
  dir: number
  dest: number
  t: number
  pause: number
  hidden: boolean
}

export const GRID_CIRCULAR = 0
export const GRID_RANDOM10 = 1
export const GRID_RANDOM = 2
export const GRID_PATROL = 3
export const GRID_ONEWAY_REPOP = 4
export const GRID_RAND5_LOS = 5
export const GRID_ONEWAY_DEPOP = 6
export const GRID_CENTER = 7
export const GRID_RANDOM_CENTER = 8
export const GRID_RANDOM_PATH = 9

export const PAUSE_RANDOM_HALF = 0
export const PAUSE_FULL = 1
export const PAUSE_RANDOM = 2

export function gridTypeLabel(type: number) {
  switch (Number(type) || 0) {
    case GRID_CIRCULAR:
      return "circular"
    case GRID_RANDOM10:
      return "random 10"
    case GRID_RANDOM:
      return "random"
    case GRID_PATROL:
      return "patrol"
    case GRID_ONEWAY_REPOP:
      return "one-way"
    case GRID_RAND5_LOS:
      return "random LoS"
    case GRID_ONEWAY_DEPOP:
      return "one-way depop"
    case GRID_CENTER:
      return "center"
    case GRID_RANDOM_CENTER:
      return "random center"
    case GRID_RANDOM_PATH:
      return "random path"
    default:
      return `type ${type}`
  }
}

export function gridMoves(type: number, pointCount: number) {
  return pointCount >= 2
}

export async function loadZonePathing(zoneId: number): Promise<Map<number, PathGrid>> {
  const grids = new Map<number, PathGrid>()
  const gridApi = new GridApi(...SpireApi.cfg())
  const entryApi = new GridEntryApi(...SpireApi.cfg())
  const headerReq = (new SpireQueryBuilder())
    .where("zoneid", "=", zoneId)
    .limit(100000)
    .get() as any
  const entryReq = (new SpireQueryBuilder())
    .where("zoneid", "=", zoneId)
    .orderBy(["gridid", "number"])
    .limit(100000)
    .get() as any
  const [headers, entries] = await Promise.all([
    gridApi.listGrids(headerReq),
    entryApi.listGridEntries(entryReq),
  ])
  for (const row of headers.data || []) {
    grids.set(Number(row.id), {
      id: Number(row.id),
      type: Number(row.type) || 0,
      type2: Number((row as any).type_2 != null ? (row as any).type_2 : (row as any).type2) || 0,
      points: [],
    })
  }
  for (const row of entries.data || []) {
    const id = Number(row.gridid)
    if (!grids.has(id)) {
      grids.set(id, {id, type: GRID_CIRCULAR, type2: 0, points: []})
    }
    grids.get(id)!.points.push({
      number: Number(row.number) || 0,
      x: Number(row.x) || 0,
      y: Number(row.y) || 0,
      z: Number(row.z) || 0,
      heading: Number(row.heading) || 0,
      pause: Number(row.pause) || 0,
      centerpoint: Number(row.centerpoint) || 0,
    })
  }
  for (const grid of grids.values()) {
    grid.points.sort((a, b) => a.number - b.number)
  }
  return grids
}

export function eqDistance(a: {x: number; y: number; z: number}, b: {x: number; y: number; z: number}) {
  const dx = a.x - b.x
  const dy = a.y - b.y
  const dz = a.z - b.z
  return Math.sqrt(dx * dx + dy * dy + dz * dz)
}

export function lerpWaypoint(a: PathWaypoint, b: PathWaypoint, t: number): PathWaypoint {
  const u = Math.max(0, Math.min(1, t))
  return {
    number: b.number,
    x: a.x + (b.x - a.x) * u,
    y: a.y + (b.y - a.y) * u,
    z: a.z + (b.z - a.z) * u,
    heading: b.heading,
    pause: 0,
    centerpoint: 0,
  }
}

export function nearestWaypointIndex(grid: PathGrid, x: number, y: number, z: number) {
  let best = 0
  let bestDist = Infinity
  for (let i = 0; i < grid.points.length; i++) {
    const d = eqDistance(grid.points[i], {x, y, z})
    if (d < bestDist) {
      bestDist = d
      best = i
    }
  }
  return best
}

function nearestIndexes(grid: PathGrid, from: number, count: number) {
  const scored = grid.points.map((p, i) => ({
    i,
    d: i === from ? Infinity : eqDistance(grid.points[from], p),
  }))
  scored.sort((a, b) => a.d - b.d)
  return scored.slice(0, Math.max(1, count)).map((s) => s.i)
}

function pickOther(n: number, from: number) {
  if (n < 2) {
    return from
  }
  let next = from
  let guard = 0
  while (next === from && guard < 12) {
    next = Math.floor(Math.random() * n)
    guard += 1
  }
  return next
}

function centerIndex(grid: PathGrid) {
  const marked = grid.points.findIndex((p) => p.centerpoint)
  return marked >= 0 ? marked : 0
}

function centerIndexes(grid: PathGrid) {
  const idx = grid.points.map((p, i) => (p.centerpoint ? i : -1)).filter((i) => i >= 0)
  return idx.length ? idx : [0]
}

export function waypointPauseSeconds(grid: PathGrid, pause: number) {
  const raw = Number(pause) || 0
  if (raw <= 0) {
    return 0
  }
  const kind = Number(grid.type2) || 0
  if (kind === PAUSE_FULL) {
    return raw
  }
  if (kind === PAUSE_RANDOM) {
    return Math.random() * raw
  }
  return raw * 0.5 + Math.random() * (raw * 0.5)
}

export function eqHeadingFromTo(from: {x: number; y: number}, to: {x: number; y: number}) {
  const dx = to.x - from.x
  const dy = to.y - from.y
  if (Math.abs(dx) < 0.01 && Math.abs(dy) < 0.01) {
    return 0
  }
  const heading = (Math.atan2(dx, dy) / (Math.PI * 2)) * 512
  return ((heading % 512) + 512) % 512
}

export function nextWaypoint(grid: PathGrid, from: number, dir: number, dest: number) {
  const n = grid.points.length
  if (n < 2) {
    return {index: from, dir, dest, done: true}
  }
  const type = Number(grid.type) || 0
  if (type === GRID_RANDOM) {
    return {index: pickOther(n, from), dir: 1, dest: -1, done: false}
  }
  if (type === GRID_RANDOM10) {
    const near = nearestIndexes(grid, from, 10)
    return {index: near[Math.floor(Math.random() * near.length)], dir: 1, dest: -1, done: false}
  }
  if (type === GRID_RAND5_LOS) {
    const near = nearestIndexes(grid, from, 5)
    return {index: near[Math.floor(Math.random() * near.length)], dir: 1, dest: -1, done: false}
  }
  if (type === GRID_PATROL) {
    let next = from + dir
    if (next >= n) {
      return {index: n - 2 >= 0 ? n - 2 : 0, dir: -1, dest: -1, done: false}
    }
    if (next < 0) {
      return {index: 1 < n ? 1 : 0, dir: 1, dest: -1, done: false}
    }
    return {index: next, dir, dest: -1, done: false}
  }
  if (type === GRID_ONEWAY_REPOP || type === GRID_ONEWAY_DEPOP) {
    if (from >= n - 1) {
      return {index: from, dir: 0, dest: -1, done: true}
    }
    return {index: from + 1, dir: 1, dest: -1, done: false}
  }
  if (type === GRID_CENTER) {
    const center = centerIndex(grid)
    if (from === center) {
      return {index: pickOther(n, center), dir: 1, dest: -1, done: false}
    }
    return {index: center, dir: 1, dest: -1, done: false}
  }
  if (type === GRID_RANDOM_CENTER) {
    const centers = centerIndexes(grid)
    const atCenter = centers.indexOf(from) >= 0
    if (atCenter) {
      const others: number[] = []
      for (let i = 0; i < n; i++) {
        if (centers.indexOf(i) < 0) {
          others.push(i)
        }
      }
      const pool = others.length ? others : nearestIndexes(grid, from, n)
      return {index: pool[Math.floor(Math.random() * pool.length)], dir: 1, dest: -1, done: false}
    }
    return {index: centers[Math.floor(Math.random() * centers.length)], dir: 1, dest: -1, done: false}
  }
  if (type === GRID_RANDOM_PATH) {
    let target = dest
    if (target < 0 || target === from) {
      target = pickOther(n, from)
    }
    if (from === target) {
      target = pickOther(n, from)
    }
    const step = target > from ? 1 : -1
    return {index: from + step, dir: step, dest: target, done: false}
  }
  return {index: (from + 1) % n, dir: 1, dest: -1, done: false}
}

export function nextWaypointIndex(grid: PathGrid, from: number, dir: number) {
  const nxt = nextWaypoint(grid, from, dir, -1)
  return {index: nxt.index, dir: nxt.dir}
}

export function makePathState(grid: PathGrid, spawn: {x: number; y: number; z: number; heading?: number}): PathState | null {
  if (!grid || !gridMoves(grid.type, grid.points.length)) {
    return null
  }
  const start = nearestWaypointIndex(grid, spawn.x, spawn.y, spawn.z)
  return {
    from: {
      number: 0,
      x: Number(spawn.x) || 0,
      y: Number(spawn.y) || 0,
      z: Number(spawn.z) || 0,
      heading: Number(spawn.heading) || 0,
      pause: 0,
      centerpoint: 0,
    },
    toIndex: start,
    dir: 1,
    dest: -1,
    t: 0,
    pause: 0,
    hidden: false,
  }
}

export function stepPathState(grid: PathGrid, state: PathState, speed: number, dt: number) {
  if (!state || !grid || !grid.points.length) {
    return state
  }
  if (state.hidden) {
    state.pause -= dt
    if (state.pause <= 0) {
      state.hidden = false
      state.toIndex = 0
      state.t = 0
      state.dir = 1
      state.dest = -1
    }
    return state
  }
  if (state.pause > 0) {
    state.pause -= dt
    return state
  }
  const to = grid.points[state.toIndex]
  if (!to) {
    return state
  }
  const dist = Math.max(0.1, eqDistance(state.from, to))
  state.t += (speed * dt) / dist
  while (state.t >= 1) {
    state.t -= 1
    state.from = {...to}
    const arrivedHeading = Number(to.heading)
    if (arrivedHeading >= 0) {
      state.from.heading = arrivedHeading
    }
    state.pause = waypointPauseSeconds(grid, to.pause)
    const type = Number(grid.type) || 0
    if ((type === GRID_ONEWAY_REPOP || type === GRID_ONEWAY_DEPOP) && state.toIndex >= grid.points.length - 1) {
      if (type === GRID_ONEWAY_DEPOP) {
        state.hidden = true
        state.pause = Math.max(state.pause, 8)
      } else {
        state.from = {...grid.points[0]}
        state.toIndex = 0
        state.t = 0
        state.pause = Math.max(state.pause, 4)
      }
      break
    }
    const nxt = nextWaypoint(grid, state.toIndex, state.dir, state.dest)
    state.toIndex = nxt.index
    state.dir = nxt.dir
    state.dest = nxt.dest
    if (nxt.done) {
      state.t = 1
      break
    }
  }
  return state
}

export function currentPathPose(grid: PathGrid, state: PathState) {
  const to = grid.points[state.toIndex] || state.from
  const here = lerpWaypoint(state.from, to, state.t)
  here.heading = eqHeadingFromTo(state.from, to)
  if (state.pause > 0 && Number(state.from.heading) >= 0) {
    here.heading = state.from.heading
  }
  return here
}

export function eqWalkSpeed(walkspeed: number, runspeed: number) {
  const walk = Number(walkspeed)
  const run = Number(runspeed)
  const base = Number.isFinite(walk) && walk > 0
    ? walk
    : ((Number.isFinite(run) && run > 0 ? run : 1.25) * 0.46)
  return Math.max(6, base * 32)
}

export function eqRunSpeed(runspeed: number) {
  const speed = Number(runspeed)
  return Math.max(12, (Number.isFinite(speed) && speed > 0 ? speed : 1.25) * 32)
}

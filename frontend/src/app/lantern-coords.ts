// Lantern zone GLBs ship with a node matrix of scale (-0.1, 0.1, 0.1).
// Client objects and spawn2 overlays must use the same scale or the
// terrain collapses to a speck at the origin while trees sit in empty space.
export const WORLD_SCALE = 0.1
export const ZONE_MIRROR_X = -1

export type CoordMapId = "x_y" | "x_-y" | "-x_y" | "-x_-y" | "y_x" | "y_-x" | "-y_x" | "-y_-x"

export type CoordTransform = {
  id: CoordMapId
  map: (x: number, y: number) => [number, number]
}

export const NPC_COORD_TRANSFORMS: CoordTransform[] = [
  {id: "x_y", map: (x, y) => [x, y]},
  {id: "x_-y", map: (x, y) => [x, -y]},
  {id: "-x_y", map: (x, y) => [-x, y]},
  {id: "-x_-y", map: (x, y) => [-x, -y]},
  {id: "y_x", map: (x, y) => [y, x]},
  {id: "y_-x", map: (x, y) => [y, -x]},
  {id: "-y_x", map: (x, y) => [-y, x]},
  {id: "-y_-x", map: (x, y) => [-y, -x]},
]

export function eqToWorld(x: number, y: number, z: number, transform: CoordTransform) {
  const [mx, mz] = transform.map(x, y)
  return {
    x: mx * WORLD_SCALE,
    y: z * WORLD_SCALE,
    z: mz * WORLD_SCALE,
  }
}

export function lanternInstanceToWorld(pos: number[]) {
  return {
    x: ZONE_MIRROR_X * (pos[0] || 0) * WORLD_SCALE,
    y: (pos[1] || 0) * WORLD_SCALE,
    z: (pos[2] || 0) * WORLD_SCALE,
  }
}

export function worldToLantern(wx: number, wy: number, wz: number) {
  return {
    pos: [
      (wx / WORLD_SCALE) / ZONE_MIRROR_X,
      wy / WORLD_SCALE,
      wz / WORLD_SCALE,
    ],
  }
}

export function degToRad(deg: number) {
  return (Number(deg) || 0) * Math.PI / 180
}

export function radToDeg(rad: number) {
  return (Number(rad) || 0) * 180 / Math.PI
}

function invertMapped(id: CoordMapId, mx: number, mz: number): [number, number] {
  switch (id) {
    case "x_y":
      return [mx, mz]
    case "x_-y":
      return [mx, -mz]
    case "-x_y":
      return [-mx, mz]
    case "-x_-y":
      return [-mx, -mz]
    case "y_x":
      return [mz, mx]
    case "y_-x":
      return [-mz, mx]
    case "-y_x":
      return [mz, -mx]
    case "-y_-x":
      return [-mz, -mx]
    default:
      return [mx, mz]
  }
}

export function worldToEq(wx: number, wy: number, wz: number, transform: CoordTransform) {
  const mapped = invertMapped(transform.id, wx / WORLD_SCALE, wz / WORLD_SCALE)
  return {x: mapped[0], y: mapped[1], z: wy / WORLD_SCALE}
}

export function eqHeadingToYaw(heading: number) {
  return -((Number(heading) || 0) / 512) * Math.PI * 2
}

export function yawToEqHeading(yaw: number) {
  const heading = (-yaw / (Math.PI * 2)) * 512
  return ((heading % 512) + 512) % 512
}

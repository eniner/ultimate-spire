import * as L from "leaflet"

/**
 * Coordinate helpers for the Spire zone map (Leaflet CRS.Simple).
 *
 * EqZoneMap.createPoint(x, y) is:
 *   lat = -y, lng = x
 *
 * Spawn2 / doors / grids / safe point are drawn with createPoint(-eqX, -eqY), so:
 *   lat = eqY, lng = -eqX
 *
 * Map-file L/P records use createPoint(eqX, eqY) without that negation. The two
 * spaces currently line up in the viewer; do not "correct" one without the other.
 * Placement saves MUST invert the spawn path so DB x/y stay in EQ spawn2 space.
 */
export function createPoint(x: number | string, y: number | string): L.LatLng {
  const nx = typeof x === "string" ? parseFloat(x) : x
  const ny = typeof y === "string" ? parseFloat(y) : y
  return L.latLng(-ny, nx)
}

export function eqSpawnToLeaflet(x: number, y: number): L.LatLng {
  return createPoint(-x, -y)
}

export function leafletToEqSpawn(lat: number, lng: number): { x: number; y: number } {
  return {
    x: roundCoord(-lng),
    y: roundCoord(lat),
  }
}

export function roundCoord(n: number): number {
  return Math.round(n * 10000) / 10000
}

<template>
  <div>
    <eq-window class="p-2 zone-editor-map-wrap" style="height: 96vh; position: relative;">

      <!-- Loader -->
      <eq-window
        class="text-center"
        style="position: absolute; right: 3%; z-index: 99; padding: 15px; padding-top: 10px;"
        v-if="!isDataLoaded()"
      >
        <div class="mb-2">
          {{ isDataLoaded() ? 'Rendering map...' : 'Loading map...' }}
        </div>
        <loader-fake-progress v-if="!isDataLoaded()"/>
        <eq-progress-bar :percent="100" v-if="isDataLoaded()"/>
      </eq-window>

      <div class="zone-editor-map-hud" v-if="isDataLoaded()">
        <div class="zone-editor-map-hud-row">
          <label class="mr-2 mb-0">
            <input type="checkbox" v-model="editLayers.npcs"> NPCs
          </label>
          <label class="mr-2 mb-0">
            <input type="checkbox" v-model="editLayers.doors"> Doors
          </label>
          <label class="mr-2 mb-0">
            <input type="checkbox" v-model="editLayers.teleports"> Teleports
          </label>
          <label class="mr-2 mb-0">
            <input type="checkbox" v-model="editLayers.zonelines"> Zone lines
          </label>
          <label class="mr-2 mb-0" v-if="zAbsMax > zAbsMin">
            <input type="checkbox" v-model="zClip"> Z clip
          </label>
          <span class="zone-editor-z-range" v-if="zClip && zAbsMax > zAbsMin">
            <input type="range" :min="zAbsMin" :max="zAbsMax" step="1" v-model.number="zMin">
            <input type="range" :min="zAbsMin" :max="zAbsMax" step="1" v-model.number="zMax">
            <span>{{ zMin }} to {{ zMax }}</span>
          </span>
        </div>
        <div class="zone-editor-mode" v-if="editMode">Edit: click to select, drag to move spawn2</div>
      </div>

      <div class="card">
        <l-map
          v-if="center"
          :crs="crs"
          style="height: 94vh"
          class="map-tiles"
          :center="center"
          :bounds="bounds"
          :min-zoom="-5"
          :zoom="zoom"
          :zoom-animation="true"
          :zoom-animation-threshold="10"
          @update:zoom="zoomUpdate"
        >

          <!-- Draw map lines -->
          <l-polyline
            v-if="visibleLines && visibleLines.length > 0"
            :lat-lngs="visibleLines"
            color="gray"
            :weight="1"
          />

          <!-- grid points -->
          <l-marker
            v-for="(m, index) in pathingGridMarkers"
            :key="index + '-' + m.point.lat + '-' + m.point.lng"
            :lat-lng="m.point"
            v-if="markers && markers.length > 0"
          >
            <l-tooltip :options="{ permanent: true, interactive: true }">
              {{ m.label }}
            </l-tooltip>
          </l-marker>

          <!-- Draw pathing grid lines -->
          <l-polyline
            v-if="pathingGridLines"
            :lat-lngs="pathingGridLines"
            color="blue"
            dashArray="5, 10"
            :opacity=".8"
            :weight="2"
          />

          <!--          &lt;!&ndash; markers from map &ndash;&gt;-->
          <!--          <l-marker-->
          <!--            v-for="(marker, index) in markers"-->
          <!--            :key="index"-->
          <!--            :lat-lng="marker.point"-->
          <!--            v-if="markers && markers.length > 0"-->
          <!--          >-->
          <!--            <l-tooltip>-->
          <!--              <eq-window>{{ marker.label }}-->
          <!--              </eq-window>-->
          <!--            </l-tooltip>-->
          <!--          </l-marker>-->

          <!-- zone points -->
          <l-marker
            v-for="(m, index) in zonelineMarkers"
            :key="index"
            :lat-lng="m.point"
            v-if="editLayers.zonelines && zonelineMarkers && zonelineMarkers.length > 0"
            @click="navigateToZone(m.zone.short_name, m.zone.version)"
          >
            <l-tooltip>
              <eq-window>{{ m.label }}
              </eq-window>
            </l-tooltip>
          </l-marker>

          <!-- door zone points -->
          <l-marker
            v-for="(m, index) in doorZonePoints"
            :key="index + '-' + m.destName + '-' + m.destInstance"
            :lat-lng="m.point"
            v-if="markers && markers.length > 0"
            @click="navigateToZone(m.destName, m.destInstance)"
          >
            <l-tooltip>
              <eq-window>{{ m.label }}
              </eq-window>
            </l-tooltip>
          </l-marker>

          <!-- NPC markers -->
          <l-marker
            v-for="(marker, index) in visibleNpcMarkers"
            :key="'spawn2-' + marker.spawn2Id + '-npc-' + marker.npc.id + '-' + index + '-' + (editMode ? 'edit' : 'view')"
            :lat-lng="marker.point"
            :draggable="editMode"
            :opacity="getNpcOpacity('spawn2-' + marker.spawn2Id + '-npc-' + marker.npc.id + '-' + index, marker.npc.id, marker.spawn2Id)"
            @mouseover="npcMarkerHover(marker, 'spawn2-' + marker.spawn2Id + '-npc-' + marker.npc.id + '-' + index)"
            @click="npcMarkerClick(marker, 'spawn2-' + marker.spawn2Id + '-npc-' + marker.npc.id + '-' + index)"
            @dragstart="onNpcDragStart"
            @dragend="onNpcDragEnd(marker, $event)"
          >

            <l-tooltip>
              <eq-window>
                {{ getCleanName(marker.npc.name) }}<span v-if="editMode"> (spawn2 {{ marker.spawn2Id }})</span>
              </eq-window>
            </l-tooltip>

            <l-icon
              icon-url="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII="
              :class-name="npcIconClass(marker)"
              :iconSize="(zoomLevel >= 1) ? marker.iconSize : calcSmallIcons(marker.iconSize)"
            >
            </l-icon>

          </l-marker>

          <!-- Door markers -->
          <l-marker
            v-for="(marker, index) in doorMarkers"
            :key="index + '-' + marker.name"
            :lat-lng="marker.point"
            v-if="editLayers.doors && doorMarkers && doorMarkers.length > 0"
          >

            <l-tooltip>
              <eq-window>
                {{ marker.label }}
              </eq-window>
            </l-tooltip>

            <l-icon
              icon-url="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII="
              :class-name="marker.iconClass"
              :iconSize="marker.iconSize"
            >
            </l-icon>
          </l-marker>

          <!-- Translocate markers -->
          <l-marker
            v-for="(m, index) in translocatePoints"
            :key="index + '-' + m.label"
            :lat-lng="m.point"
            @mouseover="spellMarkerHover(m.spell)"
            v-if="editLayers.teleports && translocatePoints && translocatePoints.length > 0"
            style="border-radius: 10px"
          >
            <l-tooltip>
              <eq-window>
                {{ m.label }}
              </eq-window>
            </l-tooltip>

            <l-icon
              icon-url="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII="
              :class-name="m.iconClass"
              :iconSize="m.iconSize"
            >
            </l-icon>
          </l-marker>

          <!-- Safe coordinate markers -->
          <l-marker
            v-for="(m, index) in safeCoordinateMarker"
            :key="index + '-' + m.label"
            :lat-lng="m.point"
            v-if="safeCoordinateMarker && safeCoordinateMarker.length > 0"
            style="border-radius: 10px"
          >
            <l-tooltip>
              <eq-window>
                {{ m.label }}
              </eq-window>
            </l-tooltip>

            <l-icon
              icon-url="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII="
              :class-name="m.iconClass"
              :iconSize="m.iconSize"
            >
            </l-icon>
          </l-marker>

        </l-map>
      </div>
    </eq-window>
  </div>
</template>

<script>
import {LIcon, LMap, LMarker, LPolyline, LPopup, LTileLayer, LTooltip} from 'vue2-leaflet';
import ContentArea                                                     from "./layout/ContentArea";
import * as L                                                          from "leaflet";
import axios                                                           from "axios";
import {GridEntryApi, Spawn2Api, SpellsNewApi, ZonePointApi} from "../app/api";
import {SpireApi}                                            from "../app/api/spire-api";
import {SpireQueryBuilder}                                   from "../app/api/spire-query-builder";
import EqNpcCardPreview                                      from "./preview/EQNpcCardPreview";
import EqWindow                                              from "./eq-ui/EQWindow";
import LoaderFakeProgress                                    from "./LoaderFakeProgress";
import EqProgressBar                                         from "./eq-ui/EQProgressBar";
import {Npcs}                                                from "../app/npcs";
import {Zones}                                               from "../app/zones";
import {DoorApi}                                             from "../app/api/api/door-api";
import {EventBus}                                            from "../app/event-bus/event-bus";
import {Spawn}                                               from "../app/spawn";
import {eqSpawnToLeaflet, leafletToEqSpawn, roundCoord}      from "../app/eq-coords";

export default {
  name: "EqZoneMap",
  props: {
    zone: {
      type: String,
      required: true
    },
    version: {
      type: String,
      required: true
    },
    editMode: {
      type: Boolean,
      default: false
    },
  },
  components: {
    EqProgressBar,
    LoaderFakeProgress,
    EqWindow,
    EqNpcCardPreview,
    ContentArea,
    LMap,
    LIcon,
    LMarker,
    LPopup,
    LTooltip,
    LTileLayer,
    LPolyline
  },
  computed: {
    visibleLines() {
      if (!this.mapSegments || this.mapSegments.length === 0) {
        return this.lines || []
      }
      if (!this.zClip) {
        return this.mapSegments.map((s) => s.latlngs)
      }
      const zMin = Math.min(this.zMin, this.zMax)
      const zMax = Math.max(this.zMin, this.zMax)
      return this.mapSegments
        .filter((s) => s.z >= zMin && s.z <= zMax)
        .map((s) => s.latlngs)
    },
    visibleNpcMarkers() {
      if (!this.editLayers.npcs || !this.npcMarkers) {
        return []
      }
      if (!this.zClip) {
        return this.npcMarkers
      }
      const zMin = Math.min(this.zMin, this.zMax)
      const zMax = Math.max(this.zMin, this.zMax)
      return this.npcMarkers.filter((m) => {
        const z = m.spawn2 && typeof m.spawn2.z === "number" ? m.spawn2.z : 0
        return z >= zMin && z <= zMax
      })
    }
  },
  watch: {
    zone() {
      this.loadMap()
    },
    version() {
      this.loadMap()
    },
  },

  methods: {

    handleNpcZoomEvent(e) {
      console.log("[EqZoneMap] Handling Zoom event")
      console.log(e)

      this.zoomedNpcId = e.id

      for (let n of this.npcMarkers) {
        if (n.npc.id === e.id) {
          console.log("found NPC marker at ", n)

          // zoom out first
          this.zoom = this.starterZoomLevel

          // center on target
          setTimeout(() => {
            this.center = n.point
          }, 600)

          // zoom in
          setTimeout(() => {
            this.zoom = 1
          }, 1000)

          break;
        }
      }

    },

    getNpcOpacity(elementKey, npcId, spawn2Id) {
      if (this.selectedSpawn2Id && spawn2Id === this.selectedSpawn2Id) {
        return 1;
      }

      if (this.selectedSpawn2Id && spawn2Id !== this.selectedSpawn2Id) {
        return .35;
      }

      if (this.zoomedNpcId === npcId) {
        return 1;
      }

      if (this.zoomedNpcId > 0 && this.zoomedNpcId !== npcId) {
        return .3;
      }

      if (this.hoveredNpc === "") {
        return 1;
      }

      if (this.hoveredNpc !== "" && this.hoveredNpc !== elementKey) {
        return .3;
      }

      return 1;
    },

    npcIconClass(marker) {
      let cls = (this.zoomLevel >= 1) ? marker.iconClass : marker.iconClass + '-sm'
      if (this.editMode && this.selectedSpawn2Id && marker.spawn2Id === this.selectedSpawn2Id) {
        cls += ' zone-editor-selected'
      }
      return cls
    },

    npcMarkerClick(marker, elementKey) {
      if (this.justDragged) {
        return
      }
      this.selectedSpawn2Id = marker.spawn2Id
      this.selectedMarkerKey = elementKey
      this.$emit("npc-marker-select", {
        npc: marker.npc,
        spawn2: Object.assign({}, marker.spawn2),
        markerKey: elementKey
      })
      this.npcMarkerHover(marker, elementKey)
    },

    onNpcDragStart() {
      this.justDragged = true
    },

    onNpcDragEnd(marker, event) {
      this.justDragged = true
      setTimeout(() => {
        this.justDragged = false
      }, 80)

      if (!this.editMode || !event || !event.target || !event.target.getLatLng) {
        return
      }

      const ll = event.target.getLatLng()
      const eq = leafletToEqSpawn(ll.lat, ll.lng)
      const from = {
        x: marker.spawn2.x,
        y: marker.spawn2.y,
        z: marker.spawn2.z,
        heading: marker.spawn2.heading
      }
      const to = {
        x: eq.x,
        y: eq.y,
        z: marker.spawn2.z,
        heading: marker.spawn2.heading
      }

      this.applySpawn2Placement(marker.spawn2Id, to)
      this.selectedSpawn2Id = marker.spawn2Id
      this.$emit("spawn2-moved", {
        id: marker.spawn2Id,
        from: from,
        to: to,
        spawn2: Object.assign({}, marker.spawn2),
        npc: marker.npc
      })
    },

    applySpawn2Placement(spawn2Id, coords) {
      if (!this.npcMarkers) {
        return
      }
      for (let m of this.npcMarkers) {
        if (m.spawn2Id !== spawn2Id) {
          continue
        }
        m.spawn2.x = roundCoord(coords.x)
        m.spawn2.y = roundCoord(coords.y)
        m.spawn2.z = roundCoord(coords.z)
        m.spawn2.heading = roundCoord(coords.heading)
        m.point = eqSpawnToLeaflet(m.spawn2.x, m.spawn2.y)
      }
      this.$forceUpdate()
    },

    clearSpawnSelection() {
      this.selectedSpawn2Id = 0
      this.selectedMarkerKey = ""
    },

    refreshZExtents() {
      let min = Number.POSITIVE_INFINITY
      let max = Number.NEGATIVE_INFINITY
      if (this.mapSegments) {
        for (let s of this.mapSegments) {
          if (typeof s.z === "number" && isFinite(s.z)) {
            min = Math.min(min, s.z)
            max = Math.max(max, s.z)
          }
        }
      }
      if (this.npcMarkers) {
        for (let m of this.npcMarkers) {
          if (m.spawn2 && typeof m.spawn2.z === "number" && isFinite(m.spawn2.z)) {
            min = Math.min(min, m.spawn2.z)
            max = Math.max(max, m.spawn2.z)
          }
        }
      }
      if (!isFinite(min) || !isFinite(max)) {
        min = 0
        max = 0
      }
      this.zAbsMin = Math.floor(min)
      this.zAbsMax = Math.ceil(max)
      this.zMin = this.zAbsMin
      this.zMax = this.zAbsMax
    },

    // this is not a computed property because the dependencies are not reactive
    isDataLoaded() {
      return this.npcMarkers
    },

    navigateToZone(shortName, version) {
      this.$router.push(
        {
          path: `/zone/${shortName}?v=${version}`
        }
      ).catch(() => {
      })
    },

    getCleanName(n) {
      return Npcs.getCleanName(n)
    },

    npcMarkerHover(e, elementKey) {
      // console.log(e)

      // reset
      this.hoveredNpc  = ""
      this.zoomedNpcId = 0
      if (this.pathingGridLines.length > 0) {
        this.pathingGridLines   = []
        this.pathingGridMarkers = []
        this.$forceUpdate()
      }

      if (e.grid > 0) {
        // transform grid entries into poly lines
        let polyLines   = []
        let gridMarkers = []
        for (const [id, g] of this.pathingGridData.entries()) {
          if (g && id === e.grid) {
            this.hoveredNpc = elementKey

            for (const [i, e] of g.entries()) {

              // make sure we have a valid entry as well as
              // a valid next point so we can draw a complete line
              if (e && e.x && g[i + 1]) {
                // console.log(i, e)
                const current = e
                const next    = g[i + 1]
                polyLines.push(
                  [
                    this.createPoint(-current.x, -current.y),
                    this.createPoint(-next.x, -next.y),
                  ]
                )
              }

              if (e && e.x) {
                // console.log(i, e)
                gridMarkers.push(
                  {
                    point: this.createPoint(-e.x, -e.y),
                    label: i,
                  }
                )
              }
            }
          }
        }

        this.pathingGridLines   = polyLines
        this.pathingGridMarkers = gridMarkers
      }

      this.$emit("npc-marker-hover", e.npc);
    },

    spellMarkerHover(s) {
      this.$emit("spell-marker-hover", s);
    },

    calcSmallIcons(xy) {
      // console.log("small icon")
      // console.log(xy)

      return [
        xy[0] / 2,
        xy[1] / 2,
      ]
    },

    getNpcIcon(npc) {
      return 'race-models-ctn-' + npc.race + '-' + npc.gender + '-' + npc.texture + '-' + npc.helmtexture;
    },

    zoomUpdate(e) {
      console.log("zoom level [%s]", e)

      if (this.starterZoomLevel === -100) {
        this.starterZoomLevel = e
      }

      this.zoomLevel = e
    },
    createPoint(x, y) {
      return L.latLng(
        (typeof (y) === "string" ? -parseFloat(y) : -y),
        (typeof (x) === "string" ? parseFloat(x) : x));
    },
    iconClass() {
      return this.zoomLevel >= 2 ? 'item-4472' : 'item-4472-sm'
    },
    iconSize() {
      return this.zoomLevel >= 2 ? [40, 40] : [12, 12]
    },
    async getMapContents() {
      const postfix = ["", "_1", "_3"]
      let contents  = ""
      for (let p of postfix) {
        try {
          const r = await axios.get(
            `/eq-asset-preview-master/assets/eq-maps/${this.zone}${p}.txt`
          )

          if (r.status === 200) {
            if (r.data.length > 0) {
              contents += r.data
            }
          }

        } catch (err) {
          // console.log("items.ts %s", err)
        }
      }
      return contents
    },
    async parseRaceIconSizes() {

      console.time("[EqZoneMap] parseRaceIconSizes");

      // parse CSS sheet to pull sizes
      let raceIconSizes = {}
      try {
        const r = await axios.get(
          `/eq-asset-preview-master/assets/sprites/race-models.css`
        )

        if (r.status === 200) {
          if (r.data.length > 0) {
            for (let line of r.data.split("\n")) {
              line               = line.replace(".", "")
              const raceClassKey = line.split(" ")[0].trim()
              const height       = line.split("height: ")[1].split(";")[0].replace("px", "").trim()
              const width        = line.split("width: ")[1].split(";")[0].replace("px", "").trim()
              // console.log(raceClassKey)
              // console.log(height)
              // console.log(width)

                  raceIconSizes[raceClassKey] = [parseInt(width, 10) || 30, parseInt(height, 10) || 100]
            }
          }
        }

      } catch (err) {
        console.log("map.vue %s", err)
      }

      this.raceIconSizes = raceIconSizes

      console.timeEnd("[EqZoneMap] parseRaceIconSizes");
    },

    async loadSafeCoordinates(gen) {
      const zone          = (await Zones.getZoneByShortName(this.zone))
      if (this.isStale(gen) || !zone) {
        return
      }
      let safeCoordinates = []
      safeCoordinates.push({
          point: this.createPoint(-zone.safe_x, -zone.safe_y),
          label: `Safe Coordinates (${zone.safe_x}, ${zone.safe_y}, ${zone.safe_z}) (xyz)`,
          iconClass: 'fade-in item-6852',
          iconSize: [40, 40]
        }
      )

      this.safeCoordinateMarker = safeCoordinates
    },

    async loadTranslocatePoints(gen) {
      const api = (new SpellsNewApi(...SpireApi.cfg()))

      try {
        const r = await api.listSpellsNews(
          (new SpireQueryBuilder())
            .where("teleport_zone", "=", this.zone)
            .get()
        )

        if (r.status === 200) {
          // used as a mechanism to stagger multiple markers on the same coordinate
          let sameCoord = {}

          let translocatePoints = []
          for (const s of r.data) {
            if (typeof sameCoord[s.effect_base_value_2 + s.effect_base_value_1] === "undefined") {
              sameCoord[s.effect_base_value_2 + s.effect_base_value_1] = 0
            }

            let sameCoordOffset = sameCoord[s.effect_base_value_2 + s.effect_base_value_1] * 1
            translocatePoints.push({
                point: this.createPoint(-s.effect_base_value_2 + sameCoordOffset, -s.effect_base_value_1 - sameCoordOffset),
                label: s.name,
                spell: s,
                iconClass: 'fade-in spell-' + s.new_icon + '-40',
                iconSize: [40, 40]
              }
            )

            sameCoord[s.effect_base_value_2 + s.effect_base_value_1]++
          }

          if (this.isStale(gen)) {
            return
          }
          this.translocatePoints = translocatePoints
          this.$forceUpdate()
        }

      } catch (err) {
        console.log("map.vue %s", err)
      }
    },

    async loadMapLines() {
      console.time("[EqZoneMap] loadMapLines");

      let map        = await this.getMapContents()
      let bounds     = [0, 0, 0, 0];
      let mapLines   = []
      let mapMarkers = []
      let mapSegments = []
      for (let line of map.split("\n")) {
        const cols = line.replaceAll(",", "").split(/\s+/)

        // lines
        if (cols[0].trim() === "L") {
          const x  = cols[1].trim()
          const y  = cols[2].trim()
          const z1 = cols.length > 3 ? parseFloat(cols[3]) : 0
          const x2 = cols[4].trim()
          const y2 = cols[5].trim()
          const z2 = cols.length > 6 ? parseFloat(cols[6]) : 0
          const p  = [
            this.createPoint(x, y),
            this.createPoint(x2, y2),
          ]
          bounds   = [
            Math.min(bounds[0], p[0].lat, p[1].lat),
            Math.min(bounds[1], p[0].lng, p[1].lng),
            Math.max(bounds[2], p[0].lat, p[1].lat),
            Math.max(bounds[3], p[0].lng, p[1].lng),
          ];

          mapLines.push(p)
          mapSegments.push({
            latlngs: p,
            z: ((isFinite(z1) ? z1 : 0) + (isFinite(z2) ? z2 : 0)) / 2
          })
        }

        // points
        if (cols[0].trim() === "P") {
          const x     = cols[1].trim()
          const y     = cols[2].trim()
          const label = cols[8].trim()

          mapMarkers.push(
            {
              point: this.createPoint(x, y),
              label: label.replaceAll("_", " "),
            }
          )
        }
      }

      this.markers = mapMarkers
      this.lines   = mapLines
      this.mapSegments = mapSegments
      this.bounds  = [
        [bounds[0], bounds[1]],
        [bounds[2], bounds[3]]
      ]

      this.center = [
        (bounds[0] + bounds[2]) / 2,
        (bounds[3] + bounds[1]) / 2
      ];

      this.$forceUpdate()

      console.timeEnd("[EqZoneMap] loadMapLines");
    },

    isStale(gen) {
      return gen != null && gen !== this.loadGeneration
    },

    async loadDoors(gen) {
      console.time("[EqZoneMap] loadDoors");

      const api       = (new DoorApi(...SpireApi.cfg()))
      let doorMarkers = []

      try {
        const r = await api.listDoors(
          (new SpireQueryBuilder())
            .where("zone", "=", this.zone)
            .where("version", "=", this.version)
            .get()
        )

        if (r.status === 200) {

          let doorZonePoints = []

          for (let d of r.data) {
            doorMarkers.push(
              {
                point: this.createPoint(-d.pos_x, -d.pos_y),
                label: d.name,
                iconClass: 'fade-in item-8057',
                iconSize: [40, 40]
              }
            )

            // zone teleport linked door
            if (d.dest_zone !== "NONE") {
              const z = (await Zones.getZoneLongNameByShortName(d.dest_zone))

              // console.log(z)

              if (z !== "") {
                doorZonePoints.push(
                  {
                    point: this.createPoint(-d.pos_x, -d.pos_y),
                    label: "(Door Click) Zone Point (" + z + ")",
                    destName: d.dest_zone,
                    destInstance: d.dest_instance,
                  }
                )

                // console.log(d)
              }

            }
          }

          if (this.isStale(gen)) {
            return
          }
          this.doorMarkers    = doorMarkers
          this.doorZonePoints = doorZonePoints
        }

      } catch (err) {
        console.log("map.vue %s", err)
      }

      console.timeEnd("[EqZoneMap] loadDoors");
    },

    async loadMapSpawns() {
      let npcMarkers       = []
      const gridEntriesApi = (new GridEntryApi(...SpireApi.cfg()))
      try {
        console.time("[EqZoneMap] loadMapSpawns");

        // grids
        const zone = (await Zones.getZoneByShortName(this.zone))
        const r    = await gridEntriesApi.listGridEntries(
          (new SpireQueryBuilder())
            .where("zoneid", "=", zone.zoneidnumber)
            .orderBy(["gridid", "number"])
            .get()
        );
        if (r.status === 200) {
          let gridEntries = []
          for (let e of r.data) {
            if (typeof gridEntries[e.gridid] === "undefined") {
              gridEntries[e.gridid] = []
            }
            if (typeof gridEntries[e.gridid][e.number] === "undefined") {
              gridEntries[e.gridid][e.number] = []
            }

            gridEntries[e.gridid][e.number] =
              {
                x: e.x,
                y: e.y,
              }
          }
          this.pathingGridData = gridEntries
        }

        const result = await Spawn.getByZone(this.zone, this.version, true)
        if (result.length > 0) {
          for (let spawn2 of result) {
            if (spawn2.spawnentries) {
              for (let spawnentry of spawn2.spawnentries) {
                if (spawnentry.npc_type) {

                  // if (spawn.pathgrid > 0) {
                  //   console.log(spawn)
                  // }

                  // make sure we have a npc associated to spawn
                  let npcName = ""

                  const n = spawnentry.npc_type
                  npcName = n.name + (n.lastname ? ` (${n.lastname})` : '')

                  // console.log(this.raceIconSizes[this.getNpcIcon(n)])

                  npcMarkers.push(
                    {
                      point: this.createPoint(-spawn2.x, -spawn2.y),
                      label: Npcs.getCleanName(npcName),
                      npc: n,
                      grid: spawn2.pathgrid,
                      spawn2Id: spawn2.id,
                      spawn2: {
                        id: spawn2.id,
                        x: spawn2.x,
                        y: spawn2.y,
                        z: spawn2.z,
                        heading: spawn2.heading,
                        zone: spawn2.zone,
                        version: spawn2.version,
                        pathgrid: spawn2.pathgrid,
                        spawngroup_id: spawn2.spawngroup_id
                      },
                      iconClass: 'fade-in ' + this.getNpcIcon(n),
                      iconSize: this.raceIconSizes[this.getNpcIcon(n)] ? this.raceIconSizes[this.getNpcIcon(n)] : [30, 100]
                    }
                  )
                }
              }
            }
          }

          this.npcMarkers = npcMarkers

          this.$forceUpdate()
        } else {
          this.npcMarkers = []
        }
        console.timeEnd("[EqZoneMap] loadMapSpawns");
      } catch (err) {
        console.log("map.vue %s", err)
        this.npcMarkers = this.npcMarkers || []
      }
    },

    async loadZonePoints(gen) {
      console.time("[EqZoneMap] loadZonePoints");

      let zonePoints = []
      const zapi     = (new ZonePointApi(...SpireApi.cfg()))
      try {
        const r = await zapi.listZonePoints(
          (new SpireQueryBuilder())
            .where("zone", "=", this.zone)
            .get()
        )
        if (r.status === 200) {
          for (let point of r.data) {
            if (this.isStale(gen)) {
              return
            }
            const z = (await Zones.getZoneById(point.target_zone_id))
            zonePoints.push({
                point: this.createPoint(-point.x, -point.y),
                label: "Zone Point to: " + (z && z.long_name ? z.long_name : point.target_zone_id),
                zone: z,
              }
            )
          }
          if (this.isStale(gen)) {
            return
          }
          this.zonelineMarkers = zonePoints
          this.$forceUpdate()
        }
      } catch (err) {
        console.log("map.vue %s", err)
      }
      console.timeEnd("[EqZoneMap] loadZonePoints");
    },

    async loadMap() {
      const gen = this.loadGeneration + 1
      this.loadGeneration = gen
      // reset
      this.markers              = null
      this.lines                = null
      this.npcMarkers           = null
      this.doorMarkers          = null
      this.safeCoordinateMarker = null
      this.zonelineMarkers      = null
      this.translocatePoints    = null
      this.lines                = []
      this.pathingGridLines     = []
      this.pathingGridMarkers   = null
      this.pathingGridData      = []
      this.mapSegments          = []
      this.selectedSpawn2Id     = 0
      this.selectedMarkerKey    = ""

      // load
      await this.parseRaceIconSizes()
      if (this.isStale(gen)) {
        return
      }
      await this.loadMapLines()
      if (this.isStale(gen)) {
        return
      }
      await this.loadMapSpawns()
      if (this.isStale(gen)) {
        return
      }
      await Promise.all([
        this.loadDoors(gen),
        this.loadZonePoints(gen),
        this.loadTranslocatePoints(gen),
        this.loadSafeCoordinates(gen),
      ])
      if (this.isStale(gen)) {
        return
      }
      this.refreshZExtents()
      this.$emit("map-loaded")

      this.$forceUpdate()
    }

  },
  async mounted() {
    this.loadMap()
  },
  beforeDestroy() {
    EventBus.$off("NPC_ZOOM", this.handleNpcZoomEvent);
  },
  created() {
    EventBus.$on("NPC_ZOOM", this.handleNpcZoomEvent);
  },
  data() {
    return {
      zoom: 0,
      center: null,

      hoveredNpc: "",

      zoomedNpcId: 0,
      starterZoomLevel: -100,

      zoomLevel: 0,

      bounds: null,
      crs: L.CRS.Simple,

      icon: L.icon({
        iconUrl: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=',
        // iconSize: [32, 37],
        // iconAnchor: [16, 37],
        className: this.zoomLevel >= 2 ? 'item-4472' : 'item-4472-sm'
      }),

      map: "",

      markers: null,
      npcMarkers: null,
      doorMarkers: null,
      zonelineMarkers: null,
      doorZonePoints: null,
      translocatePoints: null,
      safeCoordinateMarker: null,
      pathingGridData: [],
      pathingGridLines: [],
      pathingGridMarkers: [],
      lines: [],

      raceIconSizes: {},
      editLayers: {
        npcs: true,
        doors: true,
        teleports: true,
        zonelines: true
      },
      zClip: false,
      zAbsMin: 0,
      zAbsMax: 0,
      zMin: 0,
      zMax: 0,
      selectedSpawn2Id: 0,
      selectedMarkerKey: "",
      mapSegments: [],
      justDragged: false,
      loadGeneration: 0
    };
  }
}
</script>

<style>
.leaflet-tooltip {
  background-color: transparent;
  border: none;
  -webkit-box-shadow: none;
  box-shadow: none;
}

.zone-editor-map-hud {
  position: absolute;
  left: 52px;
  top: 12px;
  z-index: 1000;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
  max-width: 78%;
  padding: 8px 10px;
  background: rgba(18, 22, 27, 0.88);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 6px;
  color: #e7eaef;
  font-size: 12px;
  pointer-events: auto;
}

.zone-editor-map-hud-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.zone-editor-map-hud label {
  white-space: nowrap;
}

.zone-editor-mode,
.zone-editor-z-range {
  white-space: nowrap;
  opacity: 0.9;
}

.zone-editor-selected {
  outline: 3px solid #c9a24a;
  outline-offset: 2px;
  border-radius: 4px;
}

.leaflet-marker-draggable {
  cursor: grab;
}
</style>

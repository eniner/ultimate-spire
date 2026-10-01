<template>
  <eq-window
    id="zone-preview"
    v-if="zone"
    class="p-0"
  >
    <div class="p-3">
      <h6 class="eq-header">{{ getZoneLongName() }}</h6>

      <eq-tabs>
        <eq-tab
          :selected="true"
          :name="'NPCs' + (npcTypes && npcTypes.length > 0 ? ` (${npcTypes.length})` : '')"
        >

          <!-- Fake Loader -->
          <div v-if="loadingNpcs" class="mt-3 text-center">
            Loading NPCs...
            <loader-fake-progress class="mt-3"/>
          </div>

          <div v-else-if="npcTypes.length === 0" class="mt-3 text-center">
            No NPCs found in this zone.
          </div>

          <!-- NPCs Table -->
          <div style="height: 85vh; overflow-y: scroll;" v-if="npcTypes.length > 0">
            <table
              id="npctable"
              class="eq-table eq-highlight-rows"
              style="display: table; font-size: 14px; "
            >
              <thead
                class="eq-table-floating-header"
              >
              <tr>
                <th class="text-center" style="vertical-align: middle !important">
                  <b-button
                    class="btn-dark btn-sm btn-dark"
                    title="NPC Grid Editor"
                    @click="npcGridEditor()"
                  >
                    <i class="fa fa-th"></i> Bulk Editor
                  </b-button>
                </th>
                <th class="text-center">
                  NPC
                </th>
              </tr>
              </thead>
              <tbody>
              <tr
                :id="'npc-' + n.short_name"
                v-for="(n, index) in npcTypes"
                :key="n.id"
              >
                <td class="text-center" style="width: 150px">
                  <b-button
                    class="btn-dark btn-sm btn-dark"
                    @click="showNpcOnMap(n.npc)"
                    title="Show on Map"
                  >
                    <i class="fa fa-map-marker"></i>
                  </b-button>

                  <b-button
                    class="btn-dark btn-sm btn-dark ml-3"
                    @click="showNpcCard(n.npc)"
                    title="Show NPC card"
                  >
                    <i class="fa fa-eye"></i>
                  </b-button>

                  <b-button
                    class="btn-dark btn-sm btn-dark ml-3"
                    @click="editNpc(n.npc)"
                    title="Edit NPC"
                  >
                    <i class="fa fa-edit"></i>
                  </b-button>
                </td>
                <td style="position: relative">
                  <npc-popover
                    :npc="n.npc"
                  />
                </td>

              </tr>
              </tbody>
            </table>
          </div>


        </eq-tab>
        <eq-tab name="Items">
          <div v-if="itemsLoading" class="mt-3 text-center">
            Loading items...
            <loader-fake-progress class="mt-3"/>
          </div>
          <div v-else-if="zoneItems.length === 0" class="mt-3 text-center">
            No loot or ground items found in this zone.
          </div>
          <div v-else class="zone-catalog-pane">
            <input
              type="text"
              class="form-control form-control-sm mb-2"
              :placeholder="'Filter items (' + zoneItems.length + ')...'"
              v-model="itemSearch"
            >
            <div v-if="filteredItems.length > catalogLimit" class="mb-2">
              Showing {{ catalogLimit }} of {{ filteredItems.length }}. Filter to narrow results.
            </div>
            <div v-if="filteredItems.length === 0" class="text-center mt-3">No matching items.</div>
            <table
              v-else
              class="eq-table eq-highlight-rows zone-catalog-table"
            >
              <thead class="eq-table-floating-header">
              <tr>
                <th class="text-center" style="width: 90px"></th>
                <th>Item</th>
                <th>Source</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="row in displayedRows(filteredItems)" :key="'item-' + row.item.id">
                <td class="text-center">
                  <b-button
                    class="btn-dark btn-sm"
                    title="Edit item"
                    @click="editItem(row.item)"
                  >
                    <i class="fa fa-edit"></i>
                  </b-button>
                </td>
                <td>
                  <item-popover
                    v-if="row.item"
                    :item="row.item"
                    size="sm"
                  />
                </td>
                <td>
                  <span v-if="row.ground">Ground spawn</span>
                  <span
                    v-for="(npc, npcIndex) in visibleNpcs(row.npcs)"
                    :key="'item-npc-' + row.item.id + '-' + npc.id"
                    class="zone-npc-link"
                    @click="showNpcOnMap(npc)"
                  >{{ npcIndex > 0 ? ', ' : '' }}{{ getCleanName(npc.name) }}</span>
                  <span v-if="row.npcs.length > 5"> +{{ row.npcs.length - 5 }} more</span>
                </td>
              </tr>
              </tbody>
            </table>
          </div>
        </eq-tab>
        <eq-tab name="Tasks">
          <div v-if="loadingTasks" class="mt-3 text-center">
            Loading tasks...
            <loader-fake-progress class="mt-3"/>
          </div>
          <div v-else-if="zoneTasks.length === 0" class="mt-3 text-center">
            No tasks reference this zone.
          </div>
          <div v-else class="zone-catalog-pane">
            <input
              type="text"
              class="form-control form-control-sm mb-2"
              :placeholder="'Filter tasks (' + zoneTasks.length + ')...'"
              v-model="taskSearch"
            >
            <div v-if="filteredTasks.length > catalogLimit" class="mb-2">
              Showing {{ catalogLimit }} of {{ filteredTasks.length }}. Filter to narrow results.
            </div>
            <div v-if="filteredTasks.length === 0" class="text-center mt-3">No matching tasks.</div>
            <table
              v-else
              class="eq-table eq-highlight-rows zone-catalog-table"
            >
              <thead class="eq-table-floating-header">
              <tr>
                <th class="text-center" style="width: 90px"></th>
                <th>Task</th>
                <th>Type</th>
                <th>Activities in zone</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="row in displayedRows(filteredTasks)" :key="'task-' + row.task.id">
                <td class="text-center">
                  <b-button
                    class="btn-dark btn-sm"
                    title="Edit task"
                    @click="editTask(row.task)"
                  >
                    <i class="fa fa-edit"></i>
                  </b-button>
                </td>
                <td>
                  <div class="font-weight-bold">{{ row.task.title }}</div>
                  <div>ID {{ row.task.id }} · Lvl {{ row.task.min_level }}-{{ row.task.max_level }}</div>
                </td>
                <td>{{ taskTypeName(row.task.type) }}</td>
                <td>{{ row.activitySummary }}</td>
              </tr>
              </tbody>
            </table>
          </div>
        </eq-tab>
        <eq-tab name="Sold">
          <div v-if="loadingNpcs" class="mt-3 text-center">
            Loading merchants...
            <loader-fake-progress class="mt-3"/>
          </div>
          <div v-else-if="soldItems.length === 0" class="mt-3 text-center">
            No merchant items found in this zone.
          </div>
          <div v-else class="zone-catalog-pane">
            <input
              type="text"
              class="form-control form-control-sm mb-2"
              :placeholder="'Filter sold items (' + soldItems.length + ')...'"
              v-model="soldSearch"
            >
            <div v-if="filteredSold.length > catalogLimit" class="mb-2">
              Showing {{ catalogLimit }} of {{ filteredSold.length }}. Filter to narrow results.
            </div>
            <div v-if="filteredSold.length === 0" class="text-center mt-3">No matching items.</div>
            <table
              v-else
              class="eq-table eq-highlight-rows zone-catalog-table"
            >
              <thead class="eq-table-floating-header">
              <tr>
                <th class="text-center" style="width: 90px"></th>
                <th>Item</th>
                <th>Price</th>
                <th>Sold by</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="row in displayedRows(filteredSold)" :key="'sold-' + row.item.id">
                <td class="text-center">
                  <b-button
                    class="btn-dark btn-sm"
                    title="Edit item"
                    @click="editItem(row.item)"
                  >
                    <i class="fa fa-edit"></i>
                  </b-button>
                </td>
                <td>
                  <item-popover
                    v-if="row.item"
                    :item="row.item"
                    size="sm"
                  />
                </td>
                <td>
                  <eq-cash-display
                    v-if="row.item && typeof row.item.price !== 'undefined'"
                    :price="parseInt(row.item.price)"
                  />
                </td>
                <td>
                  <span
                    v-for="(npc, npcIndex) in visibleNpcs(row.npcs)"
                    :key="'sold-npc-' + row.item.id + '-' + npc.id"
                    class="zone-npc-link"
                    @click="showNpcOnMap(npc)"
                  >{{ npcIndex > 0 ? ', ' : '' }}{{ getCleanName(npc.name) }}</span>
                  <span v-if="row.npcs.length > 5"> +{{ row.npcs.length - 5 }} more</span>
                </td>
              </tr>
              </tbody>
            </table>
          </div>
        </eq-tab>
        <eq-tab name="Spells">
          <div v-if="spellsLoading" class="mt-3 text-center">
            Loading spells...
            <loader-fake-progress class="mt-3"/>
          </div>
          <div v-else-if="zoneSpells.length === 0" class="mt-3 text-center">
            No NPC or teleport spells found for this zone.
          </div>
          <div v-else class="zone-catalog-pane">
            <input
              type="text"
              class="form-control form-control-sm mb-2"
              :placeholder="'Filter spells (' + zoneSpells.length + ')...'"
              v-model="spellSearch"
            >
            <div v-if="filteredSpells.length > catalogLimit" class="mb-2">
              Showing {{ catalogLimit }} of {{ filteredSpells.length }}. Filter to narrow results.
            </div>
            <div v-if="filteredSpells.length === 0" class="text-center mt-3">No matching spells.</div>
            <table
              v-else
              class="eq-table eq-highlight-rows zone-catalog-table"
            >
              <thead class="eq-table-floating-header">
              <tr>
                <th class="text-center" style="width: 90px"></th>
                <th>Spell</th>
                <th>Used by</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="row in displayedRows(filteredSpells)" :key="'spell-' + row.spell.id">
                <td class="text-center">
                  <b-button
                    class="btn-dark btn-sm"
                    title="Edit spell"
                    @click="editSpell(row.spell)"
                  >
                    <i class="fa fa-edit"></i>
                  </b-button>
                </td>
                <td>
                  <spell-popover
                    v-if="row.spell"
                    :spell="row.spell"
                    :size="20"
                    :spell-name-length="40"
                    :annotation="row.isTeleport ? '(Teleport here)' : ''"
                  />
                </td>
                <td>
                  <span v-if="row.isTeleport && row.npcs.length === 0">Teleport / translocate into this zone</span>
                  <span
                    v-for="(npc, npcIndex) in visibleNpcs(row.npcs)"
                    :key="'spell-npc-' + row.spell.id + '-' + npc.id"
                    class="zone-npc-link"
                    @click="showNpcOnMap(npc)"
                  >{{ npcIndex > 0 ? ', ' : '' }}{{ getCleanName(npc.name) }}</span>
                  <span v-if="row.npcs.length > 5"> +{{ row.npcs.length - 5 }} more</span>
                </td>
              </tr>
              </tbody>
            </table>
          </div>
        </eq-tab>
        <eq-tab name="Zone Connections">
          <div v-if="loadingConnections" class="mt-3 text-center">
            Loading zone connections...
            <loader-fake-progress class="mt-3"/>
          </div>
          <div v-else-if="connectionRows.length === 0" class="mt-3 text-center">
            No zone points, doors, or teleport connections found.
          </div>
          <div v-else class="zone-catalog-pane">
            <input
              type="text"
              class="form-control form-control-sm mb-2"
              :placeholder="'Filter connections (' + connectionRows.length + ')...'"
              v-model="connectionSearch"
            >
            <div v-if="filteredConnections.length > catalogLimit" class="mb-2">
              Showing {{ catalogLimit }} of {{ filteredConnections.length }}. Filter to narrow results.
            </div>
            <div v-if="filteredConnections.length === 0" class="text-center mt-3">No matching connections.</div>
            <table
              v-else
              class="eq-table eq-highlight-rows zone-catalog-table"
            >
              <thead class="eq-table-floating-header">
              <tr>
                <th class="text-center" style="width: 90px"></th>
                <th>Type</th>
                <th>Direction</th>
                <th>Zone</th>
                <th>Detail</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="row in displayedRows(filteredConnections)" :key="row.key">
                <td class="text-center">
                  <b-button
                    v-if="row.destShort && row.destShort !== zone.short_name"
                    class="btn-dark btn-sm"
                    title="Open zone"
                    @click="openZone(row.destShort, row.destVersion)"
                  >
                    <i class="fa fa-external-link"></i>
                  </b-button>
                </td>
                <td>{{ row.type }}</td>
                <td>{{ row.direction }}</td>
                <td>{{ row.zoneLabel }}</td>
                <td>{{ row.detail }}</td>
              </tr>
              </tbody>
            </table>
          </div>
        </eq-tab>
        <eq-tab name="Zone">
          <div class="row mr-4">
            <div class="col-6">
              <div
                v-for="f in [
              { field: 'id', description: 'DB ID' },
              { field: 'zoneidnumber', description: 'Game ID' },
              { field: 'short_name', description: 'Short Name', break: true },
              { field: 'long_name', description: 'Long Name' },
              { field: 'version', description: 'Version' },


              { field: 'ztype', description: 'Zone Type', break: true },

              // fog
              // { field: 'fog_minclip', description: 'Fog Min Clip' },
              // { field: 'fog_maxclip', description: 'Fog Max Clip' },
              // { field: 'fog_blue', description: 'Fog Blue' },
              // { field: 'fog_red', description: 'Fog Red' },
              // { field: 'fog_green', description: 'Fog Green' },
              // { field: 'fog_red_1', description: 'fog_red_1' },
              // { field: 'fog_green_1', description: 'fog_green_1' },
              // { field: 'fog_blue_1', description: 'fog_blue_1' },
              // { field: 'fog_minclip_1', description: 'fog_minclip_1' },
              // { field: 'fog_maxclip_1', description: 'fog_maxclip_1' },
              // { field: 'fog_red_2', description: 'fog_red_2' },
              // { field: 'fog_green_2', description: 'fog_green_2' },
              // { field: 'fog_blue_2', description: 'fog_blue_2' },
              // { field: 'fog_minclip_2', description: 'fog_minclip_2' },
              // { field: 'fog_maxclip_2', description: 'fog_maxclip_2' },
              // { field: 'fog_red_3', description: 'fog_red_3' },
              // { field: 'fog_green_3', description: 'fog_green_3' },
              // { field: 'fog_blue_3', description: 'fog_blue_3' },
              // { field: 'fog_minclip_3', description: 'fog_minclip_3' },
              // { field: 'fog_maxclip_3', description: 'fog_maxclip_3' },
              // { field: 'fog_red_4', description: 'fog_red_4' },
              // { field: 'fog_green_4', description: 'fog_green_4' },
              // { field: 'fog_blue_4', description: 'fog_blue_4' },
              // { field: 'fog_minclip_4', description: 'fog_minclip_4' },
              // { field: 'fog_maxclip_4', description: 'fog_maxclip_4' },
              // { field: 'fog_density', description: 'fog_density' },

              // sky
              { field: 'sky', description: 'Sky Type' },
              { field: 'skylock', description: 'Sky Lock' },

              // weather
              // { field: 'rain_chance_1', description: 'rain_chance_1' },
              // { field: 'rain_chance_2', description: 'rain_chance_2' },
              // { field: 'rain_chance_3', description: 'rain_chance_3' },
              // { field: 'rain_chance_4', description: 'rain_chance_4' },
              // { field: 'rain_duration_1', description: 'rain_duration_1' },
              // { field: 'rain_duration_2', description: 'rain_duration_2' },
              // { field: 'rain_duration_3', description: 'rain_duration_3' },
              // { field: 'rain_duration_4', description: 'rain_duration_4' },
              // { field: 'snow_chance_1', description: 'snow_chance_1' },
              // { field: 'snow_chance_2', description: 'snow_chance_2' },
              // { field: 'snow_chance_3', description: 'snow_chance_3' },
              // { field: 'snow_chance_4', description: 'snow_chance_4' },
              // { field: 'snow_duration_1', description: 'snow_duration_1' },
              // { field: 'snow_duration_2', description: 'snow_duration_2' },
              // { field: 'snow_duration_3', description: 'snow_duration_3' },
              // { field: 'snow_duration_4', description: 'snow_duration_4' },

              // { field: 'file_name', description: 'File Name' },
              { field: 'map_file_name', description: 'Map File', break: true },
              { field: 'graveyard_id', description: 'Graveyard ID' },
              { field: 'min_level', description: 'Min. Lvl', break: true },
              { field: 'min_status', description: 'Min. Status' },

              { field: 'timezone', description: 'Timezone', break: true },
              { field: 'time_type', description: 'Time Type' },

              { field: 'note', description: 'Note' },

              { field: 'walkspeed', description: 'Walkspeed' },
              { field: 'flag_needed', description: 'Flag Needed' },

              { field: 'insttype', description: 'Instance Type' },
              { field: 'shutdowndelay', description: 'Shutdown Delay' },
              { field: 'expansion', description: 'Expansion' },


              // content filtering
              { field: 'min_expansion', description: 'Min Expansion', break: true },
              { field: 'max_expansion', description: 'Max Expansion' },
              { field: 'content_flags', description: 'Content Flags Enabled' },
              { field: 'content_flags_disabled', description: 'Content Flags Disabled' },
          ]"
                v-if="typeof zone[f.field] !== 'undefined'"
                :key="f.field"
                :class="'row ' + (f.break ? 'mt-3' : '')"
              >

                <div class="col-6 text-right">
                  <span class="font-weight-bold">{{ f.description }}</span>
                </div>
                <div class="col-6 pl-0">
                  {{ zone[f.field] }}

                </div>
              </div>

            </div>

            <div class="col-6">

              <!-- Zone Settings (Bool) -->
              <div
                class="col-12"
                v-for="f in [
              { field: 'canbind', description: 'Can Bind' },
              { field: 'cancombat', description: 'Can Combat' },
              { field: 'canlevitate', description: 'Can Levitate' },
              { field: 'castoutdoor', description: 'Can Cast Outdoor' },
              { field: 'hotzone', description: 'Is Hotzone' },
              { field: 'peqzone', description: 'Is PEQ Zone Enabled' },
              { field: 'suspendbuffs', description: 'Suspend Buffs' },
          ]"
                :key="f.field"
                v-if="typeof zone[f.field] !== 'undefined'"
              >
                <div class="row">
                  <div class="col-1 pl-0">
                    <eq-checkbox
                      :disabled="true"
                      :value="zone[f.field]"
                    />
                  </div>
                  <div class="col-11">
                <span
                  class="font-weight-bold"
                  style="position: relative; bottom: 2px"
                >{{ f.description }}</span>
                  </div>

                </div>
              </div>

              <!-- Zone Settings -->
              <div class="mt-3">
                <div
                  class="col-12"
                  v-for="f in [
                    // zone level settings
                    { field: 'fast_regen_hp', description: 'Fast Regen HP' },
                    { field: 'fast_regen_mana', description: 'Fast Regen Mana' },
                    { field: 'fast_regen_endurance', description: 'Fast Regen Endurance' },

                    { field: 'npc_max_aggro_dist', description: 'NPC Max Aggro Dist', break: true },
                    { field: 'max_movement_update_range', description: 'Max Move Update Range' },
                    { field: 'underworld_teleport_index', description: 'Underworld Teleport' },
                    { field: 'lava_damage', description: 'Lava Damage', break: true },
                    { field: 'min_lava_damage', description: 'Min. Lava Damage' },
                    { field: 'gravity', description: 'Gravity', break: true },
                    { field: 'type', description: 'Zone Type', break: true },
                    { field: 'zone_exp_multiplier', description: 'Zone EXP Multiplier' },

                    { field: 'maxclients', description: 'Max Clients', break: true },
                    { field: 'ruleset', description: 'Ruleset' },

                    // clipping
                    { field: 'underworld', description: 'Underworld', break: true },
                    { field: 'minclip', description: 'Min Clip' },
                    { field: 'maxclip', description: 'Max Clip' },

                    // safe
                    { field: 'safe_x', description: 'Safe X', break: true },
                    { field: 'safe_y', description: 'Safe Y' },
                    { field: 'safe_z', description: 'Safe Z' },
                    { field: 'safe_heading', description: 'Safe Heading' },
                ]"
                  :key="f.field"
                  v-if="typeof zone[f.field] !== 'undefined'"
                >
                  <div :class="'row ' + (f.break ? 'mt-3' : '')">
                    <div class="col-1 pl-0">
                      {{ zone[f.field] }}
                    </div>

                    <div class="col-11">
                <span
                  class="font-weight-bold"
                >{{ f.description }}</span>
                    </div>
                  </div>
                </div>
              </div>


            </div>
          </div>
        </eq-tab>
      </eq-tabs>

    </div>

  </eq-window>
</template>

<script>

import EqWindow   from "../eq-ui/EQWindow";
import {SpireApi} from "../../app/api/spire-api";
import EqTabs     from "../eq-ui/EQTabs";
import EqTab               from "../eq-ui/EQTab";
import EqCheckbox          from "../eq-ui/EQCheckbox";
import {
  DoorApi,
  GroundSpawnApi,
  ItemApi,
  SpellsNewApi,
  TaskActivityApi,
  TaskApi,
  ZonePointApi
}                          from "../../app/api";
import {SpireQueryBuilder} from "../../app/api/spire-query-builder";
import {Npcs}              from "../../app/npcs";
import NpcPopover          from "../NpcPopover";
import {EventBus}          from "../../app/event-bus/event-bus";
import LoaderFakeProgress  from "../LoaderFakeProgress";
import util                from "util";
import {ROUTE}             from "../../routes";
import {Spawn}             from "../../app/spawn";
import ItemPopover         from "../ItemPopover";
import SpellPopover        from "../SpellPopover";
import EqCashDisplay       from "../eq-ui/EqCashDisplay";
import {Zones}             from "../../app/zones";
import {Tasks}             from "../../app/tasks";
import {TASK_ACTIVITY_TYPES, TASK_TYPES} from "../../app/constants/eq-task-constants";

export default {
  name: "EqZoneCardPreview",
  components: {
    LoaderFakeProgress,
    NpcPopover,
    ItemPopover,
    SpellPopover,
    EqCashDisplay,
    EqCheckbox,
    EqTab,
    EqTabs,
    EqWindow
  },
  props: {
    zone: Object,
    required: true,
  },
  data() {
    return {
      npcTypes: [],
      lootItems: [],
      groundItems: [],
      soldItems: [],
      npcSpells: [],
      teleportSpells: [],
      zoneTasks: [],
      zoneConnections: [],
      itemSearch: "",
      soldSearch: "",
      spellSearch: "",
      taskSearch: "",
      connectionSearch: "",
      loadingNpcs: true,
      loadingGround: true,
      loadingTeleports: true,
      loadingTasks: true,
      loadingConnections: true,
      loadSeq: 0,
      catalogLimit: 200,
    }
  },
  computed: {
    itemsLoading() {
      return this.loadingNpcs || this.loadingGround
    },
    spellsLoading() {
      return this.loadingNpcs || this.loadingTeleports
    },
    zoneItems() {
      const byId = {}
      for (let i = 0; i < this.lootItems.length; i++) {
        const row = this.lootItems[i]
        byId[row.item.id] = {
          item: row.item,
          npcs: row.npcs.slice(),
          ground: false
        }
      }
      for (let i = 0; i < this.groundItems.length; i++) {
        const item = this.groundItems[i]
        if (!item || !item.id) {
          continue
        }
        if (byId[item.id]) {
          byId[item.id].ground = true
        } else {
          byId[item.id] = { item: item, npcs: [], ground: true }
        }
      }
      return Object.keys(byId).map((id) => byId[id]).sort((a, b) => {
        return this.nameSort(a.item, b.item)
      })
    },
    zoneSpells() {
      const byId = {}
      for (let i = 0; i < this.npcSpells.length; i++) {
        const row = this.npcSpells[i]
        byId[row.spell.id] = {
          spell: row.spell,
          npcs: row.npcs.slice(),
          isTeleport: false
        }
      }
      for (let i = 0; i < this.teleportSpells.length; i++) {
        const spell = this.teleportSpells[i]
        if (!spell || !spell.id) {
          continue
        }
        if (byId[spell.id]) {
          byId[spell.id].isTeleport = true
        } else {
          byId[spell.id] = { spell: spell, npcs: [], isTeleport: true }
        }
      }
      return Object.keys(byId).map((id) => byId[id]).sort((a, b) => {
        return this.nameSort(a.spell, b.spell)
      })
    },
    filteredItems() {
      return this.filterBySearch(this.zoneItems, this.itemSearch, (row) => {
        return [row.item.name, row.item.id, this.npcNames(row.npcs), row.ground ? "ground" : ""]
      })
    },
    filteredSold() {
      return this.filterBySearch(this.soldItems, this.soldSearch, (row) => {
        return [row.item.name, row.item.id, this.npcNames(row.npcs)]
      })
    },
    filteredSpells() {
      return this.filterBySearch(this.zoneSpells, this.spellSearch, (row) => {
        return [row.spell.name, row.spell.id, this.npcNames(row.npcs), row.isTeleport ? "teleport translocate" : ""]
      })
    },
    filteredTasks() {
      return this.filterBySearch(this.zoneTasks, this.taskSearch, (row) => {
        return [row.task.title, row.task.id, row.activitySummary, this.taskTypeName(row.task.type)]
      })
    },
    connectionRows() {
      const rows = this.zoneConnections.slice()
      for (let i = 0; i < this.teleportSpells.length; i++) {
        const spell = this.teleportSpells[i]
        if (!spell || !spell.id) {
          continue
        }
        rows.push({
          key: "teleport-" + spell.id,
          type: "Teleport",
          direction: "In",
          destShort: "",
          destVersion: 0,
          zoneLabel: this.zone && this.zone.long_name ? this.zone.long_name : "",
          detail: spell.name + " (#" + spell.id + ")"
        })
      }
      return rows
    },
    filteredConnections() {
      return this.filterBySearch(this.connectionRows, this.connectionSearch, (row) => {
        return [row.type, row.direction, row.zoneLabel, row.detail, row.destShort]
      })
    }
  },
  created() {
    this.backgroundImages  = []
    this.currentImageIndex = 0

    // cycle background images
    this.interval = setInterval(this.setBackgroundImage, 3 * 1000)
  },
  beforeDestroy() {
    if (this.interval) {
      clearInterval(this.interval)
    }
  },
  mounted() {
    this.init()
  },
  watch: {
    zone: {
      handler: function (val, oldVal) {
        this.init()
      },
    },
  },
  methods: {
    npcGridEditor() {
      this.$router.push(
        {
          path: ROUTE.NPCS_EDIT.replaceAll(":zone", this.zone.short_name),
          query: {
            v: this.zone.version
          }
        }
      ).catch(() => {
      })
    },

    editNpc(n) {
      this.$router.push(
        {
          path: ROUTE.NPC_EDIT.replaceAll(":npc", n.id)
        }
      ).catch(() => {
      })
    },

    showNpcCard(n) {
      EventBus.$emit('NPC_SHOW_CARD', n);
    },

    showNpcOnMap(n) {
      console.log(n)
      EventBus.$emit('NPC_ZOOM', n);
    },

    getCleanName(name) {
      if (!name) {
        return ""
      }
      return Npcs.getCleanName(name)
    },

    tabLabel(base, count, loading) {
      if (loading) {
        return base
      }
      return count ? (base + " (" + count + ")") : base
    },

    displayedRows(rows) {
      if (!rows) {
        return []
      }
      return rows.slice(0, this.catalogLimit)
    },

    nameSort(a, b) {
      const an = a && a.name ? a.name : ""
      const bn = b && b.name ? b.name : ""
      return an.localeCompare(bn)
    },

    npcNames(npcs) {
      if (!npcs || npcs.length === 0) {
        return ""
      }
      const names = []
      for (let i = 0; i < npcs.length; i++) {
        names.push(this.getCleanName(npcs[i].name))
      }
      return names.join(" ")
    },

    visibleNpcs(npcs) {
      if (!npcs) {
        return []
      }
      return npcs.slice(0, 5)
    },

    filterBySearch(rows, search, fieldsFn) {
      const q = (search || "").toLowerCase().trim()
      if (!q) {
        return rows
      }
      const matches = []
      for (let i = 0; i < rows.length; i++) {
        const fields = fieldsFn(rows[i])
        let hit = false
        for (let f = 0; f < fields.length; f++) {
          if (String(fields[f] == null ? "" : fields[f]).toLowerCase().indexOf(q) !== -1) {
            hit = true
            break
          }
        }
        if (hit) {
          matches.push(rows[i])
        }
      }
      return matches
    },

    taskTypeName(type) {
      return TASK_TYPES[type] || "Task"
    },

    addUniqueNpc(npcs, npc) {
      if (!npc || !npc.id) {
        return
      }
      for (let i = 0; i < npcs.length; i++) {
        if (npcs[i].id === npc.id) {
          return
        }
      }
      npcs.push(npc)
    },

    editItem(item) {
      if (!item || !item.id) {
        return
      }
      this.$router.push({ path: util.format(ROUTE.ITEM_EDIT, item.id) }).catch(() => {})
    },

    editSpell(spell) {
      if (!spell || !spell.id) {
        return
      }
      this.$router.push({ path: util.format(ROUTE.SPELL_EDIT, spell.id) }).catch(() => {})
    },

    editTask(task) {
      if (!task || !task.id) {
        return
      }
      this.$router.push({ path: util.format(ROUTE.TASK_EDIT, task.id) }).catch(() => {})
    },

    openZone(shortName, version) {
      if (!shortName) {
        return
      }
      this.$router.push({
        path: "/zone/" + shortName,
        query: { v: version != null ? version : 0 }
      }).catch(() => {})
    },

    activityTouchesZone(activity, zoneId) {
      if (!activity || activity.zones == null || activity.zones === "") {
        return false
      }
      const id = String(zoneId)
      const parts = String(activity.zones).split(",")
      for (let i = 0; i < parts.length; i++) {
        if (parts[i].trim() === id) {
          return true
        }
      }
      return false
    },

    resetCatalogs() {
      this.npcTypes = []
      this.lootItems = []
      this.groundItems = []
      this.soldItems = []
      this.npcSpells = []
      this.teleportSpells = []
      this.zoneTasks = []
      this.zoneConnections = []
      this.itemSearch = ""
      this.soldSearch = ""
      this.spellSearch = ""
      this.taskSearch = ""
      this.connectionSearch = ""
      this.loadingNpcs = true
      this.loadingGround = true
      this.loadingTeleports = true
      this.loadingTasks = true
      this.loadingConnections = true
    },

    init() {
      this.loadSeq = this.loadSeq + 1
      const seq = this.loadSeq
      this.resetCatalogs()

      // get zone wallpaper
      this.loadBackgroundImages().then(() => {
        this.setBackgroundImage()
      })

      this.loadNpcTypes(seq)
      this.loadTeleportSpells(seq)
      this.loadGroundItems(seq)
      this.loadZoneTasks(seq)
      this.loadZoneConnections(seq)
    },

    getZoneLongName() {
      return this.zone.long_name
    },

    shuffle(array) {
      let currentIndex = array.length, randomIndex;

      // While there remain elements to shuffle.
      while (currentIndex !== 0) {

        // Pick a remaining element.
        randomIndex = Math.floor(Math.random() * currentIndex);
        currentIndex--;

        // And swap it with the current element.
        [array[currentIndex], array[randomIndex]] = [
          array[randomIndex], array[currentIndex]];
      }

      return array;
    },

    async loadBackgroundImages() {
      console.log("[EQZoneCardPreview] loadBackgroundImages")

      document.body.style.setProperty("--zone-background", "none");
      document.body.style.setProperty("--zone-background-size", "auto");

      // get zone wallpaper
      await SpireApi.v1().get('/assets/zone-images/' + encodeURIComponent(this.zone.long_name)).then((r) => {
        if (r.status === 200) {
          this.backgroundImages = this.shuffle(r.data.images)
        }
      })
    },

    setBackgroundImage() {
      if (this.backgroundImages && this.backgroundImages.length > 0) {
        const image = this.backgroundImages[this.currentImageIndex];
        // console.log("IMAGE ", image)

        // console.log(
        //   "[EQZoneCardPreview] loadBackgroundImages Playing index [%s] out of [%s]",
        //   this.currentImageIndex,
        //   this.backgroundImages.length
        // )

        if (image.length > 0) {
          let img     = new Image();
          img.src     = image;
          img.onload  = () => {
            // document.body.style.setProperty("--image", "url(" + image + ")");
            document.body.style.setProperty("--zone-background", "url(" + image + ")");
            document.body.style.setProperty("--zone-background-size", "cover");

            // increment
            this.currentImageIndex++;

            // reset if rollover
            if (this.currentImageIndex >= this.backgroundImages.length) {
              // console.log("[EQZoneCardPreview] loadBackgroundImages resetting")
              this.currentImageIndex = 0;
            }
          }
          img.onerror = () => {
            // console.log(
            //   "[EQZoneCardPreview] loadBackgroundImages Failed to load index [%s] out of [%s]",
            //   this.currentImageIndex,
            //   this.backgroundImages.length
            // )

            this.currentImageIndex++
            this.setBackgroundImage()
          }

        }
      }
    },

    startsWithUppercase(str) {
      if (!str) {
        return false
      }
      return str.substr(0, 1).match(/[A-Z\u00C0-\u00DC]/);
    },

    async loadNpcTypes(seq) {
      let npcTypes = [];
      try {
        const r = await Spawn.getByZone(this.zone.short_name, this.zone.version, true)
        if (r && r.length > 0) {
          for (let spawn2 of r) {
            if (spawn2.spawnentries) {
              for (let spawnentry of spawn2.spawnentries) {
                if (spawnentry.npc_type) {

                  // make sure we only add unique NPC IDs since spawns can use multiple
                  // of the same NPC ID
                  if (npcTypes.filter(f => f.npc.id === spawnentry.npc_type.id).length === 0) {
                    npcTypes.push(
                      {
                        npc: spawnentry.npc_type,
                        spawn: {
                          x: spawn2.x,
                          y: spawn2.y,
                        }
                      }
                    )
                  }
                }
              }
            }
          }

          // sort alpha, upper case first
          npcTypes = npcTypes.sort((a, b) => {
            if (this.startsWithUppercase(a.npc.name) && !this.startsWithUppercase(b.npc.name)) {
              return -1;
            } else if (this.startsWithUppercase(b.npc.name) && !this.startsWithUppercase(a.npc.name)) {
              return 1;
            }
            return a.npc.name.localeCompare(b.npc.name);
          });
        }
      } catch (err) {
        console.log("[EQZoneCardPreview] loadNpcTypes %s", err)
      }

      if (seq !== this.loadSeq) {
        return
      }

      this.npcTypes = npcTypes
      this.buildNpcCatalogs(npcTypes)
      this.loadingNpcs = false
      this.$forceUpdate()
    },

    buildNpcCatalogs(npcTypes) {
      const lootMap = {}
      const soldMap = {}
      const spellMap = {}

      for (let i = 0; i < npcTypes.length; i++) {
        const npc = npcTypes[i].npc
        if (!npc) {
          continue
        }

        if (npc.loottable && npc.loottable.loottable_entries) {
          for (let le = 0; le < npc.loottable.loottable_entries.length; le++) {
            const entry = npc.loottable.loottable_entries[le]
            if (!entry.lootdrop || !entry.lootdrop.lootdrop_entries) {
              continue
            }
            for (let de = 0; de < entry.lootdrop.lootdrop_entries.length; de++) {
              const drop = entry.lootdrop.lootdrop_entries[de]
              if (!drop.item || !drop.item.id) {
                continue
              }
              if (!lootMap[drop.item.id]) {
                lootMap[drop.item.id] = { item: drop.item, npcs: [] }
              }
              this.addUniqueNpc(lootMap[drop.item.id].npcs, npc)
            }
          }
        }

        if (npc.merchantlists && npc.merchantlists.length > 0) {
          for (let m = 0; m < npc.merchantlists.length; m++) {
            const listitem = npc.merchantlists[m]
            const item = listitem.items && listitem.items.length > 0 ? listitem.items[0] : null
            if (!item || !item.id) {
              continue
            }
            if (!soldMap[item.id]) {
              soldMap[item.id] = { item: item, npcs: [] }
            }
            this.addUniqueNpc(soldMap[item.id].npcs, npc)
          }
        }

        if (npc.npc_spell && npc.npc_spell.npc_spells_entries) {
          for (let s = 0; s < npc.npc_spell.npc_spells_entries.length; s++) {
            const spellEntry = npc.npc_spell.npc_spells_entries[s]
            if (!spellEntry.spells_new || !spellEntry.spells_new.id) {
              continue
            }
            const spell = spellEntry.spells_new
            if (!spellMap[spell.id]) {
              spellMap[spell.id] = { spell: spell, npcs: [] }
            }
            this.addUniqueNpc(spellMap[spell.id].npcs, npc)
          }
        }
      }

      this.lootItems = Object.keys(lootMap).map((id) => lootMap[id]).sort((a, b) => {
        return this.nameSort(a.item, b.item)
      })
      this.soldItems = Object.keys(soldMap).map((id) => soldMap[id]).sort((a, b) => {
        return this.nameSort(a.item, b.item)
      })
      this.npcSpells = Object.keys(spellMap).map((id) => spellMap[id]).sort((a, b) => {
        return this.nameSort(a.spell, b.spell)
      })
    },

    async loadTeleportSpells(seq) {
      try {
        const r = await (new SpellsNewApi(...SpireApi.cfg())).listSpellsNews(
          (new SpireQueryBuilder())
            .where("teleport_zone", "=", this.zone.short_name)
            .limit(5000)
            .get()
        )
        if (seq !== this.loadSeq) {
          return
        }
        if (r.status === 200 && r.data) {
          this.teleportSpells = r.data
        } else {
          this.teleportSpells = []
        }
      } catch (err) {
        console.log("[EQZoneCardPreview] loadTeleportSpells %s", err)
        if (seq === this.loadSeq) {
          this.teleportSpells = []
        }
      }
      if (seq === this.loadSeq) {
        this.loadingTeleports = false
      }
    },

    async loadGroundItems(seq) {
      try {
        const r = await (new GroundSpawnApi(...SpireApi.cfg())).listGroundSpawns(
          (new SpireQueryBuilder())
            .where("zoneid", "=", this.zone.zoneidnumber)
            .limit(5000)
            .get()
        )
        if (seq !== this.loadSeq) {
          return
        }

        const ids = []
        const seen = {}
        const zoneVersion = parseInt(this.zone.version, 10)
        if (r.status === 200 && r.data) {
          for (let i = 0; i < r.data.length; i++) {
            const spawn = r.data[i]
            const spawnVersion = parseInt(spawn.version, 10)
            if (spawnVersion && spawnVersion !== zoneVersion) {
              continue
            }
            if (spawn.item && !seen[spawn.item]) {
              seen[spawn.item] = true
              ids.push(spawn.item)
            }
          }
        }

        if (ids.length === 0) {
          this.groundItems = []
        } else {
          const items = await (new ItemApi(...SpireApi.cfg())).getItemsBulk({ body: { ids: ids } })
          if (seq !== this.loadSeq) {
            return
          }
          this.groundItems = items.status === 200 && items.data ? items.data : []
        }
      } catch (err) {
        console.log("[EQZoneCardPreview] loadGroundItems %s", err)
        if (seq === this.loadSeq) {
          this.groundItems = []
        }
      }
      if (seq === this.loadSeq) {
        this.loadingGround = false
      }
    },

    async loadZoneTasks(seq) {
      try {
        const r = await (new TaskActivityApi(...SpireApi.cfg())).listTaskActivities(
          (new SpireQueryBuilder())
            .where("zones", "like", String(this.zone.zoneidnumber))
            .limit(10000)
            .get()
        )
        if (seq !== this.loadSeq) {
          return
        }

        const byTask = {}
        if (r.status === 200 && r.data) {
          for (let i = 0; i < r.data.length; i++) {
            const activity = r.data[i]
            if (!this.activityTouchesZone(activity, this.zone.zoneidnumber)) {
              continue
            }
            if (!byTask[activity.taskid]) {
              byTask[activity.taskid] = []
            }
            byTask[activity.taskid].push(activity)
          }
        }

        const ids = Object.keys(byTask).map((id) => parseInt(id, 10)).filter((id) => id > 0)
        if (ids.length === 0) {
          this.zoneTasks = []
        } else {
          const chunkSize = 200
          const tasksById = {}
          for (let i = 0; i < ids.length; i += chunkSize) {
            const chunk = ids.slice(i, i + chunkSize)
            const bulk = await (new TaskApi(...SpireApi.cfg())).getTasksBulk({ body: { ids: chunk } })
            if (seq !== this.loadSeq) {
              return
            }
            if (bulk.status === 200 && bulk.data) {
              for (let t = 0; t < bulk.data.length; t++) {
                tasksById[bulk.data[t].id] = bulk.data[t]
              }
            }
          }

          const rows = []
          for (let i = 0; i < ids.length; i++) {
            const task = tasksById[ids[i]] || { id: ids[i], title: "Task #" + ids[i], type: 0, min_level: 0, max_level: 0 }
            const activities = byTask[ids[i]]
            const types = []
            const seenType = {}
            for (let a = 0; a < activities.length; a++) {
              const label = TASK_ACTIVITY_TYPES[activities[a].activitytype] || ("Type " + activities[a].activitytype)
              if (!seenType[label]) {
                seenType[label] = true
                types.push(label)
              }
            }
            rows.push({
              task: task,
              activities: activities,
              activitySummary: types.join(", ") + (activities[0] ? " — " + Tasks.buildActivityDescription(activities[0]) : "")
            })
          }
          rows.sort((a, b) => {
            const at = a.task.title || ""
            const bt = b.task.title || ""
            return at.localeCompare(bt)
          })
          this.zoneTasks = rows
        }
      } catch (err) {
        console.log("[EQZoneCardPreview] loadZoneTasks %s", err)
        if (seq === this.loadSeq) {
          this.zoneTasks = []
        }
      }
      if (seq === this.loadSeq) {
        this.loadingTasks = false
      }
    },

    connectionZoneLabel(shortName, longName) {
      if (longName && shortName) {
        return longName + " (" + shortName + ")"
      }
      return longName || shortName || "Unknown zone"
    },

    async loadZoneConnections(seq) {
      try {
        await Zones.getZones()
        if (seq !== this.loadSeq) {
          return
        }

        const shortName = this.zone.short_name
        const zoneId = this.zone.zoneidnumber
        const version = this.zone.version
        const rows = []

        const zpApi = new ZonePointApi(...SpireApi.cfg())
        const doorApi = new DoorApi(...SpireApi.cfg())

        const outboundZp = zpApi.listZonePoints(
          (new SpireQueryBuilder()).where("zone", "=", shortName).limit(5000).get()
        )
        const inboundZp = zpApi.listZonePoints(
          (new SpireQueryBuilder()).where("target_zone_id", "=", zoneId).limit(5000).get()
        )
        const outboundDoors = doorApi.listDoors(
          (new SpireQueryBuilder()).where("zone", "=", shortName).where("version", "=", version).limit(5000).get()
        )
        const inboundDoors = doorApi.listDoors(
          (new SpireQueryBuilder()).where("dest_zone", "=", shortName).limit(5000).get()
        )

        const zpOut = await outboundZp
        const zpIn = await inboundZp
        const doorOut = await outboundDoors
        const doorIn = await inboundDoors
        if (seq !== this.loadSeq) {
          return
        }

        if (zpOut.status === 200 && zpOut.data) {
          for (let i = 0; i < zpOut.data.length; i++) {
            const point = zpOut.data[i]
            const dest = Zones.getZoneByIdSync(point.target_zone_id) || {}
            const destShort = dest.short_name || ""
            const same = destShort === shortName
            rows.push({
              key: "zp-out-" + point.id,
              type: "Zone Point",
              direction: same ? "Internal" : "Out",
              destShort: destShort,
              destVersion: dest.version || 0,
              zoneLabel: this.connectionZoneLabel(destShort, dest.long_name),
              detail: "ZP #" + point.number + " @ " + Math.round(point.x) + ", " + Math.round(point.y) + ", " + Math.round(point.z)
            })
          }
        }

        if (zpIn.status === 200 && zpIn.data) {
          for (let i = 0; i < zpIn.data.length; i++) {
            const point = zpIn.data[i]
            if (point.zone === shortName) {
              continue
            }
            const src = await Zones.getZoneByShortName(point.zone)
            const srcShort = src && src.short_name ? src.short_name : point.zone
            rows.push({
              key: "zp-in-" + point.id,
              type: "Zone Point",
              direction: "In",
              destShort: srcShort,
              destVersion: src && src.version != null ? src.version : 0,
              zoneLabel: this.connectionZoneLabel(srcShort, src && src.long_name ? src.long_name : ""),
              detail: "From ZP #" + point.number
            })
          }
        }

        if (doorOut.status === 200 && doorOut.data) {
          for (let i = 0; i < doorOut.data.length; i++) {
            const door = doorOut.data[i]
            const destZone = door.dest_zone ? String(door.dest_zone) : ""
            if (!destZone || destZone.toUpperCase() === "NONE") {
              continue
            }
            const destLong = await Zones.getZoneLongNameByShortName(destZone)
            const same = destZone.toLowerCase() === String(shortName).toLowerCase()
            rows.push({
              key: "door-out-" + door.id,
              type: "Door",
              direction: same ? "Internal" : "Out",
              destShort: destZone,
              destVersion: 0,
              zoneLabel: this.connectionZoneLabel(destZone, destLong),
              detail: door.name || ("Door #" + door.doorid)
            })
          }
        }

        if (doorIn.status === 200 && doorIn.data) {
          for (let i = 0; i < doorIn.data.length; i++) {
            const door = doorIn.data[i]
            if (door.zone === shortName) {
              continue
            }
            const src = await Zones.getZoneByShortName(door.zone)
            const srcShort = src && src.short_name ? src.short_name : door.zone
            rows.push({
              key: "door-in-" + door.id,
              type: "Door",
              direction: "In",
              destShort: srcShort,
              destVersion: src && src.version != null ? src.version : 0,
              zoneLabel: this.connectionZoneLabel(srcShort, src && src.long_name ? src.long_name : ""),
              detail: door.name || ("Door #" + door.doorid)
            })
          }
        }

        if (seq !== this.loadSeq) {
          return
        }

        this.zoneConnections = rows
      } catch (err) {
        console.log("[EQZoneCardPreview] loadZoneConnections %s", err)
        if (seq === this.loadSeq) {
          this.zoneConnections = []
        }
      }
      if (seq === this.loadSeq) {
        this.loadingConnections = false
      }
    }
  }
}
</script>

<style>
:root {
  --zone-background-size: auto;
  --zone-background: none;
}

#zone-preview::before {
  content: "";

  background-size: var(--zone-background-size) !important;
  background-repeat: no-repeat !important;
  background-attachment: fixed !important;
  background-position: center !important;

  z-index: -99999;

  top: 0;
  right: 0;
  bottom: 0;
  left: 0;

  background: var(--zone-background);
  opacity: .1;

  --webkit-transition: background-image 1s ease-in-out;
  transition: background-image 1s ease-in-out;
}

#npctable td, #npctable th {
  vertical-align: middle;
  padding: 10px;
  height: 60px;
}

.zone-catalog-pane {
  height: 85vh;
  overflow-y: scroll;
}

.zone-catalog-table {
  display: table;
  font-size: 14px;
  width: 100%;
}

.zone-catalog-table td, .zone-catalog-table th {
  vertical-align: middle;
  padding: 10px;
}

.zone-npc-link {
  cursor: pointer;
  text-decoration: underline;
}
</style>

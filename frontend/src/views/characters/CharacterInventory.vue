<template>
  <div class="char-inv">
    <eq-window :title="windowTitle" class="p-3">
      <div class="char-inv-top">
        <div class="char-inv-toolbar">
          <input
            class="form-control form-control-sm"
            v-model="charSearch"
            @keyup.enter="searchCharacters"
            placeholder="Character name or id"
          >
          <button class="btn btn-sm btn-dark" @click="searchCharacters">Find</button>
          <router-link class="btn btn-sm btn-dark" to="/editors/players">Players</router-link>
          <router-link class="btn btn-sm btn-dark" to="/editors/inventory-raw">Raw rows</router-link>
          <span class="char-inv-status" v-if="status">{{ status }}</span>
        </div>
        <div class="char-inv-hero" v-if="character">
          <div class="char-inv-faces">
            <span class="char-inv-face" :title="className(character)">
              <span v-if="classIcon(character)" :class="'item-' + classIcon(character)"></span>
            </span>
            <span class="char-inv-face" :title="raceName(character)">
              <span v-if="raceIcon(character)" :class="'item-' + raceIcon(character)"></span>
            </span>
          </div>
          <div class="char-inv-hero-text">
            <div class="char-inv-name">{{ displayName(character) }}</div>
            <div class="char-inv-meta">
              L{{ character.level }} {{ className(character) }}
              <template v-if="raceName(character)"> · {{ raceName(character) }}</template>
              <template v-if="genderName(character)"> · {{ genderName(character) }}</template>
              <span v-if="character.gm"> · GM</span>
            </div>
            <div class="char-inv-meta">
              #{{ character.id }}
              <template v-if="character.account_id"> · Acc {{ character.account_id }}</template>
              · {{ filledCount }} items
            </div>
            <div class="char-inv-meta" v-if="zoneLabel(character) || deityName(character)">
              <template v-if="zoneLabel(character)">{{ zoneLabel(character) }}</template>
              <template v-if="zoneLabel(character) && deityName(character)"> · </template>
              <template v-if="deityName(character)">{{ deityName(character) }}</template>
            </div>
          </div>
        </div>
      </div>

      <div class="char-inv-hits" v-if="charHits.length > 1">
        <div
          v-for="hit in charHits"
          :key="hit.id"
          class="char-inv-hit"
          role="button"
          tabindex="0"
          :class="{ 'is-on': character && character.id === hit.id }"
          :title="displayName(hit)"
          @click="selectCharacter(hit.id)"
          @keydown.enter.prevent="selectCharacter(hit.id)"
        >
          <div class="char-inv-faces">
            <span class="char-inv-face" :title="className(hit)">
              <span v-if="classIcon(hit)" :class="'item-' + classIcon(hit)"></span>
            </span>
            <span class="char-inv-face" :title="raceName(hit)">
              <span v-if="raceIcon(hit)" :class="'item-' + raceIcon(hit)"></span>
            </span>
          </div>
          <div class="char-inv-hit-text">
            <b>{{ displayName(hit) }}</b>
            <span>L{{ hit.level }} {{ className(hit) }}<template v-if="raceName(hit)"> · {{ raceName(hit) }}</template><template v-if="genderName(hit)"> · {{ genderName(hit) }}</template></span>
            <span>#{{ hit.id }}<template v-if="hit.account_id"> · Acc {{ hit.account_id }}</template><template v-if="hit.gm"> · GM</template></span>
            <span v-if="zoneLabel(hit) || deityName(hit)">{{ zoneLabel(hit) }}<template v-if="zoneLabel(hit) && deityName(hit)"> · </template>{{ deityName(hit) }}</span>
          </div>
        </div>
      </div>

      <div class="char-inv-body" v-if="character">
        <div class="char-inv-main">
          <div class="char-inv-tabs">
            <button
              v-for="tab in visibleTabs"
              :key="tab.id"
              type="button"
              class="char-inv-tab ui-plain"
              :class="{ 'is-on': activeTab === tab.id }"
              @click="activeTab = tab.id"
            >{{ tab.label }}</button>
          </div>

          <div v-if="loading" class="char-inv-empty">Loading…</div>

          <div v-else-if="activeTab === 'worn'" class="char-inv-worn-wrap">
            <div class="inv-panel inv-paper-panel">
              <div class="inv-panel-head">Equipment</div>
              <div class="inv-paper">
                <template v-for="(slotId, i) in wornPaperdoll">
                  <inventory-slot-cell
                    v-if="slotId != null"
                    :key="'w-' + slotId"
                    :slot-id="slotId"
                    :row="rowAt(slotId)"
                    :selected="selectedSlot === slotId"
                    :label="shortName(slotId)"
                    @pick="selectInvSlot"
                  />
                  <div v-else :key="'sp-' + i" class="inv-slot-spacer"></div>
                </template>
              </div>
            </div>
            <div class="inv-side-col">
              <div class="inv-panel">
                <div class="inv-panel-head">Bags &amp; cursor</div>
                <div class="inv-grid inv-grid-4">
                  <inventory-slot-cell
                    v-for="slotId in generalSlots"
                    :key="'g-' + slotId"
                    :slot-id="slotId"
                    :row="rowAt(slotId)"
                    :selected="selectedSlot === slotId || openBagParent === slotId"
                    :label="shortName(slotId)"
                    @pick="selectInvSlot"
                  />
                </div>
                <div class="inv-cursor-row">
                  <inventory-slot-cell
                    :slot-id="cursorSlot"
                    :row="rowAt(cursorSlot)"
                    :selected="selectedSlot === cursorSlot || openBagParent === cursorSlot"
                    label="Cursor"
                    @pick="selectInvSlot"
                  />
                  <inventory-slot-cell
                    v-for="slotId in extraSlots"
                    :key="'x-' + slotId"
                    :slot-id="slotId"
                    :row="rowAt(slotId)"
                    :selected="selectedSlot === slotId || openBagParent === slotId"
                    :label="shortName(slotId)"
                    @pick="selectInvSlot"
                  />
                </div>
              </div>
              <div class="inv-panel" v-if="openBagSlots.length">
                <div class="inv-panel-head">{{ openBagTitle }}</div>
                <div class="inv-grid" :class="openBagGridClass">
                  <inventory-slot-cell
                    v-for="slotId in openBagSlots"
                    :key="'b-' + slotId"
                    :slot-id="slotId"
                    :row="rowAt(slotId)"
                    :selected="selectedSlot === slotId"
                    :label="bagSlotCaption(slotId)"
                    @pick="selectInvSlot"
                  />
                </div>
              </div>
            </div>
          </div>

          <div v-else-if="activeTab === 'bags'" class="inv-split">
            <div class="inv-panel">
              <div class="inv-panel-head">General bags</div>
              <div class="inv-grid inv-grid-4">
                <inventory-slot-cell
                  v-for="slotId in generalSlots"
                  :key="'bg-' + slotId"
                  :slot-id="slotId"
                  :row="rowAt(slotId)"
                  :selected="selectedSlot === slotId || openBagParent === slotId"
                  :label="shortName(slotId)"
                  @pick="selectInvSlot"
                />
              </div>
              <div class="inv-cursor-row">
                <inventory-slot-cell
                  :slot-id="cursorSlot"
                  :row="rowAt(cursorSlot)"
                  :selected="selectedSlot === cursorSlot || openBagParent === cursorSlot"
                  label="Cursor"
                  @pick="selectInvSlot"
                />
                <inventory-slot-cell
                  v-for="slotId in extraSlots"
                  :key="'xb-' + slotId"
                  :slot-id="slotId"
                  :row="rowAt(slotId)"
                  :selected="selectedSlot === slotId || openBagParent === slotId"
                  :label="shortName(slotId)"
                  @pick="selectInvSlot"
                />
              </div>
            </div>
            <div class="inv-panel" v-if="openBagSlots.length">
              <div class="inv-panel-head">{{ openBagTitle }}</div>
              <div class="inv-grid" :class="openBagGridClass">
                <inventory-slot-cell
                  v-for="slotId in openBagSlots"
                  :key="'bb-' + slotId"
                  :slot-id="slotId"
                  :row="rowAt(slotId)"
                  :selected="selectedSlot === slotId"
                  :label="bagSlotCaption(slotId)"
                  @pick="selectInvSlot"
                />
              </div>
            </div>
            <div class="char-inv-empty" v-else>Select a bag to open it.</div>
          </div>

          <div v-else-if="activeTab === 'bank'" class="inv-split">
            <div class="inv-panel">
              <div class="inv-panel-head">Bank</div>
              <div class="inv-grid inv-grid-6">
                <inventory-slot-cell
                  v-for="slotId in bankSlots"
                  :key="'k-' + slotId"
                  :slot-id="slotId"
                  :row="rowAt(slotId)"
                  :selected="selectedSlot === slotId || openBagParent === slotId"
                  :label="shortName(slotId)"
                  @pick="selectInvSlot"
                />
              </div>
            </div>
            <div class="inv-panel" v-if="openBagSlots.length">
              <div class="inv-panel-head">{{ openBagTitle }}</div>
              <div class="inv-grid" :class="openBagGridClass">
                <inventory-slot-cell
                  v-for="slotId in openBagSlots"
                  :key="'kb-' + slotId"
                  :slot-id="slotId"
                  :row="rowAt(slotId)"
                  :selected="selectedSlot === slotId"
                  :label="bagSlotCaption(slotId)"
                  @pick="selectInvSlot"
                />
              </div>
            </div>
          </div>

          <div v-else-if="activeTab === 'shared'" class="inv-split">
            <div class="inv-panel">
              <div class="inv-panel-head">Shared bank</div>
              <div class="inv-warn" v-if="sharedError">{{ sharedError }}</div>
              <div class="inv-grid inv-grid-4">
                <inventory-slot-cell
                  v-for="slotId in sharedSlots"
                  :key="'s-' + slotId"
                  :slot-id="slotId"
                  :row="rowAt(slotId)"
                  :selected="selectedSlot === slotId || openBagParent === slotId"
                  :label="shortName(slotId)"
                  @pick="selectInvSlot"
                />
              </div>
            </div>
            <div class="inv-panel" v-if="openBagSlots.length">
              <div class="inv-panel-head">{{ openBagTitle }}</div>
              <div class="inv-grid" :class="openBagGridClass">
                <inventory-slot-cell
                  v-for="slotId in openBagSlots"
                  :key="'sb-' + slotId"
                  :slot-id="slotId"
                  :row="rowAt(slotId)"
                  :selected="selectedSlot === slotId"
                  :label="bagSlotCaption(slotId)"
                  @pick="selectInvSlot"
                />
              </div>
            </div>
          </div>

          <div v-else-if="activeTab === 'parcels'">
            <div v-if="!parcels.length" class="char-inv-empty">No parcels. Search an item on the right to add one.</div>
            <div class="inv-parcel-list">
              <div
                v-for="parcel in parcels"
                :key="parcel.id"
                class="inv-parcel"
                :class="{ 'is-on': selectedParcel && selectedParcel.id === parcel.id }"
                role="button"
                tabindex="0"
                @click="selectParcel(parcel)"
                @keydown.enter.prevent="selectParcel(parcel)"
              >
                <span class="inv-pick-icon">
                  <span v-if="parcelIcon(parcel)" :class="'item-' + parcelIcon(parcel) + '-sm'"></span>
                </span>
                <span class="inv-pick-text">
                  <b>{{ parcelName(parcel) }}</b>
                  <span class="inv-parcel-meta">×{{ parcel.quantity || 1 }} · {{ parcel.from_name || "—" }}</span>
                </span>
              </div>
            </div>
          </div>

          <div v-else-if="activeTab === 'other'">
            <div v-if="!otherRows.length" class="char-inv-empty">No extra slots.</div>
            <div class="inv-grid inv-grid-6" v-else>
              <inventory-slot-cell
                v-for="row in otherRows"
                :key="'o-' + row.slot_id"
                :slot-id="row.slot_id"
                :row="row"
                :selected="selectedSlot === row.slot_id"
                :label="shortName(row.slot_id)"
                @pick="selectInvSlot"
              />
            </div>
          </div>
        </div>

        <aside class="char-inv-side">
          <div class="inv-panel">
            <div class="inv-panel-head">{{ inspectorTitle }}</div>

            <div v-if="activeTab === 'parcels' && selectedParcel">
              <item-popover v-if="selectedParcelItem" :item="selectedParcelItem" size="sm"/>
              <div v-else class="mb-2">{{ parcelName(selectedParcel) }} (#{{ selectedParcel.item_id }})</div>
              <label class="inv-field">Quantity</label>
              <input class="form-control form-control-sm" type="number" v-model.number="parcelDraft.quantity">
              <label class="inv-field">From</label>
              <input class="form-control form-control-sm" v-model="parcelDraft.from_name">
              <label class="inv-field">Note</label>
              <input class="form-control form-control-sm" v-model="parcelDraft.note">
              <div class="inv-actions">
                <button class="btn btn-sm btn-dark" :disabled="saving" @click="saveParcel">Save</button>
                <button class="btn btn-sm btn-outline-danger" :disabled="saving" @click="clearParcel">Delete</button>
              </div>
            </div>

            <div v-else-if="selectedSlot != null">
              <div v-if="selectedItem" class="inv-selected">
                <item-popover :item="selectedItem" size="sm"/>
                <router-link class="inv-item-link" :to="'/item/' + selectedItem.id">#{{ selectedItem.id }}</router-link>
                <div class="inv-warn" v-if="slotMismatch">Does not fit {{ shortName(selectedSlot) }}.</div>
                <label class="inv-field" v-if="selectedSlot <= 22">
                  <input type="checkbox" v-model="forceEquip"> Force equip anyway
                </label>
              </div>
              <div v-else class="inv-empty-slot">
                <div>Empty {{ slotName(selectedSlot) }}.</div>
                <div>Search an item below and click it to place it here.</div>
              </div>

              <template v-if="selectedRow">
                <label class="inv-field">Charges <span v-if="invDraft.charges >= 10000" class="inv-unlimited">unlimited</span></label>
                <input class="form-control form-control-sm" type="number" v-model.number="invDraft.charges">
                <div class="inv-augs">
                  <div v-for="n in 6" :key="'aug-' + n">
                    <label class="inv-field">Aug {{ n }}</label>
                    <input class="form-control form-control-sm" type="number" v-model.number="invDraft['augment_' + augKey(n)]">
                  </div>
                </div>
                <div class="inv-actions">
                  <button class="btn btn-sm btn-dark" :disabled="saving" @click="saveInvDraft">Save</button>
                  <button class="btn btn-sm btn-outline-danger" :disabled="saving" @click="clearInvSlot">Remove</button>
                </div>
              </template>
            </div>

            <div v-else class="char-inv-empty">
              {{ activeTab === 'parcels' ? 'Select a parcel or search an item to add one.' : 'Select a slot to edit it.' }}
            </div>
          </div>

          <div class="inv-panel">
            <div class="inv-panel-head">Find item</div>
            <div class="char-inv-toolbar">
              <input
                ref="itemSearch"
                class="form-control form-control-sm"
                v-model="itemSearch"
                @keyup.enter="searchItems"
                placeholder="Item name or id"
              >
              <button class="btn btn-sm btn-dark" @click="searchItems">Search</button>
            </div>
            <div class="char-inv-empty" v-if="activeTab === 'parcels'">Choosing an item adds a parcel.</div>
            <div class="char-inv-empty" v-else-if="selectedSlot == null">Click an empty or filled slot first, then pick an item to place or replace it.</div>
            <div class="inv-picker" v-if="itemHits.length">
              <div
                v-for="item in itemHits"
                :key="item.id"
                class="inv-pick"
                role="button"
                tabindex="0"
                :title="item.name + ' #' + item.id"
                @click="pickItem(item)"
                @keydown.enter.prevent="pickItem(item)"
              >
                <span class="inv-pick-icon">
                  <span v-if="item.icon" :class="'item-' + item.icon + '-sm'"></span>
                </span>
                <span class="inv-pick-text">
                  <span class="inv-pick-name">{{ item.name }}</span>
                  <span class="inv-pick-id">#{{ item.id }}</span>
                </span>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </eq-window>
  </div>
</template>

<script>
import EqWindow from "../../components/eq-ui/EQWindow"
import ItemPopover from "../../components/ItemPopover"
import InventorySlotCell from "../../components/inventory/InventorySlotCell.vue"
import axios from "axios"
import {SpireApi} from "../../app/api/spire-api"
import {SpireQueryBuilder} from "../../app/api/spire-query-builder"
import {CharacterDatumApi} from "../../app/api/api/character-datum-api"
import {InventoryApi} from "../../app/api/api/inventory-api"
import {ItemApi} from "../../app/api/api/item-api"
import {CharacterParcelApi} from "../../app/api/api/character-parcel-api"
import {DB_PLAYER_CLASSES_ALL} from "../../app/constants/eq-classes-constants"
import {DB_CLASSES_ICONS} from "../../app/constants/eq-class-icon-constants"
import {DB_PLAYER_RACES} from "../../app/constants/eq-races-constants"
import {DB_RACES_ICONS} from "../../app/constants/eq-race-icon-constants"
import {DB_DIETIES_FULL} from "../../app/constants/eq-deities-constants"
import {GENDER} from "../../app/constants/eq-gender-constants"
import {Zones} from "../../app/zones"

const CHAR_IDENTITY_FIELDS = [
  "id", "account_id", "name", "last_name", "title", "suffix",
  "level", "class", "race", "gender", "deity", "zone_id", "zone_instance", "gm",
]
import {
  WORN_PAPERDOLL,
  WORN_SLOTS,
  GENERAL_SLOTS,
  EXTRA_SLOTS,
  BANK_SLOTS,
  SHARED_SLOTS,
  CURSOR_SLOT,
  bagChildSlots,
  bagBegin,
  bagGridClass,
  slotLabel,
  shortSlotLabel,
  isKnownSlot,
  isBank,
  isShared,
  isInventoryParent,
  isSharedContext,
  itemFitsWorn,
} from "../../app/eq-inventory-slots"

export default {
  name: "CharacterInventory",
  components: {EqWindow, ItemPopover, InventorySlotCell},
  data() {
    return {
      charSearch: "",
      charHits: [],
      character: null,
      rows: [],
      sharedRows: [],
      parcels: [],
      parcelItems: {},
      loading: false,
      saving: false,
      status: "",
      activeTab: "worn",
      selectedSlot: null,
      selectedParcel: null,
      openBagParent: null,
      invDraft: {},
      parcelDraft: {},
      itemSearch: "",
      itemHits: [],
      forceEquip: false,
      sharedError: "",
      wornPaperdoll: WORN_PAPERDOLL,
      generalSlots: GENERAL_SLOTS,
      zoneNames: {},
      extraSlots: EXTRA_SLOTS,
      bankSlots: BANK_SLOTS,
      sharedSlots: SHARED_SLOTS,
      cursorSlot: CURSOR_SLOT,
      tabs: [
        {id: "worn", label: "Worn"},
        {id: "bags", label: "Bags"},
        {id: "bank", label: "Bank"},
        {id: "shared", label: "Shared"},
        {id: "parcels", label: "Parcels"},
        {id: "other", label: "Other"},
      ],
    }
  },
  computed: {
    bySlot() {
      const map = {}
      this.rows.forEach((row) => {
        map[row.slot_id] = row
      })
      return map
    },
    sharedBySlot() {
      const map = {}
      this.sharedRows.forEach((row) => {
        map[row.slot_id] = row
      })
      return map
    },
    filledCount() {
      return this.rows.length
    },
    selectedRow() {
      return this.selectedSlot == null ? null : this.rowAt(this.selectedSlot)
    },
    selectedItem() {
      return this.selectedRow && this.selectedRow.item ? this.selectedRow.item : null
    },
    selectedParcelItem() {
      if (!this.selectedParcel) {
        return null
      }
      return this.parcelItems[this.selectedParcel.item_id] || null
    },
    slotMismatch() {
      return this.selectedItem && this.selectedSlot != null && !itemFitsWorn(this.selectedItem, this.selectedSlot)
    },
    windowTitle() {
      return this.character ? (this.character.name + " — Inventory") : "Character Inventory"
    },
    inspectorTitle() {
      if (this.activeTab === "parcels") {
        return this.selectedParcel ? "Parcel" : "Parcels"
      }
      return this.selectedSlot == null ? "Slot" : slotLabel(this.selectedSlot)
    },
    openBagTitle() {
      if (this.openBagParent == null) {
        return ""
      }
      const row = this.rowAt(this.openBagParent)
      const name = row && row.item && row.item.name ? row.item.name : shortSlotLabel(this.openBagParent)
      return name
    },
    openBagSlots() {
      if (this.openBagParent == null) {
        return []
      }
      const row = this.rowAt(this.openBagParent)
      const bagslots = row && row.item && row.item.bagslots ? row.item.bagslots : 10
      return bagChildSlots(this.openBagParent, bagslots)
    },
    openBagGridClass() {
      return bagGridClass(this.openBagSlots.length)
    },
    otherRows() {
      return this.rows.filter((row) => !isKnownSlot(row.slot_id))
    },
    visibleTabs() {
      return this.tabs.filter((tab) => tab.id !== "other" || this.otherRows.length)
    },
  },
  watch: {
    "$route.query.c"() {
      this.bootFromRoute()
    },
    activeTab(tab) {
      if (tab === "parcels") {
        this.selectedSlot = null
      } else {
        this.selectedParcel = null
      }
      if (tab === "worn" || tab === "bags") {
        if (this.openBagParent != null && !isInventoryParent(this.openBagParent)) {
          this.openBagParent = null
        }
      } else if (tab === "bank") {
        if (this.openBagParent != null && !isBank(this.openBagParent)) {
          this.openBagParent = null
        }
      } else if (tab === "shared") {
        if (this.openBagParent != null && !isShared(this.openBagParent)) {
          this.openBagParent = null
        }
      } else {
        this.openBagParent = null
      }
    },
  },
  mounted() {
    this.preloadZones()
    this.bootFromRoute()
  },
  methods: {
    charApi() {
      return new CharacterDatumApi(...SpireApi.cfg())
    },
    invApi() {
      return new InventoryApi(...SpireApi.cfg())
    },
    itemApi() {
      return new ItemApi(...SpireApi.cfg())
    },
    parcelApi() {
      return new CharacterParcelApi(...SpireApi.cfg())
    },
    classId(row) {
      if (!row) {
        return 0
      }
      return row.class != null ? row.class : (row._class || 0)
    },
    className(row) {
      const rec = DB_PLAYER_CLASSES_ALL[this.classId(row)]
      return rec ? rec.class : ""
    },
    classIcon(row) {
      const id = this.classId(row)
      const rec = DB_PLAYER_CLASSES_ALL[id]
      return (rec && rec.icon) || DB_CLASSES_ICONS[id] || 0
    },
    raceName(row) {
      const rec = row && DB_PLAYER_RACES[row.race]
      return rec ? rec.race : ""
    },
    raceIcon(row) {
      if (!row || row.race == null) {
        return 0
      }
      const rec = DB_PLAYER_RACES[row.race]
      return (rec && rec.icon) || DB_RACES_ICONS[row.race] || 0
    },
    genderName(row) {
      if (!row || row.gender == null) {
        return ""
      }
      return GENDER[row.gender] || ""
    },
    deityName(row) {
      if (!row || !row.deity) {
        return ""
      }
      const rec = DB_DIETIES_FULL[row.deity]
      return rec ? rec.name : ""
    },
    displayName(row) {
      if (!row) {
        return ""
      }
      const parts = []
      if (row.title) {
        parts.push(row.title)
      }
      if (row.name) {
        parts.push(row.name)
      }
      if (row.last_name) {
        parts.push(row.last_name)
      }
      let name = parts.join(" ")
      if (row.suffix) {
        name += (name ? " " : "") + row.suffix
      }
      return name
    },
    zoneLabel(row) {
      if (!row || !row.zone_id) {
        return ""
      }
      const name = this.zoneNames[row.zone_id] || ("Zone " + row.zone_id)
      if (row.zone_instance) {
        return name + " (" + row.zone_instance + ")"
      }
      return name
    },
    async preloadZones() {
      try {
        await Zones.getZones()
        this.refreshZoneNames()
      } catch (err) {
        // zone names stay as Zone <id>
      }
    },
    refreshZoneNames() {
      const ids = {}
      this.charHits.forEach((hit) => {
        if (hit && hit.zone_id) {
          ids[hit.zone_id] = true
        }
      })
      if (this.character && this.character.zone_id) {
        ids[this.character.zone_id] = true
      }
      const names = Object.assign({}, this.zoneNames)
      Object.keys(ids).forEach((key) => {
        const id = parseInt(key, 10)
        const zone = Zones.getZoneByIdSync(id)
        names[id] = (zone && (zone.long_name || zone.short_name)) || ("Zone " + id)
      })
      this.zoneNames = names
    },
    slotName(slotId) {
      return slotLabel(slotId)
    },
    shortName(slotId) {
      return shortSlotLabel(slotId)
    },
    bagSlotCaption(slotId) {
      const begin = this.openBagParent != null ? bagBegin(this.openBagParent) : null
      if (begin == null) {
        return shortSlotLabel(slotId)
      }
      return String(slotId - begin + 1)
    },
    rowAt(slotId) {
      if (this.activeTab === "shared" || isSharedContext(slotId)) {
        return this.sharedBySlot[slotId] || null
      }
      return this.bySlot[slotId] || null
    },
    augKey(n) {
      const names = ["one", "two", "three", "four", "five", "six"]
      return names[n - 1]
    },
    parcelIcon(parcel) {
      const item = this.parcelItems[parcel.item_id]
      return item && item.icon ? item.icon : 0
    },
    parcelName(parcel) {
      const item = this.parcelItems[parcel.item_id]
      return item && item.name ? item.name : ("Item " + parcel.item_id)
    },
    errorText(err) {
      if (err && err.response && err.response.data && err.response.data.error) {
        return err.response.data.error
      }
      return err && err.message ? err.message : "Request failed"
    },
    pushQuery(id) {
      const query = {}
      if (id) {
        query.c = String(id)
      }
      this.$router.replace({path: this.$route.path, query}).catch(() => {})
    },
    bootFromRoute() {
      const raw = this.$route.query.c
      const id = raw ? parseInt(raw, 10) : 0
      if (id) {
        this.selectCharacter(id)
      }
    },
    async searchCharacters() {
      this.status = ""
      const q = (this.charSearch || "").trim()
      if (!q) {
        this.charHits = []
        return
      }
      try {
        const builder = new SpireQueryBuilder()
        builder.limit(20)
        builder.orderBy(["name"])
        builder.select(CHAR_IDENTITY_FIELDS)
        if (/^\d+$/.test(q)) {
          builder.where("id", "=", q)
        } else {
          builder.where("name", "like", q)
        }
        const r = await this.charApi().listCharacterData(builder.get())
        this.charHits = (r && r.data) ? r.data : []
        this.refreshZoneNames()
        if (this.charHits.length === 1) {
          this.selectCharacter(this.charHits[0].id)
        }
      } catch (err) {
        this.status = this.errorText(err)
      }
    },
    async selectCharacter(id) {
      this.status = ""
      this.loading = true
      this.selectedSlot = null
      this.selectedParcel = null
      this.openBagParent = null
      this.pushQuery(id)
      try {
        const cr = await this.charApi().getCharacterDatum({
          id,
          select: CHAR_IDENTITY_FIELDS.join("."),
        })
        this.character = cr && cr.data ? cr.data : null
        if (this.character && !this.charHits.find((h) => h.id === this.character.id)) {
          this.charHits = [this.character].concat(this.charHits)
        }
        this.refreshZoneNames()
        await Promise.all([this.loadInventory(), this.loadParcels(), this.loadShared()])
        if (this.selectedSlot == null) {
          const firstWorn = WORN_SLOTS.find((id) => this.bySlot[id])
          if (firstWorn != null) {
            this.selectInvSlot(firstWorn)
          }
        }
      } catch (err) {
        this.status = this.errorText(err)
        this.character = null
      } finally {
        this.loading = false
      }
    },
    async loadShared() {
      if (!this.character || !this.character.account_id) {
        this.sharedRows = []
        this.sharedError = ""
        return
      }
      this.sharedError = ""
      try {
        const client = axios.create(SpireApi.getAxiosConfig())
        const r = await client.get("/sharedbanks", {
          params: {
            where: "account_id__" + this.character.account_id,
            limit: 500,
            orderBy: "slot_id",
          },
        })
        const rows = (r && r.data) ? r.data : []
        const ids = Array.from(new Set(rows.map((row) => row.item_id).filter(Boolean)))
        let items = {}
        if (ids.length) {
          const ir = await this.itemApi().getItemsBulk({body: {ids}})
          ;((ir && ir.data) || []).forEach((item) => {
            items[item.id] = item
          })
        }
        this.sharedRows = rows.map((row) => Object.assign({}, row, {item: items[row.item_id] || null}))
      } catch (err) {
        this.sharedRows = []
        this.sharedError = this.errorText(err)
      }
    },
    async loadInventory() {
      if (!this.character) {
        this.rows = []
        return
      }
      const builder = new SpireQueryBuilder()
      builder.where("character_id", "=", this.character.id)
      builder.includes(["Item"])
      builder.limit(2000)
      builder.orderBy(["slot_id"])
      const r = await this.invApi().listInventories(builder.get())
      this.rows = (r && r.data) ? r.data : []
      if (this.activeTab === "other" && !this.otherRows.length) {
        this.activeTab = "worn"
      }
    },
    async loadParcels() {
      if (!this.character) {
        this.parcels = []
        return
      }
      const builder = new SpireQueryBuilder()
      builder.where("char_id", "=", this.character.id)
      builder.limit(500)
      builder.orderBy(["id"])
      const r = await this.parcelApi().listCharacterParcels(builder.get())
      this.parcels = (r && r.data) ? r.data : []
      const ids = Array.from(new Set(this.parcels.map((p) => p.item_id).filter(Boolean)))
      if (!ids.length) {
        this.parcelItems = {}
        return
      }
      try {
        const ir = await this.itemApi().getItemsBulk({body: {ids}})
        const map = {}
        ;((ir && ir.data) || []).forEach((item) => {
          map[item.id] = item
        })
        this.parcelItems = map
      } catch (err) {
        this.parcelItems = {}
      }
    },
    selectInvSlot(slotId) {
      this.selectedSlot = slotId
      this.selectedParcel = null
      const row = this.rowAt(slotId)
      this.invDraft = row ? this.draftFromRow(row) : {}
      if (row && row.item && row.item.bagslots > 0 && bagBegin(slotId) != null) {
        this.openBagParent = slotId
      }
      this.$nextTick(() => {
        if (!row && this.$refs.itemSearch) {
          this.$refs.itemSearch.focus()
        }
      })
    },
    selectParcel(parcel) {
      this.selectedParcel = parcel
      this.selectedSlot = null
      this.parcelDraft = {
        quantity: parcel.quantity || 1,
        from_name: parcel.from_name || "",
        note: parcel.note || "",
      }
    },
    draftFromRow(row) {
      return {
        charges: row.charges != null ? row.charges : 1,
        augment_one: row.augment_one || 0,
        augment_two: row.augment_two || 0,
        augment_three: row.augment_three || 0,
        augment_four: row.augment_four || 0,
        augment_five: row.augment_five || 0,
        augment_six: row.augment_six || 0,
        instnodrop: row.instnodrop || 0,
        ornament_icon: row.ornament_icon || 0,
      }
    },
    sanitizeRow(row) {
      return {
        character_id: row.character_id,
        slot_id: row.slot_id,
        item_id: row.item_id,
        charges: row.charges != null ? row.charges : 1,
        color: row.color || 0,
        augment_one: row.augment_one || 0,
        augment_two: row.augment_two || 0,
        augment_three: row.augment_three || 0,
        augment_four: row.augment_four || 0,
        augment_five: row.augment_five || 0,
        augment_six: row.augment_six || 0,
        instnodrop: row.instnodrop || 0,
        custom_data: row.custom_data || "",
        ornament_icon: row.ornament_icon || 0,
        ornament_idfile: row.ornament_idfile || 0,
        ornament_hero_model: row.ornament_hero_model || 0,
      }
    },
    blankRow(slotId, item) {
      return this.sanitizeRow({
        character_id: this.character.id,
        slot_id: slotId,
        item_id: item.id,
        charges: item.stackable ? (item.stacksize || 1) : 1,
      })
    },
    sanitizeShared(row) {
      return {
        account_id: row.account_id,
        slot_id: row.slot_id,
        item_id: row.item_id,
        charges: row.charges != null ? row.charges : 1,
        color: row.color || 0,
        augment_one: row.augment_one || 0,
        augment_two: row.augment_two || 0,
        augment_three: row.augment_three || 0,
        augment_four: row.augment_four || 0,
        augment_five: row.augment_five || 0,
        augment_six: row.augment_six || 0,
        custom_data: row.custom_data || "",
        ornament_icon: row.ornament_icon || 0,
        ornament_idfile: row.ornament_idfile || 0,
        ornament_hero_model: row.ornament_hero_model || 0,
      }
    },
    blankShared(slotId, item) {
      return this.sanitizeShared({
        account_id: this.character.account_id,
        slot_id: slotId,
        item_id: item.id,
        charges: item.stackable ? (item.stacksize || 1) : 1,
      })
    },
    usesShared(slotId) {
      return this.activeTab === "shared" || isSharedContext(slotId)
    },
    sharedHttp() {
      return axios.create(SpireApi.getAxiosConfig())
    },
    slotOpts(slotId) {
      const query = {slot_id: slotId}
      if (this.forceEquip) {
        query.force = "1"
      }
      return {query}
    },
    createOpts() {
      return this.forceEquip ? {query: {force: "1"}} : undefined
    },
    async searchItems() {
      this.status = ""
      const q = (this.itemSearch || "").trim()
      if (!q) {
        this.itemHits = []
        return
      }
      try {
        const builder = new SpireQueryBuilder()
        builder.limit(40)
        builder.orderBy(["id"])
        if (/^\d+$/.test(q)) {
          builder.where("id", "=", q)
        } else {
          builder.where("name", "like", q)
        }
        const r = await this.itemApi().listItems(builder.get())
        this.itemHits = (r && r.data) ? r.data : []
      } catch (err) {
        this.status = this.errorText(err)
      }
    },
    async pickItem(item) {
      if (this.activeTab === "parcels") {
        await this.addParcel(item)
        return
      }
      if (this.selectedSlot == null || !this.character) {
        this.status = "Select a slot first."
        return
      }
      await this.assignItem(item)
    },
    async assignItem(item) {
      const slotId = this.selectedSlot
      const existing = this.rowAt(slotId)
      if (!this.forceEquip && !itemFitsWorn(item, slotId)) {
        this.status = "Does not fit that worn slot. Enable Force to override."
        return
      }
      this.saving = true
      this.status = ""
      try {
        if (this.usesShared(slotId)) {
          const client = this.sharedHttp()
          if (existing) {
            const body = this.sanitizeShared(Object.assign({}, existing, {item_id: item.id}))
            await client.patch("/sharedbank/" + this.character.account_id, body, {params: {slot_id: slotId}})
          } else {
            await client.put("/sharedbank", this.blankShared(slotId, item))
          }
          await this.loadShared()
        } else if (existing) {
          const body = this.sanitizeRow(Object.assign({}, existing, {item_id: item.id}))
          await this.invApi().updateInventory({id: this.character.id, inventory: body}, this.slotOpts(slotId))
          await this.loadInventory()
        } else {
          await this.invApi().createInventory({inventory: this.blankRow(slotId, item)}, this.createOpts())
          await this.loadInventory()
        }
        this.selectInvSlot(slotId)
        this.status = "Placed " + item.name
      } catch (err) {
        this.status = this.errorText(err)
      } finally {
        this.saving = false
      }
    },
    async saveInvDraft() {
      if (!this.selectedRow || this.selectedSlot == null) {
        return
      }
      this.saving = true
      this.status = ""
      try {
        if (this.usesShared(this.selectedSlot)) {
          const body = this.sanitizeShared(Object.assign({}, this.selectedRow, this.invDraft))
          await this.sharedHttp().patch("/sharedbank/" + this.character.account_id, body, {params: {slot_id: this.selectedSlot}})
          await this.loadShared()
        } else {
          const body = this.sanitizeRow(Object.assign({}, this.selectedRow, this.invDraft))
          await this.invApi().updateInventory({id: this.character.id, inventory: body}, this.slotOpts(this.selectedSlot))
          await this.loadInventory()
        }
        this.selectInvSlot(this.selectedSlot)
        this.status = "Saved"
      } catch (err) {
        this.status = this.errorText(err)
      } finally {
        this.saving = false
      }
    },
    async clearInvSlot() {
      if (this.selectedSlot == null || !this.rowAt(this.selectedSlot)) {
        return
      }
      if (!confirm("Remove item from " + slotLabel(this.selectedSlot) + "?")) {
        return
      }
      this.saving = true
      this.status = ""
      try {
        if (this.usesShared(this.selectedSlot)) {
          await this.sharedHttp().delete("/sharedbank/" + this.character.account_id, {params: {slot_id: this.selectedSlot}})
          await this.loadShared()
        } else {
          await this.invApi().deleteInventory({id: this.character.id}, this.slotOpts(this.selectedSlot))
          await this.loadInventory()
        }
        const keep = this.selectedSlot
        this.selectInvSlot(keep)
        this.status = "Removed"
      } catch (err) {
        this.status = this.errorText(err)
      } finally {
        this.saving = false
      }
    },
    async addParcel(item) {
      if (!this.character) {
        return
      }
      this.saving = true
      this.status = ""
      try {
        const nextSlot = this.parcels.reduce((max, p) => Math.max(max, Number(p.slot_id) || 0), -1) + 1
        await this.parcelApi().createCharacterParcel({
          characterParcel: {
            char_id: this.character.id,
            item_id: item.id,
            quantity: 1,
            slot_id: nextSlot,
            from_name: "Spire",
            note: "",
          },
        })
        await this.loadParcels()
        this.status = "Added parcel " + item.name
      } catch (err) {
        this.status = this.errorText(err)
      } finally {
        this.saving = false
      }
    },
    async saveParcel() {
      if (!this.selectedParcel) {
        return
      }
      this.saving = true
      this.status = ""
      try {
        const body = Object.assign({}, this.selectedParcel, this.parcelDraft)
        await this.parcelApi().updateCharacterParcel({id: this.selectedParcel.id, characterParcel: body})
        await this.loadParcels()
        const still = this.parcels.find((p) => p.id === this.selectedParcel.id)
        if (still) {
          this.selectParcel(still)
        }
        this.status = "Saved"
      } catch (err) {
        this.status = this.errorText(err)
      } finally {
        this.saving = false
      }
    },
    async clearParcel() {
      if (!this.selectedParcel) {
        return
      }
      if (!confirm("Delete this parcel?")) {
        return
      }
      this.saving = true
      this.status = ""
      try {
        await this.parcelApi().deleteCharacterParcel({id: this.selectedParcel.id})
        this.selectedParcel = null
        await this.loadParcels()
        this.status = "Deleted"
      } catch (err) {
        this.status = this.errorText(err)
      } finally {
        this.saving = false
      }
    },
  },
}
</script>

<style scoped>
.char-inv-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}
.char-inv-toolbar {
  display: flex;
  flex-wrap: nowrap;
  gap: 8px;
  align-items: center;
  position: relative;
  z-index: 2;
}
.char-inv-toolbar > * {
  white-space: nowrap;
  flex-shrink: 0;
}
.char-inv-side .char-inv-toolbar {
  flex-wrap: wrap;
}
.char-inv-side .char-inv-toolbar input {
  flex: 1 1 140px;
  max-width: none;
  min-width: 0;
}
.char-inv-toolbar input {
  max-width: 240px;
}
.char-inv-status {
  font-size: 12px;
  color: var(--text-muted, #888);
}
.char-inv-hits {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 0 0 12px;
}
.char-inv-hit {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  min-width: 220px;
  max-width: 280px;
  box-sizing: border-box;
  border: 1px solid var(--border, #444);
  background: var(--surface-2, rgba(255,255,255,0.04));
  color: inherit;
  border-radius: 8px;
  padding: 8px 10px;
  text-align: left;
  cursor: pointer;
}
.char-inv-faces {
  display: flex;
  flex-shrink: 0;
  gap: 3px;
}
.char-inv-face {
  width: 40px;
  height: 40px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.25);
}
.char-inv-face [class*="item-"] {
  display: block !important;
  margin: 0 !important;
}
.char-inv-hit-text {
  min-width: 0;
  flex: 1;
}
.char-inv-hit-text b {
  display: block;
  font-size: 13px;
  line-height: 1.25;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.char-inv-hit-text span {
  display: block;
  font-size: 11px;
  line-height: 1.35;
  color: var(--text-muted, #8b93a0);
}
.char-inv-hit.is-on,
.char-inv-tab.is-on,
.inv-parcel.is-on {
  border-color: var(--accent, #c9a24a);
  box-shadow: inset 0 0 0 1px var(--accent, #c9a24a);
}
.char-inv-hero {
  display: flex;
  align-items: flex-start;
  justify-content: flex-end;
  gap: 8px;
  text-align: right;
}
.char-inv-hero-text {
  min-width: 0;
}
.char-inv-name {
  font-size: 16px;
  font-weight: 650;
  letter-spacing: -0.02em;
}
.char-inv-meta,
.char-inv-empty,
.inv-parcel-meta,
.inv-pick-id,
.inv-field {
  font-size: 12px;
  color: var(--text-muted, #8b93a0);
}
.char-inv-body {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 280px;
  gap: 14px;
  align-items: start;
}
.char-inv-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-bottom: 12px;
}
.char-inv-tab {
  border: 1px solid var(--border, #444);
  background: var(--surface-2, rgba(255,255,255,0.04));
  color: inherit;
  border-radius: 6px;
  padding: 4px 10px;
  cursor: pointer;
  font-size: 12px;
}
.char-inv-worn-wrap,
.inv-split {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: start;
}
.inv-paper-panel {
  width: max-content;
}
.inv-paper {
  display: grid;
  grid-template-columns: repeat(3, 46px);
  gap: 6px;
}
.inv-side-col {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-width: 220px;
}
.inv-panel {
  border: 1px solid var(--border-strong, #37404d);
  border-radius: 8px;
  background: var(--surface-3, #252b35);
  padding: 10px 12px 12px;
}
.inv-panel-head {
  font-size: 11px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--text-muted, #8b93a0);
  margin-bottom: 8px;
}
.inv-grid {
  display: grid;
  gap: 6px;
}
.inv-grid-4 { grid-template-columns: repeat(4, 46px); }
.inv-grid-5 { grid-template-columns: repeat(5, 46px); }
.inv-grid-6 { grid-template-columns: repeat(6, 46px); }
.inv-cursor-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid var(--border, #333);
}
.inv-empty-slot {
  font-size: 12px;
  color: var(--text-muted, #8b93a0);
  line-height: 1.4;
  margin-bottom: 8px;
}
.inv-parcel-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.inv-parcel-meta,
.inv-pick-id {
  display: block;
}
.inv-picker {
  display: block;
  max-height: 320px;
  overflow: auto;
  margin-top: 8px;
}
.inv-field {
  display: block;
  margin: 8px 0 3px;
}
.inv-augs {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 6px;
}
.inv-actions {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}
.inv-selected {
  margin-bottom: 8px;
}
.inv-item-link {
  display: inline-block;
  margin-top: 4px;
  font-size: 12px;
}
.inv-warn {
  color: #d4a017;
  font-size: 12px;
  margin-top: 4px;
}
.inv-unlimited {
  text-transform: none;
  letter-spacing: 0;
  font-size: 11px;
  color: var(--accent, #c9a24a);
}
@media (max-width: 820px) {
  .char-inv-body {
    grid-template-columns: 1fr;
  }
}
</style>

<style>
.inv-slot-spacer {
  width: 46px;
  height: 46px;
  border-radius: 4px;
  background: rgba(255,255,255,0.02);
}
.char-inv-face {
  width: 40px;
  height: 40px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}
.char-inv-face [class*="item-"] {
  display: block !important;
  margin: 0 !important;
}
.inv-parcel,
.inv-pick {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  box-sizing: border-box;
  padding: 6px 8px;
  margin: 0 0 4px;
  min-height: 40px;
  height: auto;
  flex: none;
  flex-shrink: 0;
  border: 1px solid #3d4654;
  border-radius: 6px;
  background: #1a2028;
  color: #e8eaed;
  text-align: left;
  cursor: pointer;
  line-height: 1.25;
  font-size: 13px;
  overflow: hidden;
}
.inv-parcel:hover,
.inv-pick:hover,
.inv-parcel.is-on,
.inv-pick.is-on {
  border-color: #c9a24a;
}
.inv-pick-icon {
  width: 24px;
  height: 24px;
  flex: 0 0 24px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}
.inv-pick-icon [class*="item-"] {
  display: block !important;
  margin: 0 !important;
}
.inv-pick-text {
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1px;
}
.inv-pick-name,
.inv-pick-text b {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  color: #e8eaed;
}
.inv-pick-id,
.inv-parcel-meta {
  font-size: 11px;
  color: #8b93a0;
}
</style>


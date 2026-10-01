<template>
  <div>
    <eq-window title="Items">
      <div class="ui-section">
        <div class="ui-section-head">
          <h6 class="ui-section-title">Search</h6>
        </div>
        <div class="item-search-row">
          <div class="item-field" style="flex: 2 1 220px">
            <label for="item_name">Name or ID</label>
            <input
              name="item_name"
              type="text"
              class="form-control form-control-sm"
              v-on:keyup.enter="search()"
              v-model="itemName"
              placeholder="e.g. Cloak of Flames or 1234"
              autofocus=""
              id="item_name"
            >
          </div>

          <div class="item-field" style="flex: 1 1 180px">
            <label for="item_type">Item type</label>
            <select id="item_type" class="form-control form-control-sm" v-model="itemType" @change="triggerState()">
              <option value="-1">Any type</option>
              <option v-for="option in itemTypeOptions" :key="option.value" :value="option.value">
                {{ option.text }}
              </option>
            </select>
          </div>

          <div class="item-field" style="flex: 0 1 110px">
            <label for="item_level">Required level</label>
            <select id="item_level" class="form-control form-control-sm" v-model="selectedLevel" @change="triggerState()">
              <option value="0">Any</option>
              <option v-for="l in 105" :key="l" :value="l">{{ l }}</option>
            </select>
          </div>

          <div class="item-field" v-if="parseInt(selectedLevel) > 0">
            <label>Level match</label>
            <div class="btn-group" role="group" aria-label="Level match">
              <b-button
                v-for="(label, type) in ['Exactly', 'And higher', 'And lower']"
                :key="type"
                size="sm"
                :variant="parseInt(selectedLevelType) === type ? 'warning' : 'outline-secondary'"
                :aria-pressed="parseInt(selectedLevelType) === type ? 'true' : 'false'"
                @click="selectedLevelType = type; triggerStateDelayed();"
              >{{ label }}</b-button>
            </div>
          </div>

          <div class="item-actions">
            <b-button size="sm" variant="primary" @click="search()">
              <i class="fa fa-search"></i> Search
            </b-button>
            <b-button size="sm" variant="link" class="item-reset" @click="resetForm()">Reset</b-button>

            <div class="btn-group" role="group" aria-label="Results per page" title="Results per page">
              <b-button
                v-for="n in [10, 100, 1000]"
                :key="n"
                size="sm"
                :variant="parseInt(limit) === n ? 'warning' : 'outline-secondary'"
                :aria-pressed="parseInt(limit) === n ? 'true' : 'false'"
                @click="limit = n; triggerStateDelayed()"
              >{{ n }}</b-button>
            </div>

            <div class="btn-group" role="group" aria-label="Result layout">
              <b-button
                size="sm"
                title="Show as table"
                :variant="listType === 'table' ? 'warning' : 'outline-secondary'"
                @click="listType = 'table'; triggerState()"
              ><i class="fa fa-table"></i></b-button>
              <b-button
                size="sm"
                title="Show as cards"
                :variant="listType === 'card' ? 'warning' : 'outline-secondary'"
                @click="listType = 'card'; triggerState()"
              ><i class="fa fa-th"></i></b-button>
            </div>
          </div>
        </div>
      </div>

      <chip-mask-selector
        title="Classes"
        :options="classOptions"
        :value="selectedClasses"
        @change="selectedClasses = $event; selectClass()"
      >
        <template #extra>
          <button
            type="button"
            :class="'ui-chip ml-auto' + (selectOnlyClassEnabled ? ' is-on' : '')"
            :aria-pressed="selectOnlyClassEnabled ? 'true' : 'false'"
            title="Only items usable by exactly these classes"
            @click="selectOnlyClassEnabled = !selectOnlyClassEnabled; selectClass()"
          >Exact match only</button>
        </template>
      </chip-mask-selector>

      <chip-mask-selector
        title="Races"
        :options="raceOptions"
        :value="selectedRaces"
        @change="selectedRaces = $event; selectRaces()"
      />

      <chip-mask-selector
        title="Slots"
        :options="slotOptions"
        :value="selectedSlots"
        @change="selectedSlots = $event; selectSlots()"
      />

      <chip-mask-selector
        title="Deities"
        :options="deityOptions"
        :value="selectedDeities"
        @change="selectedDeities = $event; selectDeities()"
      />

      <div class="ui-section">
        <div class="ui-section-head">
          <h6 class="ui-section-title">Flags</h6>
          <span class="ui-section-links">
            <a href="#" @click.prevent="clearFlags()">None</a>
          </span>
        </div>
        <button
          v-for="f in getCheckboxFilters()"
          :key="f.field"
          type="button"
          :class="'ui-chip' + (isFlagOn(f) ? ' is-on' : '')"
          :aria-pressed="isFlagOn(f) ? 'true' : 'false'"
          @click="toggleFlag(f)"
        >{{ f.description }}</button>
      </div>

      <div class="ui-section">
        <div class="ui-section-head">
          <h6 class="ui-section-title">Column filters</h6>
          <span class="text-muted small">Filter on any column of the items table</span>
        </div>
        <db-column-filter
          v-if="itemFields && filters"
          :set-filters="filters"
          @input="handleDbColumnFilters($event);"
          :columns="itemFields"
        />
      </div>
    </eq-window>

    <app-loader :is-loading="!loaded" padding="4"/>

    <!-- card rendering -->
    <div class="row" style="justify-content: center" v-if="loaded && listType === 'card'">
      <div
        v-for="(item, index) in items"
        class="col-lg-4 col-sm-9 mb-3"
        :key="item.id"
        style="display: inline-block; vertical-align: top"
      >
        <eq-window style="margin-right: 10px; width: auto; height: 100%">
          <eq-item-card-preview
            :item-data="item"
            :show-edit="true"
            :show-related-data="true"
          />
        </eq-window>
      </div>
    </div>

    <info-error-banner
      :notification="notification"
      :error="error"
      @dismiss-error="error = ''"
      @dismiss-notification="notification = ''"
      class="mt-0 pt-3"
    />

    <div v-if="loaded && !items" class="ui-empty mt-3">
      <div class="ui-empty-title">Search for items</div>
      Type a name or item ID and press Search, or pick classes, races, slots or flags above.
    </div>

    <div v-if="loaded && items && items.length === 0" class="ui-empty mt-3">
      <div class="ui-empty-title">No items found</div>
      Try fewer filters, a shorter name, or press Reset to start over.
    </div>

    <!-- table -->
    <item-preview-table
      :items="items"
      @reload-list="listItems"
      v-if="loaded && listType === 'table' && items && items.length"
    />

    <!--          <eq-spell-preview-table :items="items" v-if="loaded && listType === 'table' && items"/>-->

  </div>
</template>

<script type="ts">
import {ItemApi} from "@/app/api/api";
import EqWindow from "@/components/eq-ui/EQWindow.vue";
import {SpireApi} from "@/app/api/spire-api";
import EqItemCardPreview from "@/components/preview/EQItemCardPreview.vue";
import {DB_CLASSES_ICONS} from "@/app/constants/eq-class-icon-constants";
import {DB_CLASSES_SHORT, DB_PLAYER_CLASSES} from "@/app/constants/eq-classes-constants";
import {DB_SPA} from "@/app/constants/eq-spell-constants";
import {ROUTE} from "@/routes";
import ClassBitmaskCalculator from "@/components/tools/ClassBitmaskCalculator.vue";
import RaceBitmaskCalculator from "@/components/tools/RaceBitmaskCalculator.vue";
import InventorySlotCalculator from "@/components/tools/InventorySlotCalculator.vue";
import DeityBitmaskCalculator from "@/components/tools/DeityCalculator.vue";
import itemTypes from "@/constants/item-types.json"
import EqCheckbox from "@/components/eq-ui/EQCheckbox.vue";
import ItemPreviewTable from "@/views/items/components/ItemPreviewTable.vue";
import {SpireQueryBuilder} from "@/app/api/spire-query-builder";
import DbColumnFilter from "@/components/DbColumnFilter.vue";
import {DbSchema} from "@/app/db-schema";
import {Zones} from "@/app/zones";
import {Items} from "@/app/items";
import ItemPopover from "@/components/ItemPopover.vue";
import ContentArea from "@/components/layout/ContentArea.vue";
import InfoErrorBanner from "@/components/InfoErrorBanner.vue";
import {WindowManager} from "@/app/window";
import ChipMaskSelector from "@/components/forms/ChipMaskSelector.vue";
import {
  CLASSIC_CLASS_BITS,
  CLASSIC_DEITY_BITS,
  CLASSIC_RACE_BITS,
  CLASSIC_SLOT_BITS
} from "@/app/constants/eq-item-classic-constants"
import {classBitIcon, deityBitIcon, raceBitIcon} from "@/app/eq-chip-icons";

const PAIRED_SLOTS = {Ear01: "Ears", Bracer01: "Wrists", Ring01: "Fingers"}
const SECOND_SLOTS = ["Ear02", "Bracer02", "Ring02"]

export default {
  components: {
    ChipMaskSelector,
    InfoErrorBanner,
    ContentArea,
    ItemPopover,
    DbColumnFilter,
    ItemPreviewTable,
    EqCheckbox,
    DeityBitmaskCalculator,
    InventorySlotCalculator,
    RaceBitmaskCalculator,
    ClassBitmaskCalculator,
    // EqItemCardPreviewTable,
    EqItemCardPreview,
    EqItemCardPreview,
    EqWindow,
    "page-header": () => import("@/components/layout/PageHeader.vue")
  },
  data() {
    return {
      loaded: false,
      limit: 100,
      dbClassIcons: DB_CLASSES_ICONS,
      dbClassesShort: DB_CLASSES_SHORT,
      dbClasses: DB_PLAYER_CLASSES,
      dbItemEffects: DB_SPA,

      // form values
      selectedClasses: 0,
      selectedRaces: 0,
      selectedDeities: 0,
      selectedSlots: 0,
      itemName: "",
      itemType: -1,
      spellEffect: "",

      // when "only" option is set
      selectOnlyClassEnabled: false,

      checkboxFilters: {},
      filters: [],

      selectedLevel: 0,
      selectedLevelType: 0,

      listType: "table",

      itemFields: [],

      itemTypeOptions: [],

      // notification / errors
      notification: "",
      error: "",
    }
  },

  computed: {
    classOptions() {
      return CLASSIC_CLASS_BITS.map(([bit, name]) => [bit, name, name, classBitIcon(bit)])
    },
    // Shroud is left out so "All" matches the 65535 all-races value items use
    raceOptions() {
      return CLASSIC_RACE_BITS
        .filter(([bit]) => bit < 65536)
        .map(([bit, name]) => [bit, name, name, raceBitIcon(bit)])
    },
    deityOptions() {
      return CLASSIC_DEITY_BITS.map(([bit, name]) => [bit, name, name, deityBitIcon(bit)])
    },
    // paired slots (both ears, wrists, fingers) are one chip covering both bits
    slotOptions() {
      const bitOf = (name) => CLASSIC_SLOT_BITS.find((s) => s[1] === name)[0]
      return CLASSIC_SLOT_BITS
        .filter(([, name]) => !SECOND_SLOTS.includes(name))
        .map(([bit, name]) => {
          if (PAIRED_SLOTS[name]) {
            const second = name.replace("01", "02")
            return [bit | bitOf(second), PAIRED_SLOTS[name]]
          }
          return [bit, name]
        })
    },
  },

  created() {
    this.items = null;
  },

  async mounted() {

    this.resetCheckboxFilters()

    if (Object.keys(this.$route.query).length !== 0) {
      this.loadQueryState()
      this.listItems()
    }

    if (Object.keys(this.$route.query).length === 0) {
      this.loaded = true;
    }

    this.itemFields = await DbSchema.getTableColumns("items")

    this.itemTypeOptions = [];
    for (const [type, description] of Object.entries(itemTypes)) {
      this.itemTypeOptions.push(
        {
          text: type + ") " + description,
          value: type
        }
      )
    }

    Zones.getZones()
  },

  watch: {
    // reset state vars when we navigate
    '$route'() {
      this.loadQueryState()
      this.listItems()
    },
  },

  methods: {

    flagTrue(f) {
      return typeof f.true !== 'undefined' ? f.true : 1
    },

    flagFalse(f) {
      return typeof f.false !== 'undefined' ? f.false : 0
    },

    isFlagOn(f) {
      return this.checkboxFilters[f.field] === this.flagTrue(f)
    },

    toggleFlag(f) {
      this.$set(this.checkboxFilters, f.field, this.isFlagOn(f) ? this.flagFalse(f) : this.flagTrue(f))
      this.triggerCheckboxFilter(f.field, this.flagFalse(f))
    },

    clearFlags() {
      this.resetCheckboxFilters()
      this.search()
    },

    // a changed query re-lists through the $route watcher; an unchanged one needs a direct list
    search() {
      const before = this.$route.fullPath
      this.updateQueryState()
      setTimeout(() => {
        if (this.$route.fullPath === before) {
          this.listItems()
        }
      }, 0)
    },

    handleDbColumnFilters(filters) {
      this.filters = filters
      this.updateQueryState()
    },

    getCheckboxFilters() {
      return [
        {
          description: 'Is Magic',
          field: 'magic'
        },
        {
          description: 'No Drop',
          field: 'nodrop',
          true: 0,
          false: 1,
        },
        {
          description: 'FV No Drop',
          field: 'fvnodrop',
        },
        {
          description: 'No Rent',
          field: 'norent',
          true: 0,
          false: 1,
        },
        {
          description: 'Tradeskill Item',
          field: 'tradeskills'
        },
        {
          description: 'Book',
          field: 'book'
        },
        {
          description: 'No Transfer',
          field: 'notransfer'
        },
        {
          description: 'Summoned',
          field: 'summonedflag'
        },
        {
          description: 'Quest',
          field: 'questitemflag'
        },
        {
          description: 'Artifact',
          field: 'artifactflag'
        },
        {
          description: 'No Pet',
          field: 'nopet'
        },
        {
          description: 'Attuneable',
          field: 'attuneable'
        },
        {
          description: 'Stackable',
          field: 'stackable'
        },
        {
          description: 'Potion Belt',
          field: 'potionbelt'
        },
        // {
        //   description: 'Placeable',
        //   field: 'placeable'
        // },
        {
          description: 'Epic Item',
          field: 'epicitem'
        },
        // {
        //   description: 'Arrow Expend',
        //   field: 'expendablearrow'
        // },
        // {
        //   description: 'Heirloom',
        //   field: 'heirloom'
        // },
      ]
    },

    getChangedFilterCount() {
      let setValues = 0
      for (let key in this.checkboxFilters) {
        const value = this.checkboxFilters[key]

        this.getCheckboxFilters().forEach((filter) => {
          if (filter.field == key && typeof filter.true !== 'undefined' && filter.true == 0 && value === 0) {
            setValues++
          } else if (filter.field == key && typeof filter.true === 'undefined') {
            setValues++
          }
        })
      }

      console.log("[getChangedFilterCount] set values [%s]", setValues)

      return setValues
    },

    getFiltersNonZeroValues() {
      let values = {}
      for (let key in this.checkboxFilters) {
        const value = this.checkboxFilters[key]

        this.getCheckboxFilters().forEach((filter) => {
          if (filter.field == key && typeof filter.true !== 'undefined' && filter.true == 0 && value === 0) {
            values[filter.field] = value
          } else if (filter.field == key && typeof filter.true === 'undefined') {
            values[filter.field] = value
          }
        })
      }

      return values
    },

    triggerCheckboxFilter(field, falseValue) {
      // console.log("hello")
      // console.log(this.checkboxFilters)

      // delete filter from checkboxFilters if false value set
      for (let key in this.checkboxFilters) {
        const value = this.checkboxFilters[key]
        if (field == key && value === falseValue) {
          delete this.checkboxFilters[key]
        }
      }

      this.updateQueryState()
      this.listItems()

      // if (Object.keys(this.checkboxFilters).length > 0) {
      //   this.listItems()
      // }

      // if no checkboxFilters set, clear items result
      // if (setValues === 0) {
      //   this.items = null
      // }

    },

    updateQueryState: function () {
      let queryState = {};

      if (this.selectedClasses !== 0) {
        queryState.classes = this.selectedClasses
      }
      if (this.selectedRaces !== 0) {
        queryState.races = this.selectedRaces
      }
      if (this.selectedDeities !== 0) {
        queryState.deities = this.selectedDeities
      }
      if (this.selectedSlots !== 0) {
        queryState.slots = this.selectedSlots
      }
      if (this.itemType !== -1) {
        queryState.itemType = this.itemType
      }
      if (this.listType !== "") {
        queryState.listType = this.listType
      }
      if (this.itemName !== "") {
        queryState.name = this.itemName
      }
      if (this.selectedLevel !== 0) {
        queryState.level = this.selectedLevel
      }
      if (this.limit !== 0) {
        queryState.limit = this.limit
      }
      if (this.selectedLevelType >= 0) {
        queryState.levelType = this.selectedLevelType
      }
      if (this.selectOnlyClassEnabled) {
        queryState.classSelectOnly = 1
      }
      if (this.getChangedFilterCount() > 0) {
        queryState.checkboxFilters = JSON.stringify(this.getFiltersNonZeroValues())
      }
      if (this.filters && this.filters.length > 0) {
        queryState.filters = JSON.stringify(this.filters)
      }

      this.$router.push(
        {
          path: ROUTE.ITEMS_LIST,
          query: queryState
        }
      ).catch(() => {
      })
    },

    resetCheckboxFilters() {
      this.checkboxFilters = {}

      // make sure that checkboxFilters that are non-zero defaults are initialized as their
      // zero-values first
      this.getCheckboxFilters().forEach((filter) => {
        if (typeof filter.true !== 'undefined' && filter.true === 0) {
          this.checkboxFilters[filter.field] = 1
        }
      })
    },

    resetForm: function () {
      this.resetCheckboxFilters()

      this.filters = []
      this.selectedClasses = 0;
      this.selectedRaces = 0;
      this.selectedSlots = 0;
      this.selectedDeities = 0;
      this.itemType = -1;
      this.itemName = "";
      this.spellEffect = "";
      this.selectedLevel = 0;
      this.limit = 100;
      this.selectedLevelType = 0;
      this.items = null;
      this.listType = "table"
      this.updateQueryState()
    },

    loadQueryState: function () {
      if (this.$route.query.classes) {
        this.selectedClasses = parseInt(this.$route.query.classes);
      }
      if (this.$route.query.races) {
        this.selectedRaces = parseInt(this.$route.query.races);
      }
      if (this.$route.query.deities) {
        this.selectedDeities = parseInt(this.$route.query.deities);
      }
      if (this.$route.query.slots) {
        this.selectedSlots = parseInt(this.$route.query.slots);
      }
      if (this.$route.query.itemType) {
        this.itemType = this.$route.query.itemType;
      }
      if (this.$route.query.name) {
        this.itemName = this.$route.query.name;
      }
      if (this.$route.query.level) {
        this.selectedLevel = this.$route.query.level;
      }
      if (this.$route.query.limit) {
        this.limit = this.$route.query.limit;
      }
      if (this.$route.query.levelType) {
        this.selectedLevelType = this.$route.query.levelType;
      }
      if (this.$route.query.listType) {
        this.listType = this.$route.query.listType;
      }
      if (parseInt(this.$route.query.classSelectOnly) === 1) {
        this.selectOnlyClassEnabled = true;
      }
      if (this.$route.query.checkboxFilters) {
        this.resetCheckboxFilters()
        let checkboxFilters = JSON.parse(this.$route.query.checkboxFilters);
        for (let key in checkboxFilters) {
          this.checkboxFilters[key] = checkboxFilters[key]
        }
      }
      if (this.$route.query.filters) {
        this.filters = JSON.parse(this.$route.query.filters);
      } else {
        this.filters = [];
      }
    },

    selectClass: function () {
      this.itemName = ""
      this.updateQueryState();
      this.listItems()
    },

    selectRaces: function () {
      this.itemName = ""
      this.updateQueryState();
      this.listItems()
    },

    selectDeities: function () {
      this.itemName = ""
      this.updateQueryState();
      this.listItems()
    },

    selectSlots: function () {
      this.itemName = ""
      this.updateQueryState();
      this.listItems()
    },

    triggerStateDelayed() {
      setTimeout(() => {
        this.triggerState()
      }, 100)
    },

    triggerState() {
      this.updateQueryState();
    },

    isClassSelected: function (eqClass) {
      return eqClass === this.selectedClasses;
    },

    listItems: async function () {
      this.loaded = false;
      const builder = new SpireQueryBuilder()
      const api = (new ItemApi(...SpireApi.cfg()))

      // filter by class
      if (this.selectedClasses && parseInt(this.selectedClasses) > 0 && parseInt(this.selectedClasses) !== 65535) {
        if (this.selectOnlyClassEnabled) {
          builder.where("classes", "=", this.selectedClasses)
        } else {
          builder.where("classes", "&", this.selectedClasses)
        }

        builder.where("classes", "!=", 65535)
      } else if (this.selectedClasses && parseInt(this.selectedClasses) > 0 && parseInt(this.selectedClasses) === 65535) {
        builder.where("classes", "=", 65535)
      }

      // filter by race
      if (this.selectedRaces && parseInt(this.selectedRaces) > 0 && parseInt(this.selectedRaces) !== 65535) {
        builder.where("races", "&", this.selectedRaces)
        builder.where("races", "!=", 65535)
      } else if (this.selectedRaces && parseInt(this.selectedRaces) > 0 && parseInt(this.selectedRaces) === 65535) {
        builder.where("races", "=", 65535)
      }

      // filter by deity
      if (this.selectedDeities && parseInt(this.selectedDeities) > 0 && parseInt(this.selectedDeities) !== 65535) {
        builder.where("deity", "&", this.selectedDeities)
        builder.where("deity", "!=", 65535)
      } else if (this.selectedDeities && parseInt(this.selectedDeities) > 0 && parseInt(this.selectedDeities) === 65535) {
        builder.where("deity", "=", 65535)
      }

      // filter by slots
      if (this.selectedSlots && parseInt(this.selectedSlots) > 0 && parseInt(this.selectedSlots) !== 65535) {
        builder.where("slots", "&", this.selectedSlots)
        builder.where("slots", "!=", 65535)

      } else if (this.selectedSlots && parseInt(this.selectedSlots) > 0 && parseInt(this.selectedSlots) === 65535) {
        builder.where("slots", "=", 65535)
      }

      for (let key in this.getFiltersNonZeroValues()) {
        builder.where(key, "=", this.checkboxFilters[key])
      }

      // item type
      if (this.itemType && this.itemType > 0) {
        builder.where("itemtype", "=", this.itemType)
      }

      if (this.filters && this.filters.length > 0) {
        this.filters.forEach((f) => {
          builder.where(f.f, f.o, f.v)
        });
      }

      // level
      if (this.selectedLevel > 0) {
        let filterType = "="
        if (parseInt(this.selectedLevelType) === 1) {
          filterType = ">=";
        }
        if (parseInt(this.selectedLevelType) === 2) {
          filterType = "<=";
        }

        builder.where("reqlevel", filterType, this.selectedLevel)
      }

      // if number, filter by id
      // else name
      if (!isNaN(this.itemName) && this.itemName) {
        builder.where("id", "=", this.itemName)
      } else if (this.itemName) {
        builder.where("name", "like", this.itemName)
      }

      if (builder.getFilterCount() === 0) {
        this.items = null
        this.loaded = true
        return;
      }

      builder.groupBy(["id"])
      builder.limit(this.limit)
      builder.includes(Items.getRelationships())

      try {
        const r = await api.listItems(builder.get())
        if (r.status === 200) {
          // set items to be rendered
          this.items = r.data
          this.loaded = true;
        }
      } catch (e) {
        if (e.response && e.response.data && e.response.data.error) {
          this.error = e.response.data.error
          this.loaded = true
        }
      }
    }
  }
}

</script>

<style scoped>
.item-search-row {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 12px;
}

.item-field label {
  display: block;
  margin-bottom: 4px;
  color: var(--text-muted);
  font-size: 12px;
  font-weight: 500;
}

.item-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}

.item-reset {
  color: var(--text-muted) !important;
  border: 0 !important;
  background: transparent !important;
}

.item-reset:hover {
  color: var(--text) !important;
}
</style>

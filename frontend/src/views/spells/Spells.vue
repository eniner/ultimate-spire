<template>
  <div>
    <eq-window-simple title="Spells">
      <div class="ui-section">
        <div class="ui-section-head">
          <h6 class="ui-section-title">Class</h6>
        </div>
        <div class="spell-class-row">
          <div v-for="(icon, index) in dbClassIcons" :key="index" class="spell-class">
            <div class="spell-class-label">{{ dbClassesShort[index] }}</div>
            <span
              @click="selectClass(index)"
              :class="'hover-highlight-inner item-' + icon + (isClassSelected(index) ? ' highlight-selected-inner' : '')"
            />
          </div>
        </div>
      </div>

      <div class="ui-toolbar">
        <div class="ui-field" style="flex: 2 1 200px; margin-bottom: 0">
          <label for="spell_name">Name or ID</label>
          <input
            name="spell_name"
            type="text"
            class="form-control form-control-sm"
            autofocus
            v-on:keyup.enter="triggerState"
            v-model="spellName"
            placeholder="Name or ID"
            id="spell_name"
          >
        </div>

        <div class="ui-field" style="flex: 2 1 220px; margin-bottom: 0">
          <label for="spell_effect">Spell effect (SPA)</label>
          <select
            name="class"
            id="spell_effect"
            class="form-control form-control-sm"
            v-model="selectedSpa"
            @change="triggerState()"
          >
            <option value="-1">Any effect</option>
            <option v-for="(spellEffect, id) in dbSpellEffects" :key="id" v-bind:value="id">
              {{ id }}) {{ spellEffect }}
            </option>
          </select>
        </div>

        <div class="ui-field" style="flex: 0 1 90px; margin-bottom: 0">
          <label for="Class">Level</label>
          <select
            name="class"
            id="Class"
            class="form-control form-control-sm"
            v-model="selectedLevel"
            @change="selectClass(selectedClass)"
          >
            <option value="0">Any</option>
            <option v-for="l in 105" :key="l" v-bind:value="l">
              {{ l }}
            </option>
          </select>
        </div>

        <div class="btn-group" role="group" v-if="selectedLevel">
          <b-button
            @click="selectedLevelType = 0; triggerStateDelayed();"
            size="sm"
            :variant="(parseInt(selectedLevelType) === 0 ? 'warning' : 'outline-secondary')"
          >Only
          </b-button>
          <b-button
            @click="selectedLevelType = 1; triggerStateDelayed();"
            size="sm"
            :variant="(parseInt(selectedLevelType) === 1 ? 'warning' : 'outline-secondary')"
          >Higher
          </b-button>
          <b-button
            @click="selectedLevelType = 2; triggerStateDelayed();"
            size="sm"
            :variant="(parseInt(selectedLevelType) === 2 ? 'warning' : 'outline-secondary')"
          >Lower
          </b-button>
        </div>

        <div class="btn-group" role="group" aria-label="Result layout">
          <b-button
            title="Show as table"
            @click="listType = 'table'; "
            size="sm"
            :variant="(listType === 'table' ? 'warning' : 'outline-secondary')"
          ><i class="fa fa-table"></i></b-button>
          <b-button
            title="Show as cards"
            @click="listType = 'card'; "
            size="sm"
            :variant="(listType === 'card' ? 'warning' : 'outline-secondary')"
          ><i class="fa fa-th"></i></b-button>
        </div>

        <div class="btn-group" role="group" aria-label="Results per page">
          <b-button
            v-for="n in [10, 100, 250, 1000]"
            :key="n"
            @click="limit = n; triggerStateDelayed()"
            size="sm"
            :variant="(parseInt(limit) === n ? 'warning' : 'outline-secondary')"
          >{{ n }}
          </b-button>
        </div>

        <b-button size="sm" variant="link" @click="resetForm()">Reset</b-button>
      </div>

      <div class="ui-section" v-if="spellFields && filters">
        <div class="ui-section-head">
          <h6 class="ui-section-title">Column filters</h6>
        </div>
        <db-column-filter
          :set-filters="filters"
          @input="handleDbColumnFilters($event);"
          :columns="spellFields"
        />
      </div>

      <div class="ui-stat-line" v-if="message">
        {{ message }}
      </div>

    </eq-window-simple>

    <app-loader :is-loading="!loaded" padding="4"/>

    <!-- card rendering -->
    <div class="row" style="justify-content: center" v-if="loaded && listType === 'card'">
      <div
        v-for="(spell, index) in spells"
        class="col-lg-4 col-sm-9 mb-3"
        :key="spell.id"
        style="display: inline-block; vertical-align: top"
      >
        <eq-window style="margin-right: 1px; width: auto; height: 100%">
          <eq-spell-preview :spell-data="spell"/>
        </eq-window>
      </div>
    </div>

    <eq-spell-preview-table
      @reload-list="listSpells()"
      :spells="spells"
      v-if="loaded && listType === 'table' && spells"
    />

  </div>
</template>

<script type="ts">
import {ItemApi, SpellsNewApi} from "@/app/api/api";
import EqWindow from "@/components/eq-ui/EQWindow.vue";
import {SpireApi} from "@/app/api/spire-api";
import EqItemCardPreview from "@/components/preview/EQItemCardPreview.vue";
import * as util from "util";
import EqSpellPreview from "@/components/preview/EQSpellCardPreview.vue";
import {DB_CLASSES_ICONS} from "@/app/constants/eq-class-icon-constants";
import {DB_CLASSES_SHORT, DB_PLAYER_CLASSES} from "@/app/constants/eq-classes-constants";
import {DB_SPA} from "@/app/constants/eq-spell-constants";
import EqSpellPreviewTable from "@/components/preview/EQSpellPreviewTable.vue";
import {Spells} from "@/app/spells";
import {Items} from "@/app/items";
import {ROUTE} from "@/routes";
import EqWindowSimple from "@/components/eq-ui/EQWindowSimple.vue";
import {SpireQueryBuilder} from "@/app/api/spire-query-builder";
import DbColumnFilter from "@/components/DbColumnFilter.vue";
import {DbSchema} from "@/app/db-schema";
import ContentArea from "@/components/layout/ContentArea.vue";

export default {
  name: "Spells",
  components: {
    ContentArea,
    DbColumnFilter,
    EqWindowSimple,
    EqSpellPreviewTable,
    EqSpellPreview,
    EqItemCardPreview,
    EqWindow,
    "page-header": () => import("@/components/layout/PageHeader.vue")
  },
  data() {
    return {
      loaded: false,
      limit: 100,
      beginRange: 10000,
      endRange: 100000,
      dbClassIcons: DB_CLASSES_ICONS,
      dbClassesShort: DB_CLASSES_SHORT,
      dbClasses: DB_PLAYER_CLASSES,
      dbSpellEffects: DB_SPA,

      // form values
      selectedClass: 0,
      spellName: "",
      spellEffect: "",
      selectedSpa: -1,
      selectedLevel: 0,
      selectedLevelType: 0,

      message: "",

      filters: [],

      spellFields: [],

      listType: "table"
    }
  },

  watch: {
    '$route'() {
      this.loadQueryState()
      this.listSpells()
    },
  },

  created() {
    this.spells = null // we don't want reactivity so we register this here
  },

  async mounted() {
    if (Object.keys(this.$route.query).length !== 0) {
      this.loadQueryState()
      Spells.preloadDbstr().then((res) => {
        this.listSpells()
      })
    }

    if (Object.keys(this.$route.query).length === 0) {
      Spells.preloadDbstr()
      this.loaded = true;
    }

    this.spellFields = await DbSchema.getTableColumns("spells_new")

  },
  methods: {

    handleDbColumnFilters(checkboxFilters) {
      this.filters = checkboxFilters
      this.updateQueryState()
    },

    updateQueryState: function () {
      let queryState = {};

      if (this.selectedClass !== 0) {
        queryState.class = this.selectedClass
      }
      if (this.listType !== "") {
        queryState.listType = this.listType
      }
      if (this.spellName !== "") {
        queryState.name = this.spellName
      }
      if (this.selectedSpa !== -1) {
        queryState.spa = this.selectedSpa
      }
      if (this.selectedLevel !== 0) {
        queryState.level = this.selectedLevel
      }
      if (this.selectedLevelType !== 0) {
        queryState.levelType = this.selectedLevelType
      }
      if (this.filters && this.filters.length > 0) {
        queryState.filters = JSON.stringify(this.filters)
      }

      this.$router.push(
        {
          path: ROUTE.SPELLS_LIST,
          query: queryState
        }
      ).catch(() => {
      })
    },

    resetForm: function () {
      this.filters           = []
      this.selectedClass     = 0;
      this.spellName         = "";
      this.spellEffect       = "";
      this.selectedSpa       = -1;
      this.selectedLevel     = 0;
      this.selectedLevelType = 0;
      this.listType          = "table"
      this.spells            = null;
      this.updateQueryState()
    },

    loadQueryState: function () {
      if (this.$route.query.class) {
        this.selectedClass = this.$route.query.class;
      }
      if (this.$route.query.spa) {
        this.selectedSpa = this.$route.query.spa;
      }
      if (this.$route.query.name) {
        this.spellName = this.$route.query.name;
      }
      if (this.$route.query.level) {
        this.selectedLevel = this.$route.query.level;
      }
      if (this.$route.query.levelType) {
        this.selectedLevelType = this.$route.query.levelType;
      }
      if (this.$route.query.listType) {
        this.listType = this.$route.query.listType;
      }
      if (this.$route.query.filters) {
        this.filters = JSON.parse(this.$route.query.filters);
      } else {
        this.filters = []
      }
    },

    selectClass: function (eqClass) {
      this.selectedClass = eqClass;
      this.spellName     = ""
      this.selectedSpa   = -1
      this.updateQueryState();
      this.listSpells()
    },

    triggerStateDelayed() {
      setTimeout(() => {
        this.triggerState()
      }, 100)
    },

    triggerState() {
      this.updateQueryState();
      this.listSpells()
    },

    isClassSelected: function (eqClass) {
      return eqClass === this.selectedClass;
    },

    listSpells: function () {
      this.loaded  = false;
      this.message = ""

      const api     = (new SpellsNewApi(...SpireApi.cfg()))
      const builder = (new SpireQueryBuilder())

      // filter by class and no level set
      if (this.selectedClass > 0 && this.selectedLevel === 0) {
        builder.where("classes" + this.selectedClass, ">=", "1")
        builder.where("classes" + this.selectedClass, "<=", "250")
      }

      let filterType = "="
      if (parseInt(this.selectedLevelType) === 1) {
        filterType = ">=";
      }
      if (parseInt(this.selectedLevelType) === 2) {
        filterType = "<=";
      }

      // filter by level if class set
      if (this.selectedLevel > 0 && this.selectedClass > 0) {
        builder.where("classes" + this.selectedClass, filterType, this.selectedLevel)
        builder.where("classes" + this.selectedClass, "<=", "250")
      }

      // when no class is set but level is greater than 0
      if (this.selectedClass === 0 && this.selectedLevel > 0) {
        for (let i = 1; i < 16; i++) {
          builder.whereOr("classes" + i, filterType, this.selectedLevel)
        }
      }

      // if number, filter by id
      // else name
      if (!isNaN(this.spellName) && this.spellName) {
        builder.where("id", "=", this.spellName)
      } else if (this.spellName) {
        builder.where("name", "like", this.spellName)
      }

      if (this.selectedSpa > 0) {
        for (let effectIndex = 1; effectIndex <= 12; effectIndex++) {
          builder.whereOr("effectid" + effectIndex, "=", this.selectedSpa)
        }
      }

      builder.limit(this.limit);

      // filter by class
      if (this.selectedClass > 0) {
        builder.orderBy([util.format("classes%s", this.selectedClass)])
      }

      if (this.filters && this.filters.length > 0) {
        this.filters.forEach((f) => {
          builder.where(f.f, f.o, f.v)
        })
      }

      api.listSpellsNews(builder.get()).then(async (result) => {
        if (result.status === 200) {
          // set spells to be rendered
          this.spells = result.data

          // fetch spell ids that might be referenced by effects to bulk preload
          let spellsToPreload = [];
          let itemsToPreload  = [];
          result.data.forEach((spell) => {
            for (let effectIndex = 1; effectIndex <= 12; effectIndex++) {
              const spellId = Spells.getSpellIdFromEffectIfExists(spell, effectIndex);
              if (spellId > 0) {
                spellsToPreload.push(spellId);
              }
              const itemId = Spells.getItemIdFromEffectIfExists(spell, effectIndex);
              if (itemId > 0) {
                itemsToPreload.push(itemId);
              }
            }
          })

          // fetch spell ids that might be referenced by effects to bulk preload
          result.data.forEach((spell) => {
            for (let i = 0; i < 4; i++) {
              if (spell["components_" + i] > 0) {
                itemsToPreload.push(spell["components_" + i])
              }
            }
          });

          // bulk fetch preload
          const response = await (new ItemApi(...SpireApi.cfg())).getItemsBulk({
            body: {
              ids: itemsToPreload
            }
          })

          if (response.status == 200 && response.data && parseInt(response.data.length) > 0) {
            response.data.forEach((item) => {
              Items.setItem(item.id, item);
            })
          }

          // spells bulk fetch preload
          api.getSpellsNewsBulk({
            body: {
              ids: spellsToPreload
            }
          }).then((response) => {
            if (response.status == 200 && response.data && parseInt(response.data.length) > 0) {
              response.data.forEach((spell) => {
                if (!Spells.isSpellSet(spell.id)) {
                  Spells.setSpell(spell.id, spell);
                }
              })
            }
            this.loaded = true;
          });

        }
      })
    }
  }
}

</script>

<style scoped>
.spell-class-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
}

.spell-class {
  text-align: center;
  min-width: 40px;
}

.spell-class-label {
  font-size: 11px;
  color: var(--text-muted);
  margin-bottom: 4px;
}
</style>

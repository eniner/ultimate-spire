<template>
  <div
    class='eq-window-simple p-0 mb-0'
    :style="'margin-bottom: 40px; ' + (title ? 'padding-top: 30px' : 'padding-top: 0px !important')"
  >
    <!--      <div class='eq-window-title-bar' v-if="title">{{ title }}</div>-->
    <div :style="'' + (title ? '' : '') ">
      <div class='eq-window-nested-blue text-center p-5' v-if="spells.length === 0">
        No spells were found
      </div>

      <div
        class='spell-table fill-screen'
        style="overflow-y: scroll; overflow-x: hidden;"
        v-if="spells.length > 0"
      >
        <!--        <div class='eq-window-nested-blue' v-if="spells.length > 0" style="overflow-y: scroll;">-->
        <table id="spell-table" class="eq-table bordered eq-highlight-rows">
          <thead class="eq-table-floating-header">
          <tr>
            <th style="width: 96px;"></th>
            <th class="text-right" style="width: 70px;">ID</th>
            <th style="width: auto; min-width: 250px">Spell</th>
            <th style="width: auto; min-width: 280px">Classes</th>
            <th>Effects</th>

            <th class="text-right">Mana</th>
            <th class="text-right" style="width: 110px">Cast / recast</th>
            <th style="width: 130px">Duration</th>
            <th>Target</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="(spell, index) in spells" :key="spell.id">
            <td class="spell-actions">
              <b-button
                size="sm"
                variant="outline-secondary"
                title="Edit"
                aria-label="Edit spell"
                @click="editSpell(spell.id)"
              >
                <i class="fa fa-pencil"></i>
              </b-button>
              <b-button
                size="sm"
                variant="outline-secondary"
                title="Clone"
                aria-label="Clone spell"
                @click="editSpell(spell.id, true)"
              >
                <i class="fa fa-clone"></i>
              </b-button>
              <b-button
                size="sm"
                variant="outline-danger"
                title="Delete"
                aria-label="Delete spell"
                @click="deleteSpell(spell)"
              >
                <i class="fa fa-trash"></i>
              </b-button>
            </td>
            <td class="text-right tabular">
              {{ spell.id }}
            </td>
            <td
              class="text-left"
            >
              <spell-popover
                :spell="spell"
                :size="30"
                :spell-name-length="25"
                v-if="Object.keys(spell).length > 0 && spell"
                class="mt-2"
              />
            </td>
            <td class="text-left">
              <div class="spell-class-chips">
                <span
                  v-for="c in classLevels(spell)"
                  :key="c.index"
                  class="spell-class-chip"
                  :title="c.name + ' level ' + c.level"
                ><span v-if="c.icon" :class="'item-' + c.icon + '-sm'"></span><b>{{ c.name }}</b>{{ c.level }}</span>
              </div>
            </td>
            <td class="text-left spell-effects">
              <eq-spell-effects :spell="spell"/>
            </td>

            <td class="text-right tabular">{{ spell["mana"] > 0 ? spell["mana"] : "" }}</td>
            <td class="text-right tabular text-nowrap">
              {{ seconds(spell["cast_time"]) }}
              <span class="text-muted">/ {{ spell["recast_time"] > 0 ? seconds(spell["recast_time"]) : "–" }}</span>
            </td>
            <td v-if="spell['buffduration']" class="tabular">
              {{ humanTime(getBuffDuration(spell) * 6) }}
              <div class="text-muted small">{{ getBuffDuration(spell) }} tics</div>
            </td>
            <td v-else class="text-muted">–</td>
            <td> {{ getTargetTypeName(spell["targettype"]) }}</td>


            <!--              <td style="text-align: left">-->
            <!--                <eq-spell-description :spell="spell"/>-->
            <!--              </td>-->
          </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script>
import {Spells}           from "@/app/spells";
import EqWindow           from "@/components/eq-ui/EQWindow.vue";
import EqSpellEffects     from "@/components/preview/EQSpellEffects";
import EqSpellPreview     from "@/components/preview/EQSpellCardPreview.vue";
import {App}              from "@/constants/app";
import EqSpellDescription from "@/components/preview/EQSpellDescription";
import {DB_SPELL_TARGETS} from "@/app/constants/eq-spell-constants";
import {DB_CLASSES_ICONS} from "@/app/constants/eq-class-icon-constants";
import {DB_CLASSES_SHORT} from "@/app/constants/eq-classes-constants";
import {ROUTE}            from "@/routes";
import * as util          from "util";
import SpellPopover       from "@/components/SpellPopover";
import {Items}            from "@/app/items";
import {WindowManager} from "@/app/window";

export default {
  name: "EqSpellPreviewTable",
  components: {
    SpellPopover,
    EqSpellDescription,
    EqSpellEffects,
    EqSpellPreview,
    EqWindow,
  },
  data() {
    return {
      debug: App.DEBUG,
      debugSpellEffects: false,
      title: "",
      dbClassIcons: DB_CLASSES_ICONS,
      dbClassesShort: DB_CLASSES_SHORT,
    }
  },
  async created() {
    this.title = "Spells (" + this.spells.length + ")";
  },
  mounted() {
    setTimeout(() => {
      WindowManager.resizeFillScreenElements()
    }, 100);
  },
  props: {
    spells: Array
  },
  methods: {
    async deleteSpell(spell) {
      if (confirm(`Are you sure you want to permanently delete this spell? [${spell.name}] (${spell.id})`)) {
        await Spells.deleteSpell(spell.id)
        this.$emit("reload-list", true);
      }
    },

    getClasses: function (spell) {
      return Spells.getClasses(spell)
    },
    // classes_N of 255 means the class can't use the spell
    classLevels(spell) {
      return Object.keys(this.dbClassIcons)
        .filter((i) => spell['classes_' + i] > 0 && spell['classes_' + i] < 255)
        .map((i) => ({
          index: i,
          name: this.dbClassesShort[i],
          level: spell['classes_' + i],
          icon: this.dbClassIcons[i],
        }))
    },
    seconds(ms) {
      const s = ms / 1000
      return (Number.isInteger(s) ? s : s.toFixed(1)) + "s"
    },
    getTargetTypeColor: function (targetType) {
      return Spells.getTargetTypeColor(targetType)
    },
    getTargetTypeName: function (targetType) {
      return DB_SPELL_TARGETS[targetType] ? DB_SPELL_TARGETS[targetType] : "Unknown Target (" + targetType + ")"
    },
    humanTime: function (sec) {
      let result = ""
      if (sec === 0) {
        result = "";
      } else {
        let h  = Math.floor(sec / 3600);
        let m  = Math.floor((sec - h * 3600) / 60);
        let s  = sec - h * 3600 - m * 60;
        result = (h > 1 ? h + " hours " : "") + (h === 1 ? "1 hour " : "") + (m > 0 ? m + " min " : "") + (s > 0 ? s + " sec" : "");
      }

      return result;
    },
    getBuffDuration: function (spell) {
      return Spells.getBuffDuration(spell)
    },
    editSpell(spellId, clone = false) {
      this.$router.push(
        {
          path: util.format(ROUTE.SPELL_EDIT + (clone ? "?clone" : ""), spellId),
          query: {}
        }
      ).catch(() => {
      })
    }
  }
}
</script>

<style scoped>

.spell-actions {
  white-space: nowrap;
}

.spell-actions .btn {
  width: 26px;
  height: 26px;
  padding: 0 !important;
  margin-right: 4px;
  font-size: 12px !important;
}

.spell-class-chips {
  display: grid;
  grid-template-columns: repeat(auto-fill, 78px);
  gap: 4px;
}

.spell-class-chip {
  display: inline-flex;
  justify-content: space-between;
  align-items: center;
  gap: 3px;
  padding: 1px 7px;
  border: 1px solid var(--border);
  border-radius: 999px;
  background: var(--surface-2);
  color: var(--text-muted);
  font-size: 11.5px;
  font-variant-numeric: tabular-nums;
}

.spell-class-chip [class*="item-"] {
  display: block !important;
  margin: 0 !important;
}

.spell-class-chip b {
  color: var(--text);
  font-weight: 600;
  margin-right: 6px;
}

.spell-effects {
  max-width: 440px;
  min-width: 260px;
  line-height: 1.45;
}

/* For Mobile */
@media screen and (max-width: 540px) {
  .spell-table {
    overflow-x: visible;
    overflow-y: scroll !important
  }
}

/* For Tablets */
@media screen and (min-width: 540px) and (max-width: 780px) {
  .spell-table {
    overflow-x: visible;
    overflow-y: scroll !important
  }
}
</style>

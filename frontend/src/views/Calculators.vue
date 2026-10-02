<template>
  <div>
    <eq-window title="Calculators">
      <div class="ui-stat-line">
        Each tool is its own window. Click a bitmask or type a number — both stay in sync.
      </div>
    </eq-window>

    <div class="page-grid">
      <eq-window id="race-bitmask-calculator" title="Race Bitmask">
        <race-bitmask-calculator :inputData.sync="racesBitmask" :mask="racesBitmask"/>
        <div class="ui-field mt-3">
          <label for="race-bitmask-value">Bitmask</label>
          <input id="race-bitmask-value" type="text" class="form-control form-control-sm" v-model="racesBitmask">
        </div>
      </eq-window>

      <eq-window id="class-bitmask-calculator" title="Class Bitmask">
        <class-bitmask-calculator :inputData.sync="classesBitmask" :mask="classesBitmask"/>
        <div class="ui-field mt-3">
          <label for="class-bitmask-value">Bitmask</label>
          <input id="class-bitmask-value" type="text" class="form-control form-control-sm" v-model="classesBitmask">
        </div>
      </eq-window>

      <eq-window id="deity-bitmask-calculator" title="Deity Bitmask">
        <deity-bitmask-calculator :inputData.sync="deityBitmask" :mask="deityBitmask"/>
        <div class="ui-field mt-3">
          <label for="deity-bitmask-value">Bitmask</label>
          <input id="deity-bitmask-value" type="text" class="form-control form-control-sm" v-model="deityBitmask">
        </div>
      </eq-window>

      <eq-window id="client-version-calculator" title="Client Version Bitmask">
        <client-version-calculator :inputData.sync="clientVersionBitmask" :mask="clientVersionBitmask"/>
        <div class="ui-field mt-3">
          <label for="client-version-value">Bitmask</label>
          <input id="client-version-value" type="text" class="form-control form-control-sm" v-model="clientVersionBitmask">
        </div>
      </eq-window>

      <eq-window id="expansions-bitmask-calculator" title="Expansion Bitmask">
        <expansion-bitmask-calculator :inputData.sync="expansionBitmask" :mask="expansionBitmask"/>
        <div class="ui-field mt-3">
          <label for="expansion-bitmask-value">Bitmask</label>
          <input id="expansion-bitmask-value" type="text" class="form-control form-control-sm" v-model="expansionBitmask">
        </div>
      </eq-window>

      <eq-window id="augment-type-calculator" title="Augment Type Bitmask">
        <aug-bitmask-calculator :inputData.sync="augBitmask" :mask="augBitmask"/>
        <div class="ui-field mt-3">
          <label for="aug-bitmask-value">Bitmask</label>
          <input id="aug-bitmask-value" type="text" class="form-control form-control-sm" v-model="augBitmask">
        </div>
      </eq-window>

      <eq-window id="inventory-slot-calculator" title="Inventory Slot Bitmask">
        <inventory-slot-calculator :inputData.sync="inventorySlotBitmask" :mask="inventorySlotBitmask"/>
        <div class="ui-field mt-3">
          <label for="inventory-slot-value">Bitmask</label>
          <input id="inventory-slot-value" type="text" class="form-control form-control-sm" v-model="inventorySlotBitmask">
        </div>
      </eq-window>
    </div>

    <eq-window id="npc-special-abilities" title="NPC Special Abilities" class="page-grid-wide">
      <npc-special-abilities :inputData.sync="specialAbilityInput" :abilities="specialAbilityInput"/>
      <div class="ui-field mt-3">
        <label for="special-ability-value">Special abilities code</label>
        <input id="special-ability-value" type="text" class="form-control form-control-sm" v-model="specialAbilityInput">
      </div>
    </eq-window>
  </div>
</template>

<script>
// @ is an alias to /src

import EqWindowSimple             from "@/components/eq-ui/EQWindowSimple";
import EqWindowFancy              from "@/components/eq-ui/EQWindowFancy";
import EqWindowComplex            from "@/components/eq-ui/EQWindowComplex";
import EqWindow                   from "@/components/eq-ui/EQWindow";
import EqTabs                     from "@/components/eq-ui/EQTabs";
import EqTab                      from "@/components/eq-ui/EQTab";
import EqProgressBar              from "@/components/eq-ui/EQProgressBar";
import PageHeader                 from "@/components/layout/PageHeader";
import EqItemCardPreview          from "@/components/preview/EQItemCardPreview";
import EqSpellPreview             from "@/components/preview/EQSpellCardPreview";
import NpcSpecialAbilities        from "@/components/tools/NpcSpecialAbilities";
import HeaderComponent            from "../components/layout/HeaderComponent";
import RaceBitmaskCalculator      from "../components/tools/RaceBitmaskCalculator";
import ClassBitmaskCalculator     from "../components/tools/ClassBitmaskCalculator";
import InventorySlotCalculator    from "../components/tools/InventorySlotCalculator";
import AugBitmaskCalculator       from "../components/tools/AugmentTypeCalculator";
import DeityBitmaskCalculator     from "../components/tools/DeityCalculator";
import ExpansionBitmaskCalculator from "../components/tools/ExpansionsCalculator";
import ClientVersionCalculator    from "../components/tools/ClientVersionCalculator";

export default {
  name: "Calculators",
  components: {
    ClientVersionCalculator,
    ExpansionBitmaskCalculator,
    DeityBitmaskCalculator,
    AugBitmaskCalculator,
    InventorySlotCalculator,
    ClassBitmaskCalculator,
    RaceBitmaskCalculator,
    HeaderComponent,
    NpcSpecialAbilities,
    EqSpellPreview,
    EqItemCardPreview,
    PageHeader,
    EqProgressBar,
    EqTab,
    EqTabs,
    EqWindow,
    EqWindowComplex,
    EqWindowFancy,
    EqWindowSimple
  },
  data() {
    return {
      augBitmask: "477396",
      racesBitmask: 1859,
      classesBitmask: 37349,
      inventorySlotBitmask: 1536,
      specialAbilityInput: "1,1,3000,50^2,1,1,1000,2340^3,1,20,0,0,0,0,100,0^4,1,0,100,0,0,0,100,0^11,1,4,150,0,0,5^29,1,50^40,1,10,10,100^7,1^10,1^14,1^19,1^22,1^25,1^26,1",
      deityBitmask: 3079,
      clientVersionBitmask: "16",
      expansionBitmask: "3079"
    }
  },
}
</script>

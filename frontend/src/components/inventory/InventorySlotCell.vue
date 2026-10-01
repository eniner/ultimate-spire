<template>
  <div class="inv-slot-wrap" :class="{ 'is-on': selected, 'is-filled': !!icon, 'is-empty': !icon }">
    <div
      class="inv-slot"
      role="button"
      tabindex="0"
      :title="title"
      :aria-label="title"
      @click="$emit('pick', slotId)"
      @keydown.enter.prevent="$emit('pick', slotId)"
      @keydown.space.prevent="$emit('pick', slotId)"
    >
      <span v-if="icon" :class="'item-' + icon + ' inv-slot-icon'"></span>
      <span v-else class="inv-slot-ghost">{{ caption }}</span>
      <span v-if="qty" class="inv-slot-qty">{{ qty }}</span>
    </div>
  </div>
</template>

<script>
import {displayCharges, shortSlotLabel} from "../../app/eq-inventory-slots"

export default {
  name: "InventorySlotCell",
  props: {
    slotId: {type: Number, required: true},
    row: {type: Object, default: null},
    selected: {type: Boolean, default: false},
    label: {type: String, default: ""},
  },
  computed: {
    icon() {
      return this.row && this.row.item && this.row.item.icon ? this.row.item.icon : 0
    },
    qty() {
      return displayCharges(this.row && this.row.charges)
    },
    caption() {
      return this.label || shortSlotLabel(this.slotId)
    },
    title() {
      const name = this.row && this.row.item && this.row.item.name
        ? this.row.item.name
        : (this.row && this.row.item_id ? ("Item " + this.row.item_id) : "Empty")
      return this.caption + " — " + name
    },
  },
}
</script>

<style>
.inv-slot-wrap {
  width: 46px;
  height: 46px;
}
.inv-slot {
  box-sizing: border-box;
  width: 46px;
  height: 46px;
  padding: 2px;
  margin: 0;
  border: 1px solid #6b7584;
  border-radius: 4px;
  background: #0d1016;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  cursor: pointer;
  overflow: hidden;
  line-height: 0;
  font-size: 0;
}
.inv-slot-wrap.is-filled .inv-slot {
  background: #1c222c;
}
.inv-slot-wrap.is-on .inv-slot {
  border-color: #c9a24a;
  box-shadow: 0 0 0 1px #c9a24a;
}
.inv-slot:hover {
  border-color: #8a7340;
}
.inv-slot-icon {
  display: block !important;
  box-sizing: border-box;
  width: 40px !important;
  height: 40px !important;
  min-width: 40px !important;
  min-height: 40px !important;
  max-width: 40px !important;
  max-height: 40px !important;
  margin: 0 !important;
  padding: 0 !important;
  border: 0 !important;
  background-repeat: no-repeat !important;
  pointer-events: none;
}
.inv-slot-ghost {
  font-size: 9px;
  line-height: 1.1;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: #6d7684;
  text-align: center;
  padding: 2px;
  pointer-events: none;
}
.inv-slot-qty {
  position: absolute;
  right: 1px;
  bottom: 1px;
  font-size: 10px;
  font-weight: 700;
  line-height: 1;
  color: #fff;
  text-shadow: 0 0 3px #000, 0 1px 2px #000;
  pointer-events: none;
}
</style>

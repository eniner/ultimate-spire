<template>
  <div class="ui-section">
    <div class="ui-section-head">
      <h6 class="ui-section-title">{{ title }}</h6>
      <span class="ui-section-links">
        <a href="#" @click.prevent="set(allMask)">All</a>
        <a href="#" @click.prevent="set(0)">None</a>
      </span>
      <span v-if="value" class="text-muted small">{{ selectedCount }} selected</span>
      <slot name="extra"></slot>
    </div>
    <div>
      <button
        v-for="o in options"
        :key="o[0]"
        type="button"
        :class="'ui-chip' + (isOn(o[0]) ? ' is-on' : '')"
        :aria-pressed="isOn(o[0]) ? 'true' : 'false'"
        :title="o[2] || o[1]"
        @click="toggle(o[0])"
      >
        <span v-if="iconOf(o)" class="ui-chip-icon">
          <span :class="'item-' + iconOf(o) + '-sm'"></span>
        </span>
        {{ o[1] }}
      </button>
    </div>
  </div>
</template>

<script>
// options: [[mask, label, tooltip?, icon?], ...] - a mask may cover several bits (e.g. both ear slots)
export default {
  name: "ChipMaskSelector",
  props: {
    title: {type: String, required: true},
    options: {type: Array, required: true},
    value: {type: Number, default: 0},
  },
  computed: {
    allMask() {
      return this.options.reduce((m, o) => m | o[0], 0)
    },
    selectedCount() {
      return this.options.filter((o) => this.isOn(o[0])).length
    },
  },
  methods: {
    iconOf(option) {
      return option && option[3] ? option[3] : 0
    },
    isOn(mask) {
      return (this.value & mask) === mask
    },
    toggle(mask) {
      this.set(this.isOn(mask) ? (this.value & ~mask) : (this.value | mask))
    },
    set(mask) {
      this.$emit("input", mask >>> 0)
      this.$emit("change", mask >>> 0)
    },
  },
}
</script>

<style>
.ui-section .ui-chip {
  overflow: visible !important;
}
.ui-chip-icon {
  width: 13px;
  height: 13px;
  flex: 0 0 13px;
  overflow: hidden;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  line-height: 0;
}
.ui-chip-icon [class*="item-"] {
  display: block !important;
  margin: 0 !important;
}
</style>

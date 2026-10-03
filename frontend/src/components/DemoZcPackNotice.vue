<template>
  <b-alert v-if="show" show variant="info" class="demo-zc-pack">
    This public demo is read-only. The files you put on a real server are
    <a :href="packUrl">zone-controller-starter.zip</a>
    — the actual <code>zone_controller.pl</code> plus blank mob / loot / item paths.
    <a :href="guideUrl">Where they go</a>.
  </b-alert>
</template>

<script>
import {demoZcGuideUrl, demoZcPackUrl, isDemoMode} from "@/app/demo-mode"

export default {
  name: "DemoZcPackNotice",
  computed: {
    show() {
      return isDemoMode()
    },
    packUrl() {
      return demoZcPackUrl()
    },
    guideUrl() {
      return demoZcGuideUrl()
    },
  },
}
</script>

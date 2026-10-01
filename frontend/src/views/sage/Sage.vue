<template>
  <div class="sage-container" ref="sage-root">
    <div v-if="status !== 'ready'" class="sage-status">
      <p v-if="status === 'loading'">Loading EQ Sage…</p>
      <p v-else class="sage-status-error">{{ error }}</p>
    </div>
    <iframe
      ref="sage-iframe"
      class="sage-iframe"
      src="/eqsage/"
      title="EQSage"
      @load="onSageLoad"
    ></iframe>
  </div>
</template>

<script type="ts">
import * as SpireApiTypes from "@/app/api";
import { SpireApi } from "../../app/api/spire-api";
import { SpireQueryBuilder } from "@/app/api/spire-query-builder";
import {Navbar}          from "../../app/navbar";
import { Zones } from "../../app/zones";
import { Spawn } from "../../app/spawn";
import { Npcs } from "../../app/npcs";
import { Grid } from "../../app/grid";
import UserContext from "../../app/user/UserContext";
export default {
  components: {

  },
  data() {
    return {
      status: "loading",
      error: "",
      loadRetries: 0,
      mountTimer: null,
    }
  },
  watch: {

  },

  destroyed() {
    if (this.mountTimer) {
      clearTimeout(this.mountTimer)
      this.mountTimer = null
    }
    Navbar.expand();
  },
  async mounted() {
    Navbar.collapse()
  },
  methods: {
    injectSpire(win) {
      if (!win) {
        return
      }
      win.Spire = {
        SpireApi,
        SpireApiTypes,
        SpireQueryBuilder,
        Grid,
        Zones,
        Spawn,
        Npcs,
        UserContext,
        backendBaseUrl: SpireApi.getBasePath(),
      }
    },
    sageDidMount(doc) {
      const root = doc && doc.getElementById("root")
      return !!(root && root.childElementCount > 0)
    },
    reloadSage() {
      const iframe = this.$refs["sage-iframe"]
      if (!iframe) {
        return
      }
      this.status = "loading"
      this.error = ""
      iframe.src = "/eqsage/?r=" + Date.now()
    },
    onSageLoad(e) {
      const iframe = e.target
      const win = iframe.contentWindow
      this.injectSpire(win)

      let doc = null
      try {
        doc = iframe.contentDocument
      } catch (err) {
        this.status = "error"
        this.error = "EQ Sage loaded cross-origin, so Spire cannot hook its APIs. Use the same-origin /eqsage proxy."
        return
      }

      const bodyText = (doc && doc.body && doc.body.innerText) ? doc.body.innerText : ""
      if (
        bodyText.indexOf("Cannot GET") !== -1 ||
        (doc && doc.title && doc.title.indexOf("Error") === 0 && bodyText.indexOf("Cannot GET") !== -1)
      ) {
        this.status = "error"
        this.error = "EQ Sage did not load (/eqsage 404). Restart the Vue frontend so the Sage proxy in vue.config.js is picked up."
        return
      }

      if (this.sageDidMount(doc)) {
        this.status = "ready"
        this.loadRetries = 0
        return
      }

      // Sage is ~120 JS modules through the Vue /static proxy. One 500 leaves
      // #root empty (white screen). Retry once after the shell has had time to boot.
      if (this.mountTimer) {
        clearTimeout(this.mountTimer)
      }
      this.mountTimer = setTimeout(() => {
        const live = iframe.contentDocument
        if (this.sageDidMount(live)) {
          this.status = "ready"
          this.loadRetries = 0
          return
        }
        if (this.loadRetries < 1) {
          this.loadRetries += 1
          this.reloadSage()
          return
        }
        this.status = "error"
        this.error = "EQ Sage stayed blank after loading. Refresh /sage — a Sage JS chunk failed through the local /static proxy."
      }, 4000)
    },
  },
}
</script>

<style>
.sage-container {
  position: relative;
}
.sage-status {
  position: absolute;
  z-index: 2;
  left: 24px;
  top: 16px;
  color: #e6dcc3;
  font-size: 14px;
}
.sage-status-error {
  max-width: 520px;
  background: rgba(20, 16, 12, 0.88);
  border: 1px solid #8a6a32;
  padding: 10px 12px;
}
.sage-iframe {
  height: calc(100vh - 10px);
  border: none;
  margin-left: -35px;
  margin-top: -10px;
  width: calc(100% + 70px);
}
.code-display {
  font-size: 14px !important;
  max-width: 100% !important;
  width: 100%;
}
</style>

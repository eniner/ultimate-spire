<template>
  <content-area>
    <eq-window title="Ultimate changelog generator">
      <p class="zc-copy">
        Builds notes for this Ultimate Spire build. Use <b>Notes</b> for
        <code>CHANGELOG.md</code>. Use <b>Release body</b> for the GitHub release.
      </p>
      <div class="ui-toolbar">
        <button type="button" class="btn btn-sm btn-dark" @click="generate('notes')">
          Generate notes
        </button>
        <button type="button" class="btn btn-sm btn-dark" @click="generate('release')">
          Generate release body
        </button>
        <button type="button" class="btn btn-sm btn-dark" :disabled="!changelog" @click="copyToClip(changelog)">
          Copy
        </button>
        <router-link class="btn btn-sm btn-dark" :to="ROUTE.ULTIMATE_CHANGELOG">Open changelog</router-link>
      </div>
    </eq-window>

    <eq-window v-if="changelog" class="mt-3" title="Generated markdown">
      <textarea v-model="changelog" class="form-control changelog-out" rows="28"></textarea>
    </eq-window>
  </content-area>
</template>

<script>
import EqWindow from "@/components/eq-ui/EQWindow.vue"
import ContentArea from "@/components/layout/ContentArea"
import ClipBoard from "@/app/clipboard/clipboard"
import {Notify} from "@/app/Notify"
import {ROUTE} from "@/routes"
import {formatUltimateChangelog, formatUltimateReleaseNotes} from "@/app/ultimate-changelog"

export default {
  name: "Changelog",
  components: {EqWindow, ContentArea},
  data() {
    return {
      ROUTE,
      changelog: "",
    }
  },
  methods: {
    generate(kind) {
      this.changelog = kind === "release" ? formatUltimateReleaseNotes() : formatUltimateChangelog()
    },
    copyToClip(s) {
      ClipBoard.copyFromText(s)
      Notify.toast("Copied to clipboard")
    },
  },
}
</script>

<style scoped>
.zc-copy {
  color: var(--text-muted, #b7c0cc);
  margin-bottom: 12px;
}

.ui-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.changelog-out {
  font-family: Consolas, "Courier New", monospace;
  font-size: 13px;
  min-height: 60vh;
}
</style>
